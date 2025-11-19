package scenarios

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"log-metrics-simulator/storage"

	"github.com/robfig/cron/v3"

	"log-metrics-simulator/generator"
	"log-metrics-simulator/models"
)

type ScenarioManager struct {
	storage          storage.Storage
	activeScenarios  map[string]*models.Scenario
	schedules        map[string]*models.Schedule
	activeChains     map[string]*models.ChainExecution // Активные выполнения цепочек
	cronScheduler    *cron.Cron
	cronEntries      map[string]cron.EntryID
	chainSchedules   map[string]*models.ChainSchedule
	chainCronEntries map[string]cron.EntryID
	mutex            sync.RWMutex
	stopChan         chan struct{}
}

func NewScenarioManager(storage storage.Storage) *ScenarioManager {
	c := cron.New(cron.WithSeconds())

	sm := &ScenarioManager{
		storage:          storage,
		activeScenarios:  make(map[string]*models.Scenario),
		schedules:        make(map[string]*models.Schedule),
		activeChains:     make(map[string]*models.ChainExecution),
		cronScheduler:    c,
		cronEntries:      make(map[string]cron.EntryID),
		chainSchedules:   make(map[string]*models.ChainSchedule),
		chainCronEntries: make(map[string]cron.EntryID),
		stopChan:         make(chan struct{}),
	}

	sm.restoreState()
	sm.cronScheduler.Start()

	log.Println("⏰ Scenario manager запущен с поддержкой цепочек")
	return sm
}

// Предопределенные сценарии
var predefinedScenarios = map[string]models.ScenarioConfig{
	"load_test": {
		Name:        "Load Test",
		Description: "Генерация высокой нагрузки",
		LogCount:    1000,
		Labels:      map[string]string{"test_type": "load", "environment": "testing"},
		Parameters:  map[string]interface{}{"interval_ms": 10},
	},
	"error_spike": {
		Name:        "Error Spike",
		Description: "Всплеск ошибок в системе",
		LogCount:    200,
		Labels:      map[string]string{"test_type": "errors", "environment": "testing"},
		Parameters:  map[string]interface{}{"error_rate": 0.5},
	},
	"slow_responses": {
		Name:        "Slow Responses",
		Description: "Имитация медленных ответов",
		LogCount:    500,
		Labels:      map[string]string{"test_type": "performance", "environment": "testing"},
		Parameters:  map[string]interface{}{"response_delay": 2000},
	},
	"normal_operation": {
		Name:        "Normal Operation",
		Description: "Нормальная работа системы",
		LogCount:    300,
		Labels:      map[string]string{"environment": "production"},
		Parameters:  map[string]interface{}{"error_rate": 0.05},
	},
	"continuous_load": {
		Name:        "Continuous Load",
		Description: "Постоянная нагрузка",
		LogCount:    100,
		Labels:      map[string]string{"test_type": "continuous", "environment": "testing"},
		Parameters:  map[string]interface{}{"interval_seconds": 5},
	},
}

// Предопределенные цепочки сценариев
var predefinedChains = map[string]models.Chain{
	"black_friday_rush": {
		Name:        "black_friday_rush",
		Description: "Комбинация высокой нагрузки и всплеска ошибок",
		Steps:       []string{"load_test", "error_spike"},
	},
	"slow_and_steady": {
		Name:        "slow_and_steady",
		Description: "Медленные ответы при нормальной нагрузке",
		Steps:       []string{"normal_operation", "slow_responses"},
	},
}

// ===== Базовые методы =====

func (sm *ScenarioManager) GetAvailableScenarios() map[string]models.ScenarioConfig {
	return predefinedScenarios
}

func (sm *ScenarioManager) GetAvailableChains() map[string]models.Chain {
	return predefinedChains
}

func (sm *ScenarioManager) GetActiveScenarios() []*models.Scenario {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	scenarios := make([]*models.Scenario, 0, len(sm.activeScenarios))
	for _, scenario := range sm.activeScenarios {
		scenarios = append(scenarios, scenario)
	}
	return scenarios
}

func (sm *ScenarioManager) Stop() {
	sm.cronScheduler.Stop()
	close(sm.stopChan)

	if err := sm.storage.Close(); err != nil {
		log.Printf("❌ Ошибка закрытия хранилища: %v", err)
	}

	log.Println("🛑 Scenario manager остановлен")
}

// ===== Методы для работы со сценариями =====

