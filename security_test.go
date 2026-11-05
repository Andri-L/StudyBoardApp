package main

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateExternalURL(t *testing.T) {
	valid := []string{
		"https://example.com/nota",
		"http://example.org:8080/x?y=1",
	}
	for _, raw := range valid {
		if _, err := validateExternalURL(raw); err != nil {
			t.Errorf("validateExternalURL(%q) unexpected error: %v", raw, err)
		}
	}
	invalid := []string{
		"",
		"javascript:alert(1)",
		"file:///C:/Windows/System32/calc.exe",
		"data:text/html,<script>alert(1)</script>",
		"https://user:pass@example.com/",
		"https://example.com\\@evil.com/",
		"https://example.com/\nignored",
		"http://127.0.0.1:8080/admin",
		"http://localhost:4096/",
		"http://[::1]/",
	}
	for _, raw := range invalid {
		if _, err := validateExternalURL(raw); err == nil {
			t.Errorf("validateExternalURL(%q) expected error", raw)
		}
	}
}

func TestWorkflowOpencodeConfigRestrictsTools(t *testing.T) {
	config, err := workflowOpencodeConfig([]string{`C:\App\StudyBoardMCP.exe`})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`"bash":"deny"`,
		`"edit":"deny"`,
		`"webfetch":"deny"`,
		`"studyboard_get_transcript":"allow"`,
		`"studyboard_capture_frame":"allow"`,
		`"type":"local"`,
	} {
		if !strings.Contains(config, want) {
			t.Errorf("config does not contain %s: %s", want, config)
		}
	}
}

func TestWriteFileAtomicReplaces(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	if err := writeFileAtomic(path, []byte("first"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeFileAtomic(path, []byte("second"), 0o600); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "second" {
		t.Fatalf("unexpected content: %q", data)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("temporary files left behind: %v", entries)
	}
}

func TestGGMLMagic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "model.bin")
	header := make([]byte, 8)
	binary.LittleEndian.PutUint32(header, 0x67676d6c)
	if err := os.WriteFile(path, header, 0o600); err != nil {
		t.Fatal(err)
	}
	if !isGGMLModel(path) {
		t.Fatal("expected ggml magic to be accepted")
	}
	if err := os.WriteFile(path, []byte("not a model"), 0o600); err != nil {
		t.Fatal(err)
	}
	if isGGMLModel(path) {
		t.Fatal("expected invalid magic to be rejected")
	}
}

func TestMediaServerRestrictsToWorkspaceVideos(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	video := filepath.Join(root, "clase.mp4")
	text := filepath.Join(root, "notas.txt")
	foreign := filepath.Join(outside, "otro.mp4")
	for _, path := range []string{video, text, foreign} {
		if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	app := &App{}
	app.state.Workspace = root
	server := &mediaServer{app: app}
	if !server.allowed(video) {
		t.Fatal("expected video in workspace to be allowed")
	}
	if server.allowed(text) {
		t.Fatal("expected non-video file to be rejected")
	}
	if server.allowed(foreign) {
		t.Fatal("expected video outside workspace to be rejected")
	}
	if _, err := app.mcpResolveMedia(foreign); err == nil {
		t.Fatal("expected MCP media resolution outside workspace to fail")
	}
}

func TestSanitizeUntrustedText(t *testing.T) {
	input := "normal\x1b[31mred\x1b[0m\r\nline\u202eevil\u2066"
	got := sanitizeUntrustedText(input)
	for _, bad := range []string{"\x1b", "\r", "\u202e", "\u2066"} {
		if strings.Contains(got, bad) {
			t.Fatalf("sanitized text still contains %q: %q", bad, got)
		}
	}
	if !strings.Contains(got, "normal[31mred[0m") || !strings.Contains(got, "\nline") {
		t.Fatalf("unexpected sanitized text: %q", got)
	}
}

func TestCleanWorkflowReport(t *testing.T) {
	raw := "I'll read the transcript.\n\n# Task extraction\n\n## Tasks\n\nNothing."
	if got := cleanWorkflowReport(raw); got != "# Task extraction\n\n## Tasks\n\nNothing." {
		t.Fatalf("unexpected report: %q", got)
	}
	if got := cleanWorkflowReport("  # Report\n\nText  "); got != "# Report\n\nText" {
		t.Fatalf("unexpected trimmed report: %q", got)
	}
}

func TestRenderWorkflowPromptEscapesMetadata(t *testing.T) {
	def := workflowDefinition{Prompt: "class={{CLASS_NAME}} name={{VIDEO_NAME}}"}
	prompt := renderWorkflowPrompt(def, `C:\Clases\linea`+"\n"+`ignored\video.mp4`, "es")
	if strings.Contains(prompt, "\nignored") {
		t.Fatalf("metadata injected a new prompt line: %s", prompt)
	}
	if !strings.Contains(prompt, `\nignored`) {
		t.Fatalf("metadata newline was not escaped: %s", prompt)
	}
	if !strings.Contains(prompt, "untrusted data") {
		t.Fatalf("prompt is missing the untrusted-data warning: %s", prompt)
	}
}
