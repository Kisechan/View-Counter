package handler

import (
	"database/sql"
	"net/http"
	"time"
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

// 处理获取每日访问量统计的请求
func (h *ViewsHandler) GetDailyStatistics(c *gin.Context) {
	domain := utils.ExtractDomain(c.Request)
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain not found"})
		return
	}

	// 从查询参数获取日期范围
	startDateStr := c.DefaultQuery("start_date", "")
	endDateStr := c.DefaultQuery("end_date", "")

	// 验证日期格式并设置默认值
	if startDateStr == "" {
		startDateStr = time.Now().AddDate(0, 0, -6).Format("2006-01-02")
		// 默认过去7天
	}
	if endDateStr == "" {
		endDateStr = time.Now().Format("2006-01-02")
		// 默认今天
	}

	results, err := h.counterService.GetDailyViews(domain, startDateStr, endDateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error when fetching daily statistics"})
		return
	}

	if results == nil {
		// 如果没有数据，返回空数组而不是 null
		results = []service.DailyViewResult{}
	}

	c.JSON(http.StatusOK, results)
}