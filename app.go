package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails backend.
type App struct {
	ctx       context.Context
	mu        sync.RWMutex
	configDir string
	cacheDir  string
	statePath string

	state         AppState
	engine        enginePaths
	selectedModel string
	language      string
	media         *mediaServer

	queue         chan *Job
	jobs          map[string]*Job
	jobOrder      []string
	cancels       map[string]context.CancelFunc
	cancelMu      sync.Mutex
	durationMu    sync.Mutex
	durationCache map[string]cachedDuration

	downloadMu      sync.Mutex
	downloads       map[string]*ModelDownload
	downloadCancels map[string]context.CancelFunc

	workflowMu      sync.Mutex
	workflowJobs    map[string]*WorkflowJob
	workflowOrder   []string
	workflowQueue   chan *WorkflowJob
	workflowCancels map[string]context.CancelFunc
	playbackMu      sync.Mutex
}

func NewApp() *App {
	return &App{
		downloads:       map[string]*ModelDownload{},
		downloadCancels: map[string]context.CancelFunc{},
		workflowJobs:    map[string]*WorkflowJob{},
		workflowQueue:   make(chan *WorkflowJob, 32),
		workflowCancels: map[string]context.CancelFunc{},
	}
}

func isMCPInvocation() bool {
	if len(os.Args) > 1 && (os.Args[1] == "--mcp" || os.Args[1] == "mcp") {
		return true
	}
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	name := strings.ToLower(filepath.Base(exe))
	return strings.Contains(name, "mcp")
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.initPaths()
	a.loadState()
	a.engine = findEngine(a.selectedModel)
	if a.selectedModel == "" && a.engine.Model != "" {
		a.selectedModel = filepath.Base(a.engine.Model)
	}
	if media, err := startMediaServer(a); err == nil {
		a.media = media
	}
	wailsRuntime.WindowSetAlwaysOnTop(ctx, false)
	a.startWorker()
	a.startWorkflowWorker()
}

func (a *App) initPaths() {
	a.initPathsWithCleanup(true)
}

func (a *App) initPathsForMCP() {
	a.initPathsWithCleanup(false)
}

func (a *App) initPathsWithCleanup(cleanJobs bool) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		cfg = os.Getenv("APPDATA")
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		cache = os.Getenv("LOCALAPPDATA")
	}
	a.configDir = filepath.Join(cfg, "StudyBoard App")
	a.cacheDir = filepath.Join(cache, "StudyBoard App")
	a.statePath = filepath.Join(a.configDir, "state.json")

	// One-time migration from the previous app name.
	migrateDir(filepath.Join(cfg, "Aula"), a.configDir)
	migrateDir(filepath.Join(cache, "Aula"), a.cacheDir)

	for _, dir := range []string{a.configDir, a.cacheDir, filepath.Join(a.cacheDir, "transcripts"), filepath.Join(a.cacheDir, "thumbs"), filepath.Join(a.cacheDir, "jobs"), filepath.Join(a.cacheDir, "models")} {
		_ = os.MkdirAll(dir, 0o755)
	}
	if cleanJobs {
		// Jobs are not resumed across launches; clear leftover audio files.
		_ = os.RemoveAll(filepath.Join(a.cacheDir, "jobs"))
		_ = os.MkdirAll(filepath.Join(a.cacheDir, "jobs"), 0o755)
	}
}

func (a *App) shutdown(ctx context.Context) {
	a.cancelMu.Lock()
	for _, cancel := range a.cancels {
		cancel()
	}
	a.cancelMu.Unlock()
	a.downloadMu.Lock()
	for _, cancel := range a.downloadCancels {
		cancel()
	}
	a.downloadMu.Unlock()
	a.workflowMu.Lock()
	for _, cancel := range a.workflowCancels {
		cancel()
	}
	a.workflowMu.Unlock()
	if a.media != nil {
		a.media.stop()
	}
}

// GetState returns the full workspace state.
func (a *App) GetState() (AppState, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.snapshotLocked(), nil
}

