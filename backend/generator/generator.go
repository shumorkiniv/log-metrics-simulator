package generator

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"log-metrics-simulator/models"
)

var (
	logs         []models.LogEntry
	metrics      []models.Metric
	logsMutex    sync.RWMutex
	metricsMutex sync.RWMutex
	// Собственные счетчики приложения (кумулятивные)
	appGeneratedLogsTotal    int64
	appGeneratedMetricsTotal int64
)

// Сервисы приложения интернет-магазина
var services = []string{
	"api-gateway", "auth-service", "user-service", "product-service",
	"cart-service", "order-service", "payment-service", "inventory-service",
	"notification-service", "search-service", "recommendation-service",
	"analytics-service", "shipping-service", "review-service",
}

// Реалистичные категории товаров
var productCategories = []string{
	"electronics", "clothing", "books", "home-garden", "sports",
	"beauty", "toys", "automotive", "health", "food-drinks",
}

// Реалистичные статусы заказов
var orderStatuses = []string{
	"pending", "confirmed", "processing", "shipped", "delivered", "cancelled", "returned",
}

// Реалистичные способы оплаты
var paymentMethods = []string{
	"credit_card", "debit_card", "paypal", "apple_pay", "google_pay", "bank_transfer",
}

// Города для доставки
var cities = []string{
	"Moscow", "Saint Petersburg", "Novosibirsk", "Yekaterinburg", "Kazan",
	"Nizhny Novgorod", "Chelyabinsk", "Samara", "Omsk", "Rostov-on-Don",
}

// HTTP методы и пути
var httpMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
var httpPaths = []string{
	"/api/v1/auth/login",
	"/api/v1/auth/register",
	"/api/v1/users/profile",
	"/api/v1/products",
	"/api/v1/products/{id}",
	"/api/v1/cart",
	"/api/v1/cart/items",
	"/api/v1/orders",
	"/api/v1/orders/{id}",
	"/api/v1/payments",
	"/api/v1/search",
	"/api/v1/inventory",
}

// User Agents (обновленные)
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Linux; Android 13; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Mobile Safari/537.36",
}

// IP адреса (российские диапазоны)
var ipAddresses = []string{
	"89.108.65.23", "95.165.133.45", "178.176.74.89", "46.138.234.56",
	"91.200.12.34", "188.162.45.78", "37.139.56.89", "85.26.234.12",
}

func GenerateLogs(logCount int, scenario string) []models.LogEntry {
	generatedLogs := make([]models.LogEntry, logCount)

	for i := 0; i < logCount; i++ {
		generatedLogs[i] = generateRealisticLog(scenario)
		// Реалистичная задержка между запросами
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(50)+10))
	}

	// Сохраняем логи
	logsMutex.Lock()
	logs = append(logs, generatedLogs...)
	if len(logs) > 50000 {
		logs = logs[len(logs)-50000:]
	}
	logsMutex.Unlock()

	// Обновляем метрики
	updateEcommerceMetrics(generatedLogs)

	// Инкрементируем собственные счетчики приложения
	// Общее количество сгенерированных логов (кумулятивно)
	appGeneratedLogsTotal += int64(len(generatedLogs))
	// Сохраняем текущий размер набора метрик как суммарно сгенерированные метрики
	metricsMutex.RLock()
	currentMetrics := len(metrics)
	metricsMutex.RUnlock()
	appGeneratedMetricsTotal += int64(currentMetrics)

	// Пишем краткую запись в лог приложения
	logsMutex.RLock()
	totalLogs := len(logs)
	logsMutex.RUnlock()
	log.Printf("[simulator] generated=%d scenario=%s total_logs=%d metrics_now=%d", len(generatedLogs), scenario, totalLogs, currentMetrics)

	for _, log := range logs {
		logJson, _ := json.Marshal(log)
		fmt.Println(string(logJson))
	}
	
	return generatedLogs
}

