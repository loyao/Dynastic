package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"dynastic/internal/repository"
)

// Handler 持有依赖并注册 HTTP 路由。
type Handler struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Handler {
	return &Handler{repo: repo}
}

// Router 构建并返回带有全部 API 路由的 http.Handler。
func (h *Handler) Router(cors func(http.Handler) http.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", h.health)
	mux.HandleFunc("GET /api/dynasties", h.listDynasties)
	mux.HandleFunc("GET /api/dynasties/{id}", h.getDynasty)
	mux.HandleFunc("GET /api/dynasties/{id}/tree", h.getDynastyTree)
	mux.HandleFunc("GET /api/emperors", h.listEmperors)
	mux.HandleFunc("GET /api/emperors/{id}", h.getEmperor)

	return cors(mux)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("写入响应失败: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listDynasties(w http.ResponseWriter, r *http.Request) {
	list, err := h.repo.ListDynasties()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "查询朝代失败: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "无效的 id 参数")
		return 0, false
	}
	return id, true
}

func (h *Handler) getDynasty(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	detail, err := h.repo.GetDynastyDetail(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "查询朝代详情失败: "+err.Error())
		return
	}
	if detail == nil {
		writeError(w, http.StatusNotFound, "朝代不存在")
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (h *Handler) getDynastyTree(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	tree, err := h.repo.GetTree(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "构建世系树失败: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tree)
}

func (h *Handler) listEmperors(w http.ResponseWriter, r *http.Request) {
	var dynastyID *int64
	if v := r.URL.Query().Get("dynastyId"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "无效的 dynastyId 参数")
			return
		}
		dynastyID = &id
	}
	list, err := h.repo.ListEmperors(dynastyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "查询帝王失败: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) getEmperor(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	e, err := h.repo.GetEmperor(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "查询帝王失败: "+err.Error())
		return
	}
	if e == nil {
		writeError(w, http.StatusNotFound, "帝王不存在")
		return
	}
	writeJSON(w, http.StatusOK, e)
}
