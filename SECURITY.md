# Security

## Supported versions

Security fixes are applied to the latest release. The current supported version
is **1.2.x**.

## Threat model

StudyBoard App is a local Windows desktop application. It is designed to reduce
the impact of untrusted class material such as downloaded videos, transcripts,
file names, and model-generated reports.

The app assumes that:

- The Windows user account is trusted.
- OpenCode, Whisper, FFmpeg, and the downloaded models come from trusted
  sources.
- The model provider configured in OpenCode is trusted with the transcript
  content that a workflow consults.

The app **does not** protect against malware that is already running with the
user's permissions, a compromised Windows account, or a compromised OpenCode
installation. In those cases the attacker can already read files, run programs,
and access the same credentials.

## Controls

### Local media server

- Listens only on `127.0.0.1` with a random port.
- Requires a 144-bit random token generated with `crypto/rand`; token
  generation fails closed.
- Serves only video extensions inside the configured workspace.
- Resolves symlinks and junctions before allowing a file, so a link inside a
  class folder cannot expose files outside the workspace.
- Does not send wildcard CORS headers.
- Sends `nosniff`, `no-referrer`, and sandbox CSP headers.
- Uses read-header, read, and idle timeouts plus a header size limit.

### Workspace and file operations

- All user-facing operations check that paths stay inside the workspace.
- Symlinks and junctions are resolved before access.
- Writes to state, transcripts, workflow reports, and model downloads use
  temporary files and atomic replacement instead of following a pre-existing
  destination link.
- State, transcript, and workflow files are written with user-only permissions.
- The app never moves, renames, or deletes source videos.

### OpenCode workflows

Unattended workflows are treated as a hostile-input boundary:

- Each run uses `opencode run --standalone` with a generated inline config.
- Only the read-only StudyBoard MCP tools are allowed. Shell, file editing,
  direct file reads, web fetch/search, subagents, skills, LSP, and todo tools
  are explicitly denied, even with `--auto`.
- The generated inline configuration denies those tools even when the user's
  global config would allow them.
- Completed tool calls are audited. If the model uses anything other than a
  StudyBoard MCP tool, the run is aborted and no report is saved.
- The process runs in a dedicated empty directory, never in the class
  workspace. Project config, `AGENTS.md`, local plugins, and planted DLLs in a
  class folder therefore do not influence the run.
- OpenCode and engine binaries are not discovered from the current directory.
- Transcript text, file names, class names, and MCP output are labeled as
  untrusted data in the prompt. Metadata is quoted so a file name cannot add
  new prompt lines.
- Workflow output is size-limited and control characters, ANSI escape
  sequences, and Unicode bidi overrides are removed before storage and MCP
  exposure.

### Rendered reports

Workflow reports are model output and are rendered as untrusted Markdown:

- Raw HTML is skipped.
- Images are blocked, so a report cannot load remote tracking pixels.
- Links are restricted to `http` and `https`, reject userinfo and local/loopback
  hosts, and open in the system browser instead of navigating the app WebView.
- `file:`, `javascript:`, `data:`, and other schemes are rejected.

### Downloads

- Models are downloaded only over HTTPS from the built-in catalog.
- Redirects are limited and must stay on HTTPS.
- Downloads are written to a random temporary file with user-only permissions.
- The final size must match the catalog, the file must have the GGML magic
  header, and its SHA256 must match the hash published in the catalog before it
  is accepted.
- The same SHA256 checks are applied by `scripts/setup-engine.ps1` for the VAD
  and model files.

### Dependencies

- `npm audit` reports no known vulnerabilities for the frontend.
- `govulncheck` reports no known vulnerabilities in the Go build.
- The project requires Go 1.26.10 or newer so the build includes the latest
  standard-library security fixes.
- GitHub Actions runs tests, vet, `govulncheck`, frontend build, `npm audit` and
  Playwright E2E tests on every push and pull request.

## Residual risks

- Running a workflow sends the transcript content consulted by the model to the
  provider configured in OpenCode. Do not run workflows on material you do not
  want to share with that provider.
- OpenCode global configuration, plugins, providers, and skills are trusted
  user-installed components. Review them if you install third-party plugins.
- FFmpeg and Whisper parse local media files. A malicious media file could
  exploit a vulnerability in those binaries. Keep the engine up to date.
- The Windows binaries are not code-signed. Verify the release source and hashes
  if you distribute builds.
- A local attacker who can modify the user's AppData or install directory has
  the same access as the user and is outside the app's protection boundary.

## Reporting a vulnerability

Open a private security advisory on the repository or contact the maintainer
through GitHub. Please include the affected version, reproduction steps, and
the expected impact. Do not publish exploit details before a fix is available.
