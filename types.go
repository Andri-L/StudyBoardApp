package main

// Viewport is the saved camera position of a board.
type Viewport struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Zoom float64 `json:"zoom"`
}

// Card is a visual reference to a folder or a video. StudyBoard never moves the
// underlying files.
type Card struct {
	ID        string  `json:"id"`
	Kind      string  `json:"kind"` // folder | video
	Path      string  `json:"path"`
	Label     string  `json:"label"`
	DateLabel string  `json:"dateLabel,omitempty"`
	Color     string  `json:"color,omitempty"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Width     float64 `json:"width"`
	Height    float64 `json:"height"`
	AddedAt   int64   `json:"addedAt"`
}

// Board is a canvas with cards and freeform annotations.
type Board struct {
	ID       string           `json:"id"`
	Name     string           `json:"name"`
	Cards    []Card           `json:"cards"`
	Notes    []Note           `json:"notes"`
	Arrows   []Arrow          `json:"arrows"`
	Results  []WorkflowResult `json:"results"`
	Viewport Viewport         `json:"viewport"`
}

// Note is a sticky note stored in the board JSON.
type Note struct {
	ID        string  `json:"id"`
	Text      string  `json:"text"`
	Color     string  `json:"color"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Width     float64 `json:"width"`
	Height    float64 `json:"height"`
	CreatedAt int64   `json:"createdAt"`
}

// Arrow is a freeform arrow annotation stored in flow coordinates.
type Arrow struct {
	ID        string  `json:"id"`
	X1        float64 `json:"x1"`
	Y1        float64 `json:"y1"`
	X2        float64 `json:"x2"`
	Y2        float64 `json:"y2"`
	Type      string  `json:"type,omitempty"`
	Color     string  `json:"color"`
	CreatedAt int64   `json:"createdAt"`
}

// WorkflowInfo is a declarative workflow exposed by the app and MCP.
type WorkflowInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ShortName   string `json:"shortName"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	OutputFile  string `json:"outputFile"`
}

// WorkflowVideo is a transcribed video that can be used in a workflow run.
type WorkflowVideo struct {
	Path          string  `json:"path"`
	Name          string  `json:"name"`
	Class         string  `json:"class"`
	Duration      float64 `json:"duration"`
	HasTranscript bool    `json:"hasTranscript"`
	Segments      int     `json:"segments"`
}

// WorkflowJob is one OpenCode execution of a workflow for a video.
type WorkflowJob struct {
	ID             string `json:"id"`
	BoardID        string `json:"boardId"`
	WorkflowID     string `json:"workflowId"`
	WorkflowName   string `json:"workflowName"`
	VideoPath      string `json:"videoPath"`
	VideoName      string `json:"videoName"`
	OutputLanguage string `json:"outputLanguage"`
	Kind           string `json:"kind,omitempty"`
	SourceResultID string `json:"sourceResultId,omitempty"`
	TargetLanguage string `json:"targetLanguage,omitempty"`
	Status         string `json:"status"`
	Stage          string `json:"stage"`
	Error          string `json:"error"`
	SessionID      string `json:"sessionId"`
	CreatedAt      int64  `json:"createdAt"`
	StartedAt      int64  `json:"startedAt"`
	FinishedAt     int64  `json:"finishedAt"`
}

// WorkflowResult is a rendered workflow output shown next to its video.
type WorkflowResult struct {
	ID             string  `json:"id"`
	WorkflowID     string  `json:"workflowId"`
	WorkflowName   string  `json:"workflowName"`
	VideoPath      string  `json:"videoPath"`
	VideoName      string  `json:"videoName"`
	ClassName      string  `json:"className"`
	Content        string  `json:"content"`
	FilePath       string  `json:"filePath"`
	AttachmentPath string  `json:"attachmentPath,omitempty"`
	Language       string  `json:"language,omitempty"`
	SourceResultID string  `json:"sourceResultId,omitempty"`
	Color          string  `json:"color,omitempty"`
	SessionID      string  `json:"sessionId"`
	X              float64 `json:"x"`
	Y              float64 `json:"y"`
	Width          float64 `json:"width"`
	Height         float64 `json:"height"`
	CreatedAt      int64   `json:"createdAt"`
}

// EngineStatus describes the local transcription toolchain.
type EngineStatus struct {
	Ready       bool        `json:"ready"`
	WhisperPath string      `json:"whisperPath"`
	FFmpegPath  string      `json:"ffmpegPath"`
	FFprobePath string      `json:"ffprobePath"`
	ModelPath   string      `json:"modelPath"`
	ModelName   string      `json:"modelName"`
	VADPath     string      `json:"vadPath"`
	GPU         string      `json:"gpu"`
	Missing     []string    `json:"missing"`
	Models      []ModelInfo `json:"models"`
}

// ModelInfo is a downloadable Whisper model shown in Quick Settings.
type ModelInfo struct {
	File        string  `json:"file"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	URL         string  `json:"url"`
	Size        int64   `json:"size"`
	Downloaded  bool    `json:"downloaded"`
	Active      bool    `json:"active"`
	Path        string  `json:"path"`
	Downloading bool    `json:"downloading"`
	Progress    float64 `json:"progress"`
	Error       string  `json:"error"`
}

