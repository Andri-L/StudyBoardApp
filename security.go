package main

import (
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// maxWorkflowOutput limits how much text a workflow may add to state.json
	// and to the generated Markdown file.
	maxWorkflowOutput = 1 << 20
	// maxWorkflowError limits the stderr tail kept from an OpenCode run.
	maxWorkflowError = 16 << 10
)

// insideRootResolved checks containment after resolving links and junctions.
// A path that is lexically inside the workspace but resolves outside it (for
// example a junction planted inside a class folder) is rejected.
func insideRootResolved(root, path string) bool {
	if !insideRoot(root, path) {
		return false
	}
	rootResolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		return false
	}
	pathResolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false
	}
	return insideRoot(rootResolved, pathResolved)
}

// writeFileAtomic writes through a temporary file in the destination directory
// and replaces the target only after the content is complete. It also avoids
// following a pre-existing symlink at the destination.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmp := file.Name()
	defer os.Remove(tmp)
	if err := file.Chmod(perm); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err == nil {
		return nil
	}
	// Windows can refuse to replace an existing read-only file. Retry after
	// removing it, still using the fully written temporary file.
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Rename(tmp, path)
}

// secureDownloadClient only allows HTTPS and bounded redirects. Model files
// are large, so there is no overall timeout; the response header timeout keeps
// a stalled server from hanging the download forever.
func secureDownloadClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
			ResponseHeaderTimeout: 30 * time.Second,
			IdleConnTimeout:       30 * time.Second,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("demasiadas redirecciones al descargar")
			}
			if req.URL.Scheme != "https" {
				return fmt.Errorf("redirección no segura al descargar")
			}
			return nil
		},
	}
}

// validateExternalURL accepts only normal web links. Local files, scripts and
// URL userinfo tricks are rejected before the system browser is opened.
func validateExternalURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 4096 {
		return nil, fmt.Errorf("enlace inválido")
	}
	for _, r := range raw {
		if r < 0x20 || r == 0x7f || r == '\\' {
			return nil, fmt.Errorf("enlace inválido")
		}
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("enlace inválido")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("solo se permiten enlaces http o https")
	}
	if parsed.Host == "" || parsed.User != nil {
		return nil, fmt.Errorf("enlace inválido")
	}
	host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return nil, fmt.Errorf("no se permiten enlaces a servicios locales")
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return nil, fmt.Errorf("no se permiten enlaces a servicios locales")
		}
	}
	return parsed, nil
}

// sanitizeUntrustedText removes terminal control sequences and Unicode bidi
// overrides from transcripts and model output. These characters can otherwise
// manipulate a terminal when the same content is read through MCP.
func sanitizeUntrustedText(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))
	for _, r := range value {
		switch {
		case r == '\n' || r == '\t':
			builder.WriteRune(r)
		case r == '\r':
			// Normalize line endings by dropping carriage returns.
		case r < 0x20 || r == 0x7f:
			// C0 controls and DEL.
		case r >= 0x80 && r <= 0x9f:
			// C1 controls.
		case r >= 0x202a && r <= 0x202e:
			// Unicode bidi embeddings and overrides.
		case r >= 0x2066 && r <= 0x2069:
			// Unicode bidi isolates.
		case r == 0x200e || r == 0x200f || r == 0xfeff:
			// Directional marks and BOM.
		default:
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

// isGGMLModel performs a cheap format check before a downloaded file is
// accepted as a Whisper model.
func isGGMLModel(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	var magic [4]byte
	if _, err := io.ReadFull(file, magic[:]); err != nil {
		return false
	}
	return binary.LittleEndian.Uint32(magic[:]) == 0x67676d6c
}

func validModelFile(path string, spec modelSpec) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || info.Size() != spec.Size {
		return false
	}
	return isGGMLModel(path)
}

// normalizeWorkflowLanguage keeps workflow output language values to the two
// supported options.
func normalizeWorkflowLanguage(code string) string {
	if strings.EqualFold(strings.TrimSpace(code), "en") ||
		strings.EqualFold(strings.TrimSpace(code), "english") {
		return "en"
	}
	return "es"
}

// workflowOpencodeConfig builds an inline OpenCode configuration for unattended
// workflow runs. It merges over the user's global config and denies every
// tool except the read-only StudyBoard MCP tools.
func workflowOpencodeConfig(mcpCommand []string) (string, error) {
	permissions := map[string]any{
		"*":                  "deny",
		"read":               "deny",
		"edit":               "deny",
		"bash":               "deny",
		"glob":               "deny",
		"grep":               "deny",
		"task":               "deny",
		"todowrite":          "deny",
		"skill":              "deny",
		"lsp":                "deny",
		"question":           "deny",
		"webfetch":           "deny",
		"websearch":          "deny",
		"external_directory": "deny",
		"doom_loop":          "deny",
	}
	for _, tool := range mcpToolDefinitions() {
		name, _ := tool["name"].(string)
		if name != "" {
			permissions["studyboard_"+name] = "allow"
		}
	}
	config := map[string]any{
		"permission": permissions,
		"mcp": map[string]any{
			"servers": map[string]any{
				"studyboard": map[string]any{
					"type":    "local",
					"command": mcpCommand,
				},
			},
		},
		"autoupdate": false,
		"snapshot":   false,
	}
	data, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// workflowEnv removes inherited OpenCode config overrides so only the config
// generated by StudyBoard applies to the child process.
func workflowEnv(configContent string) []string {
	env := make([]string, 0, len(os.Environ())+1)
	for _, item := range os.Environ() {
		if strings.HasPrefix(item, "OPENCODE_CONFIG=") ||
			strings.HasPrefix(item, "OPENCODE_CONFIG_CONTENT=") {
			continue
		}
		env = append(env, item)
	}
	return append(env, "OPENCODE_CONFIG_CONTENT="+configContent)
}

// limitedWriter keeps only the first part of a stream.
type limitedWriter struct {
	writer    io.Writer
	remaining int
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if w.remaining <= 0 {
		return len(p), nil
	}
	if len(p) > w.remaining {
		p = p[:w.remaining]
	}
	n, err := w.writer.Write(p)
	w.remaining -= n
	return n, err
}
