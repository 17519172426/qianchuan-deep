import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Table, Tag, Button, Card, Typography, Modal, Form, Input, Select, Switch, message, Space, Popconfirm, Drawer } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, HistoryOutlined } from '@ant-design/icons'
import TextArea from 'antd/es/input/TextArea'
import api from '../api/client'
import type { Rule, RuleExecution } from '../types'

const { Title } = Typography

const actionOptions = [
  { label: '暂停计划', value: 'pause_ad' },
  { label: '恢复计划', value: 'resume_ad' },
  { label: '调整预算', value: 'update_budget' },
  { label: '调整ROI目标', value: 'update_roi_goal' },
  { label: '通知提醒', value: 'notify' },
]

const metricOptions = [
  { label: 'ROI', value: 'roi' },
  { label: '消耗', value: 'cost' },
  { label: '点击率(CTR)', value: 'ctr' },
  { label: '转化数', value: 'conversions' },
  { label: '展示量', value: 'impressions' },
]

const operatorOptions = [
  { label: '大于 (>)', value: 'gt' },
  { label: '小于 (<)', value: 'lt' },
  { label: '大于等于 (>=)', value: 'gte' },
  { label: '小于等于 (<=)', value: 'lte' },
  { label: '等于 (=)', value: 'eq' },
]