// ChooseWorkspace opens the folder picker and sets the workspace root.
func (a *App) ChooseWorkspace() (AppState, error) {
	a.mu.RLock()
	initial := a.state.Workspace
	a.mu.RUnlock()
	dir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title:            "Elige tu carpeta de clases",
		DefaultDirectory: initial,
	})
	if err != nil {
		return AppState{}, err
	}
	if dir == "" {
		a.mu.RLock()
		defer a.mu.RUnlock()
		return a.snapshotLocked(), nil
	}
	a.mu.Lock()
	newRoot := absClean(dir)
	changed := a.state.Workspace != "" && !strings.EqualFold(a.state.Workspace, newRoot)
	a.state.Workspace = newRoot
	if changed || len(a.state.Boards) == 0 {
		b := newBoard("Clases")
		a.state.Boards = []Board{b}
		a.state.ActiveBoardID = b.ID
	}
	a.saveStateLocked()
	state := a.snapshotLocked()
	a.mu.Unlock()
	return state, nil
}

// CreateBoard creates a new board and makes it active.
func (a *App) CreateBoard(name string) (AppState, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	b := newBoard(name)
	a.state.Boards = append(a.state.Boards, b)
	a.state.ActiveBoardID = b.ID
	a.saveStateLocked()
	return a.snapshotLocked(), nil
}

// RenameBoard changes a board name.
func (a *App) RenameBoard(id, name string) (AppState, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return AppState{}, fmt.Errorf("el nombre no puede estar vacío")
	}
	if len([]rune(name)) > 60 {
		return AppState{}, fmt.Errorf("el nombre es demasiado largo")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	i := a.boardIndex(id)
	if i < 0 {
		return AppState{}, fmt.Errorf("tablero no encontrado")
	}
	a.state.Boards[i].Name = name
	a.saveStateLocked()
	return a.snapshotLocked(), nil
}

// DeleteBoard removes a board (never removes files).
func (a *App) DeleteBoard(id string) (AppState, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.state.Boards) <= 1 {
		return AppState{}, fmt.Errorf("no puedes eliminar el único tablero")
	}
	i := a.boardIndex(id)
	if i < 0 {
		return AppState{}, fmt.Errorf("tablero no encontrado")
	}
	a.state.Boards = append(a.state.Boards[:i], a.state.Boards[i+1:]...)
	if a.state.ActiveBoardID == id {
		a.state.ActiveBoardID = a.state.Boards[0].ID
	}
	a.saveStateLocked()
	return a.snapshotLocked(), nil
}

// SetActiveBoard switches the visible board.
func (a *App) SetActiveBoard(id string) (AppState, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.boardIndex(id) < 0 {
		return AppState{}, fmt.Errorf("tablero no encontrado")
	}
	a.state.ActiveBoardID = id
	a.saveStateLocked()
	return a.snapshotLocked(), nil
}