func (sm *ScenarioManager) StartScenario(scenarioType string, customConfig map[string]interface{}) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	config, exists := predefinedScenarios[scenarioType]
	if !exists {
		return fmt.Errorf("сценарий не найден: %s", scenarioType)
	}

	// Создаем копию конфигурации
	scenarioConfig := models.ScenarioConfig{
		Name:        config.Name,
		Description: config.Description,
		LogCount:    config.LogCount,
		Labels:      make(map[string]string),
		Parameters:  make(map[string]interface{}),
	}

	for k, v := range config.Labels {
		scenarioConfig.Labels[k] = v
	}
	for k, v := range config.Parameters {
		scenarioConfig.Parameters[k] = v
	}

	// Применяем кастомную конфигурацию
	duration := 0 * time.Second
	interval := 0 * time.Second
	var startDate, endDate *time.Time

	if customConfig != nil {
		if logCount, ok := customConfig["log_count"].(float64); ok {
			scenarioConfig.LogCount = int(logCount)
		}
		if labels, ok := customConfig["labels"].(map[string]interface{}); ok {
			for k, v := range labels {
				if strVal, ok := v.(string); ok {
					scenarioConfig.Labels[k] = strVal
				}
			}
		}
		if dur, ok := customConfig["duration_minutes"].(float64); ok {
			duration = time.Duration(dur) * time.Minute
		}
		if dur, ok := customConfig["duration_seconds"].(float64); ok {
			duration = time.Duration(dur) * time.Second
		}
		if interv, ok := customConfig["interval_seconds"].(float64); ok {
			interval = time.Duration(interv) * time.Second
		}
		if interv, ok := customConfig["interval_minutes"].(float64); ok {
			interval = time.Duration(interv) * time.Minute
		}
		if start, ok := customConfig["start_date"].(string); ok {
			if parsedStart, err := time.Parse(time.RFC3339, start); err == nil {
				startDate = &parsedStart
			}
		}
		if end, ok := customConfig["end_date"].(string); ok {
			if parsedEnd, err := time.Parse(time.RFC3339, end); err == nil {
				endDate = &parsedEnd
			}
		}
	}

	scenario := &models.Scenario{
		Type:      scenarioType,
		Active:    true,
		Config:    scenarioConfig,
		Started:   time.Now(),
		Duration:  duration,
		Interval:  interval,
		StartDate: startDate,
		EndDate:   endDate,
	}

	sm.activeScenarios[scenarioType] = scenario

	if err := sm.storage.SaveScenario(scenario); err != nil {
		log.Printf("❌ Ошибка сохранения сценария: %v", err)
	}

	go sm.executeScenario(scenario)

	return nil
}

func (sm *ScenarioManager) StopScenario(scenarioType string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	scenario, exists := sm.activeScenarios[scenarioType]
	if !exists {
		return fmt.Errorf("сценарий не активен: %s", scenarioType)
	}

	scenario.Active = false

	if err := sm.storage.UpdateScenario(scenario); err != nil {
		log.Printf("❌ Ошибка обновления сценария: %v", err)
	}

	log.Printf("⏹️ Остановлен сценарий: %s", scenario.Config.Name)
	return nil
}

func (sm *ScenarioManager) executeScenario(scenario *models.Scenario) {
	config := scenario.Config

	log.Printf("🔧 Выполнение сценария %s", config.Name)

	// Проверяем дату начала
	if scenario.StartDate != nil {
		now := time.Now()
		if now.Before(*scenario.StartDate) {
			waitTime := scenario.StartDate.Sub(now)
			log.Printf("⏰ Ожидание до даты начала: %v (осталось: %v)",
				scenario.StartDate.Format("2006-01-02 15:04:05"), waitTime)

			select {
			case <-time.After(waitTime):
				if !scenario.Active {
					return
				}
			case <-sm.stopChan:
				return
			}
		}
	}

	// Определяем режим выполнения
	if scenario.Interval > 0 {
		sm.executePeriodicScenario(scenario)
	} else if scenario.Duration > 0 {
		sm.executeTimedScenario(scenario)
	} else {
		sm.executeSingleScenario(scenario)
	}

	sm.mutex.Lock()
	scenario.Active = false

	if err := sm.storage.UpdateScenario(scenario); err != nil {
		log.Printf("❌ Ошибка обновления сценария: %v", err)
	}

	delete(sm.activeScenarios, scenario.Type)

	if err := sm.storage.DeleteScenario(scenario.Type); err != nil {
		log.Printf("❌ Ошибка удаления сценария: %v", err)
	}

	sm.mutex.Unlock()

	log.Printf("✅ Завершен сценарий: %s", config.Name)
}

