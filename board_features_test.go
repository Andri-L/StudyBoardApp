package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newFeatureTestApp(t *testing.T) *App {
	t.Helper()
	root := t.TempDir()
	cache := t.TempDir()
	config := t.TempDir()
	app := NewApp()
	app.state.Workspace = root
	app.state.ActiveBoardID = "board-1"
	app.state.Boards = []Board{{
		ID:       "board-1",
		Name:     "IA",
		Cards:    []Card{},
		Notes:    []Note{},
		Arrows:   []Arrow{},
		Results:  []WorkflowResult{},
		Viewport: Viewport{Zoom: 1},
	}}
	app.cacheDir = cache
	app.configDir = config
	app.statePath = filepath.Join(config, "state.json")
	app.workflowJobs = map[string]*WorkflowJob{}
	return app
}

func TestRenamePathUpdatesCardsResultsAndTranscripts(t *testing.T) {
	app := newFeatureTestApp(t)
	root := app.state.Workspace
	oldPath := filepath.Join(root, "clase.mp4")
	if err := os.WriteFile(oldPath, []byte("video"), 0o600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(oldPath)
	if err != nil {
		t.Fatal(err)
	}
	tr := &Transcript{
		Version:      1,
		MediaPath:    oldPath,
		MediaSize:    info.Size(),
		MediaModTime: info.ModTime().UnixMilli(),
		Duration:     10,
		Segments:     []Segment{{Start: 0, End: 1, Text: "hola"}},
	}
	data, err := json.Marshal(tr)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(app.cacheDir, "transcripts"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(app.transcriptFile(oldPath), data, 0o600); err != nil {
		t.Fatal(err)
	}
	resultFile := filepath.Join(root, "clase - tareas.md")
	if err := os.WriteFile(resultFile, []byte("# reporte"), 0o600); err != nil {
		t.Fatal(err)
	}
	app.state.Boards[0].Cards = append(app.state.Boards[0].Cards, Card{
		ID: "card-1", Kind: "video", Path: oldPath, Label: "clase.mp4",
	})
	app.state.Boards[0].Results = append(app.state.Boards[0].Results, WorkflowResult{
		ID: "result-1", WorkflowID: "task-extraction", WorkflowName: "Tareas",
		VideoPath: oldPath, VideoName: "clase.mp4", ClassName: "IA",
		FilePath: resultFile,
	})

	state, err := app.RenamePath(oldPath, "clase nueva")
	if err != nil {
		t.Fatal(err)
	}
	newPath := filepath.Join(root, "clase nueva.mp4")
	if _, err := os.Stat(newPath); err != nil {
		t.Fatalf("renamed video missing: %v", err)
	}
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("old video still exists: %v", err)
	}
	if state.Boards[0].Cards[0].Path != newPath || state.Boards[0].Cards[0].Label != "clase nueva.mp4" {
		t.Fatalf("card not updated: %+v", state.Boards[0].Cards[0])
	}
	result := state.Boards[0].Results[0]
	if result.VideoPath != newPath || result.VideoName != "clase nueva.mp4" {
		t.Fatalf("result not updated: %+v", result)
	}
	newResultFile := filepath.Join(root, "clase nueva - tareas.md")
	if result.FilePath != newResultFile {
		t.Fatalf("result file path not updated: %s", result.FilePath)
	}
	if _, err := os.Stat(newResultFile); err != nil {
		t.Fatalf("renamed report missing: %v", err)
	}
	moved, err := app.loadTranscript(newPath)
	if err != nil || moved == nil || moved.MediaPath != newPath {
		t.Fatalf("transcript not migrated: %+v, %v", moved, err)
	}
	if prior, _ := app.loadTranscript(oldPath); prior != nil {
		t.Fatal("old transcript still resolves")
	}
}

func TestSetItemColorAndArrows(t *testing.T) {
	app := newFeatureTestApp(t)
	app.state.Boards[0].Cards = append(app.state.Boards[0].Cards, Card{ID: "card-1", Kind: "video", Path: "C:\\x.mp4"})
	if _, err := app.SetItemColor("board-1", "card-1", "#fff4b8"); err != nil {
		t.Fatal(err)
	}
	if app.state.Boards[0].Cards[0].Color != "#fff4b8" {
		t.Fatal("card color not set")
	}
	if _, err := app.SetItemColor("board-1", "card-1", "#ff0000"); err == nil {
		t.Fatal("expected arbitrary color rejection")
	}

	state, err := app.AddArrow("board-1", 10, 10, 100, 60, "s", "#0067c0")
	if err != nil {
		t.Fatal(err)
	}
	arrow := state.Boards[0].Arrows[0]
	if arrow.Type != "s" || arrow.Color != "#0067c0" {
		t.Fatalf("arrow not stored: %+v", arrow)
	}
	state, err = app.UpdateArrow("board-1", arrow.ID, 20, 20, 120, 90, "straight", "#c42b1c")
	if err != nil {
		t.Fatal(err)
	}
	updated := state.Boards[0].Arrows[0]
	if updated.Type != "straight" || updated.Color != "#c42b1c" || updated.X1 != 20 {
		t.Fatalf("arrow not updated: %+v", updated)
	}
	if _, err := app.AddArrow("board-1", 10, 10, 100, 60, "zigzag", "#0067c0"); err == nil {
		t.Fatal("expected invalid arrow type rejection")
	}
}

func TestExtractICSBlock(t *testing.T) {
	content := "# Fechas\n\n## Resultado\n\n| Fecha | Evento |\n\n## Archivo ICS\n\n```ics\nBEGIN:VCALENDAR\nEND:VCALENDAR\n```\n"
	cleaned, ics := extractICSBlock(content)
	if !strings.Contains(ics, "BEGIN:VCALENDAR") || !strings.Contains(ics, "END:VCALENDAR") {
		t.Fatalf("ICS not extracted: %q", ics)
	}
	if strings.Contains(cleaned, "BEGIN:VCALENDAR") || !strings.Contains(cleaned, "# Fechas") {
		t.Fatalf("cleaned report invalid: %q", cleaned)
	}
}

func TestCleanTranslatedReport(t *testing.T) {
	raw := "# Extracción de tareas — clase\n\n- Clase: IA\n\n## Resultado\n\n# Tareas\n\nSin tareas."
	want := "# Tareas\n\nSin tareas."
	if got := cleanTranslatedReport(raw); got != want {
		t.Fatalf("unexpected translated report: %q", got)
	}
}

func TestModelCatalogHashes(t *testing.T) {
	for _, spec := range modelCatalog {
		if len(spec.SHA256) != 64 {
			t.Errorf("%s: missing SHA256", spec.File)
		}
	}
}

func TestPlaybackCacheIsOutsideWorkspace(t *testing.T) {
	app := newFeatureTestApp(t)
	video := filepath.Join(app.state.Workspace, "clase.mkv")
	if err := os.WriteFile(video, []byte("video"), 0o600); err != nil {
		t.Fatal(err)
	}
	cache, err := app.playbackCacheFile(video)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(strings.ToLower(cache), ".mp4") {
		t.Fatalf("unexpected playback cache path: %s", cache)
	}
	if insideRoot(app.state.Workspace, cache) {
		t.Fatal("playback cache must not be inside the workspace")
	}
}
