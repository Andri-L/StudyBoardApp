package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

//go:embed workflows/*.json workflows/prompts/*.md
var workflowFiles embed.FS

type workflowDefinition struct {
	WorkflowInfo
	PromptFile string `json:"promptFile"`
	Prompt     string `json:"-"`
}

var workflowCache struct {
	sync.Once
	defs []workflowDefinition
	err  error
}

func loadWorkflowDefinitions() ([]workflowDefinition, error) {
	workflowCache.Do(func() {
		entries, err := workflowFiles.ReadDir("workflows")
		if err != nil {
			workflowCache.err = err
			return
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			data, err := workflowFiles.ReadFile(filepath.ToSlash(filepath.Join("workflows", entry.Name())))
			if err != nil {
				workflowCache.err = err
				return
			}
			var def workflowDefinition
			if err := json.Unmarshal(data, &def); err != nil {
				workflowCache.err = fmt.Errorf("%s: %w", entry.Name(), err)
				return
			}
			if def.ID == "" || def.Name == "" {
				continue
			}
			promptData, err := workflowFiles.ReadFile(filepath.ToSlash(filepath.Join("workflows", def.PromptFile)))
			if err != nil {
				workflowCache.err = fmt.Errorf("%s: %w", def.PromptFile, err)
				return
			}
			def.Prompt = string(promptData)
			workflowCache.defs = append(workflowCache.defs, def)
		}
	})
	return workflowCache.defs, workflowCache.err
}

func workflowByID(id string) (workflowDefinition, bool) {
	defs, err := loadWorkflowDefinitions()
	if err != nil {
		return workflowDefinition{}, false
	}
	for _, def := range defs {
		if def.ID == id {
			return def, true
		}
	}
	return workflowDefinition{}, false
}

// GetWorkflows returns the workflows available to the UI and MCP.
func (a *App) GetWorkflows() ([]WorkflowInfo, error) {
	defs, err := loadWorkflowDefinitions()
	if err != nil {
		return nil, err
	}
	out := make([]WorkflowInfo, 0, len(defs))
	for _, def := range defs {
		out = append(out, def.WorkflowInfo)
	}
	return out, nil
}

