import React, { useState, useEffect } from 'react'
import {
    Card,
    Table,
    Button,
    Tag,
    Space,
    Modal,
    Form,
    Input,
    InputNumber,
    Select,
    DatePicker,
    TimePicker,
    message,
    Row,
    Col,
    Divider,
    Collapse,
    Tabs,
    List,
    Progress,
    Popconfirm
} from 'antd'
import {
    PlayCircle,
    PauseCircle,
    Plus,
    Clock,
    Calendar,
    Link,
    Settings,
    Trash2,
    Power,
    PowerOff,
    History
} from 'lucide-react'
import { simulatorAPI } from '../services/api'
import dayjs from 'dayjs'

const { Option } = Select
const { Panel } = Collapse
const { TabPane } = Tabs
const { TextArea } = Input

const ScenarioManager = () => {
    const [scenarios, setScenarios] = useState({ available: [], active: [], chains: {} })
    const [schedules, setSchedules] = useState([])
    const [chains, setChains] = useState([])
    const [chainSchedules, setChainSchedules] = useState([])
    const [activeTab, setActiveTab] = useState('scenarios')
    const [createModalVisible, setCreateModalVisible] = useState(false)
    const [chainModalVisible, setChainModalVisible] = useState(false)
    const [scenarioModalVisible, setScenarioModalVisible] = useState(false)
    const [chainScheduleModalVisible, setChainScheduleModalVisible] = useState(false)
    const [loading, setLoading] = useState(false)
    const [scenarioForm] = Form.useForm()
    const [chainForm] = Form.useForm()
    const [chainScheduleForm] = Form.useForm()

    const loadData = async () => {
        try {
            const [scenariosResponse, schedulesResponse, chainsResponse, chainSchedulesResponse] = await Promise.all([
                simulatorAPI.listScenarios(),
                simulatorAPI.listSchedules(),
                simulatorAPI.listChains(),
                simulatorAPI.listChainSchedules()
            ])
            setScenarios(scenariosResponse.data)
            setSchedules(schedulesResponse.data.schedules || [])
            setChains(chainsResponse.data.chains || [])
            setChainSchedules(chainSchedulesResponse.data.schedules || [])
        } catch (error) {
            console.error('Error loading data:', error)
            message.error('Ошибка загрузки данных')
        }
    }

    useEffect(() => {
        loadData()
        const interval = setInterval(loadData, 5000)
        return () => clearInterval(interval)
    }, [])

    const handleStartScenario = async (scenarioType, config = {}) => {
        try {
            await simulatorAPI.startScenario({ type: scenarioType, config })
            message.success('Сценарий запущен')
            loadData()
        } catch (error) {
            message.error('Ошибка запуска сценария')
        }
    }

    const handleStopScenario = async (scenarioType) => {
        try {
            await simulatorAPI.stopScenario({ type: scenarioType })
            message.success('Сценарий остановлен')
            loadData()
        } catch (error) {
            message.error('Ошибка остановки сценария')
        }
    }

    const handleCreateSchedule = async (values) => {
        try {
            await simulatorAPI.createSchedule(values)
            message.success('Расписание создано')
            setCreateModalVisible(false)
            scenarioForm.resetFields()
            loadData()
        } catch (error) {
            message.error('Ошибка создания расписания')
        }
    }

    const handleAdvancedScenario = async (values) => {
        try {
            const config = {}

            if (values.log_count) {
                config.log_count = values.log_count
            }

            if (values.duration_unit && values.duration_value) {
                config[`duration_${values.duration_unit}`] = values.duration_value
            }

            if (values.interval_unit && values.interval_value) {
                config[`interval_${values.interval_unit}`] = values.interval_value
            }

            if (values.start_date) {
                config.start_date = values.start_date.format('YYYY-MM-DDTHH:mm:ssZ')
            }

            if (values.end_date) {
                config.end_date = values.end_date.format('YYYY-MM-DDTHH:mm:ssZ')
            }

            if (values.labels) {
                const labels = {}
                values.labels.split(',').forEach(label => {
                    const [key, value] = label.split('=')
                    if (key && value) labels[key.trim()] = value.trim()
                })
                config.labels = labels
            }

            await simulatorAPI.startScenario({
                type: values.scenario_type,
                config
            })

            message.success('Сценарий запущен с расширенными настройками')
            setScenarioModalVisible(false)
            scenarioForm.resetFields()
            loadData()
        } catch (error) {
            message.error('Ошибка запуска сценария')
        }
    }

    const handleCreateChain = async (values) => {
        try {
            const chainData = {
                name: values.name,
                description: values.description,
                steps: values.steps.map(step => ({
                    name: step.name,
                    scenario_type: step.scenario_type,
                    delay_before: step.delay_before || 0,
                    config: step.duration_value && step.duration_unit ? {
                        [`duration_${step.duration_unit}`]: step.duration_value
                    } : {}
                }))
            }

            await simulatorAPI.createChain(chainData)
            message.success('Цепочка создана')
            setChainModalVisible(false)
            chainForm.resetFields()
            loadData()
        } catch (error) {
            message.error('Ошибка создания цепочки')
        }
    }

    const handleStartChain = async (chainId) => {
        try {
            await simulatorAPI.startChain(chainId)
            message.success('Цепочка запущена')
            loadData()
        } catch (error) {
            message.error('Ошибка запуска цепочки')
        }
    }

    const handleDeleteChain = async (chainId) => {
        try {
            await simulatorAPI.deleteChain(chainId)
            message.success('Цепочка удалена')
            loadData()
        } catch (error) {
            message.error('Ошибка удаления цепочки')
        }
    }

    const handleCreateChainSchedule = async (values) => {
        try {
            await simulatorAPI.createChainSchedule(values)
            message.success('Расписание цепочки создано')
            setChainScheduleModalVisible(false)
            chainScheduleForm.resetFields()
            loadData()
        } catch (error) {
            message.error('Ошибка создания расписания цепочки')
        }
    }

    const handleEnableChainSchedule = async (id) => {
        try {
            await simulatorAPI.enableChainSchedule(id)
            message.success('Расписание цепочки включено')
            loadData()
        } catch (error) {
            message.error('Ошибка включения расписания цепочки')
        }
    }

    const handleDisableChainSchedule = async (id) => {
        try {
            await simulatorAPI.disableChainSchedule(id)
            message.success('Расписание цепочки отключено')
            loadData()
        } catch (error) {
            message.error('Ошибка отключения расписания цепочки')
        }
    }

    const handleDeleteChainSchedule = async (id) => {
        try {
            await simulatorAPI.deleteChainSchedule(id)
            message.success('Расписание цепочки удалено')
            loadData()
        } catch (error) {
            message.error('Ошибка удаления расписания цепочки')
        }
    }

    const scenarioColumns = [
        {
            title: 'Название',
            dataIndex: 'name',
            key: 'name',
        },
        {
            title: 'Описание',
            dataIndex: 'description',
            key: 'description',
        },
        {
            title: 'Логов за запуск',
            dataIndex: 'log_count',
            key: 'log_count',
        },
        {
            title: 'Действия',
            key: 'actions',
            render: (_, record) => (
                <Space>
                    <Button
                        type="primary"
                        size="small"
                        icon={<PlayCircle size={12} />}
                        onClick={() => handleStartScenario(record.type)}
                    >
                        Быстрый запуск
                    </Button>
                    <Button
                        size="small"
                        icon={<Settings size={12} />}
                        onClick={() => {
                            scenarioForm.setFieldsValue({ scenario_type: record.type })
                            setScenarioModalVisible(true)
                        }}
                    >
                        Расширенно
                    </Button>
                </Space>
            ),
        },
    ]

    const activeScenarioColumns = [
        {
            title: 'Тип',
            dataIndex: 'type',
            key: 'type',
        },
        {
            title: 'Название',
            dataIndex: ['config', 'name'],
            key: 'name',
        },
        {
            title: 'Статус',
            dataIndex: 'active',
            key: 'active',
            render: (active) => (
                <Tag color={active ? 'green' : 'red'}>
                    {active ? 'Активен' : 'Неактивен'}
                </Tag>
            ),
        },
        {
            title: 'Запущен',
            dataIndex: 'started',
            key: 'started',
            render: (started) => dayjs(started).format('DD.MM.YYYY HH:mm:ss'),
        },
        {
            title: 'Действия',
            key: 'actions',
            render: (_, record) => (
                <Button
                    danger
                    size="small"
                    icon={<PauseCircle size={12} />}
                    onClick={() => handleStopScenario(record.type)}
                    disabled={!record.active}
                >
                    Остановить
                </Button>
            ),
        },
    ]

    const scheduleColumns = [
        {
            title: 'Название',
            dataIndex: 'name',
            key: 'name',
        },
        {
            title: 'Сценарий',
            dataIndex: 'scenario_type',
            key: 'scenario_type',
        },
        {
            title: 'Cron',
            dataIndex: 'cron_expr',
            key: 'cron_expr',
        },
        {
            title: 'Статус',
            dataIndex: 'enabled',
            key: 'enabled',
            render: (enabled) => (
                <Tag color={enabled ? 'green' : 'red'}>
                    {enabled ? 'Активно' : 'Отключено'}
                </Tag>
            ),
        },
        {
            title: 'Следующий запуск',
            dataIndex: 'next_run',
            key: 'next_run',
            render: (nextRun) => nextRun ? dayjs(nextRun).format('DD.MM.YYYY HH:mm:ss') : '-',
        },
    ]

    const chainColumns = [
        {
            title: 'Название',
            dataIndex: 'name',
            key: 'name',
        },
        {
            title: 'Описание',
            dataIndex: 'description',
            key: 'description',
        },
        {
            title: 'Шагов',
            dataIndex: 'steps',
            key: 'steps',
            render: (steps) => steps?.length || 0,
        },
        {
            title: 'Статус',
            dataIndex: 'status',
            key: 'status',
            render: (status) => (
                <Tag color={
                    status === 'running' ? 'green' :
                        status === 'pending' ? 'blue' :
                            status === 'completed' ? 'green' : 'red'
                }>
                    {status === 'running' ? 'Выполняется' :
                        status === 'pending' ? 'Ожидание' :
                            status === 'completed' ? 'Завершено' : 'Ошибка'}
                </Tag>
            ),
        },
        {
            title: 'Действия',
            key: 'actions',
            render: (_, record) => (
                <Space>
                    <Button
                        type="primary"
                        size="small"
                        icon={<PlayCircle size={12} />}
                        onClick={() => handleStartChain(record.id)}
                        disabled={record.status === 'running'}
                    >
                        Запустить
                    </Button>
                    <Popconfirm
                        title="Удалить цепочку?"
                        onConfirm={() => handleDeleteChain(record.id)}
                        okText="Да"
                        cancelText="Нет"
                    >
                        <Button
                            danger
                            size="small"
                            icon={<Trash2 size={12} />}
                        >
                            Удалить
                        </Button>
                    </Popconfirm>
                </Space>
            ),
        },
    ]

    const chainScheduleColumns = [
        {
            title: 'Название',
            dataIndex: 'name',
            key: 'name',
        },
        {
            title: 'Цепочка',
            dataIndex: 'chain_name',
            key: 'chain_name',
        },
        {
            title: 'Cron',
            dataIndex: 'cron_expr',
            key: 'cron_expr',
        },
        {
            title: 'Статус',
            dataIndex: 'enabled',
            key: 'enabled',
            render: (enabled) => (
                <Tag color={enabled ? 'green' : 'red'}>
                    {enabled ? 'Активно' : 'Отключено'}
                </Tag>
            ),
        },
        {
            title: 'Следующий запуск',
            dataIndex: 'next_run',
            key: 'next_run',
            render: (nextRun) => nextRun ? dayjs(nextRun).format('DD.MM.YYYY HH:mm:ss') : '-',
        },
        {
            title: 'Действия',
            key: 'actions',
            render: (_, record) => (
                <Space>
                    {record.enabled ? (
                        <Button
                            size="small"
                            icon={<PowerOff size={12} />}
                            onClick={() => handleDisableChainSchedule(record.id)}
                        >
                            Отключить
                        </Button>
                    ) : (
                        <Button
                            size="small"
                            icon={<Power size={12} />}
                            onClick={() => handleEnableChainSchedule(record.id)}
                        >
                            Включить
                        </Button>
                    )}
                    <Popconfirm
                        title="Удалить расписание?"
                        onConfirm={() => handleDeleteChainSchedule(record.id)}
                        okText="Да"
                        cancelText="Нет"
                    >
                        <Button
                            danger
                            size="small"
                            icon={<Trash2 size={12} />}
                        >
                            Удалить
                        </Button>
                    </Popconfirm>
                </Space>
            ),
        },
    ]

    const durationUnits = [
        { value: 'seconds', label: 'Секунды' },
        { value: 'minutes', label: 'Минуты' },
        { value: 'hours', label: 'Часы' },
    ]

    const predefinedChains = scenarios.chains || {}

    return (
        <div style={{ padding: '20px' }}>
            <h1 style={{ marginBottom: '24px' }}>⚙️ Управление сценариями и цепочками</h1>

            <Tabs activeKey={activeTab} onChange={setActiveTab}>
                <TabPane tab="📋 Сценарии" key="scenarios">
                    <Space style={{ marginBottom: '16px' }} wrap>
                        <Button
                            type="primary"
                            icon={<Plus size={14} />}
                            onClick={() => setCreateModalVisible(true)}
                        >
                            Создать расписание
                        </Button>
                    </Space>

                    <Row gutter={[16, 16]}>
                        <Col xs={24} lg={12}>
                            <Card title="Доступные сценарии">
                                <Table
                                    dataSource={Object.values(scenarios.available || {})}
                                    columns={scenarioColumns}
                                    rowKey="type"
                                    pagination={false}
                                    size="small"
                                />
                            </Card>
                        </Col>

                        <Col xs={24} lg={12}>
                            <Card title="Активные сценарии">
                                <Table
                                    dataSource={scenarios.active || []}
                                    columns={activeScenarioColumns}
                                    rowKey="type"
                                    pagination={false}
                                    size="small"
                                />
                            </Card>
                        </Col>
                    </Row>

                    <Card title="Расписания сценариев" style={{ marginTop: '16px' }}>
                        <Table
                            dataSource={schedules}
                            columns={scheduleColumns}
                            rowKey="id"
                            pagination={false}
                            size="small"
                        />
                    </Card>
                </TabPane>

                <TabPane tab="🔗 Цепочки" key="chains">
                    <Space style={{ marginBottom: '16px' }} wrap>
                        <Button
                            type="primary"
                            icon={<Plus size={14} />}
                            onClick={() => setChainModalVisible(true)}
                        >
                            Создать цепочку
                        </Button>
                        <Button
                            icon={<Plus size={14} />}
                            onClick={() => setChainScheduleModalVisible(true)}
                        >
                            Создать расписание цепочки
                        </Button>
                    </Space>

                    <Row gutter={[16, 16]}>
                        <Col xs={24} lg={12}>
                            <Card title="Пользовательские цепочки">
                                <Table
                                    dataSource={chains}
                                    columns={chainColumns}
                                    rowKey="id"
                                    pagination={false}
                                    size="small"
                                />
                            </Card>
                        </Col>

                        <Col xs={24} lg={12}>
                            <Card title="Предопределенные цепочки">
                                <List
                                    dataSource={Object.values(predefinedChains)}
                                    renderItem={chain => (
                                        <List.Item
                                            actions={[
                                                <Button
                                                    type="primary"
                                                    size="small"
                                                    onClick={() => {
                                                        // Запуск предопределенной цепочки
                                                        handleStartScenario(chain.steps[0])
                                                    }}
                                                >
                                                    Запустить
                                                </Button>
                                            ]}
                                        >
                                            <List.Item.Meta
                                                title={chain.name}
                                                description={chain.description}
                                            />
                                        </List.Item>
                                    )}
                                />
                            </Card>
                        </Col>
                    </Row>

                    <Card title="Расписания цепочек" style={{ marginTop: '16px' }}>
                        <Table
                            dataSource={chainSchedules}
                            columns={chainScheduleColumns}
                            rowKey="id"
                            pagination={false}
                            size="small"
                        />
                    </Card>
                </TabPane>
            </Tabs>

            {/* Модальное окно создания расписания сценария */}
            <Modal
                title="Создание расписания"
                open={createModalVisible}
                onCancel={() => setCreateModalVisible(false)}
                footer={null}
                width={600}
            >
                <Form
                    form={scenarioForm}
                    layout="vertical"
                    onFinish={handleCreateSchedule}
                >
                    <Form.Item
                        name="name"
                        label="Название расписания"
                        rules={[{ required: true, message: 'Введите название' }]}
                    >
                        <Input placeholder="Например: Ночные тесты" />
                    </Form.Item>

                    <Form.Item
                        name="scenario_type"
                        label="Сценарий"
                        rules={[{ required: true, message: 'Выберите сценарий' }]}
                    >
                        <Select placeholder="Выберите сценарий">
                            {Object.values(scenarios.available || {}).map(scenario => (
                                <Option key={scenario.type} value={scenario.type}>
                                    {scenario.name}
                                </Option>
                            ))}
                        </Select>
                    </Form.Item>

                    <Form.Item
                        name="cron_expr"
                        label="Cron выражение"
                        rules={[{ required: true, message: 'Введите cron выражение' }]}
                    >
                        <Input placeholder="0 2 * * * - каждый день в 2:00" />
                    </Form.Item>

                    <Form.Item
                        name="enabled"
                        label="Статус"
                        initialValue={true}
                    >
                        <Select>
                            <Option value={true}>Активно</Option>
                            <Option value={false}>Отключено</Option>
                        </Select>
                    </Form.Item>

                    <Form.Item>
                        <Button type="primary" htmlType="submit" style={{ width: '100%' }}>
                            Создать расписание
                        </Button>
                    </Form.Item>
                </Form>
            </Modal>

            {/* Модальное окно расширенного запуска сценария */}
            <Modal
                title="Расширенный запуск сценария"
                open={scenarioModalVisible}
                onCancel={() => setScenarioModalVisible(false)}
                footer={null}
                width={700}
            >
                <Form
                    form={scenarioForm}
                    layout="vertical"
                    onFinish={handleAdvancedScenario}
                >
                    <Form.Item
                        name="scenario_type"
                        label="Сценарий"
                        rules={[{ required: true, message: 'Выберите сценарий' }]}
                    >
                        <Select placeholder="Выберите сценарий">
                            {Object.values(scenarios.available || {}).map(scenario => (
                                <Option key={scenario.type} value={scenario.type}>
                                    {scenario.name}
                                </Option>
                            ))}
                        </Select>
                    </Form.Item>

                    <Row gutter={16}>
                        <Col span={12}>
                            <Form.Item
                                name="log_count"
                                label="Количество логов"
                            >
                                <InputNumber
                                    min={1}
                                    max={10000}
                                    style={{ width: '100%' }}
                                    placeholder="По умолчанию из сценария"
                                />
                            </Form.Item>
                        </Col>
                        <Col span={12}>
                            <Form.Item
                                name="labels"
                                label="Лейблы (key=value, разделенные запятыми)"
                            >
                                <Input placeholder="environment=test,team=devops" />
                            </Form.Item>
                        </Col>
                    </Row>

                    <Collapse
                        ghost
                        style={{ marginBottom: '16px' }}
                        items={[
                            {
                                key: '1',
                                label: '⚙️ Расширенные настройки',
                                children: (
                                    <>
                                        <Divider orientation="left">Время выполнения</Divider>
                                        <Row gutter={16}>
                                            <Col span={8}>
                                                <Form.Item name="duration_value" label="Продолжительность">
                                                    <InputNumber min={1} style={{ width: '100%' }} placeholder="Значение" />
                                                </Form.Item>
                                            </Col>
                                            <Col span={8}>
                                                <Form.Item name="duration_unit" label="Единица измерения">
                                                    <Select placeholder="Выберите единицу">
                                                        {durationUnits.map(unit => (
                                                            <Option key={unit.value} value={unit.value}>
                                                                {unit.label}
                                                            </Option>
                                                        ))}
                                                    </Select>
                                                </Form.Item>
                                            </Col>
                                            <Col span={8}>
                                                <Form.Item name="interval_value" label="Интервал повтора">
                                                    <InputNumber min={1} style={{ width: '100%' }} placeholder="Значение" />
                                                </Form.Item>
                                            </Col>
                                            <Col span={8}>
                                                <Form.Item name="interval_unit" label="Единица измерения">
                                                    <Select placeholder="Выберите единицу">
                                                        {durationUnits.map(unit => (
                                                            <Option key={unit.value} value={unit.value}>
                                                                {unit.label}
                                                            </Option>
                                                        ))}
                                                    </Select>
                                                </Form.Item>
                                            </Col>
                                        </Row>

                                        <Divider orientation="left">Временные ограничения</Divider>
                                        <Row gutter={16}>
                                            <Col span={12}>
                                                <Form.Item name="start_date" label="Дата начала">
                                                    <DatePicker
                                                        showTime
                                                        style={{ width: '100%' }}
                                                        placeholder="Выберите дату и время начала"
                                                    />
                                                </Form.Item>
                                            </Col>
                                            <Col span={12}>
                                                <Form.Item name="end_date" label="Дата окончания">
                                                    <DatePicker
                                                        showTime
                                                        style={{ width: '100%' }}
                                                        placeholder="Выберите дату и время окончания"
                                                    />
                                                </Form.Item>
                                            </Col>
                                        </Row>
                                    </>
                                )
                            }
                        ]}
                    />

                    <Form.Item>
                        <Button type="primary" htmlType="submit" style={{ width: '100%' }} size="large">
                            🚀 Запустить сценарий
                        </Button>
                    </Form.Item>
                </Form>
            </Modal>

            {/* Модальное окно создания цепочки сценариев */}
            <Modal
                title="Создание цепочки сценариев"
                open={chainModalVisible}
                onCancel={() => setChainModalVisible(false)}
                footer={null}
                width={800}
            >
                <Form
                    form={chainForm}
                    layout="vertical"
                    onFinish={handleCreateChain}
                >
                    <Form.Item
                        name="name"
                        label="Название цепочки"
                        rules={[{ required: true, message: 'Введите название цепочки' }]}
                    >
                        <Input placeholder="Например: Полное нагрузочное тестирование" />
                    </Form.Item>

                    <Form.Item
                        name="description"
                        label="Описание цепочки"
                    >
                        <TextArea placeholder="Описание цепочки сценариев" rows={3} />
                    </Form.Item>

                    <Form.List name="steps">
                        {(fields, { add, remove }) => (
                            <>
                                {fields.map(({ key, name, ...restField }) => (
                                    <Card
                                        key={key}
                                        title={`Шаг ${name + 1}`}
                                        style={{ marginBottom: '16px' }}
                                        extra={
                                            <Button type="link" danger onClick={() => remove(name)}>
                                                Удалить
                                            </Button>
                                        }
                                    >
                                        <Row gutter={16}>
                                            <Col span={12}>
                                                <Form.Item
                                                    {...restField}
                                                    name={[name, 'scenario_type']}
                                                    label="Сценарий"
                                                    rules={[{ required: true, message: 'Выберите сценарий' }]}
                                                >
                                                    <Select placeholder="Выберите сценарий">
                                                        {Object.values(scenarios.available || {}).map(scenario => (
                                                            <Option key={scenario.type} value={scenario.type}>
                                                                {scenario.name}
                                                            </Option>
                                                        ))}
                                                    </Select>
                                                </Form.Item>
                                            </Col>
                                            <Col span={12}>
                                                <Form.Item
                                                    {...restField}
                                                    name={[name, 'name']}
                                                    label="Название шага"
                                                >
                                                    <Input placeholder="Например: Пиковая нагрузка" />
                                                </Form.Item>
                                            </Col>
                                        </Row>

                                        <Row gutter={16}>
                                            <Col span={8}>
                                                <Form.Item
                                                    {...restField}
                                                    name={[name, 'delay_before']}
                                                    label="Задержка перед запуском (секунды)"
                                                >
                                                    <InputNumber min={0} style={{ width: '100%' }} placeholder="0" />
                                                </Form.Item>
                                            </Col>
                                            <Col span={8}>
                                                <Form.Item
                                                    {...restField}
                                                    name={[name, 'duration_value']}
                                                    label="Продолжительность"
                                                >
                                                    <InputNumber min={1} style={{ width: '100%' }} placeholder="Значение" />
                                                </Form.Item>
                                            </Col>
                                            <Col span={8}>
                                                <Form.Item
                                                    {...restField}
                                                    name={[name, 'duration_unit']}
                                                    label="Единица измерения"
                                                >
                                                    <Select placeholder="Выберите единицу">
                                                        {durationUnits.map(unit => (
                                                            <Option key={unit.value} value={unit.value}>
                                                                {unit.label}
                                                            </Option>
                                                        ))}
                                                    </Select>
                                                </Form.Item>
                                            </Col>
                                        </Row>
                                    </Card>
                                ))}

                                <Form.Item>
                                    <Button type="dashed" onClick={() => add()} block icon={<Plus />}>
                                        Добавить шаг
                                    </Button>
                                </Form.Item>
                            </>
                        )}
                    </Form.List>

                    <Form.Item>
                        <Button type="primary" htmlType="submit" style={{ width: '100%' }} size="large">
                            🔗 Создать цепочку
                        </Button>
                    </Form.Item>
                </Form>
            </Modal>

            {/* Модальное окно создания расписания цепочки */}
            <Modal
                title="Создание расписания цепочки"
                open={chainScheduleModalVisible}
                onCancel={() => setChainScheduleModalVisible(false)}
                footer={null}
                width={600}
            >
                <Form
                    form={chainScheduleForm}
                    layout="vertical"
                    onFinish={handleCreateChainSchedule}
                >
                    <Form.Item
                        name="name"
                        label="Название расписания"
                        rules={[{ required: true, message: 'Введите название' }]}
                    >
                        <Input placeholder="Например: Еженедельное тестирование" />
                    </Form.Item>

                    <Form.Item
                        name="chain_name"
                        label="Цепочка"
                        rules={[{ required: true, message: 'Выберите цепочку' }]}
                    >
                        <Select placeholder="Выберите цепочку">
                            {Object.keys(predefinedChains).map(chainName => (
                                <Option key={chainName} value={chainName}>
                                    {predefinedChains[chainName].name}
                                </Option>
                            ))}
                        </Select>
                    </Form.Item>

                    <Form.Item
                        name="cron_expr"
                        label="Cron выражение"
                        rules={[{ required: true, message: 'Введите cron выражение' }]}
                    >
                        <Input placeholder="0 2 * * 1 - каждый понедельник в 2:00" />
                    </Form.Item>

                    <Form.Item
                        name="enabled"
                        label="Статус"
                        initialValue={true}
                    >
                        <Select>
                            <Option value={true}>Активно</Option>
                            <Option value={false}>Отключено</Option>
                        </Select>
                    </Form.Item>

                    <Form.Item>
                        <Button type="primary" htmlType="submit" style={{ width: '100%' }}>
                            Создать расписание
                        </Button>
                    </Form.Item>
                </Form>
            </Modal>
        </div>
    )
}

export default ScenarioManager