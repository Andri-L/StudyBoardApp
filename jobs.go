package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var progressPattern = regexp.MustCompile(`progress\s*=\s*(\d+)%`)

type cachedDuration struct {
	Size    int64
	ModTime int64
	Seconds float64
}

func (a *App) startWorker() {
	a.queue = make(chan *Job, 64)
	a.jobs = map[string]*Job{}
	a.jobOrder = []string{}
	a.cancels = map[string]context.CancelFunc{}
	a.durationCache = map[string]cachedDuration{}
	go func() {
		for job := range a.queue {
			a.mu.Lock()
			canceled := job.Status == "canceled"
			a.mu.Unlock()
			if canceled {
				continue
			}
			ctx, cancel := context.WithCancel(context.Background())
			a.cancelMu.Lock()
			a.cancels[job.ID] = cancel
			a.cancelMu.Unlock()

			a.updateJob(job.ID, func(j *Job) {
				j.Status = "extracting"
				j.Stage = "Extrayendo audio"
				j.Progress = 1
				j.StartedAt = time.Now().UnixMilli()
			})
			err := a.runJob(ctx, job.ID)
			wasCanceled := ctx.Err() == context.Canceled

			cancel()
			a.cancelMu.Lock()
			delete(a.cancels, job.ID)
			a.cancelMu.Unlock()

			if wasCanceled {
				a.updateJob(job.ID, func(j *Job) {
					j.Status = "canceled"
					j.Stage = "Cancelado"
					j.FinishedAt = time.Now().UnixMilli()
				})
				continue
			}
			if err != nil {
				a.updateJob(job.ID, func(j *Job) {
					j.Status = "error"
					j.Stage = "Error"
					j.Error = err.Error()
					j.FinishedAt = time.Now().UnixMilli()
				})
				continue
			}
			a.updateJob(job.ID, func(j *Job) {
				j.Status = "done"
				j.Stage = "Listo"
				j.Progress = 100
				j.FinishedAt = time.Now().UnixMilli()
			})
			if a.ctx != nil {
				var path string
				a.mu.RLock()
				if j := a.jobs[job.ID]; j != nil {
					path = j.MediaPath
				}
				a.mu.RUnlock()
				wailsRuntime.EventsEmit(a.ctx, "transcript:ready", path)
			}
		}
	}()
}

func (a *App) enqueueJob(mediaPath, cardID string) (*Job, error) {
	if err := a.engine.validate(); err != nil {
		return nil, fmt.Errorf("el motor de transcripción no está listo: %v", err)
	}
	a.mu.RLock()
	root := a.state.Workspace
	a.mu.RUnlock()
	if !insideRoot(root, mediaPath) {
		return nil, fmt.Errorf("el video está fuera del espacio de trabajo")
	}
	if _, err := os.Stat(mediaPath); err != nil {
		return nil, fmt.Errorf("no se encontró el video")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, id := range a.jobOrder {
		j := a.jobs[id]
		if j != nil && samePath(j.MediaPath, mediaPath) && isActiveStatus(j.Status) {
			cp := *j
			return &cp, nil
		}
	}
	job := &Job{
		ID:        newID("job"),
		MediaPath: absClean(mediaPath),
		CardID:    cardID,
		Status:    "queued",
		Stage:     "En cola",
		Progress:  0,
		CreatedAt: time.Now().UnixMilli(),
	}
	a.jobs[job.ID] = job
	a.jobOrder = append([]string{job.ID}, a.jobOrder...)
	if len(a.jobOrder) > 40 {
		for _, old := range a.jobOrder[30:] {
			if j := a.jobs[old]; j != nil && !isActiveStatus(j.Status) {
				delete(a.jobs, old)
			}
		}
		a.jobOrder = a.jobOrder[:30]
	}
	cp := *job
	// Send outside of the critical section, but while the queue has room.
	go func(j *Job) {
		a.queue <- j
	}(job)
	return &cp, nil
}

func isActiveStatus(s string) bool {
	switch s {
	case "queued", "extracting", "transcribing", "saving":
		return true
	}
	return false
}

func (a *App) updateJob(id string, fn func(*Job)) {
	a.mu.Lock()
	j := a.jobs[id]
	if j == nil {
		a.mu.Unlock()
		return
	}
	fn(j)
	cp := *j
	a.mu.Unlock()
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "job:update", cp)
	}
}

func (a *App) jobCopy(id string) (Job, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	j := a.jobs[id]
	if j == nil {
		return Job{}, false
	}
	return *j, true
}

