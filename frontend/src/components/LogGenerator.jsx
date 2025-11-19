import React, { useState } from 'react'
import { Card, Form, InputNumber, Button, Slider, Select, Alert, Space, Statistic, Row, Col } from 'antd'
import { PlayCircle, Zap, AlertTriangle } from 'lucide-react'
import { simulatorAPI } from '../services/api'

const { Option } = Select

const LogGenerator = () => {
    const [form] = Form.useForm()
    const [loading, setLoading] = useState(false)
    const [result, setResult] = useState(null)

    const scenarios = [
        { value: 'load_test', label: '🚀 Нагрузочное тестирование', description: 'Высокая нагрузка' },
        { value: 'error_spike', label: '🔴 Всплеск ошибок', description: 'Увеличение ошибок' },
        { value: 'slow_responses', label: '🐌 Медленные ответы', description: 'Увеличение времени ответа' },
        { value: 'normal_operation', label: '✅ Нормальная работа', description: 'Стандартная нагрузка' },
    ]

    const onGenerate = async (values) => {
        setLoading(true)
        setResult(null)
        try {
            const response = await simulatorAPI.generateLogs(values)
            setResult(response.data)
        } catch (error) {
            console.error('Error generating logs:', error)
        } finally {
            setLoading(false)
        }
    }

    const quickGenerate = (logCount) => {
        form.setFieldsValue({ log_count: logCount })
        form.submit()
    }

    return (
        <div style={{ padding: '20px' }}>
            <h1 style={{ marginBottom: '24px' }}>🎮 Генератор логов</h1>

            <Row gutter={[16, 16]}>
                <Col xs={24} lg={12}>
                    <Card title="Быстрая генерация" style={{ marginBottom: '16px' }}>
                        <Space wrap>
                            <Button
                                icon={<Zap size={16} />}
                                onClick={() => quickGenerate(100)}
                            >
                                100 логов
                            </Button>
                            <Button
                                icon={<Zap size={16} />}
                                onClick={() => quickGenerate(1000)}
                            >
                                1,000 логов
                            </Button>
                            <Button
                                icon={<Zap size={16} />}
                                onClick={() => quickGenerate(5000)}
                            >
                                5,000 логов
                            </Button>
                        </Space>
                    </Card>

                    <Card title="Настройка генерации">
                        <Form
                            form={form}
                            layout="vertical"
                            onFinish={onGenerate}
                            initialValues={{ log_count: 1000 }}
                        >
                            <Form.Item
                                name="log_count"
                                label="Количество логов"
                                rules={[{ required: true, message: 'Введите количество логов' }]}
                            >
                                <InputNumber
                                    min={1}
                                    max={10000}
                                    style={{ width: '100%' }}
                                    placeholder="От 1 до 10,000"
                                />
                            </Form.Item>

                            <Form.Item
                                name="scenario"
                                label="Сценарий"
                            >
                                <Select placeholder="Выберите сценарий (опционально)">
                                    {scenarios.map(scenario => (
                                        <Option key={scenario.value} value={scenario.value}>
                                            <div>
                                                <div>{scenario.label}</div>
                                                <div style={{ fontSize: '12px', color: '#666' }}>
                                                    {scenario.description}
                                                </div>
                                            </div>
                                        </Option>
                                    ))}
                                </Select>
                            </Form.Item>

                            <Form.Item>
                                <Button
                                    type="primary"
                                    htmlType="submit"
                                    loading={loading}
                                    icon={<PlayCircle size={16} />}
                                    size="large"
                                    style={{ width: '100%' }}
                                >
                                    Сгенерировать логи
                                </Button>
                            </Form.Item>
                        </Form>
                    </Card>
                </Col>

                <Col xs={24} lg={12}>
                    {result && (
                        <Card title="Результат генерации" className="fade-in">
                            <Row gutter={[16, 16]}>
                                <Col xs={12}>
                                    <Statistic title="Сгенерировано логов" value={result.generated} />
                                </Col>
                                <Col xs={12}>
                                    <Statistic title="Метрик обновлено" value={result.metrics_count} />
                                </Col>
                            </Row>

                            <Alert
                                message="Пример сгенерированного лога"
                                description={
                                    <div style={{ marginTop: '8px' }}>
                                        <div><strong>Сервис:</strong> {result.sample_log?.service}</div>
                                        <div><strong>Уровень:</strong> {result.sample_log?.level}</div>
                                        <div><strong>Сообщение:</strong> {result.sample_log?.message}</div>
                                        {result.sample_log?.duration && (
                                            <div><strong>Время:</strong> {result.sample_log.duration}ms</div>
                                        )}
                                    </div>
                                }
                                type="info"
                                style={{ marginTop: '16px' }}
                            />
                        </Card>
                    )}

                    <Card title="Статистика в реальном времени" style={{ marginTop: '16px' }}>
                        <Row gutter={[16, 16]}>
                            <Col xs={8}>
                                <Statistic title="RPS" value={Math.floor(Math.random() * 100) + 50} suffix="req/s" />
                            </Col>
                            <Col xs={8}>
                                <Statistic title="Ошибки" value={Math.floor(Math.random() * 5)} suffix="%" />
                            </Col>
                            <Col xs={8}>
                                <Statistic title="Задержка" value={Math.floor(Math.random() * 200) + 50} suffix="ms" />
                            </Col>
                        </Row>
                    </Card>
                </Col>
            </Row>
        </div>
    )
}

export default LogGenerator