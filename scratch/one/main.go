package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

var (
	userMessage      = flag.String("m", "what is the weather in rijeka? what is 2 + 2?", "user message")
	requireConfirm   = flag.Bool("confirm", true, "require confirmation before running commands")
	autoApproveReads = flag.Bool("auto-approve-reads", true, "auto-approve read-only commands without confirmation")
	timeout          = flag.Duration("T", 30*time.Second, "timeout for requests")
	dumpTools        = flag.Bool("t", false, "dump tools")
	debugRenderOnly  = flag.Bool("d", false, "debug render only")
)

type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// ChatRequest, cf. https://github.com/ollama/ollama/blob/47e272c35a9d9b5780826a4965f3115908187a7b/openai/openai.go#L98-L117
type ChatRequest struct {
	Model           string    `json:"model"`
	Messages        []Message `json:"messages"`
	Tools           []Tool    `json:"tools,omitempty"`
	Stream          bool      `json:"stream"`
	DebugRenderOnly bool      `json:"_debug_render_only"`
}

type ChatResponse struct {
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

type ToolHandler func(args map[string]any) (string, error)

type ToolRegistry struct {
	definitions []Tool
	handlers    map[string]ToolHandler
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		definitions: []Tool{},
		handlers:    make(map[string]ToolHandler),
	}
}

func (r *ToolRegistry) Register(name, description string, parameters map[string]any, handler ToolHandler) {
	tool := Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        name,
			Description: description,
			Parameters:  parameters,
		},
	}
	r.definitions = append(r.definitions, tool)
	r.handlers[name] = handler
}

func (r *ToolRegistry) GetTools() []Tool {
	return r.definitions
}

func (r *ToolRegistry) Execute(name string, args map[string]any) (string, error) {
	handler, ok := r.handlers[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	return handler(args)
}

// ---- LLM Client ----

type LlmClient struct {
	baseURL string
	client  *http.Client
}

func NewLlmClient(baseURL string) *LlmClient {
	return &LlmClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: *timeout,
		},
	}
}

