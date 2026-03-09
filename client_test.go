package openrouter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(srv *httptest.Server) *Client {
	c := New("test-key")
	// Override the default HTTP client to route requests to the test server.
	http.DefaultClient = &http.Client{
		Transport: rewriteHostTransport{base: srv.URL, inner: http.DefaultTransport},
	}
	return c
}

// rewriteHostTransport rewrites requests to point at a test server.
type rewriteHostTransport struct {
	base  string
	inner http.RoundTripper
}

func (t rewriteHostTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.URL.Host = strings.TrimPrefix(t.base, "http://")
	req.URL.Scheme = "http"
	return t.inner.RoundTrip(req)
}

func TestChat_Success(t *testing.T) {
	want := ChatResponse{
		Choices: []Choice{{
			Message:      Message{Role: "assistant", Content: "Hello!"},
			FinishReason: "stop",
		}},
		Usage: Usage{PromptTokens: 10, CompletionTokens: 5, Cost: 0.001},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	got, err := c.Chat(context.Background(), ChatRequest{
		Model:    "openai/gpt-4o",
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Hello!", got.Choices[0].Message.Content)
	assert.Equal(t, "stop", got.Choices[0].FinishReason)
	assert.Equal(t, 10, got.Usage.PromptTokens)
}

func TestChat_NonTwoXX(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("rate limited"))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Chat(context.Background(), ChatRequest{
		Model:    "openai/gpt-4o",
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "429")
	assert.Contains(t, err.Error(), "rate limited")
}

func TestChat_ToolCallRoundTrip(t *testing.T) {
	toolCallResp := ChatResponse{
		Choices: []Choice{{
			Message: Message{
				Role:    "assistant",
				Content: "",
				ToolCalls: []ToolCall{{
					ID:   "call_abc",
					Type: "function",
					Function: ToolCallFunction{
						Name:      "do_something",
						Arguments: `{"key":"value"}`,
					},
				}},
			},
			FinishReason: "tool_calls",
		}},
		Usage: Usage{PromptTokens: 20, CompletionTokens: 15},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(toolCallResp)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	got, err := c.Chat(context.Background(), ChatRequest{
		Model:    "openai/gpt-4o",
		Messages: []Message{{Role: "user", Content: "Do something"}},
		Tools: []Tool{{
			Type: "function",
			Function: ToolFunction{
				Name:        "do_something",
				Description: "Does something",
				Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
			},
		}},
	})
	require.NoError(t, err)
	require.Len(t, got.Choices, 1)
	assert.Equal(t, "tool_calls", got.Choices[0].FinishReason)
	require.Len(t, got.Choices[0].Message.ToolCalls, 1)
	tc := got.Choices[0].Message.ToolCalls[0]
	assert.Equal(t, "call_abc", tc.ID)
	assert.Equal(t, "do_something", tc.Function.Name)
	assert.Equal(t, `{"key":"value"}`, tc.Function.Arguments)
}

func TestSchemaFor(t *testing.T) {
	type SampleInput struct {
		Name  string `json:"name" jsonschema:"description=The name"`
		Count int    `json:"count"`
	}

	schema := SchemaFor[SampleInput]()
	require.NotNil(t, schema)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(schema, &parsed))

	// Schema should be valid JSON containing field information
	raw, _ := json.Marshal(parsed)
	assert.Contains(t, string(raw), "name")
	assert.Contains(t, string(raw), "count")
}
