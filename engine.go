package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const preferredModel = "ggml-large-v3-turbo-q5_0.bin"

type enginePaths struct {
	BaseDir    string
	WhisperDir string
	Whisper    string
	FFmpeg     string
	FFprobe    string
	Model      string
	VAD        string
}

func executableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exe)
}

// engineCandidates returns possible locations of the portable engine folder.
// The first one that contains bin/whisper/whisper-cli.exe wins.
func engineCandidates() []string {
	seen := map[string]bool{}
	var out []string
	add := func(p string) {
		if p == "" {
			return
		}
		p = filepath.Clean(p)
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	if env := os.Getenv("STUDYBOARD_ENGINE_DIR"); env != "" {
		add(env)
	}
	// Only trust locations derived from the executable or explicit environment
	// variables. Never discover engine binaries in the current directory: the
	// app can be started from a class workspace, and a planted ffmpeg/whisper
	// executable must not run.
	exe := executableDir()
	add(exe)
	add(filepath.Join(exe, ".."))
	add(filepath.Join(exe, "..", ".."))
	add(filepath.Join(exe, "..", "..", ".."))
	if local := os.Getenv("LOCALAPPDATA"); local != "" {
		add(filepath.Join(local, "StudyBoard App"))
	}
	return out
}

func findEngine(selected string) enginePaths {
	for _, base := range engineCandidates() {
		whisper := filepath.Join(base, "bin", "whisper", "whisper-cli.exe")
		if _, err := os.Stat(whisper); err == nil {
			return buildEnginePaths(base, selected)
		}
	}
	return buildEnginePaths("", selected)
}

func buildEnginePaths(base, selected string) enginePaths {
	p := enginePaths{BaseDir: base}
	if base == "" {
		return p
	}
	p.WhisperDir = filepath.Join(base, "bin", "whisper")
	p.Whisper = filepath.Join(p.WhisperDir, "whisper-cli.exe")
	p.FFmpeg = filepath.Join(base, "bin", "ffmpeg", "ffmpeg.exe")
	p.FFprobe = filepath.Join(base, "bin", "ffmpeg", "ffprobe.exe")
	p.Model = chooseModel(modelDirs(base), selected)
	p.VAD = findFileInDirs(modelDirs(base), "ggml-silero-v5.1.2.bin")
	return p
}

func modelDirs(base string) []string {
	var dirs []string
	if env := os.Getenv("STUDYBOARD_MODELS_DIR"); env != "" {
		dirs = append(dirs, env)
	}
	if base != "" {
		dirs = append(dirs, filepath.Join(base, "bin", "models"))
	}
	if appData := os.Getenv("APPDATA"); appData != "" {
		dirs = append(dirs, filepath.Join(appData, "StudyBoard App", "models"))
	}
	if local := os.Getenv("LOCALAPPDATA"); local != "" {
		dirs = append(dirs, filepath.Join(local, "StudyBoard App", "models"))
	}
	return uniqueStrings(dirs)
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		clean := filepath.Clean(value)
		if !seen[clean] {
			seen[clean] = true
			out = append(out, clean)
		}
	}
	return out
}

func chooseModel(dirs []string, selected string) string {
	if selected != "" {
		if path := findFileInDirs(dirs, selected); path != "" {
			return path
		}
	}
	preferred := []string{
		preferredModel,
		"ggml-small-q5_1.bin",
		"ggml-small.bin",
		"ggml-medium-q5_0.bin",
		"ggml-base-q5_1.bin",
		"ggml-base.bin",
		"ggml-tiny.bin",
		"ggml-distil-large-v3.bin",
	}
	for _, name := range preferred {
		if path := findFileInDirs(dirs, name); path != "" {
			return path
		}
	}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), "ggml-") && strings.HasSuffix(entry.Name(), ".bin") {
				names = append(names, entry.Name())
			}
		}
		if len(names) > 0 {
			sort.Strings(names)
			return filepath.Join(dir, names[0])
		}
	}
	return ""
}

func findFileInDirs(dirs []string, name string) string {
	for _, dir := range dirs {
		candidate := filepath.Join(dir, name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func buildEngineStatus(p enginePaths) EngineStatus {
	s := EngineStatus{
		WhisperPath: p.Whisper,
		FFmpegPath:  p.FFmpeg,
		FFprobePath: p.FFprobe,
		ModelPath:   p.Model,
		VADPath:     p.VAD,
		Missing:     []string{},
	}
	if p.Model != "" {
		s.ModelName = strings.TrimSuffix(filepath.Base(p.Model), ".bin")
	}
	s.GPU = p.GPU()
	checks := []struct {
		label string
		path  string
	}{
		{"Whisper", p.Whisper},
		{"ffmpeg", p.FFmpeg},
		{"ffprobe", p.FFprobe},
		{"modelo de voz", p.Model},
	}
	for _, c := range checks {
		if c.path == "" {
			s.Missing = append(s.Missing, c.label)
			continue
		}
		if _, err := os.Stat(c.path); err != nil {
			s.Missing = append(s.Missing, c.label)
		}
	}
	s.Ready = len(s.Missing) == 0
	return s
}

func (e enginePaths) GPU() string {
	if e.Whisper == "" {
		return ""
	}
	if _, err := os.Stat(filepath.Join(e.WhisperDir, "ggml-vulkan.dll")); err == nil {
		return "Vulkan"
	}
	return "CPU"
}

func (e enginePaths) validate() error {
	if e.Whisper == "" {
		return fmt.Errorf("no se encontró whisper-cli.exe")
	}
	if e.FFmpeg == "" {
		return fmt.Errorf("no se encontró ffmpeg.exe")
	}
	if e.FFprobe == "" {
		return fmt.Errorf("no se encontró ffprobe.exe")
	}
	if e.Model == "" {
		return fmt.Errorf("no se encontró el modelo de Whisper")
	}
	return nil
}