func (a *App) jobsForPath(mediaPath string) *Job {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, id := range a.jobOrder {
		j := a.jobs[id]
		if j != nil && samePath(j.MediaPath, mediaPath) {
			cp := *j
			return &cp
		}
	}
	return nil
}

func (a *App) runJob(ctx context.Context, jobID string) error {
	job, ok := a.jobCopy(jobID)
	if !ok {
		return fmt.Errorf("tarea no encontrada")
	}
	mediaPath := job.MediaPath
	duration := a.probeDuration(mediaPath)

	jobDir := filepath.Join(a.cacheDir, "jobs", jobID)
	if err := os.MkdirAll(jobDir, 0o755); err != nil {
		return err
	}
	defer os.RemoveAll(jobDir)
	audioPath := filepath.Join(jobDir, "audio.wav")
	outPrefix := filepath.Join(jobDir, "transcript")

	if err := a.extractAudio(ctx, jobID, mediaPath, audioPath, duration); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	a.updateJob(jobID, func(j *Job) {
		j.Status = "transcribing"
		j.Stage = "Transcribiendo audio"
		j.Progress = 15
	})
	if err := a.transcribeAudio(ctx, jobID, audioPath, outPrefix); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	a.updateJob(jobID, func(j *Job) {
		j.Status = "saving"
		j.Stage = "Guardando transcripción"
		j.Progress = 97
	})
	return a.saveTranscript(mediaPath, outPrefix+".json", duration)
}

func (a *App) extractAudio(ctx context.Context, jobID, mediaPath, audioPath string, duration float64) error {
	args := []string{
		"-hide_banner", "-loglevel", "error", "-y",
		"-i", mediaPath,
		"-vn", "-ac", "1", "-ar", "16000", "-c:a", "pcm_s16le",
		"-progress", "pipe:1", "-nostats",
		audioPath,
	}
	cmd := exec.CommandContext(ctx, a.engine.FFmpeg, args...)
	hideWindow(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("no se pudo iniciar ffmpeg: %v", err)
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "out_time_us=") && !strings.HasPrefix(line, "out_time_ms=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		us, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || us <= 0 {
			continue
		}
		if duration > 0 {
			pct := float64(us) / (duration * 1_000_000)
			if pct > 1 {
				pct = 1
			}
			value := 2 + pct*10
			a.updateJob(jobID, func(j *Job) {
				if j.Progress < value {
					j.Progress = value
				}
				j.Stage = "Extrayendo audio"
			})
		}
	}
	waitErr := cmd.Wait()
	if waitErr != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = waitErr.Error()
		}
		return fmt.Errorf("ffmpeg no pudo extraer el audio: %s", lastLines(msg, 4))
	}
	return nil
}

func (a *App) transcribeAudio(ctx context.Context, jobID, audioPath, outPrefix string) error {
	threads := runtime.NumCPU() / 2
	if threads < 4 {
		threads = 4
	}
	if threads > 12 {
		threads = 12
	}
	args := []string{
		"-m", a.engine.Model,
		"-f", audioPath,
		"-l", a.language,
		"-ojf", "-osrt",
		"-of", outPrefix,
		"-pp",
		"-t", strconv.Itoa(threads),
	}
	if a.engine.VAD != "" {
		args = append(args, "--vad", "-vm", a.engine.VAD)
	}
	cmd := exec.CommandContext(ctx, a.engine.Whisper, args...)
	hideWindow(cmd)
	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	var logLines []string
	var logMu sync.Mutex
	scanDone := make(chan struct{})
	go func() {
		defer close(scanDone)
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			logMu.Lock()
			logLines = append(logLines, line)
			if len(logLines) > 20 {
				logLines = logLines[len(logLines)-20:]
			}
			logMu.Unlock()
			if m := progressPattern.FindStringSubmatch(line); m != nil {
				pct, _ := strconv.Atoi(m[1])
				value := 15 + float64(pct)*0.80
				a.updateJob(jobID, func(j *Job) {
					if j.Progress < value {
						j.Progress = value
					}
					j.Stage = "Transcribiendo audio"
				})
			}
		}
	}()

	runErr := cmd.Run()
	_ = pw.Close()
	<-scanDone
	if runErr != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		logMu.Lock()
		tail := strings.Join(logLines, "\n")
		logMu.Unlock()
		if tail == "" {
			tail = runErr.Error()
		}
		return fmt.Errorf("Whisper falló: %s", lastLines(tail, 5))
	}
	return nil
}

