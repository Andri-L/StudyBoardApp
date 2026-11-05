package main

import (
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
func (a *App) AddArrow(boardID string, x1, y1, x2, y2 float64) (AppState, error) {
	if math.Hypot(x2-x1, y2-y1) < 8 {
		return AppState{}, fmt.Errorf("la flecha es demasiado corta")
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
		Color:     "#0067c0",
		CreatedAt: time.Now().UnixMilli(),
	})
	a.saveStateLocked()
	return a.snapshotLocked(), nil
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
