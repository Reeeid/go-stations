package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/TechBowl-japan/go-stations/model"
	"github.com/TechBowl-japan/go-stations/service"
)

// A TODOHandler implements handling REST endpoints.
type TODOHandler struct {
	svc *service.TODOService
}

// NewTODOHandler returns TODOHandler based http.Handler.
func NewTODOHandler(svc *service.TODOService) *TODOHandler {
	return &TODOHandler{
		svc: svc,
	}
}

func (h *TODOHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		req := model.DeleteTODORequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, "invalid request Body", http.StatusBadRequest)
			return
		}

		if len(req.IDs) == 0 {
			http.Error(w, "ID is required", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		// TODO を削除する処理を呼び出す
		ids := make([]int64, len(req.IDs))
		for i, id := range req.IDs {
			ids[i] = int64(id)
		}

		err := h.svc.DeleteTODO(ctx, ids)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to delete TODO"})
			return
		}
		w.WriteHeader(http.StatusOK)
		resp := model.DeleteTODOResponse{}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	case http.MethodGet:
		var req model.ReadTODORequest
		// クエリパラメータから prev_id と size を取得
		prevID := r.URL.Query().Get("prev_id")
		size := r.URL.Query().Get("size")

		// prev_id の変換
		prevIDint, err := strconv.ParseInt(prevID, 10, 64)
		if err != nil && prevID != "" { // prev_id が空でない場合のみエラーを返す
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid prev_id parameter"})
			return
		}

		// size の変換
		sizeint, err := strconv.ParseInt(size, 10, 64)
		if err != nil && size != "" { // size が空でない場合のみエラーを返す
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid size parameter"})
			return
		}
		if sizeint == 0 {
			sizeint = 3 // デフォルト値を設定
		}

		// ReadTODORequest に値を代入
		req = model.ReadTODORequest{
			PrevID: int(prevIDint),
			Size:   int(sizeint),
		}

		// ReadTODO メソッドを呼び出し
		ctx := r.Context()
		todos, err := h.svc.ReadTODO(ctx, int64(req.PrevID), int64(req.Size))
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			if err == service.ErrInvalidInput {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid input"})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to read TODO"})
			return
		}

		// TODOs をレスポンスに変換
		convertedTODOs := make([]model.TODO, len(todos))
		for i, todo := range todos {
			convertedTODOs[i] = *todo
		}

		// ReadTODOResponse を作成
		resp := model.ReadTODOResponse{
			TODOs: convertedTODOs,
		}

		// JSON Encode を行い HTTP Response を返す
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to encode response"})
		}
	case http.MethodPut:
		var req model.UpdateTODORequest
		// リクエストボディをデコード
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request Body", http.StatusBadRequest)
			return
		}
		if req.ID == 0 {
			http.Error(w, "ID is required", http.StatusBadRequest)
			return
		}
		if req.Subject == "" {
			http.Error(w, "Subject is required", http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		// TODO を更新する処理を呼び出す
		todo, err := h.svc.UpdateTODO(ctx, int64(req.ID), req.Subject, req.Description)
		if err != nil {
			if err == service.ErrInvalidInput {
				http.Error(w, "invalid input", http.StatusBadRequest)
				return
			}
			http.Error(w, "failed to update TODO", http.StatusInternalServerError)
			return

		}
		resp := model.UpdateTODOResponse{
			TODO: *todo,
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	case http.MethodPost:
		var req model.CreateTODORequest
		// リクエストボディをデコード
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request Body", http.StatusBadRequest)
			return
		}

		// Subject が空の場合は 400 を返す
		if req.Subject == "" {
			http.Error(w, "Subject is required", http.StatusBadRequest)
			return
		}

		// Context を取得して Service に渡す
		ctx := r.Context()
		todo, err := h.svc.CreateTODO(ctx, req.Subject, req.Description)
		if err != nil {
			if err == service.ErrInvalidInput {
				http.Error(w, "invalid input", http.StatusBadRequest)
				return
			}
			http.Error(w, "failed to create TODO", http.StatusInternalServerError)
			return
		}

		// 正常系レスポンスを作成
		resp := model.CreateTODOResponse{
			TODO: *todo,
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Create handles the endpoint that creates the TODO.
func (h *TODOHandler) Create(ctx context.Context, req *model.CreateTODORequest) (*model.CreateTODOResponse, error) {
	todo, err := h.svc.CreateTODO(ctx, req.Subject, req.Description)
	if err != nil {
		return nil, err
	}

	return &model.CreateTODOResponse{
		TODO: *todo,
	}, nil
}

// Read handles the endpoint that reads the TODOs.
func (h *TODOHandler) Read(ctx context.Context, req *model.ReadTODORequest) (*model.ReadTODOResponse, error) {
	_, _ = h.svc.ReadTODO(ctx, 0, 0)
	return &model.ReadTODOResponse{}, nil
}

// Update handles the endpoint that updates the TODO.
func (h *TODOHandler) Update(ctx context.Context, req *model.UpdateTODORequest) (*model.UpdateTODOResponse, error) {
	_, _ = h.svc.UpdateTODO(ctx, 0, "", "")
	return &model.UpdateTODOResponse{}, nil

}

// Delete handles the endpoint that deletes the TODOs.
func (h *TODOHandler) Delete(ctx context.Context, req *model.DeleteTODORequest) (*model.DeleteTODOResponse, error) {
	_ = h.svc.DeleteTODO(ctx, nil)
	return &model.DeleteTODOResponse{}, nil
}
