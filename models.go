package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type modelSpec struct {
	File        string
	Name        string
	Description string
	URL         string
	Size        int64
}

var modelCatalog = []modelSpec{
	{
		File:        "ggml-large-v3-turbo-q5_0.bin",
		Name:        "large-v3-turbo · q5",
		Description: "Equilibrado y recomendado para clases",
		URL:         "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3-turbo-q5_0.bin",
		Size:        574041195,
	},
	{
		File:        "ggml-small-q5_1.bin",
		Name:        "small · q5",
		Description: "Rápido para pruebas de velocidad",
		URL:         "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small-q5_1.bin",
		Size:        190085487,
	},
	{
		File:        "ggml-small.bin",
		Name:        "small",
		Description: "Más preciso que el small q5, más lento",
		URL:         "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin",
		Size:        487601967,
	},
	{
		File:        "ggml-medium-q5_0.bin",
		Name:        "medium · q5",
		Description: "Más calidad, bastante más lento",
		URL:         "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-medium-q5_0.bin",
		Size:        539212467,
	},
	{
		File:        "ggml-base-q5_1.bin",
		Name:        "base · q5",
		Description: "Muy rápido, calidad básica",
		URL:         "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base-q5_1.bin",
		Size:        59707625,
	},
	{
		File:        "ggml-tiny.bin",
		Name:        "tiny",
		Description: "Solo para pruebas rápidas",
		URL:         "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.bin",
		Size:        77691713,
	},
	{
		File:        "ggml-distil-large-v3.bin",
		Name:        "distil-large-v3",
		Description: "Experimental: evaluar velocidad y calidad",
		URL:         "https://huggingface.co/distil-whisper/distil-large-v3-ggml/resolve/main/ggml-distil-large-v3.bin",
		Size:        1519521155,
	},
}

func modelSpecByFile(file string) (modelSpec, bool) {
	for _, spec := range modelCatalog {
		if spec.File == file {
			return spec, true
		}
	}
	return modelSpec{}, false
}

func (a *App) engineStatus() EngineStatus {
	status := buildEngineStatus(a.engine)
	status.Models = a.modelInfos()
	return status
}

func (a *App) modelInfos() []ModelInfo {
	dirs := modelDirs(a.engine.BaseDir)
	a.downloadMu.Lock()
	downloads := make(map[string]ModelDownload, len(a.downloads))
	for file, download := range a.downloads {
		downloads[file] = *download
	}
	a.downloadMu.Unlock()

	activeFile := ""
	if a.engine.Model != "" {
		activeFile = filepath.Base(a.engine.Model)
	}
	out := make([]ModelInfo, 0, len(modelCatalog))
	for _, spec := range modelCatalog {
		path := findFileInDirs(dirs, spec.File)
		info := ModelInfo{
			File:        spec.File,
			Name:        spec.Name,
			Description: spec.Description,
			URL:         spec.URL,
			Size:        spec.Size,
			Downloaded:  path != "",
			Active:      spec.File == activeFile,
			Path:        path,
		}
		if download, ok := downloads[spec.File]; ok {
			info.Downloading = download.Status == "downloading"
			info.Progress = download.Progress
			info.Error = download.Error
		}
		out = append(out, info)
	}
	return out
}

func (a *App) downloadDir() string {
	return filepath.Join(a.cacheDir, "models")
}

// SelectModel switches the active Whisper model.
func (a *App) SelectModel(file string) (AppState, error) {
	if _, ok := modelSpecByFile(file); !ok {
		return AppState{}, fmt.Errorf("modelo desconocido")
	}
	path := findFileInDirs(modelDirs(a.engine.BaseDir), file)
	if path == "" {
		return AppState{}, fmt.Errorf("el modelo todavía no está descargado")
	}
	a.mu.Lock()
	a.selectedModel = file
	a.engine = findEngine(file)
	a.saveStateLocked()
	state := a.snapshotLocked()
	a.mu.Unlock()
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "engine:update", state)
	}
	return state, nil
}