func (sm *ScenarioManager) executeSingleScenario(scenario *models.Scenario) {
	generator.GenerateLogs(scenario.Config.LogCount, scenario.Config.Name)
}

func (sm *ScenarioManager) executeTimedScenario(scenario *models.Scenario) {
	var endTime time.Time
	if scenario.EndDate != nil {
		endTime = *scenario.EndDate
	} else if scenario.Duration > 0 {
		endTime = time.Now().Add(scenario.Duration)
	} else {
		endTime = time.Now().Add(365 * 24 * time.Hour)
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !scenario.Active || time.Now().After(endTime) {
				if time.Now().After(endTime) {
					log.Printf("⏰ Достигнуто время окончания сценария: %v",
						endTime.Format("2006-01-02 15:04:05"))
				}
				return
			}

			timeUntilEnd := time.Until(endTime).Seconds()
			if timeUntilEnd <= 0 {
				return
			}

			batchSize := scenario.Config.LogCount / int(timeUntilEnd/10)
			if batchSize < 1 {
				batchSize = 1
			}
			generator.GenerateLogs(batchSize, scenario.Config.Name)
		case <-sm.stopChan:
			return
		}
	}
}

func (sm *ScenarioManager) executePeriodicScenario(scenario *models.Scenario) {
	ticker := time.NewTicker(scenario.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !scenario.Active {
				return
			}

			if scenario.EndDate != nil && time.Now().After(*scenario.EndDate) {
				log.Printf("⏰ Достигнута дата окончания сценария: %v",
					scenario.EndDate.Format("2006-01-02 15:04:05"))
				return
			}

			generator.GenerateLogs(scenario.Config.LogCount, scenario.Config.Name)
		case <-sm.stopChan:
			return
		}
	}
}

// ===== Методы для работы с расписаниями сценариев =====

func (sm *ScenarioManager) CreateSchedule(schedule *models.Schedule) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if schedule.ID == "" {
		schedule.ID = generateID()
	}

	if _, err := cron.ParseStandard(schedule.CronExpr); err != nil {
		return fmt.Errorf("неверное cron выражение: %v", err)
	}

	if schedule.StartDate != nil && schedule.EndDate != nil {
		if schedule.EndDate.Before(*schedule.StartDate) {
			return fmt.Errorf("дата окончания не может быть раньше даты начала")
		}
	}

	schedule.CreatedAt = time.Now()
	sm.schedules[schedule.ID] = schedule

	if err := sm.storage.SaveSchedule(schedule); err != nil {
		return fmt.Errorf("ошибка сохранения расписания: %v", err)
	}

	if schedule.Enabled {
		if err := sm.scheduleCronJob(schedule); err != nil {
			return err
		}
	}

	log.Printf("📅 Создано расписание: %s", schedule.Name)
	return nil
}

func (sm *ScenarioManager) scheduleCronJob(schedule *models.Schedule) error {
	now := time.Now()
	if schedule.StartDate != nil && now.Before(*schedule.StartDate) {
		log.Printf("⏰ Расписание %s начнет действовать с: %v",
			schedule.Name, schedule.StartDate.Format("2006-01-02 15:04:05"))
	}

	if schedule.EndDate != nil && now.After(*schedule.EndDate) {
		log.Printf("⏰ Расписание %s закончило действие: %v",
			schedule.Name, schedule.EndDate.Format("2006-01-02 15:04:05"))
		schedule.Enabled = false
		return nil
	}

	if entryID, exists := sm.cronEntries[schedule.ID]; exists {
		sm.cronScheduler.Remove(entryID)
		delete(sm.cronEntries, schedule.ID)
	}

	entryID, err := sm.cronScheduler.AddFunc(schedule.CronExpr, func() {
		sm.executeScheduledScenario(schedule)
	})

	if err != nil {
		return fmt.Errorf("ошибка добавления в cron: %v", err)
	}

	sm.cronEntries[schedule.ID] = entryID

	nextRun := sm.cronScheduler.Entry(entryID).Next
	schedule.NextRun = &nextRun

	log.Printf("⏰ Расписание %s добавлено в cron. Следующий запуск: %v",
		schedule.Name, schedule.NextRun.Format("2006-01-02 15:04:05"))

	return nil
}