export default function Rules() {
  const queryClient = useQueryClient()
  const [modalOpen, setModalOpen] = useState(false)
  const [execDrawer, setExecDrawer] = useState(false)
  const [editingRule, setEditingRule] = useState<Rule | null>(null)
  const [viewRuleId, setViewRuleId] = useState<number | null>(null)
  const [form] = Form.useForm()

  const { data: rules = [], isLoading } = useQuery<Rule[]>({
    queryKey: ['rules'],
    queryFn: () => api.get('/rules').then(r => r.data),
  })

  const { data: accounts = [] } = useQuery<{ id: number; account_name: string }[]>({
    queryKey: ['accounts'],
    queryFn: () => api.get('/accounts').then(r => r.data),
  })

  const { data: executions = [] } = useQuery<RuleExecution[]>({
    queryKey: ['rule-executions', viewRuleId],
    queryFn: () => api.get('/rules/executions', { params: { rule_id: viewRuleId } }).then(r => r.data),
    enabled: execDrawer && viewRuleId !== null,
  })

  const saveMutation = useMutation({
    mutationFn: (values: Record<string, unknown>) =>
      editingRule
        ? api.put(`/rules/${editingRule.id}`, values)
        : api.post('/rules', values),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rules'] })
      message.success(editingRule ? '规则已更新' : '规则已创建')
      setModalOpen(false)
      setEditingRule(null)
      form.resetFields()
    },
    onError: (err: any) => message.error(err?.response?.data?.error || '操作失败'),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => api.delete(`/rules/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rules'] })
      message.success('规则已删除')
    },
  })

  const toggleMutation = useMutation({
    mutationFn: ({ id, enabled }: { id: number; enabled: boolean }) =>
      api.put(`/rules/${id}`, { enabled }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['rules'] }),
  })

  const openCreate = () => {
    setEditingRule(null)
    form.resetFields()
    setModalOpen(true)
  }

  const openEdit = (rule: Rule) => {
    setEditingRule(rule)
    form.setFieldsValue({
      name: rule.name,
      description: rule.description,
      account_id: rule.account_id,
      condition_metric: (rule.condition_json as any)?.metric || 'roi',
      condition_operator: (rule.condition_json as any)?.operator || 'lt',
      condition_threshold: (rule.condition_json as any)?.threshold || 0,
      action_type: (rule.action_json as any)?.type || 'pause_ad',
      action_value: (rule.action_json as any)?.value || 0,
      action_value_type: (rule.action_json as any)?.value_type || 'absolute',
    })
    setModalOpen(true)
  }

  const handleSave = () => {
    form.validateFields().then((values) => {
      saveMutation.mutate({
        ...values,
        account_id: Number(values.account_id),
        condition_json: {
          metric: values.condition_metric,
          operator: values.condition_operator,
          threshold: Number(values.condition_threshold),
          duration: '30m',
        },
        action_json: {
          type: values.action_type,
          value: Number(values.action_value),
          value_type: values.action_value_type,
        },
      })
    })
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>规则管理</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建规则</Button>
      </div>

      <Card>
        <Table
          rowKey="id"
          loading={isLoading}
          dataSource={rules}
          pagination={{ pageSize: 20 }}
          columns={[
            { title: '规则名称', dataIndex: 'name', key: 'name', width: 180 },
            {
              title: '状态', dataIndex: 'enabled', key: 'enabled', width: 80,
              render: (v: boolean, record: Rule) => (
                <Switch checked={v} size="small"
                  onChange={(enabled) => toggleMutation.mutate({ id: record.id, enabled })} />
              ),
            },
            {
              title: '条件', key: 'condition', width: 200,
              render: (_: unknown, r: Rule) => {
                const c = r.condition_json as any
                const opLabel = operatorOptions.find(o => o.value === c?.operator)?.label || c?.operator
                return `${c?.metric || '-'} ${opLabel} ${c?.threshold || ''}`
              },
            },
            {
              title: '动作', key: 'action', width: 150,
              render: (_: unknown, r: Rule) => {
                const a = r.action_json as any
                return actionOptions.find(o => o.value === a?.type)?.label || a?.type || '-'
              },
            },
            { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 170,
              render: (v: string) => new Date(v).toLocaleString('zh-CN') },
            {
              title: '操作', key: 'actions', width: 200,
              render: (_: unknown, record: Rule) => (
                <Space>
                  <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(record)}>编辑</Button>
                  <Button size="small" icon={<HistoryOutlined />}
                    onClick={() => { setViewRuleId(record.id); setExecDrawer(true) }}>日志</Button>
                  <Popconfirm title="确认删除？" onConfirm={() => deleteMutation.mutate(record.id)}>
                    <Button size="small" danger icon={<DeleteOutlined />} />
                  </Popconfirm>
                </Space>
              ),
            },
          ]}
        />
      </Card>

      <Modal
        title={editingRule ? '编辑规则' : '新建规则'}
        open={modalOpen}
        onCancel={() => { setModalOpen(false); setEditingRule(null); form.resetFields() }}
        onOk={handleSave}
        confirmLoading={saveMutation.isPending}
        width={560}
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="name" label="规则名称" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="如：低ROI自动暂停" />
          </Form.Item>
          <Form.Item name="description" label="规则说明">
            <TextArea rows={2} placeholder="可选说明" />
          </Form.Item>
          <Form.Item name="account_id" label="所属账户" rules={[{ required: true, message: '请选择账户' }]}>
            <Select options={accounts.map(a => ({ label: a.account_name, value: a.id }))} placeholder="选择千川账户" />
          </Form.Item>
          <Form.Item label="触发条件">
            <Space.Compact style={{ width: '100%' }}>
              <Form.Item name="condition_metric" noStyle initialValue="roi">
                <Select options={metricOptions} style={{ width: 130 }} />
              </Form.Item>
              <Form.Item name="condition_operator" noStyle initialValue="lt">
                <Select options={operatorOptions} style={{ width: 120 }} />
              </Form.Item>
              <Form.Item name="condition_threshold" noStyle rules={[{ required: true, message: '请输入阈值' }]}>
                <Input type="number" step={0.1} placeholder="阈值" style={{ flex: 1 }} />
              </Form.Item>
            </Space.Compact>
          </Form.Item>
          <Form.Item label="执行动作">
            <Space.Compact style={{ width: '100%' }}>
              <Form.Item name="action_type" noStyle initialValue="pause_ad">
                <Select options={actionOptions} style={{ width: 160 }} />
              </Form.Item>
              <Form.Item name="action_value" noStyle initialValue={0}>
                <Input type="number" step={0.1} placeholder="参数值（可选）" style={{ flex: 1 }} />
              </Form.Item>
              <Form.Item name="action_value_type" noStyle initialValue="absolute">
                <Select options={[
                  { label: '绝对值', value: 'absolute' },
                  { label: '百分比', value: 'percentage' },
                ]} style={{ width: 110 }} />
              </Form.Item>
            </Space.Compact>
          </Form.Item>
        </Form>
      </Modal>

      <Drawer
        title={`执行日志 (规则 #${viewRuleId})`}
        open={execDrawer}
        onClose={() => { setExecDrawer(false); setViewRuleId(null) }}
        width={600}
      >
        {executions.map(e => (
          <Card key={e.id} size="small" style={{ marginBottom: 8 }}>
            <p><strong>时间:</strong> {new Date(e.triggered_at).toLocaleString('zh-CN')}</p>
            <p><strong>计划ID:</strong> {e.uni_ad_id}</p>
            <p><strong>动作:</strong> {(e.action_json as any)?.type || '-'}</p>
            <p><strong>状态:</strong> <Tag color={e.status === 'success' ? 'green' : 'red'}>{e.status}</Tag></p>
            {e.result_json && Object.keys(e.result_json as any).length > 0 && (
              <p><strong>结果:</strong> {JSON.stringify(e.result_json)}</p>
            )}
          </Card>
        ))}
        {executions.length === 0 && <p style={{ color: '#999' }}>暂无执行记录</p>}
      </Drawer>
    </div>
  )
}
