package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func validateFolderName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("el nombre de la carpeta no puede estar vacío")
	}
	if name == "." || name == ".." {
		return fmt.Errorf("nombre de carpeta inválido")
	}
	if strings.ContainsAny(name, `/\:*?"<>|`) {
		return fmt.Errorf("el nombre no puede contener / \\ : * ? \" < > |")
	}
	if strings.HasSuffix(name, ".") || strings.HasSuffix(name, " ") {
		return fmt.Errorf("el nombre no puede terminar en punto o espacio")
	}
	if len([]rune(name)) > 80 {
		return fmt.Errorf("el nombre es demasiado largo")
	}
	base := strings.ToUpper(strings.TrimSuffix(name, filepath.Ext(name)))
	reserved := map[string]bool{
		"CON": true, "PRN": true, "AUX": true, "NUL": true,
		"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
		"COM6": true, "COM7": true, "COM8": true, "COM9": true,
		"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
		"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
	}
	if reserved[base] {
		return fmt.Errorf("ese nombre está reservado por Windows")
	}
	return nil
}

// CreateFolderCard creates a real directory inside the workspace and adds the
// corresponding folder card to the board.
func (a *App) CreateFolderCard(boardID, parent, name string, x, y float64) (AppState, error) {
	name = strings.TrimSpace(name)
	if err := validateFolderName(name); err != nil {
		return AppState{}, err
	}
	a.mu.RLock()
	root := a.state.Workspace
	a.mu.RUnlock()
	if root == "" {
		return AppState{}, fmt.Errorf("elige una carpeta de clases primero")
	}
	if parent == "" {
		parent = root
	}
	parent = absClean(parent)
	if !insideRootResolved(root, parent) {
		return AppState{}, fmt.Errorf("la carpeta padre está fuera del workspace")
	}
	info, err := os.Stat(parent)
	if err != nil || !info.IsDir() {
		return AppState{}, fmt.Errorf("la carpeta padre no existe")
	}
	full := filepath.Join(parent, name)
	if !insideRoot(root, full) {
		return AppState{}, fmt.Errorf("ruta de carpeta inválida")
	}
	if _, err := os.Stat(full); err == nil {
		return AppState{}, fmt.Errorf("ya existe una carpeta con ese nombre")
	}
	if err := os.Mkdir(full, 0o755); err != nil {
		return AppState{}, fmt.Errorf("no se pudo crear la carpeta: %v", err)
	}
	return a.AddCard(boardID, full, x, y)
}