func generateRealisticLog(scenario string) models.LogEntry {
	service := services[rand.Intn(len(services))]
	level := getLogLevel()
	traceID := generateTraceID()
	spanID := generateSpanID()

	logEntry := models.LogEntry{
		Timestamp: time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second), // Логи за последний час
		Level:     level,
		Service:   service,
		TraceID:   traceID,
		SpanID:    spanID,
		UserID:    fmt.Sprintf("user-%d", rand.Intn(50000)+1),
		SessionID: generateSessionID(),
		IP:        ipAddresses[rand.Intn(len(ipAddresses))],
		UserAgent: userAgents[rand.Intn(len(userAgents))],
	}

	// Применяем сценарий
	switch scenario {
	case "black_friday":
		logEntry = applyBlackFridayScenario(logEntry)
	case "normal_load":
		logEntry = applyNormalLoadScenario(logEntry)
	case "high_load":
		logEntry = applyHighLoadScenario(logEntry)
	case "payment_issues":
		logEntry = applyPaymentIssuesScenario(logEntry)
	default:
		logEntry = applyNormalLoadScenario(logEntry)
	}

	// Генерируем лог в зависимости от сервиса
	switch service {
	case "api-gateway":
		logEntry = generateApiGatewayLog(logEntry, level)
	case "auth-service":
		logEntry = generateAuthLog(logEntry, level)
	case "user-service":
		logEntry = generateUserLog(logEntry, level)
	case "product-service":
		logEntry = generateProductLog(logEntry, level)
	case "cart-service":
		logEntry = generateCartLog(logEntry, level)
	case "order-service":
		logEntry = generateOrderLog(logEntry, level)
	case "payment-service":
		logEntry = generatePaymentLog(logEntry, level)
	case "inventory-service":
		logEntry = generateInventoryLog(logEntry, level)
	case "search-service":
		logEntry = generateSearchLog(logEntry, level)
	case "recommendation-service":
		logEntry = generateRecommendationLog(logEntry, level)
	default:
		logEntry = generateGenericLog(logEntry, level)
	}

	return logEntry
}

func generateAuthLog(logEntry models.LogEntry, level string) models.LogEntry {
	actions := []string{"login", "register", "logout", "token_refresh", "password_reset"}
	action := actions[rand.Intn(len(actions))]

	logEntry.Method = "POST"
	logEntry.Path = fmt.Sprintf("/api/v1/auth/%s", action)
	logEntry.Duration = int64(rand.Intn(200) + 50)

	switch level {
	case "INFO":
		logEntry.Message = fmt.Sprintf("User %s completed successfully", action)
		logEntry.Status = http.StatusOK
	case "WARN":
		logEntry.Message = fmt.Sprintf("User %s attempt with suspicious activity", action)
		logEntry.Status = http.StatusBadRequest
	case "ERROR":
		logEntry.Message = fmt.Sprintf("User %s failed", action)
		logEntry.Status = http.StatusUnauthorized
		logEntry.Error = "Invalid credentials"
	default:
		logEntry.Message = fmt.Sprintf("User %s completed", action)
		logEntry.Status = http.StatusOK
	}

	return logEntry
}

func generateUserLog(logEntry models.LogEntry, level string) models.LogEntry {
	actions := []string{"get_profile", "update_profile", "get_preferences", "update_preferences"}
	action := actions[rand.Intn(len(actions))]

	logEntry.Method = "GET"
	if action[:6] == "update" {
		logEntry.Method = "PUT"
	}
	logEntry.Path = fmt.Sprintf("/api/v1/users/%s", action)
	logEntry.Duration = int64(rand.Intn(150) + 30)

	switch level {
	case "INFO":
		logEntry.Message = fmt.Sprintf("User profile operation %s completed", action)
		logEntry.Status = http.StatusOK
	case "ERROR":
		logEntry.Message = fmt.Sprintf("User profile operation %s failed", action)
		logEntry.Status = http.StatusNotFound
		logEntry.Error = "User not found"
	default:
		logEntry.Message = fmt.Sprintf("User profile operation %s completed", action)
		logEntry.Status = http.StatusOK
	}

	return logEntry
}

