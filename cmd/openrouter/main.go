package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/alecthomas/kong"
	openrouter "github.com/pancomputech/openrouter"
)

var cli struct {
	Chat ChatCmd `cmd:"" help:"Send a chat completion request (JSON from stdin)."`
}

type ChatCmd struct {
	Model   string `cmd:"" short:"m" help:"Model name to override in the request body."`
	Example bool   `cmd:"" short:"e" help:"Print the example request format."`
}

func (c *ChatCmd) Run() error {
	if c.Example {
		printRequestExample()
		os.Exit(0)
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENROUTER_API_KEY is not set")
	}

	var req openrouter.ChatRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		return fmt.Errorf("decode request: %w", err)
	}
	if c.Model != "" {
		req.Model = c.Model
	}

	client := openrouter.New(apiKey)
	resp, err := client.Chat(context.Background(), req)
	if err != nil {
		return fmt.Errorf("chat: %w", err)
	}

	pp(resp)
	return nil
}

func main() {
	ctx := kong.Parse(&cli, kong.UsageOnError())
	if err := ctx.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func pp(x any) {
	bs, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("jq")
	cmd.Stdin = bytes.NewReader(bs)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("ERROR: %v", err)
		fmt.Println(string(bs))
	}
}

func printRequestExample() {
	pp(openrouter.ChatRequest{
		Model: openrouter.DefaultModel,
		Messages: []openrouter.Message{
			{
				Role: "user",
				Content: "Hello, I'm Alice, an AI assistant. " +
					"I'd like to greet a new user named \"Bob\", how should I do that? " +
					"Are there any tools that might help?",
			},
		},
		Tools: []openrouter.Tool{
			{
				Type: "function",
				Function: openrouter.ToolFunction{
					Name:        "greeter",
					Description: "Send a friendly greeting to someone whose name you know.",
					Parameters: json.RawMessage(`{
							"type": "object",
							"properties": {
								"name": {
									"type": "string",
									"items": {"type": "string"},
									"description": "The name of the person to greet."
								}
							},
							"required": ["name"]
						}`),
				},
			},
		},
	})

}
