package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	domain "github.com/hextec/yolyolakay-backend/internal/media/domain"
)

type Handler struct {
	useCase domain.MediaUseCase
}

func NewHandler(useCase domain.MediaUseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) InitRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	public := router.Group("/api/v1/media")
	{
		public.GET("/:filename", h.GetMedia)
	}

	protected := router.Group("/api/v1/media")
	protected.Use(authMiddleware)
	{
		protected.POST("/upload", h.UploadMedia)
	}
}

// UploadMedia godoc
// @Summary Upload a media file
// @Tags media
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Success 201 {object} domain.Media
// @Router /api/v1/media/upload [post]
// @Security BearerAuth
func (h *Handler) UploadMedia(c *gin.Context) {
	// Multipart form parsing
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get file from request"})
		return
	}
	defer file.Close()

	// Call UseCase
	mediaInfo, err := h.useCase.UploadFile(c.Request.Context(), file, header.Size, header.Filename)
	if err != nil {
		// Differentiate between validation errors and internal errors
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, mediaInfo)
}

// GetMedia godoc
// @Summary Get a media file
// @Tags media
// @Produce octet-stream
// @Param filename path string true "Filename"
// @Success 200 {file} file
// @Router /api/v1/media/{filename} [get]
func (h *Handler) GetMedia(c *gin.Context) {
	filename := c.Param("filename")
	
	reader, contentType, err := h.useCase.GetFile(c.Request.Context(), filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	defer reader.Close()

	c.DataFromReader(http.StatusOK, -1, contentType, reader, nil)
}
