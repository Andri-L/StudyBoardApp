package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) startWorkflowWorker() {
	if a.workflowQueue == nil {
		a.workflowQueue = make(chan *WorkflowJob, 32)
	}
	go func() {
		for job := range a.workflowQueue {
			ctx, cancel := context.WithCancel(context.Background())
			a.workflowMu.Lock()
			a.workflowCancels[job.ID] = cancel
			a.workflowMu.Unlock()

			a.updateWorkflowJob(job.ID, func(j *WorkflowJob) {
				j.Status = "running"
				j.Stage = "Preparando OpenCode"
				j.StartedAt = time.Now().UnixMilli()
			})
			err := a.runWorkflow(ctx, job)
			wasCanceled := ctx.Err() == context.Canceled
			cancel()
			a.workflowMu.Lock()
			delete(a.workflowCancels, job.ID)
			a.workflowMu.Unlock()

			switch {
			case wasCanceled:
				a.updateWorkflowJob(job.ID, func(j *WorkflowJob) {
					j.Status = "canceled"
					j.Stage = "Cancelado"
					j.FinishedAt = time.Now().UnixMilli()
				})
			case err != nil:
				a.updateWorkflowJob(job.ID, func(j *WorkflowJob) {
					j.Status = "error"
					j.Stage = "Error"
					j.Error = err.Error()
					j.FinishedAt = time.Now().UnixMilli()
				})
			default:
				a.updateWorkflowJob(job.ID, func(j *WorkflowJob) {
					j.Status = "done"
					j.Stage = "Listo"
					j.FinishedAt = time.Now().UnixMilli()
				})
			}
		}
	}()
}

// StartWorkflow queues an OpenCode workflow for a transcribed video.
func (a *App) StartWorkflow(boardID, workflowID, videoPath, outputLanguage string) (WorkflowJob, error) {
	def, ok := workflowByID(workflowID)
	if !ok {
		return WorkflowJob{}, fmt.Errorf("workflow desconocido")
	}
	videoPath = absClean(videoPath)
	a.mu.RLock()
	root := a.state.Workspace
	a.mu.RUnlock()
	if !insideRoot(root, videoPath) {
		return WorkflowJob{}, fmt.Errorf("el video está fuera del workspace")
	}
	if tr, err := a.loadTranscript(videoPath); err != nil || tr == nil {
		return WorkflowJob{}, fmt.Errorf("el video todavía no tiene transcripción")
	}
	a.mu.RLock()
	boardIndex := a.boardIndex(boardID)
	a.mu.RUnlock()
	if boardIndex < 0 {
		return WorkflowJob{}, fmt.Errorf("tablero no encontrado")
	}
	language := strings.ToLower(strings.TrimSpace(outputLanguage))
	if language != "en" {
		language = "es"
	}
	job := &WorkflowJob{
		ID:             newID("wf"),
		BoardID:        boardID,
		WorkflowID:     def.ID,
		WorkflowName:   def.Name,
		VideoPath:      videoPath,
		VideoName:      filepath.Base(videoPath),
		OutputLanguage: language,
		Status:         "queued",
		Stage:          "En cola",
		CreatedAt:      time.Now().UnixMilli(),
	}
	a.workflowMu.Lock()
	a.workflowJobs[job.ID] = job
	a.workflowOrder = append([]string{job.ID}, a.workflowOrder...)
	if len(a.workflowOrder) > 40 {
		a.workflowOrder = a.workflowOrder[:40]
	}
	cp := *job
	a.workflowMu.Unlock()
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "workflow:update", cp)
	}
	go func() {
		a.workflowQueue <- job
	}()
	return cp, nil
}

func (a *App) updateWorkflowJob(id string, fn func(*WorkflowJob)) {
	a.workflowMu.Lock()
	job := a.workflowJobs[id]
	if job == nil {
		a.workflowMu.Unlock()
		return
	}
	fn(job)
	cp := *job
	a.workflowMu.Unlock()
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "workflow:update", cp)
	}
}

