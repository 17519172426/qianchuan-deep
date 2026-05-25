import { useQuery } from '@tanstack/react-query'
import { Table, Button, Tag, Typography, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import api from '../api/client'
import type { QianchuanAccount } from '../types'

const { Title } = Typography

export default function Accounts() {
  const { data: accounts = [], isLoading } = useQuery<QianchuanAccount[]>({
    queryKey: ['accounts'],
    queryFn: () => api.get('/accounts').then(r => r.data),
    refetchInterval: 60000,
  })

  const handleAddAccount = async () => {
    try {
      const { data } = await api.get('/accounts/auth-url', {
        params: { redirect_uri: `${window.location.origin}/oauth/callback` }
      })
      window.location.href = data.url
    } catch {
      message.error('获取授权链接失败')
    }
  }

  const statusMap: Record<string, { color: string; label: string }> = {
    active: { color: 'green', label: '正常' },
    inactive: { color: 'red', label: '已禁用' },
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>账户管理</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleAddAccount}>
          添加千川账户
        </Button>
      </div>
      <Table<QianchuanAccount>
        rowKey="id"
        loading={isLoading}
        dataSource={accounts}
        columns={[
          { title: '账户名称', dataIndex: 'account_name', key: 'account_name' },
          { title: '广告主 ID', dataIndex: 'advertiser_id', key: 'advertiser_id' },
          { title: '余额', dataIndex: 'balance', key: 'balance',
            render: (v: number) => `¥${v.toLocaleString(undefined, { minimumFractionDigits: 2 })}` },
          { title: '状态', dataIndex: 'status', key: 'status',
            render: (s: string) => {
              const info = statusMap[s] || { color: 'default', label: s }
              return <Tag color={info.color}>{info.label}</Tag>
            } },
          { title: '最近同步', dataIndex: 'last_sync_at', key: 'last_sync_at',
            render: (v: string) => v ? new Date(v).toLocaleString() : '-' },
          { title: '创建时间', dataIndex: 'created_at', key: 'created_at',
            render: (v: string) => new Date(v).toLocaleString() },
        ]}
      />
    </div>
  )
}