func generateProductLog(logEntry models.LogEntry, level string) models.LogEntry {
	actions := []string{"list", "get", "search", "create", "update"}
	action := actions[rand.Intn(len(actions))]

	logEntry.Method = "GET"
	if action == "create" {
		logEntry.Method = "POST"
	} else if action == "update" {
		logEntry.Method = "PUT"
	}

	productID := fmt.Sprintf("prod-%d", rand.Intn(1000)+1)
	logEntry.Path = fmt.Sprintf("/api/v1/products/%s", productID)
	if action == "list" || action == "search" {
		logEntry.Path = "/api/v1/products"
	}

	logEntry.Duration = int64(rand.Intn(300) + 100)

	switch level {
	case "INFO":
		logEntry.Message = fmt.Sprintf("Product %s operation completed", action)
		logEntry.Status = http.StatusOK
	case "ERROR":
		logEntry.Message = fmt.Sprintf("Product %s operation failed", action)
		logEntry.Status = http.StatusInternalServerError
		logEntry.Error = "Database connection failed"
	default:
		logEntry.Message = fmt.Sprintf("Product %s operation completed", action)
		logEntry.Status = http.StatusOK
	}

	return logEntry
}

func generateCartLog(logEntry models.LogEntry, level string) models.LogEntry {
	actions := []string{"get", "add_item", "remove_item", "update_quantity", "clear"}
	action := actions[rand.Intn(len(actions))]

	logEntry.Method = "GET"
	if action != "get" {
		logEntry.Method = "POST"
	}
	logEntry.Path = "/api/v1/cart/items"
	logEntry.Duration = int64(rand.Intn(200) + 50)

	switch level {
	case "INFO":
		logEntry.Message = fmt.Sprintf("Cart %s operation completed", action)
		logEntry.Status = http.StatusOK
	case "ERROR":
		logEntry.Message = fmt.Sprintf("Cart %s operation failed", action)
		logEntry.Status = http.StatusBadRequest
		logEntry.Error = "Invalid product ID"
	default:
		logEntry.Message = fmt.Sprintf("Cart %s operation completed", action)
		logEntry.Status = http.StatusOK
	}

	return logEntry
}

func generateOrderLog(logEntry models.LogEntry, level string) models.LogEntry {
	actions := []string{"create", "get", "list", "cancel", "update_status"}
	action := actions[rand.Intn(len(actions))]

	logEntry.Method = "POST"
	if action == "get" || action == "list" {
		logEntry.Method = "GET"
	}
	logEntry.Path = "/api/v1/orders"
	logEntry.Duration = int64(rand.Intn(500) + 200)

	switch level {
	case "INFO":
		logEntry.Message = fmt.Sprintf("Order %s operation completed", action)
		logEntry.Status = http.StatusOK
	case "ERROR":
		logEntry.Message = fmt.Sprintf("Order %s operation failed", action)
		logEntry.Status = http.StatusInternalServerError
		logEntry.Error = "Payment gateway timeout"
	default:
		logEntry.Message = fmt.Sprintf("Order %s operation completed", action)
		logEntry.Status = http.StatusOK
	}

	return logEntry
}

func generatePaymentLog(logEntry models.LogEntry, level string) models.LogEntry {
	actions := []string{"process", "refund", "get_status", "create_intent"}
	action := actions[rand.Intn(len(actions))]

	logEntry.Method = "POST"
	if action == "get_status" {
		logEntry.Method = "GET"
	}
	logEntry.Path = "/api/v1/payments"
	logEntry.Duration = int64(rand.Intn(1000) + 500)

	switch level {
	case "INFO":
		logEntry.Message = fmt.Sprintf("Payment %s completed successfully", action)
		logEntry.Status = http.StatusOK
	case "ERROR":
		logEntry.Message = fmt.Sprintf("Payment %s failed", action)
		logEntry.Status = http.StatusPaymentRequired
		logEntry.Error = "Insufficient funds"
		logEntry.Stack = "payment_gateway.ProcessPayment: line 145\npayment_handler.Handle: line 89"
	default:
		logEntry.Message = fmt.Sprintf("Payment %s completed", action)
		logEntry.Status = http.StatusOK
	}

	return logEntry
}

