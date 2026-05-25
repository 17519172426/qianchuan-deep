import { useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Layout as AntLayout, Menu, Button, Typography } from 'antd'
import {
  DashboardOutlined,
  UnorderedListOutlined,
  SettingOutlined,
  BulbOutlined,
  PictureOutlined,
  BarChartOutlined,
  LogoutOutlined,
} from '@ant-design/icons'

const { Header, Sider, Content } = AntLayout

const menuItems = [
  { key: '/dashboard', icon: <DashboardOutlined />, label: '首页看板' },
  { key: '/ads', icon: <UnorderedListOutlined />, label: '全域计划' },
  { key: '/rules', icon: <SettingOutlined />, label: '规则管理' },
  { key: '/recommendations', icon: <BulbOutlined />, label: 'AI 推荐' },
  { key: '/creatives', icon: <PictureOutlined />, label: '素材库' },
  { key: '/reports', icon: <BarChartOutlined />, label: '数据报表' },
]

export default function Layout() {
  const navigate = useNavigate()
  const location = useLocation()
  const token = localStorage.getItem('token')
  const [collapsed, setCollapsed] = useState(false)

  if (!token) {
    navigate('/login')
    return null
  }

  const handleLogout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    navigate('/login')
  }

  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Sider collapsible collapsed={collapsed} onCollapse={setCollapsed}>
        <div style={{ height: 48, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <Typography.Text style={{ color: '#fff', fontWeight: 'bold', fontSize: collapsed ? 14 : 16 }}>
            {collapsed ? '千川' : '千川投流助手'}
          </Typography.Text>
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <AntLayout>
        <Header style={{ background: '#fff', padding: '0 24px', display: 'flex', justifyContent: 'flex-end', alignItems: 'center' }}>
          <Button type="text" icon={<LogoutOutlined />} onClick={handleLogout}>退出</Button>
        </Header>
        <Content style={{ margin: 16, padding: 24, background: '#fff', borderRadius: 8 }}>
          <Outlet />
        </Content>
      </AntLayout>
    </AntLayout>
  )
}
