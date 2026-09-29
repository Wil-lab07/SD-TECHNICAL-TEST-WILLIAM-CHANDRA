package matches

import (
	"errors"
	"log/slog"
	"net/http"
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
	rg.POST("/:id/result", h.ReportResult)
	rg.PUT("/:id/result", h.UpdateResult)
	rg.GET("/:id/report", h.GetMatchReport)
	rg.DELETE("/:id", h.SoftDelete)
}

func (h *Handler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrMatchNotFound):
		respond.NotFound(c, "Match not found")
	case errors.Is(err, ErrSelfPlayNotAllowed):
		respond.BadRequest(c, "A team cannot play against itself")
	case errors.Is(err, ErrTeamDeactivated):
		respond.BadRequest(c, "One or both teams are deactivated")
	case errors.Is(err, ErrMatchAlreadyFinished):
		respond.Conflict(c, "Match result already reported, use PUT to update")
	case errors.Is(err, ErrScoreMismatch):
		respond.CustomError(c, http.StatusUnprocessableEntity, "SCORE_MISMATCH", "Reported score does not match the number of goals recorded")
	case errors.Is(err, ErrPlayerNotOnTeam):
		respond.BadRequest(c, "Goal scorer does not belong to specified team in this match")
	case errors.Is(err, ErrPlayerDeactivated):
		respond.BadRequest(c, "Goal scorer is deactivated")
	default:
		h.log.Error("matches module error", "error", err, "path", c.Request.URL.Path)
		respond.InternalError(c, "An unexpected error occurred")
	}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.BadRequest(c, "Invalid JSON payload")
		return
	}

	if err := validation.Validate(&req); err != nil {
		respond.CustomError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	res, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	respond.Created(c, "Match scheduled successfully", res)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if !uuidutil.IsValid(id) {
		respond.BadRequest(c, "Invalid match ID format")
		return
	}

	res, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	respond.Success(c, "Match fetched successfully", res)
}

func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	res, meta, err := h.svc.List(c.Request.Context(), page, limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	respond.PaginatedSuccess(c, "Matches fetched successfully", res, meta)
}

func (h *Handler) ReportResult(c *gin.Context) {
	id := c.Param("id")
	if !uuidutil.IsValid(id) {
		respond.BadRequest(c, "Invalid match ID format")
		return
	}

	var req ReportResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.BadRequest(c, "Invalid JSON payload")
		return
	}

	if err := validation.Validate(&req); err != nil {
		respond.CustomError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	res, err := h.svc.ReportResult(c.Request.Context(), id, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	respond.Success(c, "Match result reported successfully", res)
}

func (h *Handler) UpdateResult(c *gin.Context) {
	id := c.Param("id")
	if !uuidutil.IsValid(id) {
		respond.BadRequest(c, "Invalid match ID format")
		return
	}

	var req ReportResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.BadRequest(c, "Invalid JSON payload")
		return
	}

	if err := validation.Validate(&req); err != nil {
		respond.CustomError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	res, err := h.svc.UpdateResult(c.Request.Context(), id, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	respond.Success(c, "Match result updated successfully", res)
}

func (h *Handler) GetMatchReport(c *gin.Context) {
	id := c.Param("id")
	if !uuidutil.IsValid(id) {
		respond.BadRequest(c, "Invalid match ID format")
		return
	}

	res, err := h.svc.GetMatchReport(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	respond.Success(c, "Match report fetched successfully", res)
}

func (h *Handler) SoftDelete(c *gin.Context) {
	id := c.Param("id")
	if !uuidutil.IsValid(id) {
		respond.BadRequest(c, "Invalid match ID format")
		return
	}

	if err := h.svc.SoftDelete(c.Request.Context(), id); err != nil {
		h.handleServiceError(c, err)
		return
	}
	respond.Success(c, "Match deleted successfully", nil)
}
