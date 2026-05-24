import { useQuery } from '@tanstack/react-query'
import { Row, Col, Card, Typography, Spin } from 'antd'
import { RiseOutlined, FallOutlined } from '@ant-design/icons'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import api from '../api/client'
import StatCard from '../components/StatCard'
import type { DashboardStats, TrendPoint } from '../types'

const { Title } = Typography

export default function Dashboard() {
  const { data: stats, isLoading } = useQuery<DashboardStats>({
    queryKey: ['dashboard-stats'],
    queryFn: () => api.get('/dashboard/stats').then(r => r.data),
    refetchInterval: 60_000,
  })

  const { data: trend = [] } = useQuery<TrendPoint[]>({
    queryKey: ['dashboard-trend'],
    queryFn: () => api.get('/dashboard/trend').then(r => r.data),
    refetchInterval: 60_000,
  })

  if (isLoading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />

  return (
    <div>
      <Title level={4} style={{ marginBottom: 16 }}>数据概览</Title>
      <Row gutter={[16, 16]}>
        <Col xs={12} sm={6}>
          <StatCard title="今日消耗" value={stats?.today_cost ?? 0} prefix="¥" precision={2} />
        </Col>
        <Col xs={12} sm={6}>
          <StatCard title="平均ROI" value={stats?.avg_roi ?? 0} precision={2}
            suffix={stats?.avg_roi && stats.avg_roi >= 1 ? <RiseOutlined /> : <FallOutlined />}
            color={stats?.avg_roi && stats.avg_roi >= 1 ? '#52c41a' : '#ff4d4f'}
          />
        </Col>
        <Col xs={12} sm={6}>
          <StatCard title="成交订单" value={stats?.total_conversions ?? 0} />
        </Col>
        <Col xs={12} sm={6}>
          <StatCard title="在投计划" value={stats?.active_ads ?? 0} />
        </Col>
      </Row>

      <Card style={{ marginTop: 24 }}>
        <Title level={5}>近7天消耗趋势</Title>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={trend}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="date" />
            <YAxis yAxisId="left" />
            <YAxis yAxisId="right" orientation="right" />
            <Tooltip />
            <Legend />
            <Line yAxisId="left" type="monotone" dataKey="cost" name="消耗(¥)" stroke="#1677ff" strokeWidth={2} />
            <Line yAxisId="right" type="monotone" dataKey="roi" name="ROI" stroke="#52c41a" strokeWidth={2} />
          </LineChart>
        </ResponsiveContainer>
      </Card>
    </div>
  )
}