func generateGenericLog(logEntry models.LogEntry, level string) models.LogEntry {
	methods := []string{"health_check", "metrics", "config_reload", "cache_clear"}
	method := methods[rand.Intn(len(methods))]

	logEntry.Method = "GET"
	logEntry.Path = fmt.Sprintf("/internal/%s", method)
	logEntry.Duration = int64(rand.Intn(100) + 10)

	switch level {
	case "INFO":
		logEntry.Message = fmt.Sprintf("Internal operation %s completed", method)
		logEntry.Status = http.StatusOK
	case "ERROR":
		logEntry.Message = fmt.Sprintf("Internal operation %s failed", method)
		logEntry.Status = http.StatusInternalServerError
		logEntry.Error = "Internal server error"
	default:
		logEntry.Message = fmt.Sprintf("Internal operation %s completed", method)
		logEntry.Status = http.StatusOK
	}

	return logEntry
}

func getLogLevel() string {
	r := rand.Float64()
	switch {
	case r < 0.65:
		return "INFO"
	case r < 0.85:
		return "WARN"
	case r < 0.95:
		return "ERROR"
	default:
		return "DEBUG"
	}
}

func updateMetrics(logs []models.LogEntry) {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()

	metrics = []models.Metric{}

	totalRequests := len(logs)
	statusCount := make(map[int]int)
	serviceCount := make(map[string]int)
	levelCount := make(map[string]int)
	totalDuration := int64(0)
	errorCount := 0

	for _, log := range logs {
		statusCount[log.Status]++
		serviceCount[log.Service]++
		levelCount[log.Level]++
		totalDuration += log.Duration

		if log.Level == "ERROR" {
			errorCount++
		}
	}

	now := time.Now()

	// HTTP метрики
	metrics = append(metrics, models.Metric{
		Name:      "http_requests_total",
		Value:     float64(totalRequests),
		Type:      "counter",
		Labels:    map[string]string{},
		Timestamp: now,
	})

	metrics = append(metrics, models.Metric{
		Name:  "http_request_duration_ms",
		Value: float64(totalDuration) / float64(totalRequests),
		Type:  "gauge",
		Labels: map[string]string{
			"quantile": "0.5",
		},
		Timestamp: now,
	})

	metrics = append(metrics, models.Metric{
		Name:      "http_errors_total",
		Value:     float64(errorCount),
		Type:      "counter",
		Labels:    map[string]string{},
		Timestamp: now,
	})

	// Метрики по сервисам
	for service, count := range serviceCount {
		metrics = append(metrics, models.Metric{
			Name:  "http_requests_total",
			Value: float64(count),
			Type:  "counter",
			Labels: map[string]string{
				"service": service,
			},
			Timestamp: now,
		})
	}

	// Метрики по статусам
	for status, count := range statusCount {
		metrics = append(metrics, models.Metric{
			Name:  "http_responses_total",
			Value: float64(count),
			Type:  "counter",
			Labels: map[string]string{
				"status": fmt.Sprintf("%d", status),
			},
			Timestamp: now,
		})
	}

	// Бизнес метрики
	metrics = append(metrics, models.Metric{
		Name:      "active_users",
		Value:     float64(rand.Intn(5000) + 1000),
		Type:      "gauge",
		Labels:    map[string]string{},
		Timestamp: now,
	})

	metrics = append(metrics, models.Metric{
		Name:      "orders_processed_total",
		Value:     float64(rand.Intn(10000) + 5000),
		Type:      "counter",
		Labels:    map[string]string{},
		Timestamp: now,
	})

	metrics = append(metrics, models.Metric{
		Name:      "revenue_total",
		Value:     float64(rand.Intn(1000000) + 500000),
		Type:      "counter",
		Labels:    map[string]string{"currency": "USD"},
		Timestamp: now,
	})
}

// Вспомогательные функции
func generateTraceID() string {
	return fmt.Sprintf("%x", rand.Uint64())
}

func generateSpanID() string {
	return fmt.Sprintf("%x", rand.Uint32())
}

func generateSessionID() string {
	return fmt.Sprintf("session-%x", rand.Uint64())
}

