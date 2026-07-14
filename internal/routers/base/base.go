package base

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
)

// ParseStatusFilter 解析可选的状态过滤查询值。
// 空字符串、非数字或负数一律返回 -1,表示"不按状态过滤";否则返回原始状态
// 枚举值(与数据库存储值一致)。调用方不得再做 ±1 调整——前端直接传枚举值
// (如任务:启用=1/禁用=0;日志:失败=0/运行中=1/成功=2/取消=3)。
func ParseStatusFilter(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return -1
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return -1
	}
	return v
}

// ParsePageAndPageSize 解析查询参数中的页数和每页数量
func ParsePageAndPageSize(c *gin.Context, params models.CommonMap) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = models.PageSize
	}

	params["Page"] = page
	params["PageSize"] = pageSize
}