// SaveViewport persists the React Flow camera.
func (a *App) SaveViewport(boardID string, v Viewport) error {
	if v.Zoom < 0.2 {
		v.Zoom = 0.2
	}
	if v.Zoom > 2.5 {
		v.Zoom = 2.5
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	i := a.boardIndex(boardID)
	if i < 0 {
		return fmt.Errorf("tablero no encontrado")
	}
	a.state.Boards[i].Viewport = v
	a.saveStateLocked()
	return nil
}

// AddCard adds a folder or video reference to a board.
func (a *App) AddCard(boardID, path string, x, y float64) (AppState, error) {
	path = absClean(path)
	a.mu.RLock()
	root := a.state.Workspace
	a.mu.RUnlock()
	if !insideRootResolved(root, path) {
		return AppState{}, fmt.Errorf("el elemento está fuera del espacio de trabajo")
	}
	info, err := os.Stat(path)
	if err != nil {
		return AppState{}, fmt.Errorf("no se encontró el elemento")
	}
	kind := "folder"
	label := info.Name()
	date := dateLabel(label)
	width, height := 286.0, 124.0
	if !info.IsDir() {
		if !isVideo(path) {
			return AppState{}, fmt.Errorf("solo se pueden añadir carpetas o videos")
		}
		kind = "video"
		width, height = 304.0, 260.0
		if d := dateLabel(filepath.Base(filepath.Dir(path))); d != "" {
			date = d
		}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	i := a.boardIndex(boardID)
	if i < 0 {
		return AppState{}, fmt.Errorf("tablero no encontrado")
	}
	for _, c := range a.state.Boards[i].Cards {
		if samePath(c.Path, path) {
			return a.snapshotLocked(), nil
		}
	}
	if x == 0 && y == 0 {
		x, y = 80, 100
	}
	card := Card{
		ID:        newID("card"),
		Kind:      kind,
		Path:      path,
		Label:     cleanLabel(label),
		DateLabel: date,
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
		AddedAt:   time.Now().UnixMilli(),
	}
	a.state.Boards[i].Cards = append(a.state.Boards[i].Cards, card)
	a.saveStateLocked()
	return a.snapshotLocked(), nil
}

// MoveCard persists a card position after dragging.
func (a *App) MoveCard(boardID, cardID string, x, y float64) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	i := a.boardIndex(boardID)
	if i < 0 {
		return fmt.Errorf("tablero no encontrado")
	}
	for j := range a.state.Boards[i].Cards {
		if a.state.Boards[i].Cards[j].ID == cardID {
			a.state.Boards[i].Cards[j].X = x
			a.state.Boards[i].Cards[j].Y = y
			a.saveStateLocked()
			return nil
		}
	}
	return fmt.Errorf("tarjeta no encontrada")
}

// RemoveCard removes a card. It never touches the underlying file.
func (a *App) RemoveCard(boardID, cardID string) (AppState, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	i := a.boardIndex(boardID)
	if i < 0 {
		return AppState{}, fmt.Errorf("tablero no encontrado")
	}
	cards := a.state.Boards[i].Cards
	for j := range cards {
		if cards[j].ID == cardID {
			a.state.Boards[i].Cards = append(cards[:j], cards[j+1:]...)
			a.saveStateLocked()
			return a.snapshotLocked(), nil
		}
	}
	return a.snapshotLocked(), nil
}

// ListDir lists direct children of a folder inside the workspace.
func (a *App) ListDir(path string) ([]DirEntry, error) {
	a.mu.RLock()
	root := a.state.Workspace
	a.mu.RUnlock()
	if root == "" {
		return nil, fmt.Errorf("elige una carpeta de clases primero")
	}
	if path == "" {
		path = root
	}
	path = absClean(path)
	if !insideRootResolved(root, path) {
		return nil, fmt.Errorf("carpeta fuera del espacio de trabajo")
	}
	items, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	out := make([]DirEntry, 0, len(items))
	for _, item := range items {
		name := item.Name()
		if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "~$") || strings.EqualFold(name, "desktop.ini") {
			continue
		}
		full := filepath.Join(path, name)
		info, err := item.Info()
		if err != nil {
			continue
		}
		kind := "other"
		if item.IsDir() {
			kind = "folder"
		} else if isVideo(full) {
			kind = "video"
		} else {
			continue
		}
		entry := DirEntry{
			Name:    name,
			Path:    full,
			IsDir:   item.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().UnixMilli(),
			Kind:    kind,
		}
		if kind == "folder" {
			entry.DateLabel = dateLabel(name)
		}
		out = append(out, entry)
	}
	sortEntries(out)
	return out, nil
}

// GetMediaInfo returns duration and transcript status for a video.
func (a *App) GetMediaInfo(path string) (MediaInfo, error) {
	path = absClean(path)
	a.mu.RLock()
	root := a.state.Workspace
	a.mu.RUnlock()
	if !insideRootResolved(root, path) {
		return MediaInfo{}, fmt.Errorf("video fuera del espacio de trabajo")
	}
	info, err := os.Stat(path)
	if err != nil {
		return MediaInfo{}, err
	}
	if info.IsDir() {
		return MediaInfo{}, fmt.Errorf("la ruta es una carpeta")
	}
	mi := MediaInfo{
		Path:     path,
		Name:     info.Name(),
		Size:     info.Size(),
		ModTime:  info.ModTime().UnixMilli(),
		Duration: a.probeDuration(path),
		Kind:     "video",
		Transcript: TranscriptMeta{
			Status: "none",
		},
	}
	if tr, err := a.loadTranscript(path); err == nil && tr != nil {
		mi.Transcript = TranscriptMeta{
			Status:    "done",
			Progress:  100,
			Segments:  len(tr.Segments),
			Language:  tr.Language,
			CreatedAt: tr.CreatedAt,
		}
	}
	if job := a.jobsForPath(path); job != nil {
		switch {
		case isActiveStatus(job.Status):
			mi.Transcript = TranscriptMeta{
				Status:   job.Status,
				Progress: job.Progress,
				Stage:    job.Stage,
				Error:    job.Error,
			}
		case job.Status == "error" && mi.Transcript.Status != "done":
			mi.Transcript = TranscriptMeta{Status: "error", Stage: job.Stage, Error: job.Error}
		}
	}
	return mi, nil
}