func (sm *ScenarioManager) executeScheduledScenario(schedule *models.Schedule) {
	now := time.Now()

	if schedule.StartDate != nil && now.Before(*schedule.StartDate) {
		return
	}

	if schedule.EndDate != nil && now.After(*schedule.EndDate) {
		sm.mutex.Lock()
		schedule.Enabled = false

		if err := sm.storage.UpdateSchedule(schedule); err != nil {
			log.Printf("❌ Ошибка обновления расписания: %v", err)
		}

		if entryID, exists := sm.cronEntries[schedule.ID]; exists {
			sm.cronScheduler.Remove(entryID)
			delete(sm.cronEntries, schedule.ID)
		}
		sm.mutex.Unlock()

		log.Printf("⏰ Расписание %s автоматически отключено", schedule.Name)
		return
	}

	execution := &models.ScheduleExecution{
		ID:           generateID(),
		ScheduleID:   schedule.ID,
		ScenarioType: schedule.ScenarioType,
		Status:       "running",
		StartedAt:    now,
	}

	if err := sm.storage.SaveExecution(execution); err != nil {
		log.Printf("❌ Ошибка сохранения выполнения: %v", err)
	}

	log.Printf("⏰ Запуск по расписанию: %s -> %s", schedule.Name, schedule.ScenarioType)

	if err := sm.StartScenario(schedule.ScenarioType, nil); err != nil {
		log.Printf("❌ Ошибка выполнения расписания %s: %v", schedule.Name, err)

		execution.Status = "failed"
		execution.Error = err.Error()
		completedAt := time.Now()
		execution.CompletedAt = &completedAt

		if err := sm.storage.SaveExecution(execution); err != nil {
			log.Printf("❌ Ошибка обновления выполнения: %v", err)
		}
		return
	}

	execution.Status = "completed"
	completedAt := time.Now()
	execution.CompletedAt = &completedAt

	if err := sm.storage.SaveExecution(execution); err != nil {
		log.Printf("❌ Ошибка обновления выполнения: %v", err)
	}

	sm.mutex.Lock()
	lastRun := time.Now()
	schedule.LastRun = &lastRun

	if entryID, exists := sm.cronEntries[schedule.ID]; exists {
		nextRun := sm.cronScheduler.Entry(entryID).Next
		schedule.NextRun = &nextRun
	}

	if err := sm.storage.UpdateSchedule(schedule); err != nil {
		log.Printf("❌ Ошибка обновления расписания: %v", err)
	}

	sm.mutex.Unlock()

	log.Printf("✅ Успешно выполнено расписание: %s", schedule.Name)
}

func (sm *ScenarioManager) UpdateSchedule(scheduleID string, updates map[string]interface{}) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	schedule, exists := sm.schedules[scheduleID]
	if !exists {
		return fmt.Errorf("расписание не найдено: %s", scheduleID)
	}

	if name, ok := updates["name"].(string); ok {
		schedule.Name = name
	}
	if cronExpr, ok := updates["cron_expr"].(string); ok {
		if _, err := cron.ParseStandard(cronExpr); err != nil {
			return fmt.Errorf("неверное cron выражение: %v", err)
		}
		schedule.CronExpr = cronExpr
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		schedule.Enabled = enabled
	}
	if startDate, ok := updates["start_date"].(string); ok {
		if parsedStart, err := time.Parse(time.RFC3339, startDate); err == nil {
			schedule.StartDate = &parsedStart
		}
	}
	if endDate, ok := updates["end_date"].(string); ok {
		if parsedEnd, err := time.Parse(time.RFC3339, endDate); err == nil {
			schedule.EndDate = &parsedEnd
		}
	}

	if schedule.StartDate != nil && schedule.EndDate != nil {
		if schedule.EndDate.Before(*schedule.StartDate) {
			return fmt.Errorf("дата окончания не может быть раньше даты начала")
		}
	}

	if schedule.Enabled {
		if err := sm.scheduleCronJob(schedule); err != nil {
			return err
		}
	} else {
		if entryID, exists := sm.cronEntries[scheduleID]; exists {
			sm.cronScheduler.Remove(entryID)
			delete(sm.cronEntries, scheduleID)
			schedule.NextRun = nil
		}
	}

	log.Printf("✏️ Обновлено расписание: %s", schedule.Name)
	return nil
}

