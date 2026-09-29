package teams

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"football-api/pkg/respond"
	"football-api/pkg/storage"
	"football-api/pkg/uuidutil"
	"football-api/pkg/validation"
)

type Handler struct {
	svc     Service
	storage storage.Storage
	log     *slog.Logger
}

func NewHandler(svc Service, storage storage.Storage, log *slog.Logger) *Handler {
	return &Handler{svc: svc, storage: storage, log: log}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("", h.Create)
	rg.GET("", h.List)
	rg.GET("/:id", h.GetByID)
	rg.PUT("/:id", h.Update)
	rg.DELETE("/:id", h.SoftDelete)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateTeamRequest

	if strings.HasPrefix(c.ContentType(), "multipart/form-data") {
		req.Name = c.PostForm("name")
		if yr, err := strconv.Atoi(c.PostForm("founded_year")); err == nil {
			req.FoundedYear = yr
		}
		req.Address = c.PostForm("address")
		req.City = c.PostForm("city")

		fileHeader, err := c.FormFile("logo")
		if err == nil && fileHeader != nil {
			if err := storage.ValidateImage(fileHeader); err != nil {
				respond.BadRequest(c, err.Error())
				return
			}
			if h.storage != nil {
				f, err := fileHeader.Open()
				if err != nil {
					respond.BadRequest(c, "failed to open uploaded file")
					return
				}
				defer f.Close()

				url, err := h.storage.UploadImage(c.Request.Context(), f, fileHeader.Filename)
				if err != nil {
					h.log.Error("failed to upload logo image", "error", err)
					respond.InternalError(c, "failed to upload logo image")
					return
				}
				req.LogoURL = url
			}
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			respond.BadRequest(c, "invalid request body")
			return
		}
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
	respond.Created(c, "team created", resp)
}

func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	teams, meta, err := h.svc.List(c.Request.Context(), page, limit)
	if err != nil {
		h.log.Error("teams.List", slog.String("error", err.Error()))
		respond.InternalError(c, "internal server error")
		return
	}
	respond.PaginatedSuccess(c, "ok", teams, meta)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if !uuidutil.IsValid(id) {
		respond.BadRequest(c, "invalid team id")
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
		respond.BadRequest(c, "invalid team id")
		return
	}

	var req UpdateTeamRequest

	if strings.HasPrefix(c.ContentType(), "multipart/form-data") {
		if val := c.PostForm("name"); val != "" {
			req.Name = &val
		}
		if val := c.PostForm("founded_year"); val != "" {
			if yr, err := strconv.Atoi(val); err == nil {
				req.FoundedYear = &yr
			}
		}
		if val := c.PostForm("address"); val != "" {
			req.Address = &val
		}
		if val := c.PostForm("city"); val != "" {
			req.City = &val
		}

		fileHeader, err := c.FormFile("logo")
		if err == nil && fileHeader != nil {
			if err := storage.ValidateImage(fileHeader); err != nil {
				respond.BadRequest(c, err.Error())
				return
			}
			if h.storage != nil {
				f, err := fileHeader.Open()
				if err != nil {
					respond.BadRequest(c, "failed to open uploaded file")
					return
				}
				defer f.Close()

				url, err := h.storage.UploadImage(c.Request.Context(), f, fileHeader.Filename)
				if err != nil {
					h.log.Error("failed to upload logo image", "error", err)
					respond.InternalError(c, "failed to upload logo image")
					return
				}
				req.LogoURL = &url
			}
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			respond.BadRequest(c, "invalid request body")
			return
		}
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
	respond.Success(c, "team updated", resp)
}

func (h *Handler) SoftDelete(c *gin.Context) {
	id := c.Param("id")
	if !uuidutil.IsValid(id) {
		respond.BadRequest(c, "invalid team id")
		return
	}

	if err := h.svc.SoftDelete(c.Request.Context(), id); err != nil {
		h.handleServiceError(c, err)
		return
	}
	respond.Success(c, "team deleted", nil)
}

func (h *Handler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrTeamNotFound):
		respond.NotFound(c, err.Error())
	case errors.Is(err, ErrTeamNameConflict):
		respond.Conflict(c, err.Error())
	case errors.Is(err, ErrInvalidFoundedYear):
		respond.BadRequest(c, err.Error())
	default:
		h.log.Error("teams handler", slog.String("error", err.Error()))
		respond.InternalError(c, "internal server error")
	}
}
