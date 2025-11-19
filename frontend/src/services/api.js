import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || 'http://localhost:8080/api/v1';

const simulatorAPI = {
    // Основные методы
    generateLogs: (data) => axios.post(`${API_BASE_URL}/generate`, data),
    getMetrics: (format = 'json') => axios.get(`${API_BASE_URL}/metrics?format=${format}`),
    getLogs: (params) => axios.get(`${API_BASE_URL}/logs`, { params }),
    getLogStats: () => axios.get(`${API_BASE_URL}/logs/stats`),

    // Сценарии
    listScenarios: () => axios.get(`${API_BASE_URL}/scenarios/list`),
    startScenario: (data) => axios.post(`${API_BASE_URL}/scenarios/start`, data),
    stopScenario: (data) => axios.post(`${API_BASE_URL}/scenarios/stop`, data),

    // Расписания
    listSchedules: () => axios.get(`${API_BASE_URL}/schedules`),
    createSchedule: (data) => axios.post(`${API_BASE_URL}/schedules`, data),
    updateSchedule: (id, data) => axios.put(`${API_BASE_URL}/schedules/${id}`, data),
    deleteSchedule: (id) => axios.delete(`${API_BASE_URL}/schedules/${id}`),
    enableSchedule: (id) => axios.post(`${API_BASE_URL}/schedules/${id}/enable`),
    disableSchedule: (id) => axios.post(`${API_BASE_URL}/schedules/${id}/disable`),

    // Цепочки
    listChains: () => axios.get(`${API_BASE_URL}/chains`),
    getChain: (id) => axios.get(`${API_BASE_URL}/chains/${id}`),
    createChain: (data) => axios.post(`${API_BASE_URL}/chains`, data),
    startChain: (id) => axios.post(`${API_BASE_URL}/chains/${id}/start`),
    stopChain: (executionId) => axios.post(`${API_BASE_URL}/chains/${executionId}/stop`),
    deleteChain: (id) => axios.delete(`${API_BASE_URL}/chains/${id}`),
    getChainExecutions: (id, limit = 10) => axios.get(`${API_BASE_URL}/chains/${id}/executions?limit=${limit}`),

    // Расписания цепочек
    listChainSchedules: () => axios.get(`${API_BASE_URL}/chains/schedules`),
    getChainSchedule: (id) => axios.get(`${API_BASE_URL}/chains/schedules/${id}`),
    createChainSchedule: (data) => axios.post(`${API_BASE_URL}/chains/schedules`, data),
    updateChainSchedule: (id, data) => axios.put(`${API_BASE_URL}/chains/schedules/${id}`, data),
    enableChainSchedule: (id) => axios.post(`${API_BASE_URL}/chains/schedules/${id}/enable`),
    disableChainSchedule: (id) => axios.post(`${API_BASE_URL}/chains/schedules/${id}/disable`),
    deleteChainSchedule: (id) => axios.delete(`${API_BASE_URL}/chains/schedules/${id}`),
};

export { simulatorAPI };