// DownloadModel downloads a model to LocalAppData and selects it on success.
func (a *App) DownloadModel(file string) (ModelDownload, error) {
	spec, ok := modelSpecByFile(file)
	if !ok {
		return ModelDownload{}, fmt.Errorf("modelo desconocido")
	}
	if path := findFileInDirs(modelDirs(a.engine.BaseDir), file); path != "" {
		if _, err := a.SelectModel(file); err != nil {
			return ModelDownload{}, err
		}
		return ModelDownload{File: file, Status: "done", Progress: 100}, nil
	}
	a.downloadMu.Lock()
	if existing := a.downloads[file]; existing != nil && existing.Status == "downloading" {
		cp := *existing
		a.downloadMu.Unlock()
		return cp, nil
	}
	download := &ModelDownload{File: file, Status: "downloading", Total: spec.Size}
	a.downloads[file] = download
	cp := *download
	a.downloadMu.Unlock()
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "model:progress", cp)
	}
	go a.runModelDownload(spec)
	return cp, nil
}

func (a *App) CancelModelDownload(file string) error {
	a.downloadMu.Lock()
	cancel := a.downloadCancels[file]
	a.downloadMu.Unlock()
	if cancel != nil {
		cancel()
	}
	return nil
}

func (a *App) runModelDownload(spec modelSpec) {
	ctx, cancel := context.WithCancel(context.Background())
	a.downloadMu.Lock()
	a.downloadCancels[spec.File] = cancel
	a.downloadMu.Unlock()
	defer func() {
		cancel()
		a.downloadMu.Lock()
		delete(a.downloadCancels, spec.File)
		a.downloadMu.Unlock()
	}()

	update := func(fn func(*ModelDownload)) {
		a.downloadMu.Lock()
		if download := a.downloads[spec.File]; download != nil {
			fn(download)
		}
		var cp ModelDownload
		if download := a.downloads[spec.File]; download != nil {
			cp = *download
		}
		a.downloadMu.Unlock()
		if a.ctx != nil {
			wailsRuntime.EventsEmit(a.ctx, "model:progress", cp)
		}
	}

	fail := func(err error) {
		if ctx.Err() == context.Canceled {
			update(func(d *ModelDownload) {
				d.Status = "canceled"
				d.Error = ""
			})
			return
		}
		update(func(d *ModelDownload) {
			d.Status = "error"
			d.Error = err.Error()
		})
	}

	if err := os.MkdirAll(a.downloadDir(), 0o755); err != nil {
		fail(err)
		return
	}
	dest := filepath.Join(a.downloadDir(), spec.File)
	tmp := dest + ".part"
	_ = os.Remove(tmp)
	defer func() {
		// If the download failed or was canceled, do not leave a partial model.
		if _, err := os.Stat(tmp); err == nil {
			_ = os.Remove(tmp)
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, spec.URL, nil)
	if err != nil {
		fail(err)
		return
	}
	req.Header.Set("User-Agent", "StudyBoardApp/"+appVersion)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fail(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fail(fmt.Errorf("la descarga respondió %s", resp.Status))
		return
	}
	out, err := os.Create(tmp)
	if err != nil {
		fail(err)
		return
	}

	total := resp.ContentLength
	if total <= 0 {
		total = spec.Size
	}
	var written int64
	lastEmit := time.Now()
	buf := make([]byte, 512*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				_ = out.Close()
				_ = os.Remove(tmp)
				fail(writeErr)
				return
			}
			written += int64(n)
			if time.Since(lastEmit) > 350*time.Millisecond {
				lastEmit = time.Now()
				progress := 0.0
				if total > 0 {
					progress = float64(written) / float64(total) * 100
					if progress > 100 {
						progress = 100
					}
				}
				update(func(d *ModelDownload) {
					d.Progress = progress
					d.Bytes = written
					d.Total = total
				})
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			_ = out.Close()
			_ = os.Remove(tmp)
			fail(readErr)
			return
		}
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmp)
		fail(err)
		return
	}
	if total > 0 && written != total {
		_ = os.Remove(tmp)
		fail(fmt.Errorf("descarga incompleta (%d de %d bytes)", written, total))
		return
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		fail(err)
		return
	}
	update(func(d *ModelDownload) {
		d.Status = "done"
		d.Progress = 100
		d.Bytes = written
		d.Total = total
		d.Error = ""
	})
	_, _ = a.SelectModel(spec.File)
}

// OpenModelsFolder opens the LocalAppData models directory in Explorer.
func (a *App) OpenModelsFolder() error {
	dir := a.downloadDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	cmd := exec.Command("explorer.exe", dir)
	return cmd.Start()
}
