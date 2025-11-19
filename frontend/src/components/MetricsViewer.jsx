import React, { useState, useEffect } from 'react'
import { Card, Table, Tag, Select, Button, Row, Col } from 'antd'
import { RefreshCw, Download } from 'lucide-react'
import { simulatorAPI } from '../services/api'

const { Option } = Select

const MetricsViewer = () => {
    const [metrics, setMetrics] = useState([])
    const [logs, setLogs] = useState([])
    const [format, setFormat] = useState('json')
    const [loading, setLoading] = useState(false)

    const loadMetrics = async () => {
        setLoading(true)
        try {
            const response = await simulatorAPI.getMetrics(format)
            if (format === 'json') {
                setMetrics(response.data.metrics || [])
            }
        } catch (error) {
            console.error('Error loading metrics:', error)
        } finally {
            setLoading(false)
        }
    }

    const loadLogs = async () => {
        try {
            const response = await simulatorAPI.getLogs({ limit: 50 })
            setLogs(response.data.logs || [])
        } catch (error) {
            console.error('Error loading logs:', error)
        }
    }

    useEffect(() => {
        loadMetrics()
        loadLogs()
    }, [format])

    const metricsColumns = [
        {
            title: 'Метрика',
            dataIndex: 'name',
            key: 'name',
            render: (name) => <Tag color="blue">{name}</Tag>,
        },
        {
            title: 'Значение',
            dataIndex: 'value',
            key: 'value',
            render: (value) => value.toFixed(2),
        },
        {
            title: 'Тип',
            dataIndex: 'type',
            key: 'type',
        },
        {
            title: 'Лейблы',
            dataIndex: 'labels',
            key: 'labels',
            render: (labels) => labels && Object.entries(labels).map(([k, v]) => (
                <Tag key={k} size="small">{k}={v}</Tag>
            )),
        },
    ]

    const logColumns = [
        {
            title: 'Время',
            dataIndex: 'timestamp',
            key: 'timestamp',
            render: (time) => new Date(time).toLocaleString(),
            width: 160,
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
            title: 'Метод',
            dataIndex: 'method',
            key: 'method',
            width: 80,
        },
        {
            title: 'Путь',
            dataIndex: 'path',
            key: 'path',
            ellipsis: true,
        },
        {
            title: 'Статус',
            dataIndex: 'status',
            key: 'status',
            render: (status) => status && <Tag>{status}</Tag>,
            width: 80,
        },
        {
            title: 'Время',
            dataIndex: 'duration',
            key: 'duration',
            render: (duration) => duration && `${duration}ms`,
            width: 80,
        },
    ]

    return (
        <div style={{ padding: '20px' }}>
            <h1 style={{ marginBottom: '24px' }}>📈 Просмотр метрик и логов</h1>

            <Row gutter={[16, 16]}>
                <Col xs={24}>
                    <Card
                        title="Метрики"
                        extra={
                            <Space>
                                <Select
                                    value={format}
                                    onChange={setFormat}
                                    style={{ width: 120 }}
                                >
                                    <Option value="json">JSON</Option>
                                    <Option value="prometheus">Prometheus</Option>
                                </Select>
                                <Button
                                    icon={<RefreshCw size={14} />}
                                    onClick={loadMetrics}
                                    loading={loading}
                                >
                                    Обновить
                                </Button>
                            </Space>
                        }
                    >
                        {format === 'json' ? (
                            <Table
                                dataSource={metrics}
                                columns={metricsColumns}
                                rowKey={(record) => `${record.name}-${JSON.stringify(record.labels)}`}
                                pagination={false}
                                size="small"
                            />
                        ) : (
                            <pre style={{ background: '#f5f5f5', padding: '16px', borderRadius: '6px', fontSize: '12px' }}>
                {metrics.length > 0 ? metrics.map(m =>
                    `${m.name}${m.labels ? `{${Object.entries(m.labels).map(([k, v]) => `${k}="${v}"`).join(',')}}` : ''} ${m.value}`
                ).join('\n') : 'Загрузка...'}
              </pre>
                        )}
                    </Card>
                </Col>

                <Col xs={24}>
                    <Card
                        title="Последние логи"
                        extra={
                            <Button
                                icon={<RefreshCw size={14} />}
                                onClick={loadLogs}
                            >
                                Обновить
                            </Button>
                        }
                    >
                        <Table
                            dataSource={logs}
                            columns={logColumns}
                            rowKey="timestamp"
                            pagination={{ pageSize: 10 }}
                            size="small"
                            scroll={{ x: 800 }}
                        />
                    </Card>
                </Col>
            </Row>
        </div>
    )
}

export default MetricsViewer