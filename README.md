# StudyBoard App

**Visual workspace and local transcription for classes and tutorials.**

[English](README.md) · [Español](README.es.md)

StudyBoard App organizes folders and videos on visual boards, transcribes
recordings locally with Whisper, and lets you query the content from
**OpenCode TUI** through an MCP server.

The app **never moves, renames, or deletes your files**: board cards are
references to real paths.

![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)
![Platform: Windows](https://img.shields.io/badge/Platform-Windows-0078D6)
![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8)
![React](https://img.shields.io/badge/React-19-61DAFB)

## Features

- Visual board with folder and video cards (React Flow).
- Right-click on the board to create **real folders**, **sticky notes**, and
  freeform **arrows**.
- Notes and arrows persist text, color, and coordinates in the board JSON.
- Multiple boards per subject and an integrated folder browser.
- Local transcription with Whisper.cpp and Vulkan GPU acceleration.
- Segment-level and word-level timestamps in the internal JSON.
- Video player synchronized with the transcript; click a timestamp to seek.
- In-transcript search, follow mode, and frame capture.
- Transcription queue with progress, cancellation, and retry.
- Markdown (with timestamps) and SRT export.
- Model manager with bundled download support for `tiny`, `base`, `small`,
  `small-q5`, `medium-q5`, `large-v3-turbo`, and `distil-large-v3`.
- Language selector: Spanish, English, or automatic detection.
- **Automatic OpenCode workflows** for task extraction and exam preparation.
- Workflow reports appear as sticky-note cards next to the video and are saved
  as Markdown beside the recording.
- Workflow picker for running several workflows across several classes at once.
- MCP server for querying classes from OpenCode without opening the app.
- The window no longer stays always on top.

## Stack

- **Go + Wails v2** desktop application (WebView2).
- **React + TypeScript + Vite** frontend.
- **React Flow** for the board.
- **Whisper.cpp + FFmpeg** as local binaries.
- Local HTTP server with a token for Range-capable video playback.
- MCP JSON-RPC server over stdio for OpenCode.

## Requirements

- Windows 10/11.
- Go 1.26.10+ (includes standard-library security fixes).
- Node.js 20+.
- Wails CLI:
  ```powershell
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```
- WebView2 Runtime (included with Windows 10/11).
- An up-to-date GPU driver if GPU acceleration is used.
- OpenCode CLI installed and authenticated (only for workflows):
  ```powershell
  npm install -g @opencode/cli
  opencode
  ```
  StudyBoard reuses your existing OpenCode authentication. It does not ask for
  credentials or store provider tokens.

## Installation

```powershell
git clone https://github.com/Andri-L/StudyBoardApp.git
cd StudyBoardApp

# Downloads Whisper Vulkan, FFmpeg, VAD, and the turbo model (~574 MB).
powershell -ExecutionPolicy Bypass -File .\scripts\setup-engine.ps1

# Builds the desktop app and the MCP server.
powershell -ExecutionPolicy Bypass -File .\scripts\build.ps1
```

The app executable is:

```text
build\bin\StudyBoardApp.exe
```

The MCP server is:

```text
build\bin\StudyBoardMCP.exe
```

### Choosing a different initial model

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\setup-engine.ps1 -Model small-q5
```

Accepted values: `turbo`, `small-q5`, `small`, `base-q5`, `none`. You can also
use `none` and download models later from the model menu inside the app.

## Development

```powershell
$env:WAILS_DEV_ALLOWED_HOSTS = "wails.localhost"
wails dev
```

If `wails` is not on your PATH:

```powershell
& "$env:USERPROFILE\go\bin\wails.exe" dev
```

## OpenCode TUI integration (MCP)

The MCP server reads the workspace and saved transcripts. It works even when the
StudyBoard UI is closed. Register it once:

```powershell
opencode mcp add studyboard --global -- "C:\path\to\StudyBoardApp\build\bin\StudyBoardMCP.exe"
opencode mcp list
```

Expected output:

```text
✓ studyboard  connected
```

Manual configuration in `~/.config/opencode/opencode.jsonc`:

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "servers": {
      "studyboard": {
        "type": "local",
        "command": [
          "C:\\path\\to\\StudyBoardApp\\build\\bin\\StudyBoardMCP.exe"
        ]
      }
    }
  }
}
```

`StudyBoardMCP.exe` automatically enters MCP mode. The desktop executable
(`StudyBoardApp.exe`) also accepts `--mcp`.

### Exposed tools

OpenCode prefixes tools with the server name:

- `studyboard_overview`: workspace, boards, videos, transcripts, and model.
- `studyboard_list_videos`: videos with class, duration, and transcript status.
- `studyboard_search_transcripts`: global accent-insensitive search.
- `studyboard_get_transcript`: full transcript or a timestamp range.
- `studyboard_capture_frame`: image from an exact second in a video.
- `studyboard_list_workflow_results`: generated task or exam reports.
- `studyboard_get_workflow_result`: full report by id, video, or workflow.

It also exposes resources (`studyboard://overview`, `videos`, `transcripts`,
`workflow-results`) and the prompts `/studyboard:resumir-clase`,
`/studyboard:buscar-clases`, `/studyboard:extraer-tareas`, and
`/studyboard:puntos-parciales`.

Examples:

```text
Find where regression was explained in my classes and give me timestamps.
Summarize class 22Jul26 and show the frame with the confusion matrix.
Compare what was said about cross-validation in two classes.
Show me the tasks extracted for class 22Jul26.
```

## Workflows with OpenCode

StudyBoard can run OpenCode automatically over a transcribed class. The app
starts `opencode run` with a new session and reuses the authentication you
already configured in OpenCode; there is no separate login inside StudyBoard.

There are two built-in workflows:

- **Task extraction** (`task-extraction`): explicit and probable assignments,
  requirements, teacher preferences, deadlines, submission channel, team mode,
  rubric hints, evidence with timestamps, and open questions.
- **Exam focus** (`exam-focus`): explicit exam information separated from
  teacher emphasis, excluded topics, question formats, a prioritized study
  plan, practice questions, and warnings. It never invents exam details.

To run one:

1. Open a transcribed class and click **Ejecutar workflow** in the board
   toolbar, or **Workflows** in the transcript viewer.
2. Select one or more workflows and one or more transcribed videos.
3. Choose the output language (Spanish or English).
4. Click **Ejecutar**.

Each result is placed next to the source video as a card and saved as a
Markdown file beside the video (`<video> - tareas.md` or
`<video> - parcial.md`). You can drag the card, copy its content, reveal the
file in Explorer, or remove the card without deleting the Markdown file.

Prompt templates are written in English for model reliability. The final report
is written in the language selected by the user. Workflows require the OpenCode
CLI to be installed and authenticated.

Workflow runs are isolated: OpenCode is started with a generated configuration
that denies shell, file editing, direct reads, web access, subagents, skills,
and LSP tools. Only the read-only StudyBoard MCP tools are allowed. The process
also runs in a dedicated directory, not in the class workspace, so project
configuration and local files cannot influence it. See `SECURITY.md` for
details.

## Models

The quick settings menu in the top toolbar shows the exact active model, for
example `large-v3-turbo-q5_0 · Vulkan`. From there you can:

- View installed models and their sizes.
- Download `small`, `small-q5`, `base-q5`, `medium-q5`, `tiny`, or
  `distil-large-v3`.
- Switch the active model for future transcriptions.
- Choose Spanish, English, or automatic detection.
- Open the models folder.

Downloads are stored in:

```text
%LOCALAPPDATA%\StudyBoard App\models\
```

Models are not committed to this repository. Whisper/FFmpeg binaries are also
installed locally by `scripts/setup-engine.ps1`.

## App data

- Board state: `%APPDATA%\StudyBoard App\state.json` (cards, notes, arrows,
  workflow results).
- Normalized transcripts: `%LOCALAPPDATA%\StudyBoard App\transcripts\`
- Thumbnails: `%LOCALAPPDATA%\StudyBoard App\thumbs\`
- Downloaded models: `%LOCALAPPDATA%\StudyBoard App\models\`
- Job temporary files: `%LOCALAPPDATA%\StudyBoard App\jobs\`

Transcripts are invalidated automatically when the source video changes size or
modification time. Video files are never copied.

## Transcription pipeline

```text
video
  └─ ffmpeg -vn -ac 1 -ar 16000 -c:a pcm_s16le → temporary audio.wav
       └─ whisper-cli -l <es|en|auto> -ojf -osrt --vad -pp
            └─ normalized JSON (segments + words + confidence)
                 ├─ viewer and timestamps
                 ├─ .md / .srt export
                 └─ on-demand frames with ffmpeg
```

The recommended model is `ggml-large-v3-turbo-q5_0.bin` with Silero VAD. The
queue processes one job at a time to avoid GPU contention. Progress is parsed
from Whisper's callback and FFmpeg's `-progress` output.

## Project structure

```text
StudyBoardApp/
├── app.go              # bindings, state, exports
├── board_items.go      # folders, sticky notes, and arrow annotations
├── security.go         # path containment, atomic writes, URL validation
├── workflows.go        # workflow definitions, prompts, results
├── workflow_jobs.go    # OpenCode queue and execution
├── workflows/          # declarative workflows and prompt templates
├── jobs.go             # queue, FFmpeg, Whisper, normalized transcripts
├── media.go            # local video/thumbnail/frame server
├── engine.go           # binary and model discovery
├── models.go           # model catalog and downloads
├── mcp.go              # MCP JSON-RPC server over stdio
├── mcp_data.go         # MCP tools, resources, and prompts
├── store.go            # JSON persistence, dates, helpers
├── types.go            # shared types
├── scripts/
│   ├── setup-engine.ps1
│   └── build.ps1
├── frontend/           # React + React Flow
└── build/              # icon and build output (build/bin ignored)
```

## Measured performance

Reference measurement on Windows 11 with an **AMD Radeon RX 6650 XT (8 GB)** and
Counter-Strike 2 running:

- 42:20 English class: **11:05** transcription time (~3.8× real time).
- `large-v3-turbo-q5_0` + VAD, 1,617 segments.
- Total adapter VRAM: 7.4–7.9 GB out of 8 GB.
- Under that GPU pressure, `small-q5` can fail to initialize Vulkan due to VRAM
  exhaustion. For long transcriptions, avoid competing with a game.
- Performance is significantly better on an idle machine.

Results depend on the video, GPU, model, and open applications.

## Current limitations

- The UI displays segment-level timestamps; word-level timestamps are already
  stored in the JSON but not yet used for navigation.
- No speaker diarization.
- Workflows require the OpenCode CLI to be installed and authenticated.
- Some MKV/AVI files may not play inside WebView2; an external player option is
  provided.
- Global search is available through MCP; the UI does not yet have a
  cross-class search page.
- The MCP server is read-only: it does not start transcriptions or workflows.

## Roadmap

1. Word-level forced alignment (WhisperX/wav2vec2 or ONNX).
2. Speaker diarization (teacher/students).
3. Global search inside the UI.
4. Batch processing of entire folders.
5. More workflows and a translation option for reports.

Technical reference for alignment and diarization:
[jpgomezv/video-transcriber](https://github.com/jpgomezv/video-transcriber).

## Privacy

Audio and transcripts are processed on your machine. StudyBoard does not upload
content on its own. It only connects to download binaries or models when you
request them.

Workflows are different: they invoke OpenCode with your configured provider, so
the transcript content consulted by the model is sent to that provider, exactly
as in any OpenCode session. If you do not want that, do not run workflows.

## Security

The app treats class material, transcripts, and model output as untrusted:

- The local media server binds to loopback, requires a random token, and serves
  only videos inside the workspace.
- Paths are resolved before access to block symlink and junction escapes.
- Workflow reports skip raw HTML, block remote images, and open links only in the
  system browser after validating `http`/`https` and rejecting local hosts.
- Model downloads require HTTPS and are checked for size and GGML format.
- State and report writes are atomic with user-only permissions.
- `npm audit` and `govulncheck` report no known vulnerabilities.

See [SECURITY.md](SECURITY.md) for the full threat model, controls, and residual
risks. The app cannot protect a machine that is already compromised by malware
running as the same Windows user.

## License

GNU General Public License v3.0. See [LICENSE](LICENSE).
