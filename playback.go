package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var nativePlaybackExts = map[string]bool{
	".mp4":  true,
	".webm": true,
	".m4v":  true,
}

func playbackSupported(path string) bool {
	return nativePlaybackExts[strings.ToLower(filepath.Ext(path))]
}

func (a *App) playbackDir() string {
	return filepath.Join(a.cacheDir, "playback")
}

func (a *App) playbackCacheFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(strings.ToLower(absClean(path)) + "|" +
		strconv.FormatInt(info.Size(), 10) + "|" +
		strconv.FormatInt(info.ModTime().UnixNano(), 10)))
	return filepath.Join(a.playbackDir(), hex.EncodeToString(sum[:])+".mp4"), nil
}

// PreparePlayback returns a WebView2-compatible copy of a video, converting it
// with FFmpeg when the container is not natively playable.
func (a *App) PreparePlayback(path string) (string, error) {
	path = absClean(path)
	a.mu.RLock()
	root := a.state.Workspace
	a.mu.RUnlock()
	if !insideRootResolved(root, path) {
		return "", fmt.Errorf("el video está fuera del espacio de trabajo")
	}
	if !isVideo(path) {
		return "", fmt.Errorf("el archivo no es un video")
	}
	if playbackSupported(path) {
		return path, nil
	}
	if a.engine.FFmpeg == "" {
		return "", fmt.Errorf("ffmpeg no está disponible")
	}
	out, err := a.playbackCacheFile(path)
	if err != nil {
		return "", err
	}
	if info, err := os.Stat(out); err == nil && !info.IsDir() {
		return out, nil
	}
	if err := os.MkdirAll(a.playbackDir(), 0o700); err != nil {
		return "", err
	}

	a.playbackMu.Lock()
	defer a.playbackMu.Unlock()
	if info, err := os.Stat(out); err == nil && !info.IsDir() {
		return out, nil
	}
	tmp := strings.TrimSuffix(out, ".mp4") + ".tmp-" + newID("playback") + ".mp4"
	defer os.Remove(tmp)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, a.engine.FFmpeg,
		"-hide_banner", "-loglevel", "error", "-y",
		"-i", path,
		"-map", "0:v:0",
		"-map", "0:a:0?",
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-crf", "23",
		"-c:a", "aac",
		"-b:a", "128k",
		"-movflags", "+faststart",
		tmp,
	)
	hideWindow(cmd)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("la conversión tardó demasiado")
		}
		return "", fmt.Errorf("no se pudo convertir el video: %s", lastLines(stderr.String(), 4))
	}
	if err := os.Rename(tmp, out); err != nil {
		return "", err
	}
	return out, nil
}