// GetTranscript returns the full timestamped transcript.
func (a *App) GetTranscript(path string) (Transcript, error) {
	path = absClean(path)
	a.mu.RLock()
	root := a.state.Workspace
	a.mu.RUnlock()
	if !insideRootResolved(root, path) {
		return Transcript{}, fmt.Errorf("video fuera del espacio de trabajo")
	}
	tr, err := a.loadTranscript(path)
	if err != nil {
		return Transcript{}, err
	}
	if tr == nil {
		return Transcript{}, fmt.Errorf("este video todavía no tiene transcripción")
	}
	return *tr, nil
}

// StartTranscription adds a video to the transcription queue.
func (a *App) StartTranscription(path, cardID string) (Job, error) {
	path = absClean(path)
	a.mu.RLock()
	root := a.state.Workspace
	a.mu.RUnlock()
	if !insideRootResolved(root, path) {
		return Job{}, fmt.Errorf("el video está fuera del espacio de trabajo")
	}
	job, err := a.enqueueJob(path, cardID)
	if err != nil {
		return Job{}, err
	}
	return *job, nil
}

// CancelJob cancels a queued or running transcription.
func (a *App) CancelJob(id string) error {
	a.cancelMu.Lock()
	cancel := a.cancels[id]
	a.cancelMu.Unlock()
	if cancel != nil {
		cancel()
	}
	a.updateJob(id, func(j *Job) {
		if j.Status == "queued" {
			j.Status = "canceled"
			j.Stage = "Cancelado"
			j.FinishedAt = time.Now().UnixMilli()
		}
	})
	return nil
}

// GetJobs returns recent jobs, newest first.
func (a *App) GetJobs() []Job {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]Job, 0, len(a.jobOrder))
	for _, id := range a.jobOrder {
		if j := a.jobs[id]; j != nil {
			out = append(out, *j)
		}
	}
	return out
}

// ExportTranscript writes a Markdown or SRT file chosen by the user.
func (a *App) ExportTranscript(mediaPath, format string) (string, error) {
	mediaPath = absClean(mediaPath)
	a.mu.RLock()
	root := a.state.Workspace
	a.mu.RUnlock()
	if !insideRootResolved(root, mediaPath) {
		return "", fmt.Errorf("video fuera del espacio de trabajo")
	}
	tr, err := a.loadTranscript(mediaPath)
	if err != nil {
		return "", err
	}
	if tr == nil {
		return "", fmt.Errorf("no hay transcripción para exportar")
	}
	base := strings.TrimSuffix(filepath.Base(mediaPath), filepath.Ext(mediaPath))
	format = strings.ToLower(strings.TrimSpace(format))
	ext := ".md"
	display := "Markdown"
	pattern := "*.md"
	if format == "srt" {
		ext = ".srt"
		display = "Subtítulos SRT"
		pattern = "*.srt"
	}
	out, err := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:            "Exportar transcripción",
		DefaultDirectory: filepath.Dir(absClean(mediaPath)),
		DefaultFilename:  base + ext,
		Filters:          []wailsRuntime.FileFilter{{DisplayName: display, Pattern: pattern}},
	})
	if err != nil {
		return "", err
	}
	if out == "" {
		return "", nil
	}
	if !strings.HasSuffix(strings.ToLower(out), ext) {
		out += ext
	}
	var content string
	if format == "srt" {
		content = toSRT(tr)
	} else {
		content = toMarkdown(tr)
	}
	if err := writeFileAtomic(out, []byte(content), 0o600); err != nil {
		return "", err
	}
	return out, nil
}

// SetLanguage changes the language used for the next transcriptions.
func (a *App) SetLanguage(language string) (AppState, error) {
	language = strings.ToLower(strings.TrimSpace(language))
	switch language {
	case "es", "en", "auto":
	default:
		return AppState{}, fmt.Errorf("idioma no soportado")
	}
	a.mu.Lock()
	a.language = language
	a.state.Language = language
	a.saveStateLocked()
	state := a.snapshotLocked()
	a.mu.Unlock()
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "engine:update", state)
	}
	return state, nil
}

// CopyText copies arbitrary text to the clipboard.
func (a *App) CopyText(text string) error {
	return wailsRuntime.ClipboardSetText(a.ctx, text)
}