func (sm *ScenarioManager) GetSchedules() []*models.Schedule {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	schedules := make([]*models.Schedule, 0, len(sm.schedules))
	for _, schedule := range sm.schedules {
		schedules = append(schedules, schedule)
	}
	return schedules
}

func (sm *ScenarioManager) GetSchedule(scheduleID string) (*models.Schedule, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	schedule, exists := sm.schedules[scheduleID]
	return schedule, exists
}

func (sm *ScenarioManager) DeleteSchedule(scheduleID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	_, exists := sm.schedules[scheduleID]
	if !exists {
		return fmt.Errorf("расписание не найдено: %s", scheduleID)
	}

	if entryID, exists := sm.cronEntries[scheduleID]; exists {
		sm.cronScheduler.Remove(entryID)
		delete(sm.cronEntries, scheduleID)
	}

	delete(sm.schedules, scheduleID)
	return nil
}

func (sm *ScenarioManager) EnableSchedule(scheduleID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	schedule, exists := sm.schedules[scheduleID]
	if !exists {
		return fmt.Errorf("расписание не найдено: %s", scheduleID)
	}

	if schedule.Enabled {
		return fmt.Errorf("расписание уже активно")
	}

	schedule.Enabled = true
	if err := sm.scheduleCronJob(schedule); err != nil {
		schedule.Enabled = false
		return err
	}

	return nil
}

func (sm *ScenarioManager) DisableSchedule(scheduleID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	schedule, exists := sm.schedules[scheduleID]
	if !exists {
		return fmt.Errorf("расписание не найдено: %s", scheduleID)
	}

	if !schedule.Enabled {
		return fmt.Errorf("расписание уже отключено")
	}

	schedule.Enabled = false
	if entryID, exists := sm.cronEntries[scheduleID]; exists {
		sm.cronScheduler.Remove(entryID)
		delete(sm.cronEntries, scheduleID)
		schedule.NextRun = nil
	}

	return nil
}

func (sm *ScenarioManager) GetExecutions(scheduleID string, limit int) ([]*models.ScheduleExecution, error) {
	return sm.storage.GetExecutions(scheduleID, limit)
}

// ===== Методы для работы с цепочками сценариев =====

func (sm *ScenarioManager) CreateChain(chain *models.ScenarioChain) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if chain.ID == "" {
		chain.ID = generateID()
	}

	chain.CreatedAt = time.Now()
	chain.Status = "pending"

	if err := sm.storage.SaveChain(chain); err != nil {
		return fmt.Errorf("ошибка сохранения цепочки: %v", err)
	}

	log.Printf("🔗 Создана цепочка: %s (%d шагов)", chain.Name, len(chain.Steps))
	return nil
}

func (sm *ScenarioManager) GetChains() ([]*models.ScenarioChain, error) {
	return sm.storage.GetChains()
}

func (sm *ScenarioManager) GetChain(chainID string) (*models.ScenarioChain, error) {
	return sm.storage.GetChain(chainID)
}

func (sm *ScenarioManager) DeleteChain(chainID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.activeChains[chainID]; exists {
		return fmt.Errorf("невозможно удалить выполняющуюся цепочку")
	}

	if err := sm.storage.DeleteChain(chainID); err != nil {
		return err
	}

	log.Printf("🗑️ Удалена цепочка: %s", chainID)
	return nil
}

func (sm *ScenarioManager) StartChain(chainID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	chain, err := sm.storage.GetChain(chainID)
	if err != nil {
		return fmt.Errorf("ошибка получения цепочки: %v", err)
	}
	if chain == nil {
		return fmt.Errorf("цепочка не найдена: %s", chainID)
	}

	execution := &models.ChainExecution{
		ID:        generateID(),
		ChainID:   chainID,
		Status:    "running",
		StartedAt: time.Now(),
		Steps:     make([]models.ChainExecutionStep, len(chain.Steps)),
	}

	for i, step := range chain.Steps {
		execution.Steps[i] = models.ChainExecutionStep{
			StepIndex:    i,
			ScenarioType: step.ScenarioType,
			Status:       "pending",
		}
	}

	if err := sm.storage.SaveChainExecution(execution); err != nil {
		return fmt.Errorf("ошибка сохранения выполнения: %v", err)
	}

	sm.activeChains[execution.ID] = execution

	go sm.executeChain(chain, execution)

	log.Printf("🎬 Запущена цепочка: %s (%d шагов)", chain.Name, len(chain.Steps))
	return nil
}

