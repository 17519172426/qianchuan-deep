import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, Tag, Button, Typography, Space, message, Empty, Radio } from 'antd'
import { CheckOutlined, CloseOutlined, BulbOutlined } from '@ant-design/icons'
import api from '../api/client'
import type { AIRecommendation } from '../types'

const { Title, Text, Paragraph } = Typography

const typeLabels: Record<string, string> = {
  budget_opt: '预算优化',
  anomaly: '异常检测',
  roi_predict: 'ROI预测',
  creative_opt: '素材优化',
}

const statusColors: Record<string, string> = {
  pending: 'blue',
  accepted: 'green',
  ignored: 'default',
}

export default function Recommendations() {
  const queryClient = useQueryClient()
  const [filter, setFilter] = useState('pending')

  const { data: recs = [], isLoading } = useQuery<AIRecommendation[]>({
    queryKey: ['recommendations', filter],
    queryFn: () => api.get('/recommendations', { params: { status: filter } }).then(r => r.data),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, status }: { id: number; status: string }) =>
      api.patch(`/recommendations/${id}/status`, { status }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['recommendations'] })
      message.success('已更新')
    },
  })

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>AI 推荐</Title>
        <Radio.Group value={filter} onChange={e => setFilter(e.target.value)}>
          <Radio.Button value="pending">待处理</Radio.Button>
          <Radio.Button value="accepted">已采纳</Radio.Button>
          <Radio.Button value="ignored">已忽略</Radio.Button>
        </Radio.Group>
      </div>

      {recs.length === 0 && !isLoading && <Empty description="暂无推荐" />}

      {recs.map(rec => (
        <Card key={rec.id} style={{ marginBottom: 12 }} size="small">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <div style={{ flex: 1 }}>
              <Space>
                <BulbOutlined style={{ color: '#faad14' }} />
                <Text strong>{rec.title}</Text>
                <Tag>{typeLabels[rec.type] || rec.type}</Tag>
                <Tag color={statusColors[rec.status]}>{rec.status}</Tag>
              </Space>
              <Paragraph style={{ marginTop: 8, color: '#666' }}>{rec.description}</Paragraph>
              <Text type="secondary">
                置信度: {(rec.confidence * 100).toFixed(0)}% | 计划ID: {rec.uni_ad_id ?? '-'}
              </Text>
            </div>
            {rec.status === 'pending' && (
              <Space style={{ marginLeft: 16 }}>
                <Button type="primary" size="small" icon={<CheckOutlined />}
                  onClick={() => updateMutation.mutate({ id: rec.id, status: 'accepted' })}>
                  采纳
                </Button>
                <Button size="small" icon={<CloseOutlined />}
                  onClick={() => updateMutation.mutate({ id: rec.id, status: 'ignored' })}>
                  忽略
                </Button>
              </Space>
            )}
          </div>
        </Card>
      ))}
    </div>
  )
}
