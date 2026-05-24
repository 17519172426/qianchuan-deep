import { Card, Statistic } from 'antd'

interface Props {
  title: string
  value: number | string
  prefix?: string
  suffix?: React.ReactNode
  precision?: number
  color?: string
}

export default function StatCard({ title, value, prefix, suffix, precision, color }: Props) {
  return (
    <Card>
      <Statistic
        title={title}
        value={value}
        prefix={prefix}
        suffix={suffix}
        precision={precision}
        valueStyle={color ? { color } : undefined}
      />
    </Card>
  )
}
