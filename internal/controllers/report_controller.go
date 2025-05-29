// internal/controllers/report_controller.go
package controllers

import (
	"modules/internal/models"
	"modules/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ReportController struct {
	service *services.ReportService
}

func NewReportController(service *services.ReportService) *ReportController {
	return &ReportController{service: service}
}

// @Summary 获取日报表
// @Description 根据指定天数获取每日停车收入和车位使用统计，默认返回最近7天的数据。
// @Tags reports
// @Produce json
// @Param days query int false "查询天数，默认为7天" default(7)
// @Security BearerAuth
// @Success 200 {object} DailyReportResponse "日报表数据"
// @Failure 400 {object} ErrorResponse "无效的查询参数"
// @Failure 401 {object} ErrorResponse "未授权访问"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /reports/daily [get]
func (c *ReportController) GetDailyReport(ctx *gin.Context) {
	days, _ := strconv.Atoi(ctx.DefaultQuery("days", "7"))

	report, err := c.service.GenerateDailyReport(ctx, days)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, ToDailyReportResponse(report))
}

// DTO转换
type DailyReportResponse struct {
	Date         string  `json:"date"`
	TotalIncome  float64 `json:"total_income"`
	TemporaryCnt int     `json:"temporary_count"`
	ShortTermCnt int     `json:"short_term_count"`
	PermanentCnt int     `json:"permanent_count"`
}

func ToDailyReportResponse(r *models.DailyReport) *DailyReportResponse {
	return &DailyReportResponse{
		Date:         r.Date.Format("2006-01-02"),
		TotalIncome:  r.TotalIncome,
		TemporaryCnt: r.TemporaryCnt,
		ShortTermCnt: r.ShortTermCnt,
		PermanentCnt: r.PermanentCnt,
	}
}