func (a *App) saveTranscript(mediaPath, whisperJSONPath string, duration float64) error {
	data, err := os.ReadFile(whisperJSONPath)
	if err != nil {
		return fmt.Errorf("no se generó el JSON de Whisper")
	}
	var w whisperJSON
	if err := json.Unmarshal(data, &w); err != nil {
		return fmt.Errorf("no se pudo leer la salida de Whisper: %v", err)
	}
	info, err := os.Stat(mediaPath)
	if err != nil {
		return err
	}
	segments := make([]Segment, 0, len(w.Transcription))
	for _, item := range w.Transcription {
		text := strings.TrimSpace(item.Text)
		if text == "" {
			continue
		}
		words, confidence := tokensToWords(item.Tokens)
		segments = append(segments, Segment{
			Start:      float64(item.Offsets.From) / 1000,
			End:        float64(item.Offsets.To) / 1000,
			Text:       text,
			Confidence: confidence,
			Words:      words,
		})
	}
	if len(segments) == 0 {
		return fmt.Errorf("Whisper no devolvió texto")
	}
	if duration <= 0 {
		duration = segments[len(segments)-1].End
	}
	tr := Transcript{
		Version:      1,
		MediaPath:    absClean(mediaPath),
		MediaSize:    info.Size(),
		MediaModTime: info.ModTime().UnixMilli(),
		Duration:     duration,
		Language:     a.language,
		Model:        strings.TrimSuffix(filepath.Base(a.engine.Model), ".bin"),
		CreatedAt:    time.Now().UnixMilli(),
		Segments:     segments,
	}
	if w.Result.Language != "" {
		tr.Language = w.Result.Language
	}
	out, err := json.Marshal(tr)
	if err != nil {
		return err
	}
	path := a.transcriptFile(mediaPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(path)
		return os.Rename(tmp, path)
	}
	return nil
}

func (a *App) transcriptFile(mediaPath string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(absClean(mediaPath))))
	return filepath.Join(a.cacheDir, "transcripts", hex.EncodeToString(sum[:])+".json")
}

func (a *App) loadTranscript(mediaPath string) (*Transcript, error) {
	data, err := os.ReadFile(a.transcriptFile(mediaPath))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var tr Transcript
	if err := json.Unmarshal(data, &tr); err != nil {
		return nil, err
	}
	info, err := os.Stat(mediaPath)
	if err != nil {
		return nil, nil
	}
	if tr.MediaSize != info.Size() || tr.MediaModTime != info.ModTime().UnixMilli() {
		return nil, nil
	}
	return &tr, nil
}

func (a *App) probeDuration(mediaPath string) float64 {
	info, err := os.Stat(mediaPath)
	if err != nil {
		return 0
	}
	a.durationMu.Lock()
	cached, ok := a.durationCache[mediaPath]
	a.durationMu.Unlock()
	if ok && cached.Size == info.Size() && cached.ModTime == info.ModTime().UnixMilli() {
		return cached.Seconds
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, a.engine.FFprobe,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		mediaPath,
	)
	hideWindow(cmd)
	out, err := cmd.Output()
	var seconds float64
	if err == nil {
		seconds, _ = strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	}
	a.durationMu.Lock()
	a.durationCache[mediaPath] = cachedDuration{Size: info.Size(), ModTime: info.ModTime().UnixMilli(), Seconds: seconds}
	a.durationMu.Unlock()
	return seconds
}

func lastLines(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, " | ")
}

// tokensToWords merges Whisper subword tokens into words, keeping the
// token-level offsets. This is the foundation for frame-accurate navigation.
func tokensToWords(tokens []whisperToken) ([]Word, float64) {
	words := make([]Word, 0, len(tokens))
	var current *Word
	flush := func() {
		if current == nil {
			return
		}
		word := *current
		word.Text = strings.TrimSpace(word.Text)
		if word.Text != "" {
			words = append(words, word)
		}
		current = nil
	}
	for _, token := range tokens {
		raw := token.Text
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "[_") {
			continue
		}
		start := float64(token.Offsets.From) / 1000
		end := float64(token.Offsets.To) / 1000
		if isPunctuation(trimmed) && current != nil {
			current.Text += trimmed
			current.End = end
			continue
		}
		if strings.HasPrefix(raw, " ") {
			flush()
		}
		if current == nil {
			current = &Word{Start: start, End: end, Text: trimmed, Confidence: token.P}
		} else {
			current.Text += trimmed
			current.End = end
			current.Confidence = (current.Confidence + token.P) / 2
		}
	}
	flush()
	if len(words) == 0 {
		return nil, 0
	}
	sum := 0.0
	for _, word := range words {
		sum += word.Confidence
	}
	return words, sum / float64(len(words))
}

func isPunctuation(s string) bool {
	if s == "" {
		return false
	}
	return strings.Trim(s, ".,;:!?¿¡\"'()[]{}…-–—") == ""
}
