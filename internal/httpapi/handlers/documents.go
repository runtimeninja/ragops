package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/runtimeninja/ragops/internal/ingest"
	"github.com/runtimeninja/ragops/internal/jobs"
)

type DocumentsHandler struct {
	pool *pgxpool.Pool
	ing  *ingest.Service
	jq   *jobs.Client
}

func NewDocumentsHandler(pool *pgxpool.Pool, ing *ingest.Service, jq *jobs.Client) *DocumentsHandler {
	return &DocumentsHandler{pool: pool, ing: ing, jq: jq}
}

type createReq struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type createResp struct {
	DocumentID string `json:"document_id"`
	Deduped    bool   `json:"deduped"`
	Enqueued   bool   `json:"enqueued"`
}

func (h *DocumentsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Text == "" {
		writeError(w, http.StatusBadRequest, "text is required")
		return
	}

	docID, deduped, err := h.ing.CreateTextDocument(r.Context(), req.Title, req.Text)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create failed")
		return
	}

	_, err = h.jq.EnqueueIngest(docID)
	writeJSON(w, http.StatusAccepted, createResp{
		DocumentID: docID,
		Deduped:    deduped,
		Enqueued:   err == nil,
	})
}

type statusResp struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func (h *DocumentsHandler) Status(w http.ResponseWriter, r *http.Request, id string) {
	var status string
	var errMsg string
	err := h.pool.QueryRow(r.Context(),
		`SELECT status, COALESCE(error_message,'') FROM documents WHERE id=$1`, id).
		Scan(&status, &errMsg)
	if err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	resp := statusResp{ID: id, Status: status}
	if errMsg != "" {
		resp.Error = errMsg
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, struct {
		Error string `json:"error"`
	}{Error: msg})
}
