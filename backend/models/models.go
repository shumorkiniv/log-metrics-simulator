package models

import "time"

// LogEntry представляет лог приложения интернет-магазина
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Service   string    `json:"service"`
	Message   string    `json:"message"`
	TraceID   string    `json:"trace_id,omitempty"`
	SpanID    string    `json:"span_id,omitempty"`
	UserID    string    `json:"user_id,omitempty"`
	SessionID string    `json:"session_id,omitempty"`
	IP        string    `json:"ip,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	Method    string    `json:"method,omitempty"`
	Path      string    `json:"path,omitempty"`
	Status    int       `json:"status,omitempty"`
	Duration  int64     `json:"duration_ms,omitempty"`
	Error     string    `json:"error,omitempty"`
	Stack     string    `json:"stack,omitempty"`
}

// Metric представляет метрику приложения
type Metric struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Type      string            `json:"type"` // counter, gauge, histogram
	Labels    map[string]string `json:"labels"`
	Timestamp time.Time         `json:"timestamp"`
}

// ScenarioConfig представляет конфигурацию сценария
type ScenarioConfig struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	LogCount    int                    `json:"log_count"`
	Parameters  map[string]interface{} `json:"parameters"`
	Labels      map[string]string
}

// Scenario представляет запущенный сценарий
type Scenario struct {
	Type      string         `json:"type"`
	Active    bool           `json:"active"`
	Config    ScenarioConfig `json:"config"`
	Started   time.Time      `json:"started"`
	Duration  time.Duration  `json:"duration,omitempty"`
	Interval  time.Duration  `json:"interval,omitempty"`
	StartDate *time.Time     `json:"start_date,omitempty"`
	EndDate   *time.Time     `json:"end_date,omitempty"`
}

// Schedule представляет расписание
type Schedule struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	ScenarioType string     `json:"scenario_type"`
	CronExpr     string     `json:"cron_expr"`
	Enabled      bool       `json:"enabled"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	LastRun      *time.Time `json:"last_run,omitempty"`
	NextRun      *time.Time `json:"next_run,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// Chain описывает цепочку сценариев
type Chain struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Steps       []string `json:"steps"` // список ScenarioType по порядку
}

// ChainSchedule представляет расписание для цепочки сценариев
type ChainSchedule struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	ChainName string     `json:"chain_name"`
	CronExpr  string     `json:"cron_expr"`
	Enabled   bool       `json:"enabled"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	LastRun   *time.Time `json:"last_run,omitempty"`
	NextRun   *time.Time `json:"next_run,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// GenerateRequest представляет запрос на генерацию
type GenerateRequest struct {
	LogCount int                    `json:"log_count" binding:"required"`
	Scenario string                 `json:"scenario,omitempty"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

// ScheduleExecution представляет выполнение расписания
type ScheduleExecution struct {
	ID           string     `json:"id"`
	ScheduleID   string     `json:"schedule_id"`
	ScenarioType string     `json:"scenario_type"`
	Status       string     `json:"status"` // running, completed, failed
	StartedAt    time.Time  `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	Error        string     `json:"error,omitempty"`
	LogsCount    int        `json:"logs_count,omitempty"`
}

// ChainStep представляет шаг в цепочке сценариев
type ChainStep struct {
	ScenarioType string                 `json:"scenario_type" binding:"required"`
	Name         string                 `json:"name"`
	Config       map[string]interface{} `json:"config"`
	DelayBefore  int                    `json:"delay_before"` // Задержка перед запуском в секундах
	Order        int                    `json:"order"`        // Порядок выполнения
}

// ScenarioChain представляет цепочку сценариев
type ScenarioChain struct {
	ID          string      `json:"id"`
	Name        string      `json:"name" binding:"required"`
	Description string      `json:"description"`
	Steps       []ChainStep `json:"steps" binding:"required"`
	Status      string      `json:"status"` // pending, running, completed, failed, stopped
	StartedAt   *time.Time  `json:"started_at,omitempty"`
	CompletedAt *time.Time  `json:"completed_at,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	CreatedBy   string      `json:"created_by,omitempty"`
}

// ChainExecution представляет выполнение цепочки
type ChainExecution struct {
	ID          string               `json:"id"`
	ChainID     string               `json:"chain_id"`
	Status      string               `json:"status"` // running, completed, failed, stopped
	StartedAt   time.Time            `json:"started_at"`
	CompletedAt *time.Time           `json:"completed_at,omitempty"`
	Error       string               `json:"error,omitempty"`
	Steps       []ChainExecutionStep `json:"steps,omitempty"`
}

// ChainExecutionStep представляет выполнение шага цепочки
type ChainExecutionStep struct {
	StepIndex    int        `json:"step_index"`
	ScenarioType string     `json:"scenario_type"`
	Status       string     `json:"status"` // pending, running, completed, failed
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	Error        string     `json:"error,omitempty"`
}