func (c *LlmClient) Chat(req ChatRequest) (*ChatResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	log.Printf("context length: %v", len(body))
	resp, err := c.client.Post(c.baseURL+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("post request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
	if req.DebugRenderOnly {
		if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
			log.Fatal(err)
		}
		return &ChatResponse{Message: Message{Content: "this is a debug message"}}, nil
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &chatResp, nil
}

func main() {
	flag.Parse()
	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
	}
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "qwen3-vl:latest"
	}
	log.Printf("using %s from %s", model, ollamaHost)
	var (
		client   = NewLlmClient(ollamaHost)
		registry = NewToolRegistry()
	)
	registerTools(registry)
	switch {
	case *dumpTools:
		b, err := json.Marshal(registry.definitions)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
	default:
		log.Printf("user: %s", *userMessage)
		err := runAgentLoop(client, model, registry, *userMessage)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func registerTools(registry *ToolRegistry) {
	registry.Register(
		"get_weather",
		"Get the current weather for a given city",
		map[string]any{
			"type":     "object",
			"required": []string{"city"},
			"properties": map[string]any{
				"city": map[string]any{
					"type":        "string",
					"description": "The city name, e.g. 'Paris' or 'New York'",
				},
			},
		},
		func(args map[string]any) (string, error) {
			city, _ := args["city"].(string)
			return fmt.Sprintf(`{"city": "%s", "temperature": "18°C", "condition": "Partly cloudy", "humidity": "65%%"}`, city), nil
		},
	)

	registry.Register(
		"add_numbers",
		"Add two numbers together and return the result",
		map[string]any{
			"type":     "object",
			"required": []string{"a", "b"},
			"properties": map[string]any{
				"a": map[string]any{
					"type":        "number",
					"description": "The first number",
				},
				"b": map[string]any{
					"type":        "number",
					"description": "The second number",
				},
			},
		},
		func(args map[string]any) (string, error) {
			a, _ := args["a"].(float64)
			b, _ := args["b"].(float64)
			return fmt.Sprintf(`{"result": %v}`, a+b), nil
		},
	)

	registry.Register(
		"get_time",
		"Get the current time in a given timezone",
		map[string]any{
			"type":     "object",
			"required": []string{"timezone"},
			"properties": map[string]any{
				"timezone": map[string]any{
					"type":        "string",
					"description": "The timezone, e.g. 'UTC', 'America/New_York', 'Europe/London'",
				},
			},
		},
		func(args map[string]any) (string, error) {
			tz, _ := args["timezone"].(string)
			return fmt.Sprintf(`{"timezone": "%s", "time": "14:30:00", "date": "2024-01-15"}`, tz), nil
		},
	)

	registry.Register(
		"search_library_catalog",
		"Search for availability of a publication in a library catalog",
		map[string]any{
			"type":     "object",
			"required": []string{"query"},
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "a query string to put into a catalog search, can be an author, title, isbn, issn, or a combination of those",
				},
			},
		},
		func(args map[string]any) (string, error) {
			return "found 4 books", nil
		},
	)

	registry.Register(
		"ping",
		"find out connectivity to a computer on the network with ping",
		map[string]any{
			"type":     "object",
			"required": []string{"hostname_or_ip"},
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "a hostname (e.g. like google.com) or an ip v4 address (like 1.2.4.5)",
				},
			},
		},
		func(args map[string]any) (string, error) {
			return "host is up", nil
		},
	)

	registry.Register(
		"list_files",
		"List files and directories in a given path",
		map[string]any{
			"type":     "object",
			"required": []string{"path"},
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "The directory path to list. Use '.' for current directory",
				},
			},
		},
		func(args map[string]any) (string, error) {
			path, ok := args["path"].(string)
			if !ok {
				return "", fmt.Errorf("path must be a string")
			}

			// Resolve to absolute path for safety
			absPath, err := filepath.Abs(path)
			if err != nil {
				return "", fmt.Errorf("invalid path: %w", err)
			}

			entries, err := os.ReadDir(absPath)
			if err != nil {
				return "", fmt.Errorf("cannot read directory: %w", err)
			}

			type FileInfo struct {
				Name  string `json:"name"`
				IsDir bool   `json:"is_dir"`
				Size  int64  `json:"size,omitempty"`
			}

			var files []FileInfo
			for _, entry := range entries {
				info, err := entry.Info()
				if err != nil {
					continue
				}
				files = append(files, FileInfo{
					Name:  entry.Name(),
					IsDir: entry.IsDir(),
					Size:  info.Size(),
				})
			}

			result, err := json.Marshal(map[string]any{
				"path":  absPath,
				"count": len(files),
				"files": files,
			})
			if err != nil {
				return "", err
			}

			return string(result), nil
		},
	)

	registry.Register(
		"read_file",
		"Read the contents of a text file",
		map[string]any{
			"type":     "object",
			"required": []string{"path"},
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "The file path to read",
				},
				"max_bytes": map[string]any{
					"type":        "number",
					"description": "Optional maximum number of bytes to read (default: 1MB)",
				},
			},
		},
		func(args map[string]any) (string, error) {
			path, ok := args["path"].(string)
			if !ok {
				return "", fmt.Errorf("path must be a string")
			}

			maxBytes := int64(1024 * 1024) // 1MB default
			if mb, ok := args["max_bytes"].(float64); ok {
				maxBytes = int64(mb)
			}

			// Resolve to absolute path
			absPath, err := filepath.Abs(path)
			if err != nil {
				return "", fmt.Errorf("invalid path: %w", err)
			}

			// Check if file exists and is a regular file
			info, err := os.Stat(absPath)
			if err != nil {
				return "", fmt.Errorf("cannot access file: %w", err)
			}
			if info.IsDir() {
				return "", fmt.Errorf("path is a directory, not a file")
			}

			// Open and read file
			file, err := os.Open(absPath)
			if err != nil {
				return "", fmt.Errorf("cannot open file: %w", err)
			}
			defer file.Close()

			// Limit reading to maxBytes
			limitedReader := io.LimitReader(file, maxBytes)
			content, err := io.ReadAll(limitedReader)
			if err != nil {
				return "", fmt.Errorf("cannot read file: %w", err)
			}

			result, err := json.Marshal(map[string]any{
				"path":      absPath,
				"size":      info.Size(),
				"read_size": len(content),
				"content":   string(content),
				"truncated": info.Size() > maxBytes,
			})
			if err != nil {
				return "", err
			}

			return string(result), nil
		},
	)

	registry.Register(
		"grep",
		"Search for a string pattern recursively in files within a directory",
		map[string]any{
			"type":     "object",
			"required": []string{"pattern"},
			"properties": map[string]any{
				"pattern": map[string]any{
					"type":        "string",
					"description": "The string pattern to search for",
				},
				"path": map[string]any{
					"type":        "string",
					"description": "The directory path to search in (default: current directory)",
				},
				"context_lines": map[string]any{
					"type":        "number",
					"description": "Number of context lines before and after match (default: 2)",
				},
				"case_sensitive": map[string]any{
					"type":        "boolean",
					"description": "Whether search should be case sensitive (default: true)",
				},
				"max_results": map[string]any{
					"type":        "number",
					"description": "Maximum number of matches to return (default: 100)",
				},
			},
		},
		func(args map[string]any) (string, error) {
			pattern, ok := args["pattern"].(string)
			if !ok || pattern == "" {
				return "", fmt.Errorf("pattern must be a non-empty string")
			}

			searchPath := "."
			if p, ok := args["path"].(string); ok && p != "" {
				searchPath = p
			}

			contextLines := 2
			if cl, ok := args["context_lines"].(float64); ok {
				contextLines = int(cl)
			}

			caseSensitive := true
			if cs, ok := args["case_sensitive"].(bool); ok {
				caseSensitive = cs
			}

			maxResults := 100
			if mr, ok := args["max_results"].(float64); ok {
				maxResults = int(mr)
			}

			// Prepare pattern for case-insensitive search
			searchPattern := pattern
			if !caseSensitive {
				searchPattern = strings.ToLower(pattern)
			}

			type Match struct {
				File        string   `json:"file"`
				Line        int      `json:"line"`
				MatchedLine string   `json:"matched_line"`
				Before      []string `json:"before,omitempty"`
				After       []string `json:"after,omitempty"`
			}

			var matches []Match
			matchCount := 0

			err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil // Skip files we can't access
				}

				// Skip directories
				if info.IsDir() {
					return nil
				}

				// Skip large files (> 10MB)
				if info.Size() > 10*1024*1024 {
					return nil
				}

				// Try to read file
				content, err := os.ReadFile(path)
				if err != nil {
					return nil // Skip files we can't read
				}

				// Skip binary files (simple heuristic: check for null bytes in first 512 bytes)
				checkLen := 512
				if len(content) < checkLen {
					checkLen = len(content)
				}
				if bytes.Contains(content[:checkLen], []byte{0}) {
					return nil
				}

				// Split content into lines
				lines := strings.Split(string(content), "\n")

				// Search through lines
				for i, line := range lines {
					if matchCount >= maxResults {
						return filepath.SkipAll
					}

					// Check for match
					checkLine := line
					if !caseSensitive {
						checkLine = strings.ToLower(line)
					}

					if strings.Contains(checkLine, searchPattern) {
						match := Match{
							File:        path,
							Line:        i + 1, // 1-indexed
							MatchedLine: line,
						}

						// Get context before
						start := i - contextLines
						if start < 0 {
							start = 0
						}
						for j := start; j < i; j++ {
							match.Before = append(match.Before, lines[j])
						}

						// Get context after
						end := i + contextLines + 1
						if end > len(lines) {
							end = len(lines)
						}
						for j := i + 1; j < end; j++ {
							match.After = append(match.After, lines[j])
						}

						matches = append(matches, match)
						matchCount++
					}
				}

				return nil
			})

			if err != nil {
				return "", fmt.Errorf("search failed: %w", err)
			}

			result, err := json.Marshal(map[string]any{
				"pattern":        pattern,
				"path":           searchPath,
				"case_sensitive": caseSensitive,
				"context_lines":  contextLines,
				"match_count":    len(matches),
				"truncated":      len(matches) >= maxResults,
				"matches":        matches,
			})
			if err != nil {
				return "", err
			}

			return string(result), nil
		},
	)

	registry.Register(
		"write_file",
		"Write content to a file. Can create new files or overwrite existing ones.",
		map[string]any{
			"type":     "object",
			"required": []string{"path", "content"},
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "The file path to write to",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "The content to write to the file",
				},
				"overwrite": map[string]any{
					"type":        "boolean",
					"description": "Whether to overwrite if file exists (default: false)",
				},
				"create_dirs": map[string]any{
					"type":        "boolean",
					"description": "Whether to create parent directories if they don't exist (default: true)",
				},
			},
		},
		func(args map[string]any) (string, error) {
			path, ok := args["path"].(string)
			if !ok || path == "" {
				return "", fmt.Errorf("path must be a non-empty string")
			}

			content, ok := args["content"].(string)
			if !ok {
				return "", fmt.Errorf("content must be a string")
			}

			overwrite := false
			if ow, ok := args["overwrite"].(bool); ok {
				overwrite = ow
			}

			createDirs := true
			if cd, ok := args["create_dirs"].(bool); ok {
				createDirs = cd
			}

			// Resolve to absolute path
			absPath, err := filepath.Abs(path)
			if err != nil {
				return "", fmt.Errorf("invalid path: %w", err)
			}

			// Check if file exists
			fileExists := false
			existingInfo, err := os.Stat(absPath)
			if err == nil {
				fileExists = true
				if existingInfo.IsDir() {
					return "", fmt.Errorf("path is a directory, not a file")
				}
				if !overwrite {
					return "", fmt.Errorf("file already exists (use overwrite=true to replace)")
				}
			} else if !os.IsNotExist(err) {
				return "", fmt.Errorf("cannot access path: %w", err)
			}

			// Create parent directories if needed
			if createDirs {
				dir := filepath.Dir(absPath)
				if err := os.MkdirAll(dir, 0755); err != nil {
					return "", fmt.Errorf("cannot create directories: %w", err)
				}
			}

			// Write the file
			if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
				return "", fmt.Errorf("cannot write file: %w", err)
			}

			// Get final file info
			finalInfo, err := os.Stat(absPath)
			if err != nil {
				return "", fmt.Errorf("file written but cannot stat: %w", err)
			}

			operation := "created"
			if fileExists {
				operation = "overwritten"
			}

			result, err := json.Marshal(map[string]any{
				"path":      absPath,
				"operation": operation,
				"size":      finalInfo.Size(),
				"success":   true,
			})
			if err != nil {
				return "", err
			}

			return string(result), nil
		},
	)

	registry.Register(
		"append_file",
		"Append content to an existing file or create a new file if it doesn't exist. Useful for adding to logs, updating lists, or incrementally building files.",
		map[string]any{
			"type":     "object",
			"required": []string{"path", "content"},
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "The file path to append to",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "The content to append to the file",
				},
				"newline_before": map[string]any{
					"type":        "boolean",
					"description": "Whether to add a newline before the content (default: true)",
				},
				"create_if_missing": map[string]any{
					"type":        "boolean",
					"description": "Whether to create the file if it doesn't exist (default: true)",
				},
				"create_dirs": map[string]any{
					"type":        "boolean",
					"description": "Whether to create parent directories if they don't exist (default: true)",
				},
			},
		},
		func(args map[string]any) (string, error) {
			path, ok := args["path"].(string)
			if !ok || path == "" {
				return "", fmt.Errorf("path must be a non-empty string")
			}

			content, ok := args["content"].(string)
			if !ok {
				return "", fmt.Errorf("content must be a string")
			}

			newlineBefore := true
			if nb, ok := args["newline_before"].(bool); ok {
				newlineBefore = nb
			}

			createIfMissing := true
			if cim, ok := args["create_if_missing"].(bool); ok {
				createIfMissing = cim
			}

			createDirs := true
			if cd, ok := args["create_dirs"].(bool); ok {
				createDirs = cd
			}

			// Resolve to absolute path
			absPath, err := filepath.Abs(path)
			if err != nil {
				return "", fmt.Errorf("invalid path: %w", err)
			}

			// Check if file exists
			fileExists := false
			var existingSize int64
			existingInfo, err := os.Stat(absPath)
			if err == nil {
				fileExists = true
				existingSize = existingInfo.Size()
				if existingInfo.IsDir() {
					return "", fmt.Errorf("path is a directory, not a file")
				}
			} else if !os.IsNotExist(err) {
				return "", fmt.Errorf("cannot access path: %w", err)
			} else if !createIfMissing {
				return "", fmt.Errorf("file does not exist and create_if_missing is false")
			}

			// Create parent directories if needed
			if createDirs {
				dir := filepath.Dir(absPath)
				if err := os.MkdirAll(dir, 0755); err != nil {
					return "", fmt.Errorf("cannot create directories: %w", err)
				}
			}

			// Prepare content to append
			appendContent := content
			if fileExists && newlineBefore && existingSize > 0 {
				// Only add newline if file exists, has content, and newlineBefore is true
				appendContent = "\n" + content
			}

			// Open file for appending (create if doesn't exist)
			file, err := os.OpenFile(absPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return "", fmt.Errorf("cannot open file for appending: %w", err)
			}
			defer file.Close()

			// Write the content
			bytesWritten, err := file.WriteString(appendContent)
			if err != nil {
				return "", fmt.Errorf("cannot write to file: %w", err)
			}

			// Get final file info
			finalInfo, err := os.Stat(absPath)
			if err != nil {
				return "", fmt.Errorf("file written but cannot stat: %w", err)
			}

			operation := "created"
			if fileExists {
				operation = "appended"
			}

			result, err := json.Marshal(map[string]any{
				"path":          absPath,
				"operation":     operation,
				"bytes_written": bytesWritten,
				"size_before":   existingSize,
				"size_after":    finalInfo.Size(),
				"success":       true,
			})
			if err != nil {
				return "", err
			}

			return string(result), nil
		},
	)

	registry.Register(
		"run_command",
		"Execute a shell command and return its output. Use this for running scripts, building projects, testing code, etc.",
		map[string]any{
			"type":     "object",
			"required": []string{"command"},
			"properties": map[string]any{
				"command": map[string]any{
					"type":        "string",
					"description": "The shell command to execute",
				},
				"working_dir": map[string]any{
					"type":        "string",
					"description": "The working directory to run the command in (default: current directory)",
				},
				"timeout_seconds": map[string]any{
					"type":        "number",
					"description": "Timeout in seconds (default: 30)",
				},
			},
		},
		func(args map[string]any) (string, error) {
			command, ok := args["command"].(string)
			if !ok || command == "" {
				return "", fmt.Errorf("command must be a non-empty string")
			}

			workingDir := "."
			if wd, ok := args["working_dir"].(string); ok && wd != "" {
				workingDir = wd
			}

			timeoutSeconds := 30
			if ts, ok := args["timeout_seconds"].(float64); ok {
				timeoutSeconds = int(ts)
			}

			// Check if command needs confirmation
			needsConfirm := *requireConfirm
			if *autoApproveReads && isReadOnlyCommand(command) {
				needsConfirm = false
			}

			if needsConfirm {
				fmt.Printf("\n⚠️  The agent wants to run a command:\n")
				fmt.Printf("   Command: %s\n", command)
				fmt.Printf("   Working Dir: %s\n\n", workingDir)
				fmt.Printf("Allow this command? [y/N]: ")

				reader := bufio.NewReader(os.Stdin)
				response, err := reader.ReadString('\n')
				if err != nil {
					return "", fmt.Errorf("confirmation failed: %w", err)
				}

				response = strings.TrimSpace(strings.ToLower(response))
				if response != "y" && response != "yes" {
					return "", fmt.Errorf("command execution denied by user")
				}
				fmt.Println()
			}

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
			defer cancel()

			// Prepare command
			cmd := exec.CommandContext(ctx, "sh", "-c", command)
			cmd.Dir = workingDir

			// Capture output
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			// Run command
			startTime := time.Now()
			err := cmd.Run()
			duration := time.Since(startTime)

			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else if ctx.Err() == context.DeadlineExceeded {
					result, _ := json.Marshal(map[string]any{
						"command":     command,
						"stdout":      stdout.String(),
						"stderr":      stderr.String(),
						"exit_code":   -1,
						"error":       "command timed out",
						"duration_ms": duration.Milliseconds(),
					})
					return string(result), nil
				} else {
					return "", fmt.Errorf("failed to execute command: %w", err)
				}
			}

			result, err := json.Marshal(map[string]any{
				"command":     command,
				"working_dir": workingDir,
				"stdout":      stdout.String(),
				"stderr":      stderr.String(),
				"exit_code":   exitCode,
				"duration_ms": duration.Milliseconds(),
				"success":     exitCode == 0,
			})
			if err != nil {
				return "", err
			}

			return string(result), nil
		},
	)

}

