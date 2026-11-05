# Changelog

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
