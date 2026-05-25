import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card, Typography, DatePicker, Table, Statistic, Row, Col, Select, Space } from 'antd'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import api from '../api/client'
import type { DailySummary } from '../types'

const { Title } = Typography
const { RangePicker } = DatePicker

export default function Reports() {
  const [dateRange, setDateRange] = useState<[string, string] | null>(null)
  const [accountFilter, setAccountFilter] = useState<string>('')

  const { data: accounts = [] } = useQuery<{ id: number; account_name: string }[]>({
    queryKey: ['accounts'],
    queryFn: () => api.get('/accounts').then(r => r.data),
  })

  const { data: summary = [], isLoading } = useQuery<DailySummary[]>({
    queryKey: ['reports-summary', dateRange, accountFilter],
    queryFn: () => api.get('/reports/summary', {
      params: {
        ...(dateRange ? { start_date: dateRange[0], end_date: dateRange[1] } : {}),
        ...(accountFilter ? { account_id: accountFilter } : {}),
      }
    }).then(r => r.data),
  })

  const totalCost = summary.reduce((s, d) => s + d.cost, 0)
  const avgROI = summary.length > 0 ? summary.reduce((s, d) => s + d.roi, 0) / summary.length : 0
  const totalConversions = summary.reduce((s, d) => s + d.conversions, 0)

  const funnelData = [
    { stage: '展示', value: summary.reduce((s, d) => s + d.impressions, 0) },
    { stage: '点击', value: summary.reduce((s, d) => s + d.clicks, 0) },
    { stage: '转化', value: totalConversions },
  ]

  return (
    <div>
      <Title level={4}>数据报表</Title>

      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}><Card><Statistic title="总消耗" value={totalCost} prefix="¥" precision={2} /></Card></Col>
        <Col span={6}><Card><Statistic title="平均 ROI" value={avgROI} precision={2} /></Card></Col>
        <Col span={6}><Card><Statistic title="总转化" value={totalConversions} /></Card></Col>
        <Col span={6}><Card><Statistic title="报表天数" value={summary.length} /></Card></Col>
      </Row>

      <Card style={{ marginBottom: 16 }}>
        <Space style={{ marginBottom: 16 }}>
          <RangePicker onChange={(_, dateStrings) => setDateRange(dateStrings as [string, string])} />
          <Select placeholder="按账户筛选" allowClear style={{ width: 200 }}
            options={accounts.map(a => ({ label: a.account_name, value: a.id }))}
            onChange={(v) => setAccountFilter(v || '')} />
        </Space>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={summary}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="date" />
            <YAxis yAxisId="left" />
            <YAxis yAxisId="right" orientation="right" />
            <Tooltip />
            <Legend />
            <Line yAxisId="left" type="monotone" dataKey="cost" stroke="#1890ff" name="消耗" />
            <Line yAxisId="right" type="monotone" dataKey="roi" stroke="#52c41a" name="ROI" />
          </LineChart>
        </ResponsiveContainer>
      </Card>

      <Card title="转化漏斗">
        <Table
          rowKey="stage"
          dataSource={funnelData}
          pagination={false}
          columns={[
            { title: '阶段', dataIndex: 'stage', key: 'stage' },
            { title: '数量', dataIndex: 'value', key: 'value',
              render: (v: number) => v.toLocaleString() },
          ]}
        />
      </Card>
    </div>
  )
}
