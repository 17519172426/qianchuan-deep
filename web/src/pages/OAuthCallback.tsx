import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { message, Spin, Result } from 'antd'
import api from '../api/client'

export default function OAuthCallback() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const [error, setError] = useState('')

  useEffect(() => {
    const authCode = searchParams.get('auth_code')
    const advertiserId = searchParams.get('advertiser_id')

    if (!authCode || !advertiserId) {
      setError('授权回调参数缺失（auth_code 或 advertiser_id）')
      return
    }

    api.post('/accounts', {
      account_name: `千川账户 ${advertiserId}`,
      advertiser_id: Number(advertiserId),
      auth_code: authCode,
    })
    .then(() => {
      message.success('千川账户授权成功')
      navigate('/accounts')
    })
    .catch(err => {
      setError(err?.response?.data?.error || '授权失败，请重试')
    })
  }, [])

  if (error) {
    return (
      <Result
        status="error"
        title="授权失败"
        subTitle={error}
        extra={<a href="/accounts">返回账户管理</a>}
      />
    )
  }

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh' }}>
      <Spin size="large" tip="正在完成千川账户授权..." />
    </div>
  )
}
