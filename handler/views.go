package handler

import (
	"database/sql"
	"net/http"
	"view-counter/service"
	"view-counter/utils"

	"github.com/gin-gonic/gin"
)

type ViewsHandler struct {
	counterService *service.CounterService
}

func NewViewsHandler(counterService *service.CounterService) *ViewsHandler {
	return &ViewsHandler{
		counterService: counterService,
	}
}

func (h *ViewsHandler) IncrementView(c *gin.Context) {
	domain := utils.ExtractDomain(c.Request)
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain not found"})
		return
	}

	if err := h.counterService.IncrementView(domain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}

	c.Status(http.StatusOK)
}

func (h *ViewsHandler) GetView(c *gin.Context) {
	domain := utils.ExtractDomain(c.Request)
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain not found"})
		return
	}

	total, err := h.counterService.GetView(domain)
	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}

	// 如果没有记录，总数返回 0
	if err == sql.ErrNoRows {
		total = 0
	}

	c.String(http.StatusOK, "%d", total)
}