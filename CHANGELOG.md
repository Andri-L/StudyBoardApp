# Changelog

## v1.1.1 — 2026-11-05

Security hardening release.

### Security

- Workflows now run OpenCode in an isolated standalone session with a generated
  inline configuration. Shell, file editing, direct file reads, web access,
  subagents, skills, and LSP tools are denied; only the read-only StudyBoard MCP
  tools are allowed, even with `--auto`.
- Workflows run in a dedicated empty directory instead of the class workspace,
  preventing project config, `AGENTS.md`, local plugins, and planted DLLs from
  influencing the unattended agent.
- Transcripts and file names are labeled as untrusted data; prompt metadata is
  quoted so a file name cannot inject new prompt lines.
- MCP absolute paths are restricted to the workspace and symlinks/junctions are
  resolved before access.
- The local media server now fails closed if it cannot generate its random
  token, serves only workspace videos, rejects paths that resolve outside the
  workspace, and no longer sends wildcard CORS headers.
- Workflow reports block remote images and raw HTML. Links are limited to
  `http`/`https`, reject userinfo and local/loopback hosts, and open in the
  system browser instead of navigating the app.
- Workflow output is size-limited and control characters, ANSI escapes, and
  Unicode bidi overrides are removed before storage and MCP exposure.
- Model downloads require HTTPS, limit redirects, verify the exact catalog size
  and GGML magic header, and use safe temporary files.
- State, transcript, and workflow files are written atomically with user-only
  permissions.
- Engine and OpenCode binaries are no longer discovered from the current
  working directory.
- The build now requires Go 1.26.10, which includes fixes for known
  standard-library vulnerabilities.

### Notes

- `npm audit` and `govulncheck` report no known vulnerabilities.
- See `SECURITY.md` for the threat model and residual risks.

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