func GetLogs(limit int, service, level string) []models.LogEntry {
	logsMutex.RLock()
	defer logsMutex.RUnlock()

	var filtered []models.LogEntry
	count := 0

	for i := len(logs) - 1; i >= 0 && count < limit; i-- {
		logEntry := logs[i]

		if service != "" && logEntry.Service != service {
			continue
		}
		if level != "" && logEntry.Level != level {
			continue
		}

		filtered = append([]models.LogEntry{logEntry}, filtered...)
		count++
	}

	return filtered
}

func GetMetrics() []models.Metric {
	metricsMutex.RLock()
	defer metricsMutex.RUnlock()

	result := make([]models.Metric, len(metrics))
	copy(result, metrics)
	return result
}

func GetMetricsPrometheus() string {
	metricsMutex.RLock()
	defer metricsMutex.RUnlock()

	var result string

	// Добавляем HELP и TYPE комментарии
	metricTypes := make(map[string]string)
	metricHelp := make(map[string]string)

	// Собираем уникальные метрики
	uniqueMetrics := make(map[string]bool)
	for _, metric := range metrics {
		if !uniqueMetrics[metric.Name] {
			uniqueMetrics[metric.Name] = true
			metricTypes[metric.Name] = metric.Type

			// Добавляем описания метрик
			switch metric.Name {
			case "ecommerce_http_requests_total":
				metricHelp[metric.Name] = "Total number of HTTP requests"
			case "ecommerce_http_responses_total":
				metricHelp[metric.Name] = "Total number of HTTP responses by status code"
			case "ecommerce_service_requests_total":
				metricHelp[metric.Name] = "Total number of requests by service"
			case "ecommerce_http_request_duration_ms":
				metricHelp[metric.Name] = "Average HTTP request duration in milliseconds"
			case "ecommerce_orders_total":
				metricHelp[metric.Name] = "Total number of orders processed"
			case "ecommerce_revenue_total":
				metricHelp[metric.Name] = "Total revenue generated"
			case "ecommerce_payments_processed":
				metricHelp[metric.Name] = "Total number of payments processed"
			case "ecommerce_search_queries":
				metricHelp[metric.Name] = "Total number of search queries"
			case "ecommerce_cart_actions":
				metricHelp[metric.Name] = "Total number of cart actions"
			case "ecommerce_auth_actions":
				metricHelp[metric.Name] = "Total number of authentication actions"
			case "ecommerce_error_rate":
				metricHelp[metric.Name] = "Error rate as a percentage"
			case "ecommerce_active_users":
				metricHelp[metric.Name] = "Current number of active users"
			case "ecommerce_inventory_items_low_stock":
				metricHelp[metric.Name] = "Number of items with low stock"
			default:
				metricHelp[metric.Name] = "Application metric"
			}
		}
	}

	// Добавляем HELP и TYPE для каждой метрики
	for metricName, help := range metricHelp {
		result += fmt.Sprintf("# HELP %s %s\n", metricName, help)
		result += fmt.Sprintf("# TYPE %s %s\n", metricName, metricTypes[metricName])

		// Добавляем значения метрик
		for _, metric := range metrics {
			if metric.Name == metricName {
				labels := ""
				for k, v := range metric.Labels {
					if labels != "" {
						labels += ","
					}
					labels += fmt.Sprintf("%s=\"%s\"", k, v)
				}

				if labels != "" {
					result += fmt.Sprintf("%s{%s} %.2f\n", metric.Name, labels, metric.Value)
				} else {
					result += fmt.Sprintf("%s %.2f\n", metric.Name, metric.Value)
				}
			}
		}
		result += "\n"
	}

	return result
}

func GetLogStatistics() map[string]interface{} {
	logsMutex.RLock()
	defer logsMutex.RUnlock()

	stats := map[string]interface{}{
		"total_logs": len(logs),
		"services":   make(map[string]int),
		"levels":     make(map[string]int),
		"statuses":   make(map[int]int),
	}

	for _, log := range logs {
		stats["services"].(map[string]int)[log.Service]++
		stats["levels"].(map[string]int)[log.Level]++
		stats["statuses"].(map[int]int)[log.Status]++
	}

	return stats
}