func (sm *ScenarioManager) StopChain(chainExecutionID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	execution, exists := sm.activeChains[chainExecutionID]
	if !exists {
		return fmt.Errorf("выполнение цепочки не найдено: %s", chainExecutionID)
	}

	execution.Status = "stopped"
	completedAt := time.Now()
	execution.CompletedAt = &completedAt

	if err := sm.storage.UpdateChainExecution(execution); err != nil {
		log.Printf("❌ Ошибка обновления выполнения цепочки: %v", err)
	}

	delete(sm.activeChains, chainExecutionID)

	log.Printf("⏹️ Остановлена цепочка: %s", chainExecutionID)
	return nil
}

func (sm *ScenarioManager) executeChain(chain *models.ScenarioChain, execution *models.ChainExecution) {
	defer func() {
		sm.mutex.Lock()
		if execution.Status == "running" {
			execution.Status = "completed"
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
		}

		if err := sm.storage.UpdateChainExecution(execution); err != nil {
			log.Printf("❌ Ошибка обновления выполнения цепочки: %v", err)
		}

		delete(sm.activeChains, execution.ID)
		sm.mutex.Unlock()

		log.Printf("✅ Завершена цепочка: %s", chain.Name)
	}()

	for i, step := range chain.Steps {
		sm.mutex.RLock()
		if execution.Status != "running" {
			sm.mutex.RUnlock()
			return
		}
		sm.mutex.RUnlock()

		sm.mutex.Lock()
		execution.Steps[i].Status = "running"
		startedAt := time.Now()
		execution.Steps[i].StartedAt = &startedAt
		sm.mutex.Unlock()

		if err := sm.storage.UpdateChainExecution(execution); err != nil {
			log.Printf("❌ Ошибка обновления шага выполнения: %v", err)
		}

		log.Printf("🔧 Выполнение шага %d/%d: %s", i+1, len(chain.Steps), step.Name)

		if step.DelayBefore > 0 {
			log.Printf("⏰ Задержка перед шагом %d: %d секунд", i+1, step.DelayBefore)
			select {
			case <-time.After(time.Duration(step.DelayBefore) * time.Second):
			case <-sm.stopChan:
				return
			}
		}

		if err := sm.StartScenario(step.ScenarioType, step.Config); err != nil {
			log.Printf("❌ Ошибка выполнения шага %d: %v", i+1, err)

			sm.mutex.Lock()
			execution.Steps[i].Status = "failed"
			execution.Steps[i].Error = err.Error()
			execution.Status = "failed"
			execution.Error = fmt.Sprintf("Ошибка на шаге %d: %v", i+1, err)
			sm.mutex.Unlock()

			return
		}

		if step.Config != nil {
			if duration, ok := getDurationFromConfig(step.Config); ok && duration > 0 {
				log.Printf("⏰ Ожидание завершения шага %d: %v", i+1, duration)
				select {
				case <-time.After(duration):
				case <-sm.stopChan:
					return
				}
			}
		}

		sm.mutex.Lock()
		execution.Steps[i].Status = "completed"
		completedAt := time.Now()
		execution.Steps[i].CompletedAt = &completedAt
		sm.mutex.Unlock()

		if err := sm.storage.UpdateChainExecution(execution); err != nil {
			log.Printf("❌ Ошибка обновления шага выполнения: %v", err)
		}

		log.Printf("✅ Завершен шаг %d/%d: %s", i+1, len(chain.Steps), step.Name)
	}
}

func (sm *ScenarioManager) GetChainExecutions(chainID string, limit int) ([]*models.ChainExecution, error) {
	return sm.storage.GetChainExecutions(chainID, limit)
}

// ===== Методы для работы с расписаниями цепочек =====

func (sm *ScenarioManager) CreateChainSchedule(schedule *models.ChainSchedule) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if schedule.ID == "" {
		schedule.ID = generateID()
	}

	if _, exists := predefinedChains[schedule.ChainName]; !exists {
		return fmt.Errorf("цепочка не найдена: %s", schedule.ChainName)
	}

	if _, err := cron.ParseStandard(schedule.CronExpr); err != nil {
		return fmt.Errorf("неверное cron выражение: %v", err)
	}

	if schedule.StartDate != nil && schedule.EndDate != nil {
		if schedule.EndDate.Before(*schedule.StartDate) {
			return fmt.Errorf("дата окончания не может быть раньше даты начала")
		}
	}

	schedule.CreatedAt = time.Now()
	sm.chainSchedules[schedule.ID] = schedule

	if err := sm.storage.SaveChainSchedule(schedule); err != nil {
		return fmt.Errorf("ошибка сохранения расписания цепочки: %v", err)
	}

	if schedule.Enabled {
		if err := sm.scheduleChainCronJob(schedule); err != nil {
			return err
		}
	}

	log.Printf("📅 Создано расписание цепочки: %s", schedule.Name)
	return nil
}

