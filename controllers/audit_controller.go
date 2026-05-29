package controllers

import (
	"net/http"
	"strconv"

	"vehicle-management-system/services"

	"github.com/gin-gonic/gin"
)

type AuditController struct {
	auditService *services.AuditService
}

func NewAuditController() *AuditController {
	return &AuditController{
		auditService: services.NewAuditService(),
	}
}

func (c *AuditController) List(ctx *gin.Context) {
	pageParam := ctx.DefaultQuery("page", "1")
	pageSizeParam := ctx.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeParam)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := c.auditService.GetAll(page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get audit logs"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
