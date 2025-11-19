package storage

import (
	"sync"

	"log-metrics-simulator/models"
)

type MemoryStorage struct {
	scenarios       map[string]*models.Scenario
	schedules       map[string]*models.Schedule
	executions      map[string]*models.ScheduleExecution
	chains          map[string]*models.ScenarioChain  // Новое: хранилище цепочек
	chainExecutions map[string]*models.ChainExecution // Новое: хранилище выполнений цепочек
	chainSchedules  map[string]*models.ChainSchedule  // Новое: хранилище расписаний цепочек
	scenarioMutex   sync.RWMutex
	scheduleMutex   sync.RWMutex
	executionMutex  sync.RWMutex
	chainMutex      sync.RWMutex // Новое: мьютекс для цепочек
	chainExecMutex  sync.RWMutex // Новое: мьютекс для выполнений цепочек
	chainSchedMutex sync.RWMutex // Новое: мьютекс для расписаний цепочек
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		scenarios:       make(map[string]*models.Scenario),
		schedules:       make(map[string]*models.Schedule),
		executions:      make(map[string]*models.ScheduleExecution),
		chains:          make(map[string]*models.ScenarioChain),  // Инициализация
		chainExecutions: make(map[string]*models.ChainExecution), // Инициализация
		chainSchedules:  make(map[string]*models.ChainSchedule),  // Инициализация расписаний цепочек
	}
}

func (m *MemoryStorage) SaveScenario(scenario *models.Scenario) error {
	m.scenarioMutex.Lock()
	defer m.scenarioMutex.Unlock()

	m.scenarios[scenario.Type] = scenario
	return nil
}

func (m *MemoryStorage) GetActiveScenarios() ([]*models.Scenario, error) {
	m.scenarioMutex.RLock()
	defer m.scenarioMutex.RUnlock()

	var active []*models.Scenario
	for _, scenario := range m.scenarios {
		if scenario.Active {
			active = append(active, scenario)
		}
	}
	return active, nil
}

func (m *MemoryStorage) UpdateScenario(scenario *models.Scenario) error {
	m.scenarioMutex.Lock()
	defer m.scenarioMutex.Unlock()

	m.scenarios[scenario.Type] = scenario
	return nil
}

func (m *MemoryStorage) DeleteScenario(scenarioType string) error {
	m.scenarioMutex.Lock()
	defer m.scenarioMutex.Unlock()

	delete(m.scenarios, scenarioType)
	return nil
}

func (m *MemoryStorage) SaveSchedule(schedule *models.Schedule) error {
	m.scheduleMutex.Lock()
	defer m.scheduleMutex.Unlock()

	m.schedules[schedule.ID] = schedule
	return nil
}

func (m *MemoryStorage) GetSchedules() ([]*models.Schedule, error) {
	m.scheduleMutex.RLock()
	defer m.scheduleMutex.RUnlock()

	schedules := make([]*models.Schedule, 0, len(m.schedules))
	for _, schedule := range m.schedules {
		schedules = append(schedules, schedule)
	}
	return schedules, nil
}

func (m *MemoryStorage) GetSchedule(id string) (*models.Schedule, error) {
	m.scheduleMutex.RLock()
	defer m.scheduleMutex.RUnlock()

	schedule, exists := m.schedules[id]
	if !exists {
		return nil, nil
	}
	return schedule, nil
}

func (m *MemoryStorage) UpdateSchedule(schedule *models.Schedule) error {
	m.scheduleMutex.Lock()
	defer m.scheduleMutex.Unlock()

	m.schedules[schedule.ID] = schedule
	return nil
}

func (m *MemoryStorage) DeleteSchedule(id string) error {
	m.scheduleMutex.Lock()
	defer m.scheduleMutex.Unlock()

	delete(m.schedules, id)
	return nil
}

func (m *MemoryStorage) SaveExecution(execution *models.ScheduleExecution) error {
	m.executionMutex.Lock()
	defer m.executionMutex.Unlock()

	m.executions[execution.ID] = execution
	return nil
}

func (m *MemoryStorage) GetExecutions(scheduleID string, limit int) ([]*models.ScheduleExecution, error) {
	m.executionMutex.RLock()
	defer m.executionMutex.RUnlock()

	var executions []*models.ScheduleExecution
	for _, exec := range m.executions {
		if exec.ScheduleID == scheduleID {
			executions = append(executions, exec)
		}
	}

	if len(executions) > limit {
		executions = executions[:limit]
	}

	return executions, nil
}

