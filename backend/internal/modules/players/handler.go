package players

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"

	"football-api/pkg/respond"
	"football-api/pkg/uuidutil"
	"football-api/pkg/validation"
)

type Handler struct {
	svc Service
	log *slog.Logger
}

func NewHandler(svc Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("", h.Create)
	rg.GET("", h.List)
	rg.GET("/:id", h.GetByID)
	rg.PUT("/:id", h.Update)
	rg.DELETE("/:id", h.SoftDelete)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreatePlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.BadRequest(c, "invalid request body")
		return
	}
	if err := validation.Validate(&req); err != nil {
		respond.BadRequest(c, err.Error())
		return
	}

	resp, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	respond.Created(c, "player created", resp)
}

func (h *Handler) List(c *gin.Context) {
	var teamID *string
	if raw := c.Query("team_id"); raw != "" {
		if !uuidutil.IsValid(raw) {
			respond.BadRequest(c, "invalid team_id")
			return
		}
		teamID = &raw
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	players, meta, err := h.svc.List(c.Request.Context(), teamID, page, limit)
	if err != nil {
		h.log.Error("players.List", slog.String("error", err.Error()))
		respond.InternalError(c, "internal server error")
		return
	}
	respond.PaginatedSuccess(c, "ok", players, meta)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if !uuidutil.IsValid(id) {
		respond.BadRequest(c, "invalid player id")
		return
	}

	resp, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	respond.Success(c, "ok", resp)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	if !uuidutil.IsValid(id) {
		respond.BadRequest(c, "invalid player id")
		return
	}

	var req UpdatePlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.BadRequest(c, "invalid request body")
		return
	}
	if err := validation.Validate(&req); err != nil {
		respond.BadRequest(c, err.Error())
		return
	}

	resp, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	respond.Success(c, "player updated", resp)
}

func (h *Handler) SoftDelete(c *gin.Context) {
	id := c.Param("id")
	if !uuidutil.IsValid(id) {
		respond.BadRequest(c, "invalid player id")
		return
	}

	if err := h.svc.SoftDelete(c.Request.Context(), id); err != nil {
		h.handleServiceError(c, err)
		return
	}
	respond.Success(c, "player deleted", nil)
}

func (h *Handler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrPlayerNotFound):
		respond.NotFound(c, err.Error())
	case errors.Is(err, ErrJerseyNumberTaken):
		respond.Conflict(c, err.Error())
	case errors.Is(err, ErrTeamNotFound):
		respond.NotFound(c, err.Error())
	default:
		h.log.Error("players handler", slog.String("error", err.Error()))
		respond.InternalError(c, "internal server error")
	}
}