// GetWorkflowJobs returns recent workflow jobs, newest first.
func (a *App) GetWorkflowJobs() []WorkflowJob {
	a.workflowMu.Lock()
	defer a.workflowMu.Unlock()
	out := make([]WorkflowJob, 0, len(a.workflowOrder))
	for _, id := range a.workflowOrder {
		if job := a.workflowJobs[id]; job != nil {
			out = append(out, *job)
		}
	}
	return out
}

// CancelWorkflow cancels a queued or running workflow.
func (a *App) CancelWorkflow(id string) error {
	a.workflowMu.Lock()
	cancel := a.workflowCancels[id]
	job := a.workflowJobs[id]
	queued := job != nil && job.Status == "queued"
	a.workflowMu.Unlock()
	if cancel != nil {
		cancel()
	}
	if queued {
		a.updateWorkflowJob(id, func(j *WorkflowJob) {
			j.Status = "canceled"
			j.Stage = "Cancelado"
			j.FinishedAt = time.Now().UnixMilli()
		})
	}
	return nil
}

func (a *App) runWorkflow(ctx context.Context, job *WorkflowJob) error {
	def, ok := workflowByID(job.WorkflowID)
	if !ok {
		return fmt.Errorf("workflow desconocido")
	}
	opencode, err := findOpencode()
	if err != nil {
		return err
	}
	prompt := renderWorkflowPrompt(def, job.VideoPath, job.OutputLanguage)
	title := fmt.Sprintf("StudyBoard · %s · %s", def.ShortName, filepath.Base(filepath.Dir(job.VideoPath)))
	args := []string{
		"run",
		"--auto",
		"--format", "json",
		"--title", title,
		prompt,
	}
	cmd := exec.CommandContext(ctx, opencode, args...)
	a.mu.RLock()
	cmd.Dir = a.state.Workspace
	a.mu.RUnlock()
	hideWindow(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("no se pudo iniciar OpenCode: %v", err)
	}
	a.updateWorkflowJob(job.ID, func(j *WorkflowJob) {
		j.Stage = "OpenCode está trabajando"
	})

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	var content strings.Builder
	var eventError string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var event struct {
			Type      string `json:"type"`
			SessionID string `json:"sessionID"`
			Part      struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"part"`
			Error any `json:"error"`
		}
		if json.Unmarshal([]byte(line), &event) != nil {
			continue
		}
		if event.SessionID != "" {
			job.SessionID = event.SessionID
			a.updateWorkflowJob(job.ID, func(j *WorkflowJob) {
				j.SessionID = event.SessionID
			})
		}
		if event.Type == "text" && event.Part.Text != "" {
			if content.Len() > 0 {
				content.WriteString("\n\n")
			}
			content.WriteString(event.Part.Text)
			a.updateWorkflowJob(job.ID, func(j *WorkflowJob) {
				j.Stage = "Redactando informe"
			})
		}
		if event.Type == "error" && event.Error != nil {
			eventError = fmt.Sprintf("%v", event.Error)
		}
	}
	runErr := cmd.Wait()
	if ctx.Err() == context.Canceled {
		return ctx.Err()
	}
	if runErr != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = eventError
		}
		if detail == "" {
			detail = runErr.Error()
		}
		return fmt.Errorf("OpenCode falló: %s. Verifica que OpenCode esté autenticado en la TUI.", lastLines(detail, 5))
	}
	if content.Len() == 0 {
		return fmt.Errorf("OpenCode no devolvió texto")
	}

	result, err := a.saveWorkflowResult(job, def, content.String())
	if err != nil {
		return err
	}
	job.SessionID = result.SessionID
	a.mu.RLock()
	state := a.snapshotLocked()
	a.mu.RUnlock()
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "workflow:result", state)
	}
	return nil
}