// ModelDownload is emitted while a model is being downloaded.
type ModelDownload struct {
	File     string  `json:"file"`
	Status   string  `json:"status"` // downloading | done | error | canceled
	Progress float64 `json:"progress"`
	Bytes    int64   `json:"bytes"`
	Total    int64   `json:"total"`
	Error    string  `json:"error"`
}

// AppState is everything the frontend needs to render the workspace.
type AppState struct {
	Version       int          `json:"version"`
	Workspace     string       `json:"workspace"`
	Boards        []Board      `json:"boards"`
	ActiveBoardID string       `json:"activeBoardId"`
	Language      string       `json:"language"`
	Engine        EngineStatus `json:"engine"`
	MediaBase     string       `json:"mediaBase"`
}

// DirEntry is one row of the folder browser.
type DirEntry struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	IsDir     bool   `json:"isDir"`
	Size      int64  `json:"size"`
	ModTime   int64  `json:"modTime"`
	Kind      string `json:"kind"` // folder | video | other
	DateLabel string `json:"dateLabel,omitempty"`
}

// Word is a token-level timestamp produced by Whisper. It is stored for
// future word-accurate navigation and visual-frame extraction.
type Word struct {
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence,omitempty"`
}

// Segment is a timestamped transcript line.
type Segment struct {
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence,omitempty"`
	Words      []Word  `json:"words,omitempty"`
}

// TranscriptMeta is the lightweight status stored with a media card.
type TranscriptMeta struct {
	Status    string  `json:"status"` // none | queued | extracting | transcribing | saving | done | error | canceled
	Progress  float64 `json:"progress"`
	Stage     string  `json:"stage"`
	Segments  int     `json:"segments"`
	Language  string  `json:"language"`
	CreatedAt int64   `json:"createdAt"`
	Error     string  `json:"error"`
}

// MediaInfo is the detailed metadata for a video card.
type MediaInfo struct {
	Path       string         `json:"path"`
	Name       string         `json:"name"`
	Size       int64          `json:"size"`
	ModTime    int64          `json:"modTime"`
	Duration   float64        `json:"duration"`
	Kind       string         `json:"kind"`
	Transcript TranscriptMeta `json:"transcript"`
}

// Transcript is the normalized, app-owned transcript file.
type Transcript struct {
	Version      int       `json:"version"`
	MediaPath    string    `json:"mediaPath"`
	MediaSize    int64     `json:"mediaSize"`
	MediaModTime int64     `json:"mediaModTime"`
	Duration     float64   `json:"duration"`
	Language     string    `json:"language"`
	Model        string    `json:"model"`
	CreatedAt    int64     `json:"createdAt"`
	Segments     []Segment `json:"segments"`
}

// Job is a transcription task. Only one runs at a time on the GPU.
type Job struct {
	ID         string  `json:"id"`
	MediaPath  string  `json:"mediaPath"`
	CardID     string  `json:"cardId"`
	Status     string  `json:"status"` // queued | extracting | transcribing | saving | done | error | canceled
	Stage      string  `json:"stage"`
	Progress   float64 `json:"progress"`
	Error      string  `json:"error"`
	CreatedAt  int64   `json:"createdAt"`
	StartedAt  int64   `json:"startedAt"`
	FinishedAt int64   `json:"finishedAt"`
}

type persistedState struct {
	Version       int     `json:"version"`
	Workspace     string  `json:"workspace"`
	Boards        []Board `json:"boards"`
	ActiveBoardID string  `json:"activeBoardId"`
	SelectedModel string  `json:"selectedModel,omitempty"`
	Language      string  `json:"language,omitempty"`
}

type whisperToken struct {
	Text    string `json:"text"`
	Offsets struct {
		From int64 `json:"from"`
		To   int64 `json:"to"`
	} `json:"offsets"`
	P float64 `json:"p"`
}

type whisperJSON struct {
	Result struct {
		Language string `json:"language"`
	} `json:"result"`
	Transcription []struct {
		Offsets struct {
			From int64 `json:"from"`
			To   int64 `json:"to"`
		} `json:"offsets"`
		Text   string         `json:"text"`
		Tokens []whisperToken `json:"tokens"`
	} `json:"transcription"`
}