func (m *MemoryStorage) Close() error {
	return nil
}

func (m *MemoryStorage) SaveChain(chain *models.ScenarioChain) error {
	m.chainMutex.Lock()
	defer m.chainMutex.Unlock()

	m.chains[chain.ID] = chain
	return nil
}

func (m *MemoryStorage) GetChains() ([]*models.ScenarioChain, error) {
	m.chainMutex.RLock()
	defer m.chainMutex.RUnlock()

	chains := make([]*models.ScenarioChain, 0, len(m.chains))
	for _, chain := range m.chains {
		chains = append(chains, chain)
	}
	return chains, nil
}

func (m *MemoryStorage) GetChain(id string) (*models.ScenarioChain, error) {
	m.chainMutex.RLock()
	defer m.chainMutex.RUnlock()

	chain, exists := m.chains[id]
	if !exists {
		return nil, nil
	}
	return chain, nil
}

func (m *MemoryStorage) UpdateChain(chain *models.ScenarioChain) error {
	m.chainMutex.Lock()
	defer m.chainMutex.Unlock()

	m.chains[chain.ID] = chain
	return nil
}

func (m *MemoryStorage) DeleteChain(id string) error {
	m.chainMutex.Lock()
	defer m.chainMutex.Unlock()

	delete(m.chains, id)
	return nil
}

func (m *MemoryStorage) SaveChainExecution(execution *models.ChainExecution) error {
	m.chainExecMutex.Lock()
	defer m.chainExecMutex.Unlock()

	m.chainExecutions[execution.ID] = execution
	return nil
}

func (m *MemoryStorage) GetChainExecutions(chainID string, limit int) ([]*models.ChainExecution, error) {
	m.chainExecMutex.RLock()
	defer m.chainExecMutex.RUnlock()

	var executions []*models.ChainExecution
	for _, exec := range m.chainExecutions {
		if exec.ChainID == chainID {
			executions = append(executions, exec)
		}
	}

	// Сортируем по времени (новые сначала) и ограничиваем
	if len(executions) > limit {
		executions = executions[:limit]
	}

	return executions, nil
}

func (m *MemoryStorage) GetChainExecution(id string) (*models.ChainExecution, error) {
	m.chainExecMutex.RLock()
	defer m.chainExecMutex.RUnlock()

	execution, exists := m.chainExecutions[id]
	if !exists {
		return nil, nil
	}
	return execution, nil
}

func (m *MemoryStorage) UpdateChainExecution(execution *models.ChainExecution) error {
	m.chainExecMutex.Lock()
	defer m.chainExecMutex.Unlock()

	m.chainExecutions[execution.ID] = execution
	return nil
}

// Методы для работы с расписаниями цепочек

func (m *MemoryStorage) SaveChainSchedule(schedule *models.ChainSchedule) error {
	m.chainSchedMutex.Lock()
	defer m.chainSchedMutex.Unlock()

	m.chainSchedules[schedule.ID] = schedule
	return nil
}

func (m *MemoryStorage) GetChainSchedules() ([]*models.ChainSchedule, error) {
	m.chainSchedMutex.RLock()
	defer m.chainSchedMutex.RUnlock()

	schedules := make([]*models.ChainSchedule, 0, len(m.chainSchedules))
	for _, schedule := range m.chainSchedules {
		schedules = append(schedules, schedule)
	}
	return schedules, nil
}

func (m *MemoryStorage) GetChainSchedule(id string) (*models.ChainSchedule, error) {
	m.chainSchedMutex.RLock()
	defer m.chainSchedMutex.RUnlock()

	schedule, exists := m.chainSchedules[id]
	if !exists {
		return nil, nil
	}
	return schedule, nil
}

func (m *MemoryStorage) UpdateChainSchedule(schedule *models.ChainSchedule) error {
	m.chainSchedMutex.Lock()
	defer m.chainSchedMutex.Unlock()

	m.chainSchedules[schedule.ID] = schedule
	return nil
}

func (m *MemoryStorage) DeleteChainSchedule(id string) error {
	m.chainSchedMutex.Lock()
	defer m.chainSchedMutex.Unlock()

	delete(m.chainSchedules, id)
	return nil
}
