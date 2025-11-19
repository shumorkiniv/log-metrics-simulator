package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"log-metrics-simulator/generator"
	"log-metrics-simulator/models"
	"log-metrics-simulator/scenarios"
)

var (
	scenarioManager *scenarios.ScenarioManager
)

func SetScenarioManager(sm *scenarios.ScenarioManager) {
	scenarioManager = sm
}

func GenerateLogsAndMetrics(c *gin.Context) {
	var req models.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос: " + err.Error()})
		return
	}

	if req.LogCount <= 0 || req.LogCount > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "log_count должен быть между 1 и 10000"})
		return
	}

	logs := generator.GenerateLogs(req.LogCount, req.Scenario)

	// Защита от потенциально пустого результата на случай будущих изменений генератора
	var sample any = nil
	if len(logs) > 0 {
		sample = logs[0]
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"generated":     len(logs),
		"sample_log":    sample,
		"metrics_count": len(generator.GetMetrics()),
	})
}

func GetMetrics(c *gin.Context) {
	format := c.Query("format")

	if format == "json" {
		metrics := generator.GetMetrics()
		c.JSON(http.StatusOK, gin.H{
			"metrics": metrics,
			"count":   len(metrics),
		})
		return
	}

	metricsText := generator.GetMetricsPrometheus()
	c.Header("Content-Type", "text/plain; version=0.0.4")
	c.String(http.StatusOK, metricsText)
}

func GetLogs(c *gin.Context) {
	limitStr := c.Query("limit")
	service := c.Query("service")
	level := c.Query("level")
	format := c.Query("format")

	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	logs := generator.GetLogs(limit, service, level)

	if format == "text" {
		logsText := generator.FormatLogsAsText(logs)
		c.Header("Content-Type", "text/plain")
		c.String(http.StatusOK, logsText)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"count": len(logs),
		"filters": gin.H{
			"service": service,
			"level":   level,
			"limit":   limit,
		},
	})
}

func StartScenario(c *gin.Context) {
	var req struct {
		Type   string                 `json:"type" binding:"required"`
		Config map[string]interface{} `json:"config,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос: " + err.Error()})
		return
	}

	if scenarioManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Scenario manager not initialized"})
		return
	}

	if err := scenarioManager.StartScenario(req.Type, req.Config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Сценарий запущен",
		"type":    req.Type,
	})
}

func StopScenario(c *gin.Context) {
	var req struct {
		Type string `json:"type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос: " + err.Error()})
		return
	}

	if scenarioManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Scenario manager not initialized"})
		return
	}

	if err := scenarioManager.StopScenario(req.Type); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Сценарий остановлен",
		"type":    req.Type,
	})
}

func ListScenarios(c *gin.Context) {
	if scenarioManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Scenario manager not initialized"})
		return
	}

	available := scenarioManager.GetAvailableScenarios()
	active := scenarioManager.GetActiveScenarios()
	chains := scenarioManager.GetAvailableChains()

	c.JSON(http.StatusOK, gin.H{
		"available": available,
		"active":    active,
		"chains":    chains,
	})
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"services": gin.H{
			"generator":        true,
			"scenario_manager": scenarioManager != nil,
		},
	})
}

func CreateSchedule(c *gin.Context) {
	var req struct {
		Name         string     `json:"name" binding:"required"`
		ScenarioType string     `json:"scenario_type" binding:"required"`
		CronExpr     string     `json:"cron_expr" binding:"required"`
		Timezone     string     `json:"timezone,omitempty"`
		Enabled      bool       `json:"enabled"`
		StartDate    *time.Time `json:"start_date,omitempty"`
		EndDate      *time.Time `json:"end_date,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос: " + err.Error()})
		return
	}

	schedule := &models.Schedule{
		Name:         req.Name,
		ScenarioType: req.ScenarioType,
		CronExpr:     req.CronExpr,
		Enabled:      req.Enabled,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
	}

	if err := scenarioManager.CreateSchedule(schedule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"message":  "Расписание создано",
		"schedule": schedule,
	})
}

func UpdateSchedule(c *gin.Context) {
	scheduleID := c.Param("id")

	var req struct {
		Name     string `json:"name,omitempty"`
		CronExpr string `json:"cron_expr,omitempty"`
		Timezone string `json:"timezone,omitempty"`
		Enabled  *bool  `json:"enabled,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос: " + err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.CronExpr != "" {
		updates["cron_expr"] = req.CronExpr
	}
	if req.Timezone != "" {
		updates["timezone"] = req.Timezone
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if err := scenarioManager.UpdateSchedule(scheduleID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Расписание обновлено",
	})
}

func GetSchedule(c *gin.Context) {
	scheduleID := c.Param("id")

	schedule, exists := scenarioManager.GetSchedule(scheduleID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Расписание не найдено"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"schedule": schedule,
	})
}

func ListSchedules(c *gin.Context) {
	schedules := scenarioManager.GetSchedules()
	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"schedules": schedules,
	})
}

func DeleteSchedule(c *gin.Context) {
	scheduleID := c.Param("id")

	if err := scenarioManager.DeleteSchedule(scheduleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Расписание удалено",
	})
}

func EnableSchedule(c *gin.Context) {
	scheduleID := c.Param("id")

	if err := scenarioManager.EnableSchedule(scheduleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Расписание активировано",
	})
}

func DisableSchedule(c *gin.Context) {
	scheduleID := c.Param("id")

	if err := scenarioManager.DisableSchedule(scheduleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Расписание отключено",
	})
}