// GetWorkflowVideos lists videos in the workspace and whether they have a transcript.
func (a *App) GetWorkflowVideos() ([]WorkflowVideo, error) {
	videos := a.mcpVideos("", 2000)
	out := make([]WorkflowVideo, 0, len(videos))
	for _, video := range videos {
		out = append(out, WorkflowVideo{
			Path:          video.Path,
			Name:          video.Name,
			Class:         video.Class,
			Duration:      video.Duration,
			HasTranscript: video.HasTranscript,
			Segments:      video.Segments,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Class != out[j].Class {
			return strings.ToLower(out[i].Class) < strings.ToLower(out[j].Class)
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}

func workflowLanguageName(code string) string {
	switch strings.ToLower(code) {
	case "en":
		return "English"
	case "es":
		return "Spanish"
	default:
		if strings.TrimSpace(code) == "" {
			return "Spanish"
		}
		return code
	}
}

func renderWorkflowPrompt(def workflowDefinition, videoPath, language string) string {
	className := filepath.Base(filepath.Dir(videoPath))
	classDate := dateLabel(className)
	if classDate == "" {
		classDate = "not available"
	}
	replacer := strings.NewReplacer(
		"{{VIDEO_PATH}}", videoPath,
		"{{VIDEO_NAME}}", filepath.Base(videoPath),
		"{{CLASS_NAME}}", className,
		"{{CLASS_DATE}}", classDate,
		"{{OUTPUT_LANGUAGE}}", workflowLanguageName(language),
	)
	prompt := replacer.Replace(def.Prompt)
	prefix := "You are executing a StudyBoard workflow from the desktop app.\n" +
		"Use the StudyBoard MCP server as the only source of transcript evidence. " +
		"Do not ask follow-up questions and do not include your reasoning.\n\n"
	suffix := "\n\nRemember: cite timestamps and verbatim evidence. Never invent information. " +
		"Return only the final Markdown report."
	return prefix + prompt + suffix
}

func renderMCPWorkflowPrompt(def workflowDefinition, classArg string) string {
	replacer := strings.NewReplacer(
		"{{VIDEO_PATH}}", "the transcribed video that best matches the class requested below",
		"{{VIDEO_NAME}}", classArg,
		"{{CLASS_NAME}}", classArg,
		"{{CLASS_DATE}}", "use the class folder date when available",
		"{{OUTPUT_LANGUAGE}}", "the language used by the user in the request (default: Spanish)",
	)
	return replacer.Replace(def.Prompt)
}

func workflowResultPosition(board *Board, videoPath string) (float64, float64) {
	for _, card := range board.Cards {
		if card.Kind == "video" && strings.EqualFold(card.Path, videoPath) {
			offset := 0.0
			for _, result := range board.Results {
				if strings.EqualFold(result.VideoPath, videoPath) {
					offset += result.Height + 24
				}
			}
			return card.X + card.Width + 28, card.Y + offset
		}
	}
	count := len(board.Results)
	return 100 + float64(count%2)*460, 120 + float64(count/2)*420
}

func workflowResultPath(videoPath, outputFile string) string {
	base := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
	safe := outputFile
	if safe == "" {
		safe = "workflow"
	}
	return filepath.Join(filepath.Dir(videoPath), base+" - "+safe+".md")
}

func findOpencode() (string, error) {
	// Prefer the real executable over the npm .cmd shim: the shim truncates
	// prompts that contain newlines on Windows.
	var candidates []string
	if appData := os.Getenv("APPDATA"); appData != "" {
		candidates = append(candidates,
			filepath.Join(appData, "npm", "node_modules", "@opencode", "cli", "bin", "opencode.exe"),
			filepath.Join(appData, "npm", "opencode.exe"),
		)
	}
	if local := os.Getenv("LOCALAPPDATA"); local != "" {
		candidates = append(candidates,
			filepath.Join(local, "Programs", "opencode", "opencode.exe"),
		)
	}
	if path, err := exec.LookPath("opencode.exe"); err == nil {
		candidates = append(candidates, path)
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	if path, err := exec.LookPath("opencode"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("no se encontró el CLI de OpenCode. Abre la TUI de OpenCode una vez o instálalo con npm")
}

// MoveWorkflowResult persists a result card position.
func (a *App) MoveWorkflowResult(boardID, resultID string, x, y float64) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	i := a.boardIndex(boardID)
	if i < 0 {
		return fmt.Errorf("tablero no encontrado")
	}
	for j := range a.state.Boards[i].Results {
		if a.state.Boards[i].Results[j].ID == resultID {
			a.state.Boards[i].Results[j].X = x
			a.state.Boards[i].Results[j].Y = y
			a.saveStateLocked()
			return nil
		}
	}
	return fmt.Errorf("resultado no encontrado")
}

// RemoveWorkflowResult removes a result card but keeps the exported Markdown file.
func (a *App) RemoveWorkflowResult(boardID, resultID string) (AppState, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	i := a.boardIndex(boardID)
	if i < 0 {
		return AppState{}, fmt.Errorf("tablero no encontrado")
	}
	results := a.state.Boards[i].Results
	for j := range results {
		if results[j].ID == resultID {
			a.state.Boards[i].Results = append(results[:j], results[j+1:]...)
			a.saveStateLocked()
			return a.snapshotLocked(), nil
		}
	}
	return a.snapshotLocked(), nil
}

func (a *App) saveWorkflowResult(job *WorkflowJob, def workflowDefinition, content string) (WorkflowResult, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return WorkflowResult{}, fmt.Errorf("OpenCode no devolvió texto")
	}
	videoPath := absClean(job.VideoPath)
	filePath := workflowResultPath(videoPath, def.OutputFile)
	if err := os.WriteFile(filePath, []byte(content+"\n"), 0o644); err != nil {
		filePath = ""
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	i := a.boardIndex(job.BoardID)
	if i < 0 {
		return WorkflowResult{}, fmt.Errorf("tablero no encontrado")
	}
	x, y := workflowResultPosition(&a.state.Boards[i], videoPath)
	result := WorkflowResult{
		ID:           newID("result"),
		WorkflowID:   def.ID,
		WorkflowName: def.Name,
		VideoPath:    videoPath,
		VideoName:    filepath.Base(videoPath),
		ClassName:    filepath.Base(filepath.Dir(videoPath)),
		Content:      content,
		FilePath:     filePath,
		SessionID:    job.SessionID,
		X:            x,
		Y:            y,
		Width:        430,
		Height:       380,
		CreatedAt:    time.Now().UnixMilli(),
	}
	a.state.Boards[i].Results = append(a.state.Boards[i].Results, result)
	a.saveStateLocked()
	return result, nil
}

func (a *App) mcpWorkflowResults(query string, limit int) []map[string]any {
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}
	needle := normalizeSearch(query)
	out := make([]map[string]any, 0)
	for _, board := range a.state.Boards {
		for _, result := range board.Results {
			if needle != "" {
				haystack := normalizeSearch(result.WorkflowName + " " + result.VideoName + " " + result.ClassName + " " + result.Content)
				if !strings.Contains(haystack, needle) {
					continue
				}
			}
			preview := result.Content
			if len([]rune(preview)) > 500 {
				preview = string([]rune(preview)[:500]) + "…"
			}
			out = append(out, map[string]any{
				"id":           result.ID,
				"workflowId":   result.WorkflowID,
				"workflowName": result.WorkflowName,
				"video":        result.VideoPath,
				"videoName":    result.VideoName,
				"class":        result.ClassName,
				"createdAt":    result.CreatedAt,
				"filePath":     result.FilePath,
				"preview":      preview,
			})
			if len(out) >= limit {
				return out
			}
		}
	}
	return out
}

func (a *App) mcpWorkflowResult(id, video, workflow string) (string, error) {
	var matches []WorkflowResult
	for _, board := range a.state.Boards {
		for _, result := range board.Results {
			if id != "" && result.ID == id {
				matches = append(matches, result)
				continue
			}
			if id == "" {
				videoMatch := video == ""
				if video != "" {
					needle := normalizeSearch(filepath.Base(video))
					videoMatch = strings.Contains(normalizeSearch(result.VideoName), needle) ||
						strings.Contains(normalizeSearch(result.VideoPath), needle)
				}
				workflowMatch := workflow == ""
				if workflow != "" {
					needle := normalizeSearch(workflow)
					workflowMatch = strings.Contains(normalizeSearch(result.WorkflowID+" "+result.WorkflowName), needle)
				}
				if videoMatch && workflowMatch {
					matches = append(matches, result)
				}
			}
		}
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no se encontró un resultado de workflow con esos criterios")
	}
	if len(matches) > 1 {
		names := make([]string, 0, len(matches))
		for i, match := range matches {
			if i >= 5 {
				names = append(names, "…")
				break
			}
			names = append(names, fmt.Sprintf("%s · %s", match.WorkflowName, match.VideoName))
		}
		return "", fmt.Errorf("varios resultados coinciden: %s. Usa el id del resultado.", strings.Join(names, ", "))
	}
	result := matches[0]
	var b strings.Builder
	b.WriteString("# " + result.WorkflowName + " — " + result.VideoName + "\n\n")
	b.WriteString("- Clase: " + result.ClassName + "\n")
	b.WriteString("- Video: " + result.VideoPath + "\n")
	if result.FilePath != "" {
		b.WriteString("- Archivo Markdown: " + result.FilePath + "\n")
	}
	if result.SessionID != "" {
		b.WriteString("- Sesión OpenCode: " + result.SessionID + "\n")
	}
	b.WriteString("\n## Resultado\n\n")
	b.WriteString(result.Content)
	return b.String(), nil
}
