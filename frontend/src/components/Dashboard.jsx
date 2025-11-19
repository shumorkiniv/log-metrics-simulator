import React, { useState, useEffect } from 'react'
import { Card, Row, Col, Statistic, Table, Tag, Space, Button, Alert } from 'antd'
import { PlayCircle, PauseCircle, BarChart3, Server, Clock, Users } from 'lucide-react'
import { simulatorAPI } from '../services/api'

const Dashboard = () => {
    const [stats, setStats] = useState({})
    const [recentLogs, setRecentLogs] = useState([])
    const [loading, setLoading] = useState(false)

    const loadData = async () => {
        setLoading(true)
        try {
            const [statsResponse, logsResponse] = await Promise.all([
                simulatorAPI.getLogStats(),
                simulatorAPI.getLogs({ limit: 10 })
            ])
            setStats(statsResponse.data.stats)
            setRecentLogs(logsResponse.data.logs)
        } catch (error) {
            console.error('Error loading dashboard data:', error)
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        loadData()
        const interval = setInterval(loadData, 10000) // Обновление каждые 10 секунд
        return () => clearInterval(interval)
    }, [])

    const logColumns = [
        {
            title: 'Время',
            dataIndex: 'timestamp',
            key: 'timestamp',
            render: (time) => new Date(time).toLocaleTimeString(),
            width: 100,
        },
        {
            title: 'Уровень',
            dataIndex: 'level',
            key: 'level',
            render: (level) => {
                const color = {
                    INFO: 'blue',
                    WARN: 'orange',
                    ERROR: 'red',
                    DEBUG: 'green',
                }[level]
                return <Tag color={color}>{level}</Tag>
            },
            width: 80,
        },
        {
            title: 'Сервис',
            dataIndex: 'service',
            key: 'service',
            width: 120,
        },
        {
            title: 'Сообщение',
            dataIndex: 'message',
            key: 'message',
            ellipsis: true,
        },
        {
            title: 'Статус',
            dataIndex: 'status',
            key: 'status',
            render: (status) => status && <Tag>{status}</Tag>,
            width: 80,
        },
    ]

    return (
        <div style={{ padding: '20px' }}>
            <h1 style={{ marginBottom: '24px' }}>📊 Metrics Simulator Dashboard</h1>

            <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
                <Col xs={24} sm={12} lg={6}>
                    <Card>
                        <Statistic
                            title="Всего логов"
                            value={stats.total_logs || 0}
                            prefix={<BarChart3 size={20} />}
                        />
                    </Card>
                </Col>
                <Col xs={24} sm={12} lg={6}>
                    <Card>
                        <Statistic
                            title="Сервисы"
                            value={stats.services ? Object.keys(stats.services).length : 0}
                            prefix={<Server size={20} />}
                        />
                    </Card>
                </Col>
                <Col xs={24} sm={12} lg={6}>
                    <Card>
                        <Statistic
                            title="Активные пользователи"
                            value={Math.floor(Math.random() * 5000) + 1000}
                            prefix={<Users size={20} />}
                        />
                    </Card>
                </Col>
                <Col xs={24} sm={12} lg={6}>
                    <Card>
                        <Statistic
                            title="Время ответа"
                            value={Math.floor(Math.random() * 200) + 50}
                            suffix="ms"
                            prefix={<Clock size={20} />}
                        />
                    </Card>
                </Col>
            </Row>

            <Row gutter={[16, 16]}>
                <Col xs={24} lg={12}>
                    <Card
                        title="Последние логи"
                        extra={
                            <Button type="link" onClick={loadData} loading={loading}>
                                Обновить
                            </Button>
                        }
                    >
                        <Table
                            dataSource={recentLogs}
                            columns={logColumns}
                            size="small"
                            pagination={false}
                            scroll={{ y: 300 }}
                            loading={loading}
                        />
                    </Card>
                </Col>

                <Col xs={24} lg={12}>
                    <Card title="Статистика по уровням">
                        {stats.levels && Object.entries(stats.levels).map(([level, count]) => (
                            <div key={level} style={{ marginBottom: '8px', display: 'flex', justifyContent: 'space-between' }}>
                                <Tag color={
                                    level === 'INFO' ? 'blue' :
                                        level === 'WARN' ? 'orange' :
                                            level === 'ERROR' ? 'red' : 'green'
                                }>
                                    {level}
                                </Tag>
                                <span>{count}</span>
                            </div>
                        ))}
                    </Card>

                    <Card title="Статистика по сервисам" style={{ marginTop: '16px' }}>
                        {stats.services && Object.entries(stats.services)
                            .sort(([,a], [,b]) => b - a)
                            .slice(0, 5)
                            .map(([service, count]) => (
                                <div key={service} style={{ marginBottom: '8px', display: 'flex', justifyContent: 'space-between' }}>
                                    <span>{service}</span>
                                    <span>{count}</span>
                                </div>
                            ))}
                    </Card>
                </Col>
            </Row>
        </div>
    )
}

export default Dashboard