func (sm *ScenarioManager) scheduleChainCronJob(schedule *models.ChainSchedule) error {
	now := time.Now()
	if schedule.StartDate != nil && now.Before(*schedule.StartDate) {
		log.Printf("⏰ Расписание цепочки %s начнет действовать с: %v", schedule.Name, schedule.StartDate.Format("2006-01-02 15:04:05"))
	}
	if schedule.EndDate != nil && now.After(*schedule.EndDate) {
		log.Printf("⏰ Расписание цепочки %s закончилось: %v", schedule.Name, schedule.EndDate.Format("2006-01-02 15:04:05"))
		schedule.Enabled = false
		if err := sm.storage.UpdateChainSchedule(schedule); err != nil {
			log.Printf("❌ Ошибка обновления расписания цепочки: %v", err)
		}
		return nil
	}

	if entryID, exists := sm.chainCronEntries[schedule.ID]; exists {
		sm.cronScheduler.Remove(entryID)
		delete(sm.chainCronEntries, schedule.ID)
	}

	entryID, err := sm.cronScheduler.AddFunc(schedule.CronExpr, func() {
		sm.executeScheduledChain(schedule)
	})
	if err != nil {
		return fmt.Errorf("ошибка добавления cron для цепочки: %v", err)
	}
	sm.chainCronEntries[schedule.ID] = entryID

	nextRun := sm.cronScheduler.Entry(entryID).Next
	schedule.NextRun = &nextRun

	if err := sm.storage.UpdateChainSchedule(schedule); err != nil {
		log.Printf("❌ Ошибка обновления расписания цепочки: %v", err)
	}

	log.Printf("⏰ Расписание цепочки %s добавлено в cron. Следующий запуск: %v", schedule.Name, schedule.NextRun.Format("2006-01-02 15:04:05"))
	return nil
}

func (sm *ScenarioManager) executeScheduledChain(schedule *models.ChainSchedule) {
	now := time.Now()
	if schedule.StartDate != nil && now.Before(*schedule.StartDate) {
		return
	}
	if schedule.EndDate != nil && now.After(*schedule.EndDate) {
		sm.mutex.Lock()
		schedule.Enabled = false
		if entryID, exists := sm.chainCronEntries[schedule.ID]; exists {
			sm.cronScheduler.Remove(entryID)
			delete(sm.chainCronEntries, schedule.ID)
		}
		if err := sm.storage.UpdateChainSchedule(schedule); err != nil {
			log.Printf("❌ Ошибка обновления расписания цепочки: %v", err)
		}
		sm.mutex.Unlock()
		log.Printf("⏰ Расписание цепочки %s автоматически отключено", schedule.Name)
		return
	}

	chain, ok := predefinedChains[schedule.ChainName]
	if !ok {
		log.Printf("❌ Цепочка не найдена: %s", schedule.ChainName)
		return
	}

	log.Printf("⏰ Запуск цепочки по расписанию: %s -> %s", schedule.Name, chain.Name)
	for idx, st := range chain.Steps {
		if err := sm.StartScenario(st, nil); err != nil {
			log.Printf("❌ Ошибка запуска шага %d цепочки %s: %v", idx+1, chain.Name, err)
			break
		}
		time.Sleep(2 * time.Second)
	}

	sm.mutex.Lock()
	lastRun := time.Now()
	schedule.LastRun = &lastRun
	if entryID, exists := sm.chainCronEntries[schedule.ID]; exists {
		nextRun := sm.cronScheduler.Entry(entryID).Next
		schedule.NextRun = &nextRun
	}
	if err := sm.storage.UpdateChainSchedule(schedule); err != nil {
		log.Printf("❌ Ошибка обновления расписания цепочки: %v", err)
	}
	sm.mutex.Unlock()

	log.Printf("✅ Успешно выполнена цепочка: %s", chain.Name)
}

func (sm *ScenarioManager) ListChainSchedules() []*models.ChainSchedule {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	result := make([]*models.ChainSchedule, 0, len(sm.chainSchedules))
	for _, s := range sm.chainSchedules {
		result = append(result, s)
	}
	return result
}

func (sm *ScenarioManager) GetChainSchedules() ([]*models.ChainSchedule, error) {
	return sm.storage.GetChainSchedules()
}

func (sm *ScenarioManager) GetChainSchedule(id string) (*models.ChainSchedule, error) {
	return sm.storage.GetChainSchedule(id)
}

func (sm *ScenarioManager) UpdateChainSchedule(schedule *models.ChainSchedule) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.chainSchedules[schedule.ID] = schedule
	return sm.storage.UpdateChainSchedule(schedule)
}

func (sm *ScenarioManager) EnableChainSchedule(id string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	s, ok := sm.chainSchedules[id]
	if !ok {
		return fmt.Errorf("расписание цепочки не найдено: %s", id)
	}
	s.Enabled = true
	if err := sm.storage.UpdateChainSchedule(s); err != nil {
		return fmt.Errorf("ошибка обновления расписания цепочки: %v", err)
	}
	return sm.scheduleChainCronJob(s)
}

func (sm *ScenarioManager) DisableChainSchedule(id string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	s, ok := sm.chainSchedules[id]
	if !ok {
		return fmt.Errorf("расписание цепочки не найдено: %s", id)
	}
	s.Enabled = false
	if entryID, exists := sm.chainCronEntries[id]; exists {
		sm.cronScheduler.Remove(entryID)
		delete(sm.chainCronEntries, id)
	}
	if err := sm.storage.UpdateChainSchedule(s); err != nil {
		return fmt.Errorf("ошибка обновления расписания цепочки: %v", err)
	}
	return nil
}

func (sm *ScenarioManager) DeleteChainSchedule(id string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	if entryID, exists := sm.chainCronEntries[id]; exists {
		sm.cronScheduler.Remove(entryID)
		delete(sm.chainCronEntries, id)
	}
	delete(sm.chainSchedules, id)
	if err := sm.storage.DeleteChainSchedule(id); err != nil {
		return fmt.Errorf("ошибка удаления расписания цепочки: %v", err)
	}
	return nil
}

// ===== Вспомогательные методы =====

func (sm *ScenarioManager) restoreState() {
	activeScenarios, err := sm.storage.GetActiveScenarios()
	if err != nil {
		log.Printf("❌ Ошибка восстановления активных сценариев: %v", err)
		return
	}

	for _, scenario := range activeScenarios {
		sm.activeScenarios[scenario.Type] = scenario
		log.Printf("🔄 Восстановлен активный сценарий: %s", scenario.Config.Name)
		go sm.executeScenario(scenario)
	}

	schedules, err := sm.storage.GetSchedules()
	if err != nil {
		log.Printf("❌ Ошибка восстановления расписаний: %v", err)
		return
	}

	for _, schedule := range schedules {
		sm.schedules[schedule.ID] = schedule
		if schedule.Enabled {
			if err := sm.scheduleCronJob(schedule); err != nil {
				log.Printf("❌ Ошибка восстановления расписания %s: %v", schedule.Name, err)
			} else {
				log.Printf("🔄 Восстановлено расписание: %s", schedule.Name)
			}
		}
	}

	chainSchedules, err := sm.storage.GetChainSchedules()
	if err != nil {
		log.Printf("❌ Ошибка восстановления расписаний цепочек: %v", err)
		return
	}

	for _, schedule := range chainSchedules {
		sm.chainSchedules[schedule.ID] = schedule
		if schedule.Enabled {
			if err := sm.scheduleChainCronJob(schedule); err != nil {
				log.Printf("❌ Ошибка восстановления расписания цепочки %s: %v", schedule.Name, err)
			} else {
				log.Printf("🔄 Восстановлено расписание цепочки: %s", schedule.Name)
			}
		}
	}
}

func getDurationFromConfig(config map[string]interface{}) (time.Duration, bool) {
	if seconds, ok := config["duration_seconds"].(float64); ok && seconds > 0 {
		return time.Duration(seconds) * time.Second, true
	}
	if minutes, ok := config["duration_minutes"].(float64); ok && minutes > 0 {
		return time.Duration(minutes) * time.Minute, true
	}
	if hours, ok := config["duration_hours"].(float64); ok && hours > 0 {
		return time.Duration(hours) * time.Hour, true
	}
	return 0, false
}

func generateID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
