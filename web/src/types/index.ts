export interface User {
  id: number
  name: string
  email: string
  role: string
}

export interface QianchuanAccount {
  id: number
  account_name: string
  advertiser_id: number
  status: string
  balance: number
  last_sync_at: string | null
  created_at: string
}

export interface UniAd {
  id: number
  account_id: number
  qianchuan_ad_id: number | null
  name: string
  marketing_goal: string
  aweme_id: number | null
  product_ids: Record<string, unknown>
  delivery_setting: Record<string, unknown>
  creative_setting: Record<string, unknown>
  status: string
  metrics_json: Record<string, unknown>
  created_at: string
  updated_at: string
  account?: QianchuanAccount
}

export interface DashboardStats {
  today_cost: number
  avg_roi: number
  total_conversions: number
  active_ads: number
  total_accounts: number
}

export interface TrendPoint {
  date: string
  cost: number
  roi: number
}

export interface UniAdReport {
  id: number
  uni_ad_id: number
  report_date: string
  report_hour: number
  impressions: number
  clicks: number
  cost: number
  conversions: number
  roi: number
  ctr: number
  ecpm: number
  pay_order_cnt: number
  pay_order_amt: number
}

export interface DailySummary {
  date: string
  cost: number
  impressions: number
  clicks: number
  conversions: number
  roi: number
}

export interface Rule {
  id: number
  name: string
  description: string
  account_id: number
  scope_json: Record<string, unknown>
  condition_json: Record<string, unknown>
  action_json: Record<string, unknown>
  schedule: string
  cooldown: string
  enabled: boolean
  created_at: string
}

export interface RuleExecution {
  id: number
  rule_id: number
  uni_ad_id: number
  triggered_at: string
  condition_json: Record<string, unknown>
  action_json: Record<string, unknown>
  status: string
  result_json: Record<string, unknown>
  executed_at: string | null
}