// isReadOnlyCommand checks if a command is likely read-only (safe to auto-approve)
func isReadOnlyCommand(command string) bool {
	command = strings.TrimSpace(strings.ToLower(command))

	// List of read-only command prefixes
	safeCommands := []string{
		"ls", "cat", "head", "tail", "grep", "find", "pwd",
		"echo", "printf", "wc", "sort", "uniq", "diff",
		"which", "whereis", "type", "file", "stat",
		"ps", "top", "df", "du", "free", "uname",
		"date", "cal", "env", "printenv",
		"git log", "git status", "git diff", "git show",
		"docker ps", "docker images", "docker inspect",
		"kubectl get", "kubectl describe",
	}

	for _, safe := range safeCommands {
		if strings.HasPrefix(command, safe+" ") || command == safe {
			return true
		}
	}

	// Check for common read-only patterns
	readOnlyPatterns := []string{
		"--help", "--version", "-h", "-v",
	}

	for _, pattern := range readOnlyPatterns {
		if strings.Contains(command, pattern) {
			return true
		}
	}

	return false
}

func runAgentLoop(client *LlmClient, model string, registry *ToolRegistry, userMessage string) error {
	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant with access to tools. Use them when needed.",
		},
		{
			Role:    "user",
			Content: userMessage,
		},
	}
	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		req := ChatRequest{
			Model:           model,
			Messages:        messages,
			Tools:           registry.GetTools(),
			Stream:          false,
			DebugRenderOnly: *debugRenderOnly,
		}
		resp, err := client.Chat(req)
		if err != nil {
			return fmt.Errorf("chat error: %w", err)
		}
		if len(resp.Message.ToolCalls) > 0 {
			log.Printf("assistant wants to call %d tool(s)", len(resp.Message.ToolCalls))
			messages = append(messages, resp.Message)
			for _, tc := range resp.Message.ToolCalls {
				log.Printf("calling tool: %s", tc.Function.Name)
				argsJSON, _ := json.Marshal(tc.Function.Arguments)
				log.Printf("args: %s", string(argsJSON))
				result, err := registry.Execute(tc.Function.Name, tc.Function.Arguments)
				if err != nil {
					result = fmt.Sprintf(`{"error": "%s"}`, err.Error())
				}
				log.Printf("    Result: %s", result)
				messages = append(messages, Message{
					Role:    "tool",
					Content: result,
				})
			}
		} else {
			log.Printf("assistant: %s", resp.Message.Content)
			return nil
		}
	}
	return fmt.Errorf("max iterations reached")
}
