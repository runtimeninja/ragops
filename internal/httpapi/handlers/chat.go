package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/runtimeninja/ragops/internal/rag"
)

type ChatHandler struct {
	emb   rag.Embedder
	ret   *rag.Retriever
	ans   rag.Answerer
	model string
}

func NewChatHandler(emb rag.Embedder, ret *rag.Retriever, ans rag.Answerer, model string) *ChatHandler {
	return &ChatHandler{emb: emb, ret: ret, ans: ans, model: model}
}

type chatReq struct {
	Question string `json:"question"`
	TopK     int    `json:"top_k"`
}

type chatResp struct {
	Answer  string       `json:"answer"`
	Sources []rag.Source `json:"sources"`
}

func (h *ChatHandler) Chat(w http.ResponseWriter, r *http.Request) {
	var req chatReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Question == "" {
		writeError(w, http.StatusBadRequest, "question is required")
		return
	}

	qvec, _, err := h.emb.Embed(r.Context(), req.Question)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "embed failed")
		return
	}

	src, err := h.ret.TopK(r.Context(), qvec, req.TopK)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "retrieve failed")
		return
	}

	answer, err := h.ans.Answer(r.Context(), h.model, req.Question, src)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "llm failed")
		return
	}

	writeJSON(w, http.StatusOK, chatResp{Answer: answer, Sources: src})
}