// AddNote creates a sticky note on a board.
func (a *App) AddNote(boardID, text string, x, y, width, height float64) (AppState, error) {
	if width <= 0 {
		width = 230
	}
	if height <= 0 {
		height = 170
	}
	if width < 160 {
		width = 160
	}
	if height < 120 {
		height = 120
	}
	if width > 700 {
		width = 700
	}
	if height > 700 {
		height = 700
	}
	text = strings.TrimSpace(text)
	if text == "" {
		text = "Nueva nota"
	}
	if len([]rune(text)) > 20000 {
		return AppState{}, fmt.Errorf("la nota es demasiado larga")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	i := a.boardIndex(boardID)
	if i < 0 {
		return AppState{}, fmt.Errorf("tablero no encontrado")
	}
	a.state.Boards[i].Notes = append(a.state.Boards[i].Notes, Note{
		ID:        newID("note"),
		Text:      text,
		Color:     "#fff4b8",
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
		CreatedAt: time.Now().UnixMilli(),
	})
	a.saveStateLocked()
	return a.snapshotLocked(), nil
}

// UpdateNote updates the text and/or color of a sticky note.
func (a *App) UpdateNote(boardID, noteID, text, color string) (AppState, error) {
	if len([]rune(text)) > 20000 {
		return AppState{}, fmt.Errorf("la nota es demasiado larga")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	i := a.boardIndex(boardID)
	if i < 0 {
		return AppState{}, fmt.Errorf("tablero no encontrado")
	}
	for j := range a.state.Boards[i].Notes {
		if a.state.Boards[i].Notes[j].ID == noteID {
			a.state.Boards[i].Notes[j].Text = text
			if strings.TrimSpace(color) != "" {
				a.state.Boards[i].Notes[j].Color = color
			}
			a.saveStateLocked()
			return a.snapshotLocked(), nil
		}
	}
	return AppState{}, fmt.Errorf("nota no encontrada")
}

// MoveNote persists a sticky note position after dragging.
func (a *App) MoveNote(boardID, noteID string, x, y float64) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	i := a.boardIndex(boardID)
	if i < 0 {
		return fmt.Errorf("tablero no encontrado")
	}
	for j := range a.state.Boards[i].Notes {
		if a.state.Boards[i].Notes[j].ID == noteID {
			a.state.Boards[i].Notes[j].X = x
			a.state.Boards[i].Notes[j].Y = y
			a.saveStateLocked()
			return nil
		}
	}
	return fmt.Errorf("nota no encontrada")
}

// RemoveNote deletes a sticky note from the board.
func (a *App) RemoveNote(boardID, noteID string) (AppState, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	i := a.boardIndex(boardID)
	if i < 0 {
		return AppState{}, fmt.Errorf("tablero no encontrado")
	}
	notes := a.state.Boards[i].Notes
	for j := range notes {
		if notes[j].ID == noteID {
			a.state.Boards[i].Notes = append(notes[:j], notes[j+1:]...)
			a.saveStateLocked()
			return a.snapshotLocked(), nil
		}
	}
	return a.snapshotLocked(), nil
}

// AddArrow creates a freeform arrow annotation on a board.
func (a *App) AddArrow(boardID string, x1, y1, x2, y2 float64, arrowType, color string) (AppState, error) {
	if math.Hypot(x2-x1, y2-y1) < 8 {
		return AppState{}, fmt.Errorf("la flecha es demasiado corta")
	}
	if arrowType == "" {
		arrowType = "straight"
	}
	if !validArrowType(arrowType) {
		return AppState{}, fmt.Errorf("tipo de flecha inválido")
	}
	if color == "" {
		color = arrowColors[0]
	}
	if !containsString(arrowColors, color) {
		return AppState{}, fmt.Errorf("color de flecha inválido")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	i := a.boardIndex(boardID)
	if i < 0 {
		return AppState{}, fmt.Errorf("tablero no encontrado")
	}
	a.state.Boards[i].Arrows = append(a.state.Boards[i].Arrows, Arrow{
		ID:        newID("arrow"),
		X1:        x1,
		Y1:        y1,
		X2:        x2,
		Y2:        y2,
		Type:      arrowType,
		Color:     color,
		CreatedAt: time.Now().UnixMilli(),
	})
	a.saveStateLocked()
	return a.snapshotLocked(), nil
}

// UpdateArrow persists the geometry and style of an arrow after dragging.
func (a *App) UpdateArrow(boardID, arrowID string, x1, y1, x2, y2 float64, arrowType, color string) (AppState, error) {
	if math.Hypot(x2-x1, y2-y1) < 4 {
		return AppState{}, fmt.Errorf("la flecha es demasiado corta")
	}
	if arrowType == "" {
		arrowType = "straight"
	}
	if !validArrowType(arrowType) {
		return AppState{}, fmt.Errorf("tipo de flecha inválido")
	}
	if !containsString(arrowColors, color) {
		return AppState{}, fmt.Errorf("color de flecha inválido")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	i := a.boardIndex(boardID)
	if i < 0 {
		return AppState{}, fmt.Errorf("tablero no encontrado")
	}
	for j := range a.state.Boards[i].Arrows {
		arrow := &a.state.Boards[i].Arrows[j]
		if arrow.ID == arrowID {
			arrow.X1 = x1
			arrow.Y1 = y1
			arrow.X2 = x2
			arrow.Y2 = y2
			arrow.Type = arrowType
			arrow.Color = color
			a.saveStateLocked()
			return a.snapshotLocked(), nil
		}
	}
	return AppState{}, fmt.Errorf("flecha no encontrada")
}

// SetItemColor sets a pastel background for a card, note or workflow result.
func (a *App) SetItemColor(boardID, itemID, color string) (AppState, error) {
	if !containsString(pastelColors, color) {
		return AppState{}, fmt.Errorf("color no permitido")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	i := a.boardIndex(boardID)
	if i < 0 {
		return AppState{}, fmt.Errorf("tablero no encontrado")
	}
	for j := range a.state.Boards[i].Cards {
		if a.state.Boards[i].Cards[j].ID == itemID {
			a.state.Boards[i].Cards[j].Color = color
			a.saveStateLocked()
			return a.snapshotLocked(), nil
		}
	}
	for j := range a.state.Boards[i].Notes {
		if a.state.Boards[i].Notes[j].ID == itemID {
			a.state.Boards[i].Notes[j].Color = color
			a.saveStateLocked()
			return a.snapshotLocked(), nil
		}
	}
	for j := range a.state.Boards[i].Results {
		if a.state.Boards[i].Results[j].ID == itemID {
			a.state.Boards[i].Results[j].Color = color
			a.saveStateLocked()
			return a.snapshotLocked(), nil
		}
	}
	return AppState{}, fmt.Errorf("elemento no encontrado")
}

// RenamePath renames a real video or folder and updates every board reference,
// transcript cache entry and generated Markdown file that points to it.
func (a *App) RenamePath(path, newName string) (AppState, error) {
	path = absClean(path)
	newName = strings.TrimSpace(newName)
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
	if job := a.jobsForPath(path); job != nil && isActiveStatus(job.Status) {
		return AppState{}, fmt.Errorf("espera a que termine la transcripción de este video")
	}
	if a.hasActiveWorkflow(path) {
		return AppState{}, fmt.Errorf("espera a que termine el workflow de este video")
	}
	parent := filepath.Dir(path)
	if !insideRootResolved(root, parent) {
		return AppState{}, fmt.Errorf("la carpeta padre está fuera del espacio de trabajo")
	}
	var finalName string
	if info.IsDir() {
		if err := validateFolderName(newName); err != nil {
			return AppState{}, err
		}
		finalName = newName
	} else {
		if !isVideo(path) {
			return AppState{}, fmt.Errorf("solo se pueden renombrar videos o carpetas")
		}
		oldExt := filepath.Ext(path)
		newExt := filepath.Ext(newName)
		switch {
		case newExt == "":
			finalName = newName + oldExt
		case strings.EqualFold(newExt, oldExt):
			finalName = newName
		default:
			return AppState{}, fmt.Errorf("mantén la extensión %s del video", oldExt)
		}
		if strings.TrimSpace(strings.TrimSuffix(finalName, filepath.Ext(finalName))) == "" {
			return AppState{}, fmt.Errorf("el nombre del video no puede estar vacío")
		}
		if err := validateFolderName(finalName); err != nil {
			return AppState{}, err
		}
	}
	newPath := filepath.Join(parent, finalName)
	if !insideRoot(root, newPath) {
		return AppState{}, fmt.Errorf("nombre inválido")
	}
	if path == newPath {
		a.mu.RLock()
		defer a.mu.RUnlock()
		return a.snapshotLocked(), nil
	}
	if _, err := os.Stat(newPath); err == nil && !samePath(path, newPath) {
		return AppState{}, fmt.Errorf("ya existe un elemento con ese nombre")
	}

	var oldTranscript *Transcript
	if !info.IsDir() {
		oldTranscript, _ = a.loadTranscript(path)
	}
	oldResultFiles := a.resultFilesForPath(path, info.IsDir())
	if err := os.Rename(path, newPath); err != nil {
		return AppState{}, fmt.Errorf("no se pudo renombrar: %v", err)
	}

	resultFileMap := map[string]string{}
	attachmentMap := map[string]string{}
	if !info.IsDir() {
		oldBase := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		newBase := strings.TrimSuffix(filepath.Base(newPath), filepath.Ext(newPath))
		for resultID, oldFile := range oldResultFiles {
			if oldFile == "" {
				continue
			}
			name := filepath.Base(oldFile)
			candidate := name
			if strings.HasPrefix(strings.ToLower(name), strings.ToLower(oldBase+" - ")) {
				candidate = newBase + name[len(oldBase):]
			}
			newFile := filepath.Join(filepath.Dir(newPath), candidate)
			if err := os.Rename(oldFile, newFile); err == nil {
				resultFileMap[resultID] = newFile
			} else {
				resultFileMap[resultID] = oldFile
			}
		}
		a.mu.RLock()
		for _, board := range a.state.Boards {
			for _, result := range board.Results {
				if !samePath(result.VideoPath, path) || result.AttachmentPath == "" {
					continue
				}
				name := filepath.Base(result.AttachmentPath)
				candidate := name
				if strings.HasPrefix(strings.ToLower(name), strings.ToLower(oldBase+" - ")) {
					candidate = newBase + name[len(oldBase):]
				}
				newFile := filepath.Join(filepath.Dir(newPath), candidate)
				if err := os.Rename(result.AttachmentPath, newFile); err == nil {
					attachmentMap[result.ID] = newFile
				} else {
					attachmentMap[result.ID] = result.AttachmentPath
				}
			}
		}
		a.mu.RUnlock()
	}

	a.migrateTranscriptsAfterRename(path, newPath, info.IsDir(), oldTranscript)

	a.mu.Lock()
	oldPrefix := path + string(filepath.Separator)
	newPrefix := newPath + string(filepath.Separator)
	for bi := range a.state.Boards {
		board := &a.state.Boards[bi]
		for ci := range board.Cards {
			card := &board.Cards[ci]
			switch {
			case samePath(card.Path, path):
				card.Path = newPath
				card.Label = cleanLabel(filepath.Base(newPath))
				if info.IsDir() {
					card.DateLabel = dateLabel(filepath.Base(newPath))
				} else {
					card.DateLabel = dateLabel(filepath.Base(filepath.Dir(newPath)))
				}
			case info.IsDir() && strings.HasPrefix(strings.ToLower(card.Path), strings.ToLower(oldPrefix)):
				card.Path = newPrefix + card.Path[len(oldPrefix):]
			}
		}
		for ri := range board.Results {
			result := &board.Results[ri]
			switch {
			case samePath(result.VideoPath, path):
				result.VideoPath = newPath
				result.VideoName = filepath.Base(newPath)
				result.ClassName = filepath.Base(filepath.Dir(newPath))
				if newFile, ok := resultFileMap[result.ID]; ok {
					result.FilePath = newFile
				}
				if newAttachment, ok := attachmentMap[result.ID]; ok {
					result.AttachmentPath = newAttachment
				}
			case info.IsDir() && strings.HasPrefix(strings.ToLower(result.VideoPath), strings.ToLower(oldPrefix)):
				result.VideoPath = newPrefix + result.VideoPath[len(oldPrefix):]
				result.ClassName = filepath.Base(filepath.Dir(result.VideoPath))
				if result.FilePath != "" && strings.HasPrefix(strings.ToLower(result.FilePath), strings.ToLower(oldPrefix)) {
					result.FilePath = newPrefix + result.FilePath[len(oldPrefix):]
				}
				if result.AttachmentPath != "" && strings.HasPrefix(strings.ToLower(result.AttachmentPath), strings.ToLower(oldPrefix)) {
					result.AttachmentPath = newPrefix + result.AttachmentPath[len(oldPrefix):]
				}
			}
		}
	}
	a.saveStateLocked()
	state := a.snapshotLocked()
	a.mu.Unlock()
	return state, nil
}

func boardHasVideo(board Board, videoPath string) bool {
	for _, card := range board.Cards {
		if card.Kind == "video" && samePath(card.Path, videoPath) {
			return true
		}
	}
	return false
}

func (a *App) hasActiveWorkflow(videoPath string) bool {
	a.workflowMu.Lock()
	defer a.workflowMu.Unlock()
	for _, job := range a.workflowJobs {
		if job == nil || !samePath(job.VideoPath, videoPath) {
			continue
		}
		if job.Status == "queued" || job.Status == "running" {
			return true
		}
	}
	return false
}

func (a *App) resultFilesForPath(path string, isDir bool) map[string]string {
	out := map[string]string{}
	a.mu.RLock()
	defer a.mu.RUnlock()
	prefix := path + string(filepath.Separator)
	for _, board := range a.state.Boards {
		for _, result := range board.Results {
			if samePath(result.VideoPath, path) ||
				(isDir && strings.HasPrefix(strings.ToLower(result.VideoPath), strings.ToLower(prefix))) {
				out[result.ID] = result.FilePath
			}
		}
	}
	return out
}

func (a *App) migrateTranscriptsAfterRename(oldPath, newPath string, isDir bool, oldTranscript *Transcript) {
	if !isDir {
		if oldTranscript == nil {
			return
		}
		oldFile := a.transcriptFile(oldPath)
		info, err := os.Stat(newPath)
		if err != nil {
			return
		}
		oldTranscript.MediaPath = newPath
		oldTranscript.MediaSize = info.Size()
		oldTranscript.MediaModTime = info.ModTime().UnixMilli()
		if data, err := json.Marshal(oldTranscript); err == nil {
			_ = writeFileAtomic(a.transcriptFile(newPath), data, 0o600)
			_ = os.Remove(oldFile)
		}
		return
	}
	oldPrefix := oldPath + string(filepath.Separator)
	newPrefix := newPath + string(filepath.Separator)
	entries, err := os.ReadDir(filepath.Join(a.cacheDir, "transcripts"))
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue
		}
		oldFile := filepath.Join(a.cacheDir, "transcripts", entry.Name())
		data, err := os.ReadFile(oldFile)
		if err != nil {
			continue
		}
		var tr Transcript
		if json.Unmarshal(data, &tr) != nil || tr.MediaPath == "" {
			continue
		}
		if !strings.HasPrefix(strings.ToLower(tr.MediaPath), strings.ToLower(oldPrefix)) {
			continue
		}
		tr.MediaPath = newPrefix + tr.MediaPath[len(oldPrefix):]
		if info, err := os.Stat(tr.MediaPath); err == nil {
			tr.MediaSize = info.Size()
			tr.MediaModTime = info.ModTime().UnixMilli()
		}
		if data, err := json.Marshal(&tr); err == nil {
			_ = writeFileAtomic(a.transcriptFile(tr.MediaPath), data, 0o600)
			_ = os.Remove(oldFile)
		}
	}
}

var pastelColors = []string{
	"#fff4b8", "#ffe2e2", "#ffe8cc", "#e8f5e9", "#e0f2fe",
	"#ede9fe", "#fce7f3", "#e5e7eb", "#fef9c3",
}

var arrowColors = []string{
	"#0067c0", "#c42b1c", "#0f7b0f", "#b8860b", "#7c3aed",
	"#0f766e", "#db2777", "#475569", "#111827",
}

func containsString(values []string, value string) bool {
	for _, item := range values {
		if strings.EqualFold(item, value) {
			return true
		}
	}
	return false
}

func validArrowType(value string) bool {
	return value == "straight" || value == "s"
}

// RemoveArrow deletes an arrow annotation.
func (a *App) RemoveArrow(boardID, arrowID string) (AppState, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	i := a.boardIndex(boardID)
	if i < 0 {
		return AppState{}, fmt.Errorf("tablero no encontrado")
	}
	arrows := a.state.Boards[i].Arrows
	for j := range arrows {
		if arrows[j].ID == arrowID {
			a.state.Boards[i].Arrows = append(arrows[:j], arrows[j+1:]...)
			a.saveStateLocked()
			return a.snapshotLocked(), nil
		}
	}
	return a.snapshotLocked(), nil
}