// Форматирование логов для вывода
func FormatLogsAsJSON(logs []models.LogEntry) string {
	jsonData, _ := json.MarshalIndent(logs, "", "  ")
	return string(jsonData)
}

func FormatLogsAsText(logs []models.LogEntry) string {
	var result string
	for _, log := range logs {
		result += fmt.Sprintf("%s [%s] %s: %s %s %d %dms\n",
			log.Timestamp.Format("2006-01-02T15:04:05.000Z"),
			log.Level,
			log.Service,
			log.Method,
			log.Path,
			log.Status,
			log.Duration,
		)
	}
	return result
}

func generateApiGatewayLog(logEntry models.LogEntry, level string) models.LogEntry {
	paths := []string{
		"/api/v1/auth/login", "/api/v1/products", "/api/v1/cart",
		"/api/v1/orders", "/api/v1/search", "/api/v1/users/profile",
	}

	logEntry.Method = []string{"GET", "POST", "PUT", "DELETE"}[rand.Intn(4)]
	logEntry.Path = paths[rand.Intn(len(paths))]
	logEntry.Duration = int64(rand.Intn(50) + 10)

	switch level {
	case "INFO":
		logEntry.Message = fmt.Sprintf("Request routed to %s", strings.Split(logEntry.Service, "-")[0])
		logEntry.Status = http.StatusOK
	case "WARN":
		logEntry.Message = "Rate limit approaching for user"
		logEntry.Status = http.StatusTooManyRequests
	case "ERROR":
		logEntry.Message = "Service unavailable"
		logEntry.Status = http.StatusServiceUnavailable
		logEntry.Error = "Downstream service timeout"
	default:
		logEntry.Message = "Request processed successfully"
		logEntry.Status = http.StatusOK
	}

	return logEntry
}

func generateSearchLog(logEntry models.LogEntry, level string) models.LogEntry {
	searchQueries := []string{
		"iphone 15", "samsung galaxy", "nike shoes", "winter jacket",
		"laptop gaming", "wireless headphones", "kitchen appliances",
	}

	query := searchQueries[rand.Intn(len(searchQueries))]
	logEntry.Method = "GET"
	logEntry.Path = fmt.Sprintf("/api/v1/search?q=%s", strings.ReplaceAll(query, " ", "%20"))
	logEntry.Duration = int64(rand.Intn(300) + 100)

	resultsCount := rand.Intn(1000) + 1

	switch level {
	case "INFO":
		logEntry.Message = fmt.Sprintf("Search query '%s' returned %d results", query, resultsCount)
		logEntry.Status = http.StatusOK
	case "WARN":
		logEntry.Message = fmt.Sprintf("Search query '%s' took longer than expected", query)
		logEntry.Status = http.StatusOK
	case "ERROR":
		logEntry.Message = fmt.Sprintf("Search index unavailable for query '%s'", query)
		logEntry.Status = http.StatusInternalServerError
		logEntry.Error = "Elasticsearch cluster unreachable"
	default:
		logEntry.Message = fmt.Sprintf("Search completed for '%s'", query)
		logEntry.Status = http.StatusOK
	}

	return logEntry
}

func generateInventoryLog(logEntry models.LogEntry, level string) models.LogEntry {
	productID := fmt.Sprintf("prod-%d", rand.Intn(10000)+1)
	actions := []string{"check_stock", "reserve_item", "release_reservation", "update_stock"}
	action := actions[rand.Intn(len(actions))]

	logEntry.Method = "POST"
	if action == "check_stock" {
		logEntry.Method = "GET"
	}
	logEntry.Path = fmt.Sprintf("/api/v1/inventory/%s", action)
	logEntry.Duration = int64(rand.Intn(200) + 50)

	quantity := rand.Intn(100) + 1

	switch level {
	case "INFO":
		logEntry.Message = fmt.Sprintf("Inventory %s for product %s, quantity: %d", action, productID, quantity)
		logEntry.Status = http.StatusOK
	case "WARN":
		logEntry.Message = fmt.Sprintf("Low stock warning for product %s, remaining: %d", productID, quantity)
		logEntry.Status = http.StatusOK
	case "ERROR":
		logEntry.Message = fmt.Sprintf("Inventory operation failed for product %s", productID)
		logEntry.Status = http.StatusConflict
		logEntry.Error = "Insufficient stock"
	default:
		logEntry.Message = fmt.Sprintf("Inventory operation %s completed", action)
		logEntry.Status = http.StatusOK
	}

	return logEntry
}

