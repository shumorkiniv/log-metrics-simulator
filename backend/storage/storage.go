package storage

import "log-metrics-simulator/models"

// Storage интерфейс определяет все методы для работы с хранилищем данных
type Storage interface {
	// Основные методы
	Close() error

	// Методы для работы со сценариями
	SaveScenario(scenario *models.Scenario) error
	GetActiveScenarios() ([]*models.Scenario, error)
	UpdateScenario(scenario *models.Scenario) error
	DeleteScenario(scenarioType string) error

	// Методы для работы с расписаниями сценариев
	SaveSchedule(schedule *models.Schedule) error
	GetSchedules() ([]*models.Schedule, error)
	GetSchedule(id string) (*models.Schedule, error)
	UpdateSchedule(schedule *models.Schedule) error
	DeleteSchedule(id string) error

	// Методы для работы с выполнениями расписаний
	SaveExecution(execution *models.ScheduleExecution) error
	GetExecutions(scheduleID string, limit int) ([]*models.ScheduleExecution, error)

	// Методы для работы с цепочками сценариев
	SaveChain(chain *models.ScenarioChain) error
	GetChains() ([]*models.ScenarioChain, error)
	GetChain(id string) (*models.ScenarioChain, error)
	UpdateChain(chain *models.ScenarioChain) error
	DeleteChain(id string) error

	// Методы для работы с выполнениями цепочек
	SaveChainExecution(execution *models.ChainExecution) error
	GetChainExecutions(chainID string, limit int) ([]*models.ChainExecution, error)
	GetChainExecution(id string) (*models.ChainExecution, error)
	UpdateChainExecution(execution *models.ChainExecution) error

	// Методы для работы с расписаниями цепочек
	SaveChainSchedule(schedule *models.ChainSchedule) error
	GetChainSchedules() ([]*models.ChainSchedule, error)
	GetChainSchedule(id string) (*models.ChainSchedule, error)
	UpdateChainSchedule(schedule *models.ChainSchedule) error
	DeleteChainSchedule(id string) error
}
