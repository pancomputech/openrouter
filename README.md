# openrouter

Go client library and CLI for the [OpenRouter](https://openrouter.ai) API.

## Library

```go
import openrouter "github.com/pancomputech/openrouter"

client := openrouter.New(os.Getenv("OPENROUTER_API_KEY"))
resp, err := client.Chat(ctx, openrouter.ChatRequest{
    Model: openrouter.DefaultModel, // "openrouter/auto"
    Messages: []openrouter.Message{
        {Role: "user", Content: "Hello!"},
    },
})
```

## CLI

```sh
go install github.com/pancomputech/openrouter/cmd/openrouter@latest
```

```sh
export OPENROUTER_API_KEY=...

# Print example request
openrouter chat --example

# Send a request from stdin
echo '{"messages":[{"role":"user","content":"Hello!"}]}' | openrouter chat
echo '{"messages":[...]}' | openrouter chat --model anthropic/claude-sonnet-4-5
```

## Development

```sh
make test
make build
```