func generateRecommendationLog(logEntry models.LogEntry, level string) models.LogEntry {
	algorithms := []string{"collaborative_filtering", "content_based", "hybrid", "trending"}
	algorithm := algorithms[rand.Intn(len(algorithms))]

	logEntry.Method = "GET"
	logEntry.Path = fmt.Sprintf("/api/v1/recommendations?user_id=%s&type=%s", logEntry.UserID, algorithm)
	logEntry.Duration = int64(rand.Intn(500) + 200)

	recommendationsCount := rand.Intn(20) + 5

	switch level {
	case "INFO":
		logEntry.Message = fmt.Sprintf("Generated %d recommendations using %s algorithm", recommendationsCount, algorithm)
		logEntry.Status = http.StatusOK
	case "WARN":
		logEntry.Message = fmt.Sprintf("Recommendation model performance degraded for algorithm %s", algorithm)
		logEntry.Status = http.StatusOK
	case "ERROR":
		logEntry.Message = fmt.Sprintf("Recommendation service failed for algorithm %s", algorithm)
		logEntry.Status = http.StatusInternalServerError
		logEntry.Error = "ML model unavailable"
	default:
		logEntry.Message = fmt.Sprintf("Recommendations generated using %s", algorithm)
		logEntry.Status = http.StatusOK
	}

	return logEntry
}

