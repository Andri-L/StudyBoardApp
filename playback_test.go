package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func findTestFFmpeg() string {
	candidates := []string{
		filepath.Join("bin", "ffmpeg", "ffmpeg.exe"),
		filepath.Join("build", "bin", "bin", "ffmpeg", "ffmpeg.exe"),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			absolute, _ := filepath.Abs(candidate)
			return absolute
		}
	}
	return ""
}

func TestPreparePlaybackConvertsUnsupportedContainer(t *testing.T) {
	ffmpeg := findTestFFmpeg()
	if ffmpeg == "" {
		t.Skip("ffmpeg not available")
	}
	app := newFeatureTestApp(t)
	app.engine.FFmpeg = ffmpeg
	source := filepath.Join(app.state.Workspace, "clase.avi")
	cmd := exec.Command(ffmpeg,
		"-hide_banner", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "color=c=red:s=160x120:d=1",
		"-c:v", "mpeg4", source,
	)
	if err := cmd.Run(); err != nil {
		t.Skipf("could not create test video: %v", err)
	}
	out, err := app.PreparePlayback(source)
	if err != nil {
		t.Fatal(err)
	}
	if out == source {
		t.Fatal("expected a converted playback file")
	}
	info, err := os.Stat(out)
	if err != nil || info.Size() == 0 {
		t.Fatalf("converted file missing: %v", err)
	}
	if !playbackSupported(out) {
		t.Fatalf("converted file has unsupported extension: %s", out)
	}
}