func GetCronExamples(c *gin.Context) {
	examples := []gin.H{
		{
			"expression":  "0 30 9 * * *",
			"description": "Каждый день в 9:30:00",
		},
		{
			"expression":  "0 0 2 * * *",
			"description": "Каждый день в 2:00:00",
		},
		{
			"expression":  "0 */5 * * * *",
			"description": "Каждые 5 минут",
		},
		{
			"expression":  "0 0 9 * * 1",
			"description": "Каждый понедельник в 9:00:00",
		},
		{
			"expression":  "0 0 6,18 * * *",
			"description": "В 6:00:00 и 18:00:00 каждый день",
		},
		{
			"expression":  "0 0 0 1 * *",
			"description": "Первое число каждого месяца в 00:00:00",
		},
		{
			"expression":  "0 0 12 * * *",
			"description": "Каждый день в 12:00:00",
		},
		{
			"expression":  "0 30 14 * * *",
			"description": "Каждый день в 14:30:00",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"examples": examples,
	})
}

func GetLogStatistics(c *gin.Context) {
	stats := generator.GetLogStatistics()
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"stats":  stats,
	})
}

// ===== Цепочки сценариев =====

func ListChains(c *gin.Context) {
	chains, err := scenarioManager.GetChains()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"chains": chains,
	})
}

func CreateChain(c *gin.Context) {
	var req struct {
		Name        string             `json:"name" binding:"required"`
		Description string             `json:"description"`
		Steps       []models.ChainStep `json:"steps" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос: " + err.Error()})
		return
	}

	// Валидация шагов
	if len(req.Steps) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Цепочка должна содержать хотя бы один шаг"})
		return
	}

	for i, step := range req.Steps {
		if step.ScenarioType == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Шаг %d: не указан тип сценария", i+1)})
			return
		}
		step.Order = i // Устанавливаем порядок выполнения
	}

	chain := &models.ScenarioChain{
		Name:        req.Name,
		Description: req.Description,
		Steps:       req.Steps,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}

	if err := scenarioManager.CreateChain(chain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"message":  "Цепочка создана",
		"chain_id": chain.ID,
		"chain":    chain,
	})
}

func GetChain(c *gin.Context) {
	chainID := c.Param("id")

	chain, err := scenarioManager.GetChain(chainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if chain == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Цепочка не найдена"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"chain":  chain,
	})
}

func StartChain(c *gin.Context) {
	chainID := c.Param("id")

	if err := scenarioManager.StartChain(chainID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"message":  "Цепочка запущена",
		"chain_id": chainID,
	})
}

func StopChain(c *gin.Context) {
	executionID := c.Param("execution_id")

	if err := scenarioManager.StopChain(executionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Цепочка остановлена",
	})
}

func DeleteChain(c *gin.Context) {
	chainID := c.Param("id")

	if err := scenarioManager.DeleteChain(chainID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Цепочка удалена",
	})
}

func GetChainExecutions(c *gin.Context) {
	chainID := c.Param("id")
	limitStr := c.Query("limit")

	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	executions, err := scenarioManager.GetChainExecutions(chainID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"executions": executions,
	})
}

func CreateChainSchedule(c *gin.Context) {
	var req struct {
		Name      string     `json:"name" binding:"required"`
		ChainName string     `json:"chain_name" binding:"required"`
		CronExpr  string     `json:"cron_expr" binding:"required"`
		Enabled   bool       `json:"enabled"`
		StartDate *time.Time `json:"start_date,omitempty"`
		EndDate   *time.Time `json:"end_date,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос: " + err.Error()})
		return
	}

	schedule := &models.ChainSchedule{
		Name:      req.Name,
		ChainName: req.ChainName,
		CronExpr:  req.CronExpr,
		Enabled:   req.Enabled,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}

	if err := scenarioManager.CreateChainSchedule(schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"schedule": schedule,
	})
}

func ListChainSchedules(c *gin.Context) {
	schedules := scenarioManager.ListChainSchedules()
	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"schedules": schedules,
	})
}

func GetChainSchedule(c *gin.Context) {
	id := c.Param("id")

	schedule, err := scenarioManager.GetChainSchedule(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if schedule == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Расписание цепочки не найдено"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"schedule": schedule,
	})
}

func UpdateChainSchedule(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name      string     `json:"name,omitempty"`
		CronExpr  string     `json:"cron_expr,omitempty"`
		Enabled   *bool      `json:"enabled,omitempty"`
		StartDate *time.Time `json:"start_date,omitempty"`
		EndDate   *time.Time `json:"end_date,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос: " + err.Error()})
		return
	}

	// Получаем текущее расписание
	schedule, err := scenarioManager.GetChainSchedule(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if schedule == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Расписание цепочки не найдено"})
		return
	}

	// Обновляем поля
	if req.Name != "" {
		schedule.Name = req.Name
	}
	if req.CronExpr != "" {
		schedule.CronExpr = req.CronExpr
	}
	if req.Enabled != nil {
		schedule.Enabled = *req.Enabled
	}
	if req.StartDate != nil {
		schedule.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		schedule.EndDate = req.EndDate
	}

	// Сохраняем изменения
	if err := scenarioManager.UpdateChainSchedule(schedule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"message":  "Расписание цепочки обновлено",
		"schedule": schedule,
	})
}

func EnableChainSchedule(c *gin.Context) {
	id := c.Param("id")
	if err := scenarioManager.EnableChainSchedule(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func DisableChainSchedule(c *gin.Context) {
	id := c.Param("id")
	if err := scenarioManager.DisableChainSchedule(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func DeleteChainSchedule(c *gin.Context) {
	id := c.Param("id")
	if err := scenarioManager.DeleteChainSchedule(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func GetActiveChainExecutions(c *gin.Context) {
	// Получаем все активные выполнения цепочек
	// Этот метод нужно будет добавить в ScenarioManager
	// Пока оставляем заглушку
	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"executions": []interface{}{},
		"count":      0,
	})
}
