import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Table, Tag, Button, Card, Typography, Modal, Form, Input, message } from 'antd'
import { EditOutlined } from '@ant-design/icons'
import api from '../api/client'
import type { Creative } from '../types'

const { Title } = Typography

export default function Creatives() {
  const queryClient = useQueryClient()
  const [modalOpen, setModalOpen] = useState(false)
  const [editingCreative, setEditingCreative] = useState<Creative | null>(null)
  const [form] = Form.useForm()

  const { data: creatives = [], isLoading } = useQuery<Creative[]>({
    queryKey: ['creatives'],
    queryFn: () => api.get('/creatives').then(r => r.data),
  })

  const updateTagsMutation = useMutation({
    mutationFn: ({ id, tags }: { id: number; tags: Record<string, unknown> }) =>
      api.put(`/creatives/${id}/tags`, { tags }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['creatives'] })
      message.success('标签已更新')
      setModalOpen(false)
      setEditingCreative(null)
    },
  })

  const openEditTags = (creative: Creative) => {
    setEditingCreative(creative)
    form.setFieldsValue({ tag_string: Object.keys(creative.tags || {}).join(', ') })
    setModalOpen(true)
  }

  const handleSaveTags = () => {
    form.validateFields().then((values) => {
      const tagArray = values.tag_string.split(',').map((t: string) => t.trim()).filter(Boolean)
      const tags: Record<string, unknown> = {}
      tagArray.forEach((t: string) => { tags[t] = true })
      updateTagsMutation.mutate({ id: editingCreative!.id, tags })
    })
  }

  const typeLabels: Record<string, string> = { video: '视频', image: '图片' }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>素材库</Title>
      </div>
      <Card>
        <Table
          rowKey="id"
          loading={isLoading}
          dataSource={creatives}
          pagination={{ pageSize: 20 }}
          columns={[
            { title: '素材名称', dataIndex: 'name', key: 'name', width: 200 },
            { title: '类型', dataIndex: 'type', key: 'type', width: 80,
              render: (v: string) => <Tag>{typeLabels[v] || v}</Tag> },
            { title: '时长(秒)', dataIndex: 'duration', key: 'duration', width: 100,
              render: (v: number) => v > 0 ? `${v}s` : '-' },
            { title: '文件大小', dataIndex: 'file_size', key: 'file_size', width: 100,
              render: (v: number) => v > 1024 * 1024 ? `${(v / (1024*1024)).toFixed(1)}MB` : `${(v / 1024).toFixed(0)}KB` },
            { title: '标签', dataIndex: 'tags', key: 'tags', width: 200,
              render: (v: Record<string, unknown>) =>
                Object.keys(v || {}).map(t => <Tag key={t}>{t}</Tag>) },
            { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 170,
              render: (v: string) => new Date(v).toLocaleString('zh-CN') },
            { title: '操作', key: 'actions', width: 100,
              render: (_: unknown, record: Creative) => (
                <Button size="small" icon={<EditOutlined />} onClick={() => openEditTags(record)}>标签</Button>
              ),
            },
          ]}
        />
      </Card>
      <Modal title="编辑标签" open={modalOpen}
        onCancel={() => { setModalOpen(false); setEditingCreative(null) }}
        onOk={handleSaveTags} confirmLoading={updateTagsMutation.isPending}>
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="tag_string" label="标签（逗号分隔）">
            <Input placeholder="品牌, 促销, 爆款" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
