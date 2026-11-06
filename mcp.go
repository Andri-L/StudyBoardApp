package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	appVersion       = "1.2.3"
	mcpServerVersion = appVersion
)

type mcpServer struct {
	app *App
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func hasRPCID(id json.RawMessage) bool {
	trimmed := strings.TrimSpace(string(id))
	return trimmed != "" && trimmed != "null"
}

func (s *mcpServer) serve(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	encoder := json.NewEncoder(out)
	encoder.SetEscapeHTML(false)

	write := func(response map[string]any) {
		if err := encoder.Encode(response); err != nil {
			fmt.Fprintf(os.Stderr, "[studyboard-mcp] write error: %v\n", err)
		}
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var req rpcRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			write(map[string]any{
				"jsonrpc": "2.0",
				"id":      nil,
				"error":   rpcError{Code: -32700, Message: "JSON inválido: " + err.Error()},
			})
			continue
		}
		if req.Method == "" {
			continue
		}
		if !hasRPCID(req.ID) {
			// Notifications (for example notifications/initialized) never get a response.
			continue
		}
		result, rpcErr := s.handle(req)
		if rpcErr != nil {
			write(map[string]any{"jsonrpc": "2.0", "id": req.ID, "error": rpcErr})
			continue
		}
		write(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
	}
	return scanner.Err()
}

func (s *mcpServer) handle(req rpcRequest) (any, *rpcError) {
	// Refresh the cheap pointers so a GUI that is open at the same time does
	// not leave the MCP server pinned to an old workspace or model.
	s.app.loadState()
	s.app.engine = findEngine(s.app.selectedModel)

	switch req.Method {
	case "initialize":
		var params struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		_ = json.Unmarshal(req.Params, &params)
		version := "2025-06-18"
		switch params.ProtocolVersion {
		case "2025-11-25", "2025-06-18", "2024-11-05":
			version = params.ProtocolVersion
		}
		return map[string]any{
			"protocolVersion": version,
			"capabilities": map[string]any{
				"tools":     map[string]any{"listChanged": false},
				"resources": map[string]any{"subscribe": false, "listChanged": false},
				"prompts":   map[string]any{"listChanged": false},
			},
			"serverInfo": map[string]any{
				"name":    "studyboard",
				"title":   "StudyBoard App",
				"version": mcpServerVersion,
			},
			"instructions": "StudyBoard App indexa clases locales y sus transcripciones con timestamps. " +
				"Usa overview para ver el workspace, list_videos para localizar una grabación, " +
				"search_transcripts para buscar una idea en todas las clases y get_transcript para leer un rango. " +
				"Cita siempre los momentos como [hh:mm:ss]. Si el usuario pregunta qué se veía en pantalla, " +
				"usa capture_frame con el segundo exacto. Las herramientas son de solo lectura.",
		}, nil

	case "ping":
		return map[string]any{}, nil

	case "tools/list":
		return map[string]any{"tools": mcpToolDefinitions()}, nil

	case "tools/call":
		var params struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, &rpcError{Code: -32602, Message: "Parámetros inválidos: " + err.Error()}
		}
		if params.Arguments == nil {
			params.Arguments = map[string]any{}
		}
		return s.callTool(params.Name, params.Arguments), nil

	case "resources/list":
		return map[string]any{"resources": mcpResourceDefinitions()}, nil

	case "resources/read":
		var params struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, &rpcError{Code: -32602, Message: "Parámetros inválidos"}
		}
		resource, err := s.readResource(params.URI)
		if err != nil {
			return nil, &rpcError{Code: -32602, Message: err.Error()}
		}
		return map[string]any{"contents": []any{resource}}, nil

	case "prompts/list":
		return map[string]any{"prompts": mcpPromptDefinitions()}, nil

	case "prompts/get":
		var params struct {
			Name      string            `json:"name"`
			Arguments map[string]string `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, &rpcError{Code: -32602, Message: "Parámetros inválidos"}
		}
		prompt, err := s.getPrompt(params.Name, params.Arguments)
		if err != nil {
			return nil, &rpcError{Code: -32602, Message: err.Error()}
		}
		return prompt, nil

	case "logging/setLevel":
		return map[string]any{}, nil

	default:
		return nil, &rpcError{Code: -32601, Message: "Método no soportado: " + req.Method}
	}
}

func runMCPServer() error {
	app := NewApp()
	app.initPathsForMCP()
	app.loadState()
	app.engine = findEngine(app.selectedModel)
	if app.selectedModel == "" && app.engine.Model != "" {
		app.selectedModel = filepath.Base(app.engine.Model)
	}
	fmt.Fprintf(os.Stderr, "[studyboard-mcp] workspace=%q model=%q protocol=stdio\n", app.state.Workspace, app.engine.Model)
	server := &mcpServer{app: app}
	return server.serve(os.Stdin, os.Stdout)
}
