package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var videoExts = map[string]bool{
	".mp4": true, ".mkv": true, ".mov": true, ".avi": true,
	".webm": true, ".m4v": true, ".wmv": true, ".mpg": true,
	".mpeg": true, ".ts": true, ".flv": true,
}

func newID(prefix string) string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(b)
}

func newBoard(name string) Board {
	if strings.TrimSpace(name) == "" {
		name = "Clases"
	}
	return Board{
		ID:       newID("board"),
		Name:     strings.TrimSpace(name),
		Cards:    []Card{},
		Notes:    []Note{},
		Arrows:   []Arrow{},
		Viewport: Viewport{Zoom: 1},
	}
}

func isVideo(path string) bool {
	return videoExts[strings.ToLower(filepath.Ext(path))]
}

func cleanLabel(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Sin nombre"
	}
	return name
}

func (a *App) loadState() {
	a.state = AppState{
		Version:  1,
		Boards:   []Board{},
		Language: "es",
	}
	a.language = "es"
	data, err := os.ReadFile(a.statePath)
	if err == nil {
		data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
		var p persistedState
		if json.Unmarshal(data, &p) == nil {
			a.state.Workspace = p.Workspace
			a.state.ActiveBoardID = p.ActiveBoardID
			a.selectedModel = p.SelectedModel
			if p.Language != "" {
				a.language = p.Language
				a.state.Language = p.Language
			}
			if p.Boards != nil {
				a.state.Boards = p.Boards
			}
		}
	}
	if len(a.state.Boards) == 0 {
		b := newBoard("Clases")
		a.state.Boards = []Board{b}
		a.state.ActiveBoardID = b.ID
	}
	found := false
	for _, b := range a.state.Boards {
		if b.ID == a.state.ActiveBoardID {
			found = true
			break
		}
	}
	if !found {
		a.state.ActiveBoardID = a.state.Boards[0].ID
	}
}

func (a *App) saveStateLocked() {
	p := persistedState{
		Version:       1,
		Workspace:     a.state.Workspace,
		Boards:        a.state.Boards,
		ActiveBoardID: a.state.ActiveBoardID,
		SelectedModel: a.selectedModel,
		Language:      a.language,
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return
	}
	tmp := a.statePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return
	}
	if err := os.Rename(tmp, a.statePath); err != nil {
		_ = os.Remove(a.statePath)
		_ = os.Rename(tmp, a.statePath)
	}
}

func (a *App) boardIndex(id string) int {
	for i := range a.state.Boards {
		if a.state.Boards[i].ID == id {
			return i
		}
	}
	return -1
}

func (a *App) activeBoard() *Board {
	if i := a.boardIndex(a.state.ActiveBoardID); i >= 0 {
		return &a.state.Boards[i]
	}
	if len(a.state.Boards) > 0 {
		return &a.state.Boards[0]
	}
	return nil
}

func (a *App) snapshotLocked() AppState {
	state := a.state
	state.Engine = a.engineStatus()
	state.Language = a.language
	if a.media != nil {
		state.MediaBase = a.media.baseURL
	}
	// Return copies of slices so callers cannot mutate stored state.
	state.Boards = make([]Board, len(a.state.Boards))
	for i, b := range a.state.Boards {
		state.Boards[i] = b
		state.Boards[i].Cards = append([]Card(nil), b.Cards...)
		state.Boards[i].Notes = append([]Note(nil), b.Notes...)
		state.Boards[i].Arrows = append([]Arrow(nil), b.Arrows...)
	}
	return state
}

func insideRoot(root, path string) bool {
	if root == "" || path == "" {
		return false
	}
	rootAbs, err1 := filepath.Abs(root)
	pathAbs, err2 := filepath.Abs(path)
	if err1 != nil || err2 != nil {
		return false
	}
	rel, err := filepath.Rel(rootAbs, pathAbs)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func samePath(a, b string) bool {
	aa, err1 := filepath.Abs(a)
	bb, err2 := filepath.Abs(b)
	if err1 != nil || err2 != nil {
		return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
	}
	return strings.EqualFold(filepath.Clean(aa), filepath.Clean(bb))
}

func absClean(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return filepath.Clean(abs)
}

var spanishMonths = map[string]time.Month{
	"ene": time.January, "feb": time.February, "mar": time.March,
	"abr": time.April, "may": time.May, "jun": time.June,
	"jul": time.July, "ago": time.August, "sep": time.September,
	"set": time.September, "oct": time.October, "nov": time.November,
	"dic": time.December,
}

var latinMonths = map[string]time.Month{
	"jan": time.January, "apr": time.April, "aug": time.August,
	"dec": time.December,
}

var datePatterns = []*regexp.Regexp{
	regexp.MustCompile(`^(\d{1,2})[ _.-]?([A-Za-zÁÉÍÓÚáéíóú]{3,4})[ _.-]?(\d{2,4})$`),
	regexp.MustCompile(`^(\d{4})[-_.](\d{1,2})[-_.](\d{1,2})$`),
	regexp.MustCompile(`^(\d{1,2})[-_.](\d{1,2})[-_.](\d{2,4})$`),
}

// dateLabel turns common class folder names into a readable date:
// 20Jul26 -> 20 jul 2026, 2026-07-20 -> 20 jul 2026.
func dateLabel(name string) string {
	n := strings.TrimSpace(name)
	if n == "" {
		return ""
	}
	var day, month, year int
	if m := datePatterns[0].FindStringSubmatch(n); m != nil {
		day, _ = strconv.Atoi(m[1])
		key := strings.ToLower(strings.TrimSpace(m[2]))
		if len(key) > 3 {
			key = key[:3]
		}
		mo, ok := spanishMonths[key]
		if !ok {
			mo, ok = latinMonths[key]
		}
		if !ok {
			return ""
		}
		month = int(mo)
		year, _ = strconv.Atoi(m[3])
	} else if m := datePatterns[1].FindStringSubmatch(n); m != nil {
		year, _ = strconv.Atoi(m[1])
		month, _ = strconv.Atoi(m[2])
		day, _ = strconv.Atoi(m[3])
	} else if m := datePatterns[2].FindStringSubmatch(n); m != nil {
		day, _ = strconv.Atoi(m[1])
		month, _ = strconv.Atoi(m[2])
		year, _ = strconv.Atoi(m[3])
	} else {
		return ""
	}
	if year < 100 {
		year += 2000
	}
	if month < 1 || month > 12 || day < 1 || day > 31 || year < 1990 || year > 2100 {
		return ""
	}
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	monthNames := []string{"ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic"}
	return fmt.Sprintf("%02d %s %d", t.Day(), monthNames[int(t.Month())-1], t.Year())
}

func sortEntries(entries []DirEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
}

// migrateDir moves the old app data directory to the new app name. If the
// rename fails (for example across volumes), it copies the tree.
func migrateDir(oldPath, newPath string) {
	if oldPath == "" || newPath == "" || samePath(oldPath, newPath) {
		return
	}
	if _, err := os.Stat(newPath); err == nil {
		return
	}
	if _, err := os.Stat(oldPath); err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		return
	}
	if err := os.Rename(oldPath, newPath); err == nil {
		return
	}
	_ = copyTree(oldPath, newPath)
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			_ = out.Close()
			return err
		}
		return out.Close()
	})
}
