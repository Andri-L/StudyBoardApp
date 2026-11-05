package main

import (
	"strings"
	"testing"
)

func TestDateLabel(t *testing.T) {
	cases := map[string]string{
		"20Jul26":     "20 jul 2026",
		"2Ene26":      "02 ene 2026",
		"2026-07-20":  "20 jul 2026",
		"20-07-2026":  "20 jul 2026",
		"no-es-fecha": "",
	}
	for input, want := range cases {
		if got := dateLabel(input); got != want {
			t.Errorf("dateLabel(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestInsideRoot(t *testing.T) {
	root := `C:\Clases`
	if !insideRoot(root, `C:\Clases\IA\20Jul26`) {
		t.Fatal("expected child path to be inside root")
	}
	if insideRoot(root, `C:\Otra\video.mp4`) {
		t.Fatal("expected sibling path to be outside root")
	}
	if insideRoot(root, `C:\Clases\..\secreto.mp4`) {
		t.Fatal("expected traversal to be outside root")
	}
}

func TestValidateFolderName(t *testing.T) {
	valid := []string{"20Jul26", "Spark Notes", "clase-01"}
	for _, name := range valid {
		if err := validateFolderName(name); err != nil {
			t.Errorf("validateFolderName(%q) unexpected error: %v", name, err)
		}
	}
	invalid := []string{"", ".", "..", `a/b`, `a\b`, "a:b", "CON", "nota."}
	for _, name := range invalid {
		if err := validateFolderName(name); err == nil {
			t.Errorf("validateFolderName(%q) expected error", name)
		}
	}
}

func TestMarkdownAndSRT(t *testing.T) {
	tr := &Transcript{
		MediaPath: `C:\Clases\IA\20Jul26\clase.mp4`,
		Duration:  87,
		Language:  "es",
		Model:     "large-v3-turbo",
		Segments: []Segment{
			{Start: 0, End: 4.2, Text: "Bienvenidos a la clase."},
			{Start: 65.5, End: 70, Text: "Nos vemos la próxima semana."},
		},
	}
	md := toMarkdown(tr)
	if !strings.Contains(md, "[00:00:00] Bienvenidos a la clase.") {
		t.Fatalf("markdown missing first timestamp: %s", md)
	}
	if !strings.Contains(md, "[00:01:05] Nos vemos la próxima semana.") {
		t.Fatalf("markdown missing second timestamp: %s", md)
	}
	srt := toSRT(tr)
	if !strings.Contains(srt, "00:00:00,000 --> 00:00:04,200") {
		t.Fatalf("srt missing first cue: %s", srt)
	}
	if !strings.Contains(srt, "00:01:05,500 --> 00:01:10,000") {
		t.Fatalf("srt missing second cue: %s", srt)
	}
}
