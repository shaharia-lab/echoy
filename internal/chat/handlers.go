package chat

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shaharia-lab/echoy/internal/chat/types"
	"github.com/shaharia-lab/goai"
	"log"
	"net/http"
)

type ChatHandler struct {
	ChatService Service
}

func NewChatHandler(chatService Service) *ChatHandler {
	return &ChatHandler{
		ChatService: chatService,
	}
}

// HandleChatRequest handles incoming chat requests
func (h *ChatHandler) HandleChatRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("failed to decode request: %v", err), http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		var chatSessionID uuid.UUID
		if req.ChatUUID != uuid.Nil {
			chatSessionID = req.ChatUUID
		}

		chatResponse, err := h.ChatService.Chat(ctx, chatSessionID, req.Question)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get chat response: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(chatResponse); err != nil {
			http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

// HandleChatStreamRequest handles streaming chat requests
func (h *ChatHandler) HandleChatStreamRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("failed to decode request: %v", err), http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		var chatSessionID uuid.UUID
		if req.ChatUUID != uuid.Nil {
			chatSessionID = req.ChatUUID
		}

		// Set proper headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		w.Header().Set("Transfer-Encoding", "chunked")

		if chatSessionID != uuid.Nil {
			w.Header().Set("X-MKit-Chat-UUID", chatSessionID.String())
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		streamChan, err := h.ChatService.ChatStreaming(ctx, chatSessionID, req.Question)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get chat stream: %v", err), http.StatusInternalServerError)
			return
		}

		// Initial response to establish the connection
		fmt.Fprintf(w, "data: %s\n\n", "{\"content\":\"\",\"done\":false}")
		flusher.Flush()

		for streamResp := range streamChan {
			if streamResp.Error != nil {
				// Send error in SSE format
				errMsg := fmt.Sprintf("{\"error\":\"%s\"}", streamResp.Error.Error())
				fmt.Fprintf(w, "data: %s\n\n", errMsg)
				flusher.Flush()
				return
			}

			if err := writeStreamChunk(w, flusher, streamResp); err != nil {
				log.Printf("error writing stream chunk: %v", err)
				return
			}
		}
	}
}

func writeStreamChunk(w http.ResponseWriter, flusher http.Flusher, streamResp goai.StreamingLLMResponse) error {
	response := struct {
		Content string `json:"content"`
		MetaKey string `json:"meta_key,omitempty"`
		Done    bool   `json:"done,omitempty"`
	}{
		Content: streamResp.Text,
		Done:    streamResp.Done,
	}

	chunkData, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal stream chunk: %w", err)
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", chunkData); err != nil {
		return fmt.Errorf("error writing response: %w", err)
	}

	flusher.Flush()
	return nil
}

func (h *ChatHandler) HandleChatHistoryRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		chatHistories, err := h.ChatService.GetListChatHistories(ctx)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get chat history: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(chatHistories); err != nil {
			http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

// HandleChatByIDRequest handles requests to get a chat by its ID
func (h *ChatHandler) HandleChatByIDRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chatUUID := chi.URLParam(r, "chatId")
		if chatUUID == "" {
			http.Error(w, `{"error": "Chat ID is required"}`, http.StatusBadRequest)
			return
		}

		// Parse the provided Chat ID as UUID
		parsedChatUUID, err := uuid.Parse(chatUUID)
		if err != nil {
			http.Error(w, `{"error": "Invalid chat ID"}`, http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		chatHistory, err := h.ChatService.GetChatHistory(ctx, parsedChatUUID)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get chat by ID: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(chatHistory); err != nil {
			http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
			return
		}
	}
}
