import React, { useState } from 'react'
import { Layout, Menu, theme, Alert } from 'antd'
import { BarChart3, PlayCircle, Settings, Server, Link } from 'lucide-react'
import Dashboard from './components/Dashboard'
import LogGenerator from './components/LogGenerator'
import ScenarioManager from './components/ScenarioManager'
import MetricsViewer from './components/MetricsViewer'

const { Header, Content, Sider } = Layout

const App = () => {
    const [collapsed, setCollapsed] = useState(false)
    const [currentPage, setCurrentPage] = useState('dashboard')
    const {
        token: { colorBgContainer, borderRadiusLG },
    } = theme.useToken()

    const menuItems = [
        {
            key: 'dashboard',
            icon: <BarChart3 size={16} />,
            label: 'Дашборд',
        },
        {
            key: 'generator',
            icon: <PlayCircle size={16} />,
            label: 'Генератор логов',
        },
        {
            key: 'scenarios',
            icon: <Settings size={16} />,
            label: 'Сценарии и цепочки',
        },
        {
            key: 'metrics',
            icon: <Server size={16} />,
            label: 'Метрики',
        },
    ]

    const renderContent = () => {
        switch (currentPage) {
            case 'dashboard':
                return <Dashboard />
            case 'generator':
                return <LogGenerator />
            case 'scenarios':
                return <ScenarioManager />
            case 'metrics':
                return <MetricsViewer />
            default:
                return <Dashboard />
        }
    }

    return (
        <Layout style={{ minHeight: '100vh' }}>
            <Sider
                collapsible
                collapsed={collapsed}
                onCollapse={setCollapsed}
                theme="light"
            >
                <div style={{
                    height: 32,
                    margin: 16,
                    background: 'rgba(255, 255, 255, 0.2)',
                    borderRadius: 6,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontWeight: 'bold',
                    fontSize: collapsed ? 12 : 14,
                    color: '#1890ff'
                }}>
                    {collapsed ? 'MS' : 'Metrics Sim'}
                </div>
                <Menu
                    theme="light"
                    selectedKeys={[currentPage]}
                    mode="inline"
                    items={menuItems}
                    onClick={({ key }) => setCurrentPage(key)}
                />
            </Sider>

            <Layout>
                <Header style={{
                    padding: '0 20px',
                    background: colorBgContainer,
                    borderBottom: '1px solid #f0f0f0',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between'
                }}>
                    <div style={{ fontWeight: 'bold', fontSize: '16px' }}>
                        Metrics Simulator
                    </div>
                    <div style={{ color: '#666', fontSize: '14px' }}>
                        Симулятор логов и метрик
                    </div>
                </Header>

                <Content style={{ margin: '0' }}>
                    <div
                        style={{
                            padding: 0,
                            minHeight: 360,
                            background: colorBgContainer,
                            borderRadius: borderRadiusLG,
                        }}
                    >
                        {renderContent()}
                    </div>
                </Content>
            </Layout>
        </Layout>
    )
}

export default App