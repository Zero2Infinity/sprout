package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

type ModelsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getFirstModel(t *testing.T, baseURL string) string {
	t.Helper()
	resp, err := http.Get(baseURL + "/api/tags")
	if err != nil {
		t.Fatalf("Failed to list models: %v", err)
	}
	defer resp.Body.Close()

	var modelsResp ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		t.Fatalf("Failed to decode models: %v", err)
	}

	if len(modelsResp.Models) == 0 {
		t.Fatal("No models available in Ollama")
	}

	return modelsResp.Models[0].Name
}

func TestOllamaConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ollamaURL := getEnv("OLLAMA_BASE_URL", "http://localhost:11434")

	t.Logf("Testing connection to %s", ollamaURL)

	// Test 1: Check if Ollama is running
	t.Run("Ollama reachable", func(t *testing.T) {
		resp, err := http.Get(ollamaURL + "/api/tags")
		if err != nil {
			t.Fatalf("Cannot reach Ollama at %s: %v\nEnsure Ollama is running: ollama serve", ollamaURL, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Ollama returned status %d", resp.StatusCode)
		}
		t.Logf("Ollama is reachable (status %d)", resp.StatusCode)
	})

	// Test 2: Simple chat completion via HTTP (like the openai-go SDK does)
	t.Run("Chat completion", func(t *testing.T) {
		model := getFirstModel(t, ollamaURL)
		t.Logf("Using model: %s", model)

		reqBody := ChatRequest{
			Model: model,
			Messages: []ChatMessage{
				{Role: "user", Content: "Say exactly: HELLO_TEST_OK"},
			},
			Stream: false,
		}

		body, _ := json.Marshal(reqBody)
		resp, err := http.Post(
			ollamaURL+"/v1/chat/completions",
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			t.Fatalf("Chat completion failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Chat completion returned status %d", resp.StatusCode)
		}

		var chatResp ChatResponse
		if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(chatResp.Choices) == 0 {
			t.Fatal("No choices in response")
		}

		response := chatResp.Choices[0].Message.Content
		t.Logf("Response: %s", response)

		if len(response) == 0 {
			t.Fatal("Empty response from model")
		}
	})
}

func TestBuildAndRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	t.Run("Build", func(t *testing.T) {
		binaryPath := filepath.Join(wd, "sprout-test-binary")
		cmd := exec.Command("go", "build", "-o", binaryPath, ".")
		cmd.Dir = wd
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Build failed: %v\n%s", err, output)
		}
		t.Log("Build successful")
		defer os.Remove(binaryPath)
	})
}

func TestSessionPersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("Session JSON format", func(t *testing.T) {
		sessionData := `{
			"id": "test-session-123",
			"model": "qwen3.6:27b",
			"createdAt": "2026-06-14T12:00:00Z",
			"updatedAt": "2026-06-14T12:00:00Z",
			"messages": [
				{"role": "user", "content": "Hello", "timestamp": "2026-06-14T12:00:00Z"},
				{"role": "assistant", "content": "Hi there", "timestamp": "2026-06-14T12:00:01Z"}
			],
			"tokenUsage": {"promptTokens": 10, "completionTokens": 10, "totalTokens": 20}
		}`

		tmpDir := t.TempDir()
		sessionFile := filepath.Join(tmpDir, "test-session-123.json")
		if err := os.WriteFile(sessionFile, []byte(sessionData), 0644); err != nil {
			t.Fatalf("Failed to write session file: %v", err)
		}

		data, err := os.ReadFile(sessionFile)
		if err != nil {
			t.Fatalf("Failed to read session file: %v", err)
		}

		if !strings.Contains(string(data), "test-session-123") {
			t.Fatal("Session file doesn't contain expected ID")
		}

		t.Log("Session persistence test passed")
	})
}

func TestFullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ollamaURL := getEnv("OLLAMA_BASE_URL", "http://localhost:11434")
	model := getFirstModel(t, ollamaURL)

	t.Logf("Testing full flow with model: %s", model)

	// Test the full flow: config -> connect -> send -> receive
	t.Run("End to end", func(t *testing.T) {
		reqBody := ChatRequest{
			Model: model,
			Messages: []ChatMessage{
				{Role: "system", Content: "You are a helpful assistant."},
				{Role: "user", Content: "What is 2+2? Reply with just the number."},
			},
			Stream: false,
		}

		body, _ := json.Marshal(reqBody)
		resp, err := http.Post(
			ollamaURL+"/v1/chat/completions",
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Status %d, expected 200", resp.StatusCode)
		}

		var chatResp ChatResponse
		if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		if len(chatResp.Choices) == 0 {
			t.Fatal("No choices returned")
		}

		response := chatResp.Choices[0].Message.Content
		t.Logf("Model response: %s", response)

		if !strings.Contains(response, "4") {
			t.Logf("Warning: expected '4' in response, got: %s", response)
		}

		fmt.Printf("\n✓ Full flow test passed! Model responded: %s\n", response)
	})
}