// Reveal shows a file or folder in Windows Explorer.
func (a *App) Reveal(path string) error {
	path = absClean(path)
	a.mu.RLock()
	root := a.state.Workspace
	a.mu.RUnlock()
	if !insideRootResolved(root, path) {
		return fmt.Errorf("ruta fuera del espacio de trabajo")
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	var cmd *exec.Cmd
	if info.IsDir() {
		cmd = exec.Command("explorer.exe", path)
	} else {
		cmd = exec.Command("explorer.exe", "/select,", path)
	}
	return cmd.Start()
}

// OpenExternal opens a video with the default Windows player.
func (a *App) OpenExternal(path string) error {
	path = absClean(path)
	a.mu.RLock()
	root := a.state.Workspace
	a.mu.RUnlock()
	if !insideRootResolved(root, path) || !isVideo(path) {
		return fmt.Errorf("ruta fuera del espacio de trabajo o no es un video")
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return fmt.Errorf("no se encontró el video")
	}
	cmd := exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", path)
	hideWindow(cmd)
	return cmd.Start()
}

// OpenURL opens an http/https link in the user's default browser. Workflow
// reports are model output and therefore untrusted: they must never be able to
// open local files, scripts, or navigate the app WebView.
func (a *App) OpenURL(rawURL string) error {
	parsed, err := validateExternalURL(rawURL)
	if err != nil {
		return err
	}
	if a.ctx == nil {
		return fmt.Errorf("la aplicación no está lista")
	}
	wailsRuntime.BrowserOpenURL(a.ctx, parsed.String())
	return nil
}

// CaptureFrame extracts a frame at a given second and returns a JPEG data URL.
func (a *App) CaptureFrame(path string, second float64) (string, error) {
	a.mu.RLock()
	root := a.state.Workspace
	a.mu.RUnlock()
	return a.captureFrameData(root, path, second)
}

func (a *App) captureFrameData(root, path string, second float64) (string, error) {
	if !insideRootResolved(root, path) {
		return "", fmt.Errorf("archivo fuera del espacio de trabajo")
	}
	return a.renderFrame(absClean(path), second)
}

// ClipInfo is a small diagnostic shown in the UI.
type ClipInfo struct {
	Path     string  `json:"path"`
	Duration float64 `json:"duration"`
}

// ProbeMedia returns metadata for a video (used by the player).
func (a *App) ProbeMedia(path string) (ClipInfo, error) {
	path = absClean(path)
	a.mu.RLock()
	root := a.state.Workspace
	a.mu.RUnlock()
	if !insideRootResolved(root, path) {
		return ClipInfo{}, fmt.Errorf("ruta fuera del espacio de trabajo")
	}
	return ClipInfo{Path: path, Duration: a.probeDuration(path)}, nil
}

func formatSeconds(s float64) string {
	if s < 0 {
		s = 0
	}
	total := int(s + 0.5)
	h := total / 3600
	m := (total % 3600) / 60
	sec := total % 60
	if h > 0 {
		return fmt.Sprintf("%dh %02dm", h, m)
	}
	return fmt.Sprintf("%dm %02ds", m, sec)
}

func timestamp(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	total := int(seconds)
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func toMarkdown(tr *Transcript) string {
	var b strings.Builder
	b.WriteString("# " + strings.TrimSuffix(filepath.Base(tr.MediaPath), filepath.Ext(tr.MediaPath)) + "\n\n")
	b.WriteString(fmt.Sprintf("- **Duración:** %s\n", formatSeconds(tr.Duration)))
	b.WriteString(fmt.Sprintf("- **Idioma:** %s\n", tr.Language))
	b.WriteString(fmt.Sprintf("- **Modelo:** %s\n", tr.Model))
	b.WriteString(fmt.Sprintf("- **Generado:** %s\n\n", time.UnixMilli(tr.CreatedAt).Format("2006-01-02 15:04")))
	b.WriteString("## Transcripción con timestamps\n\n")
	for _, s := range tr.Segments {
		b.WriteString(fmt.Sprintf("[%s] %s\n\n", timestamp(s.Start), s.Text))
	}
	return b.String()
}

func toSRT(tr *Transcript) string {
	var b strings.Builder
	for i, s := range tr.Segments {
		end := s.End
		if end <= s.Start {
			end = s.Start + 0.5
		}
		b.WriteString(strconv.Itoa(i+1) + "\n")
		b.WriteString(srtTime(s.Start) + " --> " + srtTime(end) + "\n")
		b.WriteString(s.Text + "\n\n")
	}
	return b.String()
}

func srtTime(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	total := int(seconds)
	ms := int((seconds - float64(total)) * 1000)
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}

func sortedJobs(jobs []Job) []Job {
	sort.SliceStable(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt > jobs[j].CreatedAt
	})
	return jobs
}

var _ = json.Marshal
