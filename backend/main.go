package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const OPENAI_API_KEY = "openapi-key"
const OPENAI_API_URL = "https://api.openai.com/v1/chat/completions"

// SearchIntent represents the parsed understanding of a search query
type SearchIntent struct {
	MainQuery    string   `json:"main_query"`
	ExactPhrases []string `json:"exact_phrases,omitempty"`
	SiteFilter   string   `json:"site_filter,omitempty"`
	FileType     string   `json:"file_type,omitempty"`
	ExcludeWords []string `json:"exclude_words,omitempty"`
	DateRange    string   `json:"date_range,omitempty"`
}

// OpenAIMessage represents a message in the OpenAI chat format
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIRequest represents the request structure for OpenAI API
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
}

// OpenAIResponse represents the response structure from OpenAI API
type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// SearchHandler processes search requests
type SearchHandler struct {
	openAIKey string
	client    *http.Client
}

func NewSearchHandler(openAIKey string) *SearchHandler {
	return &SearchHandler{
		openAIKey: openAIKey,
		client:    &http.Client{},
	}
}

// analyzePromptWithOpenAI sends the search prompt to OpenAI for understanding
func (h *SearchHandler) analyzePromptWithOpenAI(ctx context.Context, prompt string) (*SearchIntent, error) {
	messages := []OpenAIMessage{
		{
			Role: "system",
			Content: `You are a search query analyzer. Extract search parameters and return ONLY a JSON object like this:
{
    "main_query": "the main search terms",
    "exact_phrases": ["exact phrase 1", "exact phrase 2"],
    "site_filter": "example.com",
    "file_type": "pdf",
    "exclude_words": ["exclude1", "exclude2"],
    "date_range": "timeframe"
}
Always include all fields, use empty arrays [] for empty lists, and empty strings "" for empty fields.`,
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	reqBody := OpenAIRequest{
		Model:       "gpt-3.5-turbo",
		Messages:    messages,
		Temperature: 0.3, // Lower temperature for more consistent output
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling OpenAI request: %v", err)
	}

	// Log the request for debugging
	log.Printf("Sending request to OpenAI: %s", string(jsonBody))

	req, err := http.NewRequestWithContext(ctx, "POST", OPENAI_API_URL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating OpenAI request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.openAIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling OpenAI: %v", err)
	}
	defer resp.Body.Close()

	// Read the full response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Log the response for debugging
	log.Printf("OpenAI response: %s", string(body))

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("error parsing OpenAI response: %v", err)
	}

	if openAIResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices from OpenAI")
	}

	// Parse the JSON response from OpenAI into SearchIntent
	var intent SearchIntent
	content := strings.TrimSpace(openAIResp.Choices[0].Message.Content)
	if err := json.Unmarshal([]byte(content), &intent); err != nil {
		return nil, fmt.Errorf("error parsing intent JSON: %v\nContent: %s", err, content)
	}

	// Initialize empty slices if they're nil
	if intent.ExactPhrases == nil {
		intent.ExactPhrases = []string{}
	}
	if intent.ExcludeWords == nil {
		intent.ExcludeWords = []string{}
	}

	return &intent, nil
}

func constructSearchQuery(intent *SearchIntent) string {
	var queryParts []string

	if intent.MainQuery != "" {
		queryParts = append(queryParts, intent.MainQuery)
	}

	for _, phrase := range intent.ExactPhrases {
		if phrase != "" {
			queryParts = append(queryParts, fmt.Sprintf(`"%s"`, phrase))
		}
	}

	if intent.SiteFilter != "" {
		queryParts = append(queryParts, fmt.Sprintf("site:%s", intent.SiteFilter))
	}

	if intent.FileType != "" {
		queryParts = append(queryParts, fmt.Sprintf("filetype:%s", intent.FileType))
	}

	for _, word := range intent.ExcludeWords {
		if word != "" {
			queryParts = append(queryParts, fmt.Sprintf("-%s", word))
		}
	}

	if intent.DateRange != "" {
		queryParts = append(queryParts, fmt.Sprintf("after:%s", intent.DateRange))
	}

	baseURL := "https://www.google.com/search"
	params := url.Values{}
	params.Add("q", strings.Join(queryParts, " "))

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

func (h *SearchHandler) handleSearch(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Log the incoming request
	log.Printf("Received request body: %s", string(body))

	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	intent, err := h.analyzePromptWithOpenAI(r.Context(), req.Prompt)
	if err != nil {
		log.Printf("Error analyzing prompt: %v", err)
		http.Error(w, fmt.Sprintf("Error analyzing prompt: %v", err), http.StatusInternalServerError)
		return
	}

	searchURL := constructSearchQuery(intent)
	response := map[string]interface{}{
		"search_url": searchURL,
		"intent":     intent,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func main() {
	// Get OpenAI API key from environment variable

	handler := NewSearchHandler(OPENAI_API_KEY)
	http.HandleFunc("/search", handler.handleSearch)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