func updateEcommerceMetrics(logs []models.LogEntry) {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()

	// Сбрасываем метрики для обновления
	metrics = []models.Metric{}

	// Счетчики
	totalRequests := len(logs)
	statusCount := make(map[int]int)
	serviceCount := make(map[string]int)
	levelCount := make(map[string]int)
	methodCount := make(map[string]int)
	totalDuration := int64(0)
	errorCount := 0

	// Бизнес-метрики
	orderCount := 0
	paymentCount := 0
	searchCount := 0
	cartActions := 0
	authActions := 0
	totalRevenue := 0.0

	for _, log := range logs {
		statusCount[log.Status]++
		serviceCount[log.Service]++
		levelCount[log.Level]++
		methodCount[log.Method]++
		totalDuration += log.Duration

		if log.Level == "ERROR" {
			errorCount++
		}

		// Подсчет бизнес-метрик
		switch log.Service {
		case "order-service":
			orderCount++
			if log.Status == http.StatusOK && strings.Contains(log.Message, "create") {
				totalRevenue += float64(rand.Intn(5000) + 100) // Случайная сумма заказа
			}
		case "payment-service":
			paymentCount++
		case "search-service":
			searchCount++
		case "cart-service":
			cartActions++
		case "auth-service":
			authActions++
		}
	}

	now := time.Now()

	// HTTP метрики
	metrics = append(metrics, models.Metric{
		Name:      "ecommerce_http_requests_total",
		Value:     float64(totalRequests),
		Type:      "counter",
		Labels:    map[string]string{"app": "ecommerce"},
		Timestamp: now,
	})

	// Метрики по статус-кодам
	for status, count := range statusCount {
		metrics = append(metrics, models.Metric{
			Name:      "ecommerce_http_responses_total",
			Value:     float64(count),
			Type:      "counter",
			Labels:    map[string]string{"status": fmt.Sprintf("%d", status), "app": "ecommerce"},
			Timestamp: now,
		})
	}

	// Метрики по сервисам
	for service, count := range serviceCount {
		metrics = append(metrics, models.Metric{
			Name:      "ecommerce_service_requests_total",
			Value:     float64(count),
			Type:      "counter",
			Labels:    map[string]string{"service": service, "app": "ecommerce"},
			Timestamp: now,
		})
	}

	// Метрики времени отклика
	if totalRequests > 0 {
		avgDuration := float64(totalDuration) / float64(totalRequests)
		metrics = append(metrics, models.Metric{
			Name:      "ecommerce_http_request_duration_ms",
			Value:     avgDuration,
			Type:      "gauge",
			Labels:    map[string]string{"app": "ecommerce"},
			Timestamp: now,
		})
	}

	// Бизнес-метрики интернет-магазина
	metrics = append(metrics, []models.Metric{
		{
			Name:      "ecommerce_orders_total",
			Value:     float64(orderCount),
			Type:      "counter",
			Labels:    map[string]string{"app": "ecommerce"},
			Timestamp: now,
		},
		{
			Name:      "ecommerce_revenue_total",
			Value:     totalRevenue,
			Type:      "counter",
			Labels:    map[string]string{"currency": "RUB", "app": "ecommerce"},
			Timestamp: now,
		},
		{
			Name:      "ecommerce_payments_processed",
			Value:     float64(paymentCount),
			Type:      "counter",
			Labels:    map[string]string{"app": "ecommerce"},
			Timestamp: now,
		},
		{
			Name:      "ecommerce_search_queries",
			Value:     float64(searchCount),
			Type:      "counter",
			Labels:    map[string]string{"app": "ecommerce"},
			Timestamp: now,
		},
		{
			Name:      "ecommerce_cart_actions",
			Value:     float64(cartActions),
			Type:      "counter",
			Labels:    map[string]string{"app": "ecommerce"},
			Timestamp: now,
		},
		{
			Name:      "ecommerce_auth_actions",
			Value:     float64(authActions),
			Type:      "counter",
			Labels:    map[string]string{"app": "ecommerce"},
			Timestamp: now,
		},
		{
			Name:      "ecommerce_error_rate",
			Value:     float64(errorCount) / float64(totalRequests),
			Type:      "gauge",
			Labels:    map[string]string{"app": "ecommerce"},
			Timestamp: now,
		},
		{
			Name:      "ecommerce_active_users",
			Value:     float64(rand.Intn(1000) + 100), // Симуляция активных пользователей
			Type:      "gauge",
			Labels:    map[string]string{"app": "ecommerce"},
			Timestamp: now,
		},
		{
			Name:      "ecommerce_inventory_items_low_stock",
			Value:     float64(rand.Intn(50) + 5), // Симуляция товаров с низким остатком
			Type:      "gauge",
			Labels:    map[string]string{"app": "ecommerce"},
			Timestamp: now,
		},
	}...)

	// Добавляем собственные метрики приложения
	// Эти счетчики кумулятивны за время работы процесса
	now = time.Now()
	metrics = append(metrics,
		models.Metric{
			Name:      "app_generated_logs_total",
			Value:     float64(appGeneratedLogsTotal),
			Type:      "counter",
			Labels:    map[string]string{"app": "simulator"},
			Timestamp: now,
		},
		models.Metric{
			Name:      "app_generated_metrics_total",
			Value:     float64(appGeneratedMetricsTotal),
			Type:      "counter",
			Labels:    map[string]string{"app": "simulator"},
			Timestamp: now,
		},
	)
}

// Сценарии нагрузки
func applyBlackFridayScenario(logEntry models.LogEntry) models.LogEntry {
	// Увеличиваем вероятность ошибок и времени отклика
	if rand.Float64() < 0.3 {
		logEntry.Level = "ERROR"
		logEntry.Duration = int64(rand.Intn(5000) + 1000) // Высокое время отклика
	}
	return logEntry
}

func applyPaymentIssuesScenario(logEntry models.LogEntry) models.LogEntry {
	if logEntry.Service == "payment-service" && rand.Float64() < 0.4 {
		logEntry.Level = "ERROR"
		logEntry.Error = "Payment gateway unavailable"
	}
	return logEntry
}

func applyHighLoadScenario(logEntry models.LogEntry) models.LogEntry {
	// Увеличиваем время отклика
	logEntry.Duration = int64(float64(logEntry.Duration) * 1.5)
	return logEntry
}

func applyNormalLoadScenario(logEntry models.LogEntry) models.LogEntry {
	// Нормальная работа
	return logEntry
}
