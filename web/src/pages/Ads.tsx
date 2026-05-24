import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Table, Tag, Select, Button, Space, message, Card, Typography } from 'antd'
import { PlayCircleOutlined, PauseCircleOutlined, DeleteOutlined } from '@ant-design/icons'
import api from '../api/client'
import type { UniAd } from '../types'

const { Title } = Typography

const statusMap: Record<string, { color: string; label: string }> = {
  enable: { color: 'green', label: '投放中' },
  disable: { color: 'orange', label: '已暂停' },
  delete: { color: 'red', label: '已删除' },
  create: { color: 'blue', label: '新建' },
}

export default function Ads() {
  const [accountFilter, setAccountFilter] = useState<string>('')
  const queryClient = useQueryClient()

  const { data: ads = [], isLoading } = useQuery<UniAd[]>({
    queryKey: ['ads', accountFilter],
    queryFn: () => api.get('/ads', { params: { account_id: accountFilter || undefined } }).then(r => r.data),
    refetchInterval: 30_000,
  })

  const { data: accounts = [] } = useQuery<{ id: number; account_name: string }[]>({
    queryKey: ['accounts'],
    queryFn: () => api.get('/accounts').then(r => r.data),
  })

  const statusMutation = useMutation({
    mutationFn: ({ id, status }: { id: number; status: string }) =>
      api.patch(`/ads/${id}/status`, { status }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ads'] })
      message.success('状态更新成功')
    },
    onError: (err: any) => message.error(err?.response?.data?.error || '操作失败'),
  })

  const columns = [
    {
      title: '计划名称',
      dataIndex: 'name',
      key: 'name',
      width: 200,
    },
    {
      title: '账户',
      dataIndex: ['account', 'account_name'],
      key: 'account',
      width: 150,
    },
    {
      title: '营销目标',
      dataIndex: 'marketing_goal',
      key: 'marketing_goal',
      width: 120,
      render: (v: string) => v === 'LIVE_PROM_GOODS' ? '直播带货' : '短视频带货',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (s: string) => {
        const cfg = statusMap[s] || { color: 'default', label: s }
        return <Tag color={cfg.color}>{cfg.label}</Tag>
      },
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (v: string) => new Date(v).toLocaleString('zh-CN'),
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      render: (_: unknown, record: UniAd) => (
        <Space>
          {record.status !== 'enable' && (
            <Button size="small" icon={<PlayCircleOutlined />}
              onClick={() => statusMutation.mutate({ id: record.id, status: 'enable' })}>
              开启
            </Button>
          )}
          {record.status === 'enable' && (
            <Button size="small" icon={<PauseCircleOutlined />}
              onClick={() => statusMutation.mutate({ id: record.id, status: 'disable' })}>
              暂停
            </Button>
          )}
          {record.status !== 'delete' && (
            <Button size="small" danger icon={<DeleteOutlined />}
              onClick={() => statusMutation.mutate({ id: record.id, status: 'delete' })}>
              删除
            </Button>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>全域计划管理</Title>
        <Select
          allowClear
          placeholder="筛选账户"
          style={{ width: 200 }}
          value={accountFilter || undefined}
          onChange={(v) => setAccountFilter(v || '')}
          options={accounts.map(a => ({ label: a.account_name, value: a.id }))}
        />
      </div>
      <Card>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={ads}
          loading={isLoading}
          pagination={{ pageSize: 20, showTotal: (total) => `共 ${total} 条` }}
        />
      </Card>
    </div>
  )
}
