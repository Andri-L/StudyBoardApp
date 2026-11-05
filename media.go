package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type mediaServer struct {
	app     *App
	server  *http.Server
	baseURL string
	token   string
}

func randomToken() string {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(b)
}

func startMediaServer(app *App) (*mediaServer, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	token := randomToken()
	m := &mediaServer{app: app, token: token}
	prefix := "/m/" + token
	mux := http.NewServeMux()
	mux.HandleFunc(prefix+"/file", m.handleFile)
	mux.HandleFunc(prefix+"/thumb", m.handleThumb)
	m.server = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 15 * time.Second,
	}
	m.baseURL = fmt.Sprintf("http://%s%s", ln.Addr().String(), prefix)
	go func() {
		_ = m.server.Serve(ln)
	}()
	return m, nil
}

func (m *mediaServer) stop() {
	if m != nil && m.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = m.server.Shutdown(ctx)
	}
}

func (m *mediaServer) allowed(path string) bool {
	if path == "" {
		return false
	}
	m.app.mu.RLock()
	root := m.app.state.Workspace
	m.app.mu.RUnlock()
	if !insideRoot(root, path) {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func (m *mediaServer) handleFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Query().Get("path")
	if !m.allowed(path) {
		http.Error(w, "archivo no permitido", http.StatusForbidden)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "private, max-age=3600")
	http.ServeFile(w, r, path)
}

func (m *mediaServer) handleThumb(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if !m.allowed(path) {
		http.Error(w, "archivo no permitido", http.StatusForbidden)
		return
	}
	info, err := os.Stat(path)
	if err != nil {
		http.Error(w, "no existe", http.StatusNotFound)
		return
	}
	sum := sha256.Sum256([]byte(strings.ToLower(absClean(path)) + "|" + strconv.FormatInt(info.ModTime().UnixNano(), 10)))
	name := hex.EncodeToString(sum[:12]) + ".jpg"
	thumbPath := filepath.Join(m.app.cacheDir, "thumbs", name)
	if _, err := os.Stat(thumbPath); err != nil {
		if err := m.app.makeThumbnail(path, thumbPath); err != nil {
			http.Error(w, "no se pudo generar miniatura", http.StatusNotFound)
			return
		}
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "private, max-age=86400")
	http.ServeFile(w, r, thumbPath)
}

func (a *App) makeThumbnail(mediaPath, outPath string) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	tmpPath := strings.TrimSuffix(outPath, filepath.Ext(outPath)) + ".tmp-" + newID("thumb") + filepath.Ext(outPath)
	try := func(seek string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, a.engine.FFmpeg,
			"-hide_banner", "-loglevel", "error", "-y",
			"-ss", seek, "-i", mediaPath,
			"-frames:v", "1",
			"-vf", "scale=720:-2",
			"-q:v", "4",
			tmpPath,
		)
		hideWindow(cmd)
		var stderr strings.Builder
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err == nil {
			return nil
		}
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(stderr.String()))
	}
	var lastErr error
	for _, seek := range []string{"00:00:02", "00:00:00.2"} {
		if err := try(seek); err == nil {
			if err := os.Rename(tmpPath, outPath); err == nil {
				return nil
			}
			if _, statErr := os.Stat(outPath); statErr == nil {
				_ = os.Remove(tmpPath)
				return nil
			}
			lastErr = fmt.Errorf("no se pudo guardar la miniatura")
			break
		} else {
			lastErr = err
		}
	}
	_ = os.Remove(tmpPath)
	return lastErr
}

// renderFrame returns a JPEG data URL for a specific second of a video.
func (a *App) renderFrame(mediaPath string, second float64) (string, error) {
	if second < 0 {
		second = 0
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	dir := filepath.Join(a.cacheDir, "frames")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	out := filepath.Join(dir, newID("frame")+".jpg")
	cmd := exec.CommandContext(ctx, a.engine.FFmpeg,
		"-hide_banner", "-loglevel", "error", "-y",
		"-ss", strconv.FormatFloat(second, 'f', 3, 64),
		"-i", mediaPath,
		"-frames:v", "1",
		"-vf", "scale=1280:-2",
		"-q:v", "3",
		out,
	)
	hideWindow(cmd)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("no se pudo capturar el frame: %s", strings.TrimSpace(stderr.String()))
	}
	data, err := os.ReadFile(out)
	_ = os.Remove(out)
	if err != nil {
		return "", err
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(data), nil
}
