package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"dynastic/internal/model"
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
	mux.HandleFunc("POST /api/dynasties", h.createDynasty)
	mux.HandleFunc("GET /api/dynasties/{id}", h.getDynasty)
	mux.HandleFunc("PUT /api/dynasties/{id}", h.updateDynasty)
	mux.HandleFunc("DELETE /api/dynasties/{id}", h.deleteDynasty)
	mux.HandleFunc("GET /api/dynasties/{id}/tree", h.getDynastyTree)

	mux.HandleFunc("GET /api/emperors", h.listEmperors)
	mux.HandleFunc("POST /api/emperors", h.createEmperor)
	mux.HandleFunc("GET /api/emperors/{id}", h.getEmperor)
	mux.HandleFunc("PUT /api/emperors/{id}", h.updateEmperor)
	mux.HandleFunc("DELETE /api/emperors/{id}", h.deleteEmperor)

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

// decodeJSON 解析请求体到目标结构，失败时直接写入 400 响应。
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "请求体 JSON 解析失败: "+err.Error())
		return false
	}
	return true
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

// ============ 写操作 handler ============

func (h *Handler) createDynasty(w http.ResponseWriter, r *http.Request) {
	var d model.Dynasty
	if !decodeJSON(w, r, &d) {
		return
	}
	d.Name = strings.TrimSpace(d.Name)
	if d.Name == "" {
		writeError(w, http.StatusBadRequest, "朝代名称不能为空")
		return
	}
	id, err := h.repo.CreateDynasty(&d)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "新增朝代失败: "+err.Error())
		return
	}
	d.ID = id
	writeJSON(w, http.StatusCreated, d)
}

func (h *Handler) updateDynasty(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var d model.Dynasty
	if !decodeJSON(w, r, &d) {
		return
	}
	d.ID = id
	d.Name = strings.TrimSpace(d.Name)
	if d.Name == "" {
		writeError(w, http.StatusBadRequest, "朝代名称不能为空")
		return
	}
	if err := h.repo.UpdateDynasty(&d); err != nil {
		writeError(w, http.StatusInternalServerError, "更新朝代失败: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (h *Handler) deleteDynasty(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.repo.DeleteDynasty(id); err != nil {
		writeError(w, http.StatusInternalServerError, "删除朝代失败: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) createEmperor(w http.ResponseWriter, r *http.Request) {
	var e model.Emperor
	if !decodeJSON(w, r, &e) {
		return
	}
	if !validateEmperor(w, &e) {
		return
	}
	id, err := h.repo.CreateEmperor(&e)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "新增帝王失败: "+err.Error())
		return
	}
	e.ID = id
	writeJSON(w, http.StatusCreated, e)
}

func (h *Handler) updateEmperor(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var e model.Emperor
	if !decodeJSON(w, r, &e) {
		return
	}
	e.ID = id
	if !validateEmperor(w, &e) {
		return
	}
	if e.FatherID != nil && *e.FatherID == id {
		writeError(w, http.StatusBadRequest, "父系不能指向自己")
		return
	}
	if err := h.repo.UpdateEmperor(&e); err != nil {
		writeError(w, http.StatusInternalServerError, "更新帝王失败: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (h *Handler) deleteEmperor(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.repo.DeleteEmperor(id); err != nil {
		writeError(w, http.StatusInternalServerError, "删除帝王失败: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// validateEmperor 校验人物写入请求的必要字段。
func validateEmperor(w http.ResponseWriter, e *model.Emperor) bool {
	e.Name = strings.TrimSpace(e.Name)
	if e.DynastyID <= 0 {
		writeError(w, http.StatusBadRequest, "必须指定所属朝代")
		return false
	}
	if e.Name == "" {
		writeError(w, http.StatusBadRequest, "人物姓名不能为空")
		return false
	}
	// 非皇帝人物不参与皇位传承排序，将在位顺序归零。
	if !e.IsEmperor {
		e.OrderIndex = 0
	}
	return true
}
