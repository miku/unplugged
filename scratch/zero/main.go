package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

var (
	userMessage = flag.String("m", "what is the weather in rijeka? what is 2 + 2?", "user message")
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

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools,omitempty"`
	Stream   bool      `json:"stream"`
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
			Timeout: 30 * time.Second,
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
		model = "qwen3:latest"
	}
	log.Printf("using %s from %s", model, ollamaHost)
	var (
		client   = NewLlmClient(ollamaHost)
		registry = NewToolRegistry()
	)
	registerTools(registry)
	log.Printf("user: %s", *userMessage)
	err := runAgentLoop(client, model, registry, *userMessage)
	if err != nil {
		log.Fatal(err)
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
			Model:    model,
			Messages: messages,
			Tools:    registry.GetTools(),
			Stream:   false,
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
