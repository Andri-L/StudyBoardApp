# Changelog

## v1.1.0 — 2026-11-05

Automatic OpenCode workflows in the workspace.

### Added

- **Run workflow** action in the board toolbar.
- Workflow picker to run one or several workflows on one or several classes at
  once, with Spanish or English output.
- **Task extraction** workflow: explicit and probable assignments, teacher
  requirements and preferences, deadlines, submission channel, team mode,
  rubric hints, evidence with timestamps, and ambiguities.
- **Exam focus** workflow: explicit exam information separated from teacher
  emphasis, excluded topics, question formats, a prioritized study plan,
  practice questions, and warnings. Never invents exam details.
- Workflow results as sticky-note cards placed next to the source video.
- Every result is also saved as a Markdown file beside the video
  (`<video> - tareas.md`, `<video> - parcial.md`).
- Workflow jobs panel with live stage, cancellation, and error details.
- Per-video **Workflows** button in the transcript viewer.
- MCP tools `list_workflow_results` and `get_workflow_result`.
- MCP resource `studyboard://workflow-results`.
- MCP prompts `extraer-tareas` and `puntos-parciales`.
- Workflow definitions as declarative JSON plus English prompt templates.

### Changed

- `overview` now reports workflow results, notes, and arrows per board.

### Notes

- Workflows use the OpenCode CLI already installed and authenticated on the
  machine. Each run starts a new OpenCode session; StudyBoard does not store
  provider credentials.
- Workflow prompts are written in English for model reliability. The report
  language is selected by the user.
- The transcript is read through the StudyBoard MCP server. When a workflow
  runs, the content consulted by the model is sent to the provider configured
  in OpenCode, just like any other OpenCode session.

## v1.0.0 — 2026-11-05

First public release of StudyBoard App.

### Added

- Visual board with folder and video cards using React Flow.
- Right-click board actions to create real folders, editable sticky notes, and
  freeform arrows.
- Cards, notes, arrows, colors, and coordinates persisted in JSON.
- Multiple boards per subject, file browser, and drag & drop.
- Local transcription with Whisper.cpp and Vulkan GPU acceleration.
- Segment-level and word-level timestamps in the internal JSON.
- Video player synchronized with the transcript.
- In-transcript search and follow mode.
- Frame capture to review what was on screen.
- Transcription queue with progress, cancellation, and retry.
- Markdown export with timestamps and SRT export.
- Copy full transcript to the clipboard.
- Model manager with integrated downloads (`tiny`, `base`, `small`, `small-q5`,
  `medium-q5`, `large-v3-turbo`, `distil-large-v3`).
- Language selector: Spanish, English, or automatic detection.
- MCP server over stdio for querying classes from OpenCode TUI.
- MCP tools: overview, list_videos, search_transcripts, get_transcript, and
  capture_frame.
- Engine setup and build scripts.
- Window no longer stays always on top.

### Notes

- Models and engine binaries are not included in the repository; install them
  with `scripts/setup-engine.ps1`.
- The MCP server is read-only in this release.
- Speaker diarization and automatic LLM summaries are planned for future
  versions.
