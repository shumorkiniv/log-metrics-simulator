package main

import (
	"log"
	"os"
	"strconv"

	"log-metrics-simulator/handlers"
	"log-metrics-simulator/scenarios"
	"log-metrics-simulator/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	// Получаем порт из переменных окружения
	port := getEnv("PORT", "8080")
	environment := getEnv("ENVIRONMENT", "development")
	logLevel := getEnv("LOG_LEVEL", "info")

	// Настраиваем режим Gin
	if environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Инициализация хранилища
	storage := storage.NewMemoryStorage()

	// Инициализация менеджера сценариев
	scenarioManager := scenarios.NewScenarioManager(storage)
	defer scenarioManager.Stop()

	handlers.SetScenarioManager(scenarioManager)

	router := gin.Default()
	router.Use(gin.Recovery())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check (без префикса для Prometheus и nginx)
	router.GET("/health", handlers.HealthCheck)

	// Metrics endpoint для Prometheus (без префикса)
	router.GET("/metrics", handlers.GetMetrics)

	// API группа с префиксом v1
	api := router.Group("/api/v1")
	{
		// Основные ручки
		api.POST("/generate", handlers.GenerateLogsAndMetrics)
		api.GET("/metrics", handlers.GetMetrics) // Дублируем для API
		api.GET("/logs", handlers.GetLogs)
		api.GET("/logs/stats", handlers.GetLogStatistics)

		// Управление сценариями
		scenarios := api.Group("/scenarios")
		{
			scenarios.POST("/start", handlers.StartScenario)
			scenarios.POST("/stop", handlers.StopScenario)
			scenarios.GET("/list", handlers.ListScenarios)
		}

		// Управление расписаниями
		schedules := api.Group("/schedules")
		{
			schedules.POST("", handlers.CreateSchedule)
			schedules.GET("", handlers.ListSchedules)
			schedules.GET("/:id", handlers.GetSchedule)
			schedules.PUT("/:id", handlers.UpdateSchedule)
			schedules.DELETE("/:id", handlers.DeleteSchedule)
			schedules.POST("/:id/enable", handlers.EnableSchedule)
			schedules.POST("/:id/disable", handlers.DisableSchedule)
			schedules.GET("/cron/examples", handlers.GetCronExamples)
		}

		// Цепочки сценариев и их расписания
		chains := api.Group("/chains")
		{
			chains.POST("", handlers.CreateChain)
			chains.GET("", handlers.ListChains)
			chains.GET("/:id", handlers.GetChain)
			chains.POST("/:id/start", handlers.StartChain)
			chains.POST("/:id/stop", handlers.StopChain)
			chains.DELETE("/:id", handlers.DeleteChain)
			chains.GET("/:id/executions", handlers.GetChainExecutions)

			// Расписания цепочек
			chainSchedules := chains.Group("/schedules")
			{
				chainSchedules.POST("", handlers.CreateChainSchedule)
				chainSchedules.GET("", handlers.ListChainSchedules)
				chainSchedules.GET("/:id", handlers.GetChainSchedule)    // Добавлен GET для конкретного расписания
				chainSchedules.PUT("/:id", handlers.UpdateChainSchedule) // Добавлен PUT для обновления расписания
				chainSchedules.POST("/:id/enable", handlers.EnableChainSchedule)
				chainSchedules.POST("/:id/disable", handlers.DisableChainSchedule)
				chainSchedules.DELETE("/:id", handlers.DeleteChainSchedule)
			}
		}
	}

	log.Printf("🚀 Metrics Simulator запущен в окружении: %s", environment)
	log.Printf("📊 Порт: %s", port)
	log.Printf("🔧 Уровень логирования: %s", logLevel)
	log.Printf("💡 Health check: http://localhost:%s/health", port)
	log.Printf("📈 Prometheus metrics: http://localhost:%s/metrics", port)
	log.Printf("📚 API: http://localhost:%s/api/v1/", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}

// Вспомогательная функция для получения переменных окружения
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
