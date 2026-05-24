# 千川投流 SaaS — Phase 2 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 构建前端基础页面（看板、全域计划管理）+ 后端报表 API + 数据同步增强，使系统可用

**Architecture:** React 19 SPA (Vite + TypeScript + Ant Design 5 + React Query) 通过 REST API 与 Go 后端通信。后端新增报表模型、看板聚合 API 和报表查询 API，同步 Worker 增强为同时拉取小时级报表数据。

**Tech Stack:** React 19, TypeScript, Vite, Ant Design 5, React Query (TanStack), Recharts, Axios; Go 1.25, Gin, GORM

**Spec:** `docs/superpowers/specs/2026-05-24-qianchuan-saas-design.md`

---

## File Structure

```
server/
├── models/
│   └── report.go               # NEW: UniAdReport 时序报表模型
├── handler/
│   ├── dashboard.go            # NEW: 看板聚合 API
│   └── report.go               # NEW: 报表查询 API
├── worker/
│   └── sync.go                 # MODIFY: 增加报表数据同步
├── main.go                     # MODIFY: AutoMigrate 加 report 表
└── router/
    └── router.go               # MODIFY: 新路由 + CORS

web/                             # NEW: React 前端项目
├── index.html
├── package.json
├── tsconfig.json
├── tsconfig.app.json
├── tsconfig.node.json
├── vite.config.ts
├── src/
│   ├── main.tsx
│   ├── App.tsx
│   ├── api/
│   │   └── client.ts           # axios 实例 + JWT 拦截器
│   ├── types/
│   │   └── index.ts            # 共享 TypeScript 类型
│   ├── pages/
│   │   ├── Login.tsx           # 登录页
│   │   ├── Dashboard.tsx       # 首页看板
│   │   └── Ads.tsx             # 全域计划管理
│   └── components/
│       ├── Layout.tsx          # 侧边栏布局 + 路由守卫
│       └── StatCard.tsx        # 统计卡片
```

---

### Task 1: UniAdReport Model

**Files:**
- Create: `server/models/report.go`

- [ ] **Step 1: Write models/report.go**

```go
package models

import "time"

type UniAdReport struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UniAdID      uint      `gorm:"index;not null" json:"uni_ad_id"`
	ReportDate   time.Time `gorm:"index;not null" json:"report_date"`
	ReportHour   int       `gorm:"default:0" json:"report_hour"`
	Impressions  int64     `gorm:"default:0" json:"impressions"`
	Clicks       int64     `gorm:"default:0" json:"clicks"`
	Cost         float64   `gorm:"type:decimal(15,2);default:0" json:"cost"`
	Conversions  int       `gorm:"default:0" json:"conversions"`
	ROI          float64   `gorm:"type:decimal(10,4);default:0" json:"roi"`
	CTR          float64   `gorm:"type:decimal(10,4);default:0" json:"ctr"`
	ECPM         float64   `gorm:"type:decimal(15,4);default:0" json:"ecpm"`
	PayOrderCnt  int       `gorm:"default:0" json:"pay_order_cnt"`
	PayOrderAmt  float64   `gorm:"type:decimal(15,2);default:0" json:"pay_order_amt"`
}
```

- [ ] **Step 2: Update main.go AutoMigrate**

Add `&models.UniAdReport{}` to the AutoMigrate call in `server/main.go`. The AutoMigrate block becomes:

```go
db.AutoMigrate(
	&models.User{},
	&models.QianchuanAccount{},
	&models.UniAd{},
	&models.Creative{},
	&models.UniAdCreative{},
	&models.Rule{},
	&models.RuleExecution{},
	&models.AIRecommendation{},
	&models.UniAdReport{},
)
```

- [ ] **Step 3: Build and commit**

```bash
cd server && go build ./...
```

Expected: no errors.

```bash
git add server/models/report.go server/main.go
git commit -m "feat: add UniAdReport model for hourly ad performance data"
```

---

### Task 2: Dashboard Stats API

**Files:**
- Create: `server/handler/dashboard.go`

- [ ] **Step 1: Write handler/dashboard.go**

```go
package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/example/qianchuan-saas/db"
	"github.com/example/qianchuan-saas/models"
)

type DashboardHandler struct{}

type DashboardStats struct {
	TodayCost      float64 `json:"today_cost"`
	AvgROI         float64 `json:"avg_roi"`
	TotalConversions int64 `json:"total_conversions"`
	ActiveAds      int64   `json:"active_ads"`
	TotalAccounts  int64   `json:"total_accounts"`
}

func (h *DashboardHandler) Stats(c *gin.Context) {
	today := time.Now().Format("2006-01-02")
	var stats DashboardStats

	db.DB.Model(&models.UniAdReport{}).
		Where("report_date = ?", today).
		Select("COALESCE(SUM(cost), 0) as today_cost").
		Scan(&stats.TodayCost)

	db.DB.Model(&models.UniAdReport{}).
		Where("report_date = ?", today).
		Select("COALESCE(AVG(roi), 0) as avg_roi").
		Scan(&stats.AvgROI)

	db.DB.Model(&models.UniAdReport{}).
		Where("report_date = ?", today).
		Select("COALESCE(SUM(conversions), 0) as total_conversions").
		Scan(&stats.TotalConversions)

	db.DB.Model(&models.UniAd{}).
		Where("status = ?", "enable").
		Count(&stats.ActiveAds)

	db.DB.Model(&models.QianchuanAccount{}).Count(&stats.TotalAccounts)

	c.JSON(http.StatusOK, stats)
}

type TrendPoint struct {
	Date string  `json:"date"`
	Cost float64 `json:"cost"`
	ROI  float64 `json:"roi"`
}

func (h *DashboardHandler) Trend(c *gin.Context) {
	var points []TrendPoint
	db.DB.Model(&models.UniAdReport{}).
		Select("report_date::text as date, SUM(cost) as cost, AVG(roi) as roi").
		Where("report_date >= CURRENT_DATE - INTERVAL '7 days'").
		Group("report_date").
		Order("report_date ASC").
		Scan(&points)
	c.JSON(http.StatusOK, points)
}
```

- [ ] **Step 2: Build and commit**

```bash
cd server && go build ./...
```

```bash
git add server/handler/dashboard.go
git commit -m "feat: add dashboard stats and trend API endpoints"
```

---

### Task 3: Reports API

**Files:**
- Create: `server/handler/report.go`

- [ ] **Step 1: Write handler/report.go**

```go
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/example/qianchuan-saas/db"
	"github.com/example/qianchuan-saas/models"
)

type ReportHandler struct{}

func (h *ReportHandler) ByAd(c *gin.Context) {
	adID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ad id"})
		return
	}

	days := c.DefaultQuery("days", "7")

	var reports []models.UniAdReport
	db.DB.Where("uni_ad_id = ? AND report_date >= CURRENT_DATE - INTERVAL '1 day' * ?::int", adID, days).
		Order("report_date DESC, report_hour DESC").
		Find(&reports)
	c.JSON(http.StatusOK, reports)
}

func (h *ReportHandler) SummaryByDate(c *gin.Context) {
	accountID := c.Query("account_id")
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0,0,-7).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	type DailySummary struct {
		Date        string  `json:"date"`
		Cost        float64 `json:"cost"`
		Impressions int64   `json:"impressions"`
		Clicks      int64   `json:"clicks"`
		Conversions int     `json:"conversions"`
		ROI         float64 `json:"roi"`
	}

	var summaries []DailySummary
	q := db.DB.Model(&models.UniAdReport{}).
		Select("report_date::text as date, SUM(cost) as cost, SUM(impressions) as impressions, SUM(clicks) as clicks, SUM(conversions) as conversions, AVG(roi) as roi").
		Where("report_date BETWEEN ? AND ?", startDate, endDate)

	if accountID != "" {
		q = q.Where("uni_ad_id IN (SELECT id FROM uni_ads WHERE account_id = ?)", accountID)
	}

	q.Group("report_date").Order("report_date ASC").Scan(&summaries)
	c.JSON(http.StatusOK, summaries)
}
```

- [ ] **Step 2: Build and commit**

```bash
cd server && go build ./...
```

```bash
git add server/handler/report.go
git commit -m "feat: add reports API endpoints (by ad, daily summary)"
```

---

### Task 4: Router Update + Sync Worker Enhancement

**Files:**
- Modify: `server/router/router.go`
- Modify: `server/worker/sync.go`

- [ ] **Step 1: Update router.go — add dashboard and report routes**

Add new handler instances and routes. Replace router.go content:

```go
package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/example/qianchuan-saas/auth"
	"github.com/example/qianchuan-saas/handler"
	"github.com/example/qianchuan-saas/qianchuan"
)

func Setup(qc *qianchuan.Client) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	authH := &handler.AuthHandler{}
	accountH := &handler.AccountHandler{QC: qc}
	adH := &handler.AdHandler{QC: qc}
	dashH := &handler.DashboardHandler{}
	reportH := &handler.ReportHandler{}

	api := r.Group("/api")
	{
		api.POST("/register", authH.Register)
		api.POST("/login", authH.Login)

		authorized := api.Group("/", auth.AuthRequired())
		{
			authorized.GET("/accounts", accountH.List)
			authorized.POST("/accounts", accountH.Create)
			authorized.GET("/accounts/:id", accountH.Get)
			authorized.DELETE("/accounts/:id", accountH.Delete)

			authorized.GET("/ads", adH.List)
			authorized.POST("/ads", adH.Create)
			authorized.GET("/ads/:id", adH.Get)
			authorized.PATCH("/ads/:id/status", adH.UpdateStatus)

			authorized.GET("/dashboard/stats", dashH.Stats)
			authorized.GET("/dashboard/trend", dashH.Trend)

			authorized.GET("/reports/ads/:id", reportH.ByAd)
			authorized.GET("/reports/summary", reportH.SummaryByDate)
		}
	}

	return r
}
```

- [ ] **Step 2: Update sync.go — add report syncing**

Add `syncReports` method to `SyncWorker` and call it in the ticker. Updated `server/worker/sync.go`:

```go
package worker

import (
	"encoding/json"
	"log"
	"time"

	"github.com/example/qianchuan-saas/db"
	"github.com/example/qianchuan-saas/models"
	"github.com/example/qianchuan-saas/qianchuan"
)

type SyncWorker struct {
	QC       *qianchuan.Client
	Interval time.Duration
}

func NewSyncWorker(qc *qianchuan.Client) *SyncWorker {
	return &SyncWorker{QC: qc, Interval: 5 * time.Minute}
}

func (w *SyncWorker) Start() {
	log.Printf("sync worker started, interval=%s", w.Interval)
	ticker := time.NewTicker(w.Interval)
	go func() {
		for range ticker.C {
			w.syncAds()
			w.syncReports()
		}
	}()
}

func (w *SyncWorker) syncAds() {
	var accounts []models.QianchuanAccount
	db.DB.Find(&accounts)

	for _, acc := range accounts {
		accRef := qianchuan.AdAccount{ID: acc.ID, AdvertiserID: acc.AdvertiserID}
		resp, err := w.QC.ListUniAds(&accRef, 1, 100)
		if err != nil {
			log.Printf("sync ads failed for account %d: %v", acc.ID, err)
			continue
		}
		if resp.Code != 0 {
			log.Printf("sync ads error for account %d: code=%d msg=%s", acc.ID, resp.Code, resp.Message)
			continue
		}
		var result struct {
			List []struct {
				AdID    int64                  `json:"ad_id"`
				Name    string                 `json:"name"`
				Status  string                 `json:"status"`
				Metrics map[string]interface{} `json:"metrics"`
			} `json:"list"`
		}
		if err := json.Unmarshal(resp.Data, &result); err != nil {
			log.Printf("parse ads response failed: %v", err)
			continue
		}
		for _, item := range result.List {
			var ad models.UniAd
			if err := db.DB.Where("qianchuan_ad_id = ? AND account_id = ?", item.AdID, acc.ID).First(&ad).Error; err != nil {
				continue
			}
			updates := map[string]interface{}{"status": item.Status}
			if item.Metrics != nil {
				m := models.JSONMap(item.Metrics)
				updates["metrics_json"] = m
			}
			db.DB.Model(&ad).Updates(updates)
		}
		db.DB.Model(&acc).Update("last_sync_at", time.Now())
	}
}

func (w *SyncWorker) syncReports() {
	var ads []models.UniAd
	if err := db.DB.Where("qianchuan_ad_id IS NOT NULL").Find(&ads).Error; err != nil {
		log.Printf("sync reports: fetch ads failed: %v", err)
		return
	}

	for _, ad := range ads {
		var acc models.QianchuanAccount
		if err := db.DB.First(&acc, ad.AccountID).Error; err != nil {
			continue
		}

		accRef := qianchuan.AdAccount{ID: acc.ID, AdvertiserID: acc.AdvertiserID}
		resp, err := w.QC.GetUniAdDetail(&accRef, *ad.QianchuanAdID)
		if err != nil || resp.Code != 0 {
			continue
		}

		var detail struct {
			Metrics struct {
				Impressions  int64   `json:"impressions"`
				Clicks       int64   `json:"clicks"`
				Cost         float64 `json:"cost"`
				Conversions  int     `json:"conversions"`
				ROI          float64 `json:"roi"`
				CTR          float64 `json:"ctr"`
				ECPM         float64 `json:"ecpm"`
				PayOrderCnt  int     `json:"pay_order_cnt"`
				PayOrderAmt  float64 `json:"pay_order_amt"`
			} `json:"metrics"`
		}
		if err := json.Unmarshal(resp.Data, &detail); err != nil {
			continue
		}

		now := time.Now()
		report := models.UniAdReport{
			UniAdID:     ad.ID,
			ReportDate:  time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
			ReportHour:  now.Hour(),
			Impressions: detail.Metrics.Impressions,
			Clicks:      detail.Metrics.Clicks,
			Cost:        detail.Metrics.Cost,
			Conversions: detail.Metrics.Conversions,
			ROI:         detail.Metrics.ROI,
			CTR:         detail.Metrics.CTR,
			ECPM:        detail.Metrics.ECPM,
			PayOrderCnt: detail.Metrics.PayOrderCnt,
			PayOrderAmt: detail.Metrics.PayOrderAmt,
		}
		db.DB.Create(&report)
	}
}
```

- [ ] **Step 3: Build and commit**

```bash
cd server && go build ./...
```

```bash
git add server/router/router.go server/worker/sync.go
git commit -m "feat: add dashboard/report routes and sync worker report polling"
```

---

### Task 5: React Frontend Scaffold

**Files:**
- Create: `web/package.json`, `web/index.html`, `web/vite.config.ts`, `web/tsconfig.json`, `web/tsconfig.app.json`, `web/tsconfig.node.json`
- Create: `web/src/main.tsx`, `web/src/App.tsx`

- [ ] **Step 1: Create package.json**

```json
{
  "name": "qianchuan-web",
  "private": true,
  "version": "1.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc -b && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "antd": "^5.22.0",
    "@ant-design/icons": "^5.5.0",
    "react": "^19.0.0",
    "react-dom": "^19.0.0",
    "react-router-dom": "^7.0.0",
    "@tanstack/react-query": "^5.60.0",
    "axios": "^1.7.0",
    "recharts": "^2.13.0",
    "dayjs": "^1.11.0"
  },
  "devDependencies": {
    "@types/react": "^19.0.0",
    "@types/react-dom": "^19.0.0",
    "@vitejs/plugin-react": "^4.3.0",
    "typescript": "~5.6.0",
    "vite": "^6.0.0"
  }
}
```

- [ ] **Step 2: Create config files**

`web/index.html`:
```html
<!DOCTYPE html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>千川投流助手</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

`web/vite.config.ts`:
```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
```

`web/tsconfig.json`:
```json
{
  "files": [],
  "references": [
    { "path": "./tsconfig.app.json" },
    { "path": "./tsconfig.node.json" }
  ]
}
```

`web/tsconfig.app.json`:
```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForExpose": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "isolatedModules": true,
    "moduleDetection": "force",
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": false,
    "noUnusedParameters": false,
    "noFallthroughCasesInSwitch": true,
    "forceConsistentCasingInFileNames": true
  },
  "include": ["src"]
}
```

`web/tsconfig.node.json`:
```json
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["ES2023"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "isolatedModules": true,
    "moduleDetection": "force",
    "noEmit": true,
    "strict": true,
    "noUnusedLocals": false,
    "noUnusedParameters": false,
    "noFallthroughCasesInSwitch": true,
    "forceConsistentCasingInFileNames": true
  },
  "include": ["vite.config.ts"]
}
```

- [ ] **Step 3: Create entry files**

`web/src/main.tsx`:
```tsx
import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import App from './App'

const queryClient = new QueryClient()

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ConfigProvider locale={zhCN}>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <App />
        </BrowserRouter>
      </QueryClientProvider>
    </ConfigProvider>
  </React.StrictMode>,
)
```

`web/src/App.tsx`:
```tsx
import { Routes, Route, Navigate } from 'react-router-dom'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Ads from './pages/Ads'
import Layout from './components/Layout'

function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route element={<Layout />}>
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/ads" element={<Ads />} />
      </Route>
    </Routes>
  )
}

export default App
```

- [ ] **Step 4: Install dependencies and verify**

```bash
cd web && npm install && npx tsc --noEmit
```

Expected: no TypeScript errors.

- [ ] **Step 5: Commit**

```bash
git add web/
git commit -m "chore: scaffold React frontend with Vite, TypeScript, Ant Design"
```

---

### Task 6: API Client + Shared Types

**Files:**
- Create: `web/src/api/client.ts`
- Create: `web/src/types/index.ts`

- [ ] **Step 1: Write types/index.ts**

```typescript
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
```

- [ ] **Step 2: Write api/client.ts**

```typescript
import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 15000,
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  },
)

export default api
```

- [ ] **Step 3: Commit**

```bash
git add web/src/api/client.ts web/src/types/index.ts
git commit -m "feat: add API client with JWT interceptor and shared types"
```

---

### Task 7: Login Page

**Files:**
- Create: `web/src/pages/Login.tsx`

- [ ] **Step 1: Write pages/Login.tsx**

```tsx
import { useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Form, Input, Button, Card, Typography, message, Tabs } from 'antd'
import { UserOutlined, LockOutlined, MailOutlined } from '@ant-design/icons'
import api from '../api/client'

const { Title } = Typography

export default function Login() {
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()

  const handleSubmit = async (values: Record<string, string>, mode: 'login' | 'register') => {
    setLoading(true)
    try {
      const endpoint = mode === 'login' ? '/login' : '/register'
      const { data } = await api.post(endpoint, values)
      if (mode === 'login') {
        localStorage.setItem('token', data.token)
        localStorage.setItem('user', JSON.stringify(data.user))
        message.success('登录成功')
        navigate(searchParams.get('redirect') || '/dashboard')
      } else {
        message.success('注册成功，请登录')
      }
    } catch (err: any) {
      const msg = err?.response?.data?.error || '操作失败'
      message.error(msg)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#f0f2f5' }}>
      <Card style={{ width: 400 }}>
        <Title level={3} style={{ textAlign: 'center', marginBottom: 24 }}>千川投流助手</Title>
        <Tabs
          centered
          items={[
            {
              key: 'login',
              label: '登录',
              children: (
                <Form onFinish={(v) => handleSubmit(v, 'login')} size="large">
                  <Form.Item name="email" rules={[{ required: true, type: 'email', message: '请输入邮箱' }]}>
                    <Input prefix={<MailOutlined />} placeholder="邮箱" />
                  </Form.Item>
                  <Form.Item name="password" rules={[{ required: true, min: 6, message: '密码至少6位' }]}>
                    <Input.Password prefix={<LockOutlined />} placeholder="密码" />
                  </Form.Item>
                  <Form.Item>
                    <Button type="primary" htmlType="submit" loading={loading} block>登录</Button>
                  </Form.Item>
                </Form>
              ),
            },
            {
              key: 'register',
              label: '注册',
              children: (
                <Form onFinish={(v) => handleSubmit(v, 'register')} size="large">
                  <Form.Item name="name" rules={[{ required: true, message: '请输入姓名' }]}>
                    <Input prefix={<UserOutlined />} placeholder="姓名" />
                  </Form.Item>
                  <Form.Item name="email" rules={[{ required: true, type: 'email', message: '请输入邮箱' }]}>
                    <Input prefix={<MailOutlined />} placeholder="邮箱" />
                  </Form.Item>
                  <Form.Item name="password" rules={[{ required: true, min: 6, message: '密码至少6位' }]}>
                    <Input.Password prefix={<LockOutlined />} placeholder="密码" />
                  </Form.Item>
                  <Form.Item>
                    <Button type="primary" htmlType="submit" loading={loading} block>注册</Button>
                  </Form.Item>
                </Form>
              ),
            },
          ]}
        />
      </Card>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add web/src/pages/Login.tsx
git commit -m "feat: add login/register page with Ant Design form"
```

---

### Task 8: Layout Component with Auth Guard

**Files:**
- Create: `web/src/components/Layout.tsx`

- [ ] **Step 1: Write components/Layout.tsx**

```tsx
import { useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Layout as AntLayout, Menu, Button, Typography } from 'antd'
import {
  DashboardOutlined,
  UnorderedListOutlined,
  LogoutOutlined,
} from '@ant-design/icons'

const { Header, Sider, Content } = AntLayout

const menuItems = [
  { key: '/dashboard', icon: <DashboardOutlined />, label: '首页看板' },
  { key: '/ads', icon: <UnorderedListOutlined />, label: '全域计划' },
]

export default function Layout() {
  const navigate = useNavigate()
  const location = useLocation()
  const token = localStorage.getItem('token')
  const [collapsed, setCollapsed] = useState(false)

  if (!token) {
    navigate('/login')
    return null
  }

  const handleLogout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    navigate('/login')
  }

  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Sider collapsible collapsed={collapsed} onCollapse={setCollapsed}>
        <div style={{ height: 48, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <Typography.Text style={{ color: '#fff', fontWeight: 'bold', fontSize: collapsed ? 14 : 16 }}>
            {collapsed ? '千川' : '千川投流助手'}
          </Typography.Text>
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <AntLayout>
        <Header style={{ background: '#fff', padding: '0 24px', display: 'flex', justifyContent: 'flex-end', alignItems: 'center' }}>
          <Button type="text" icon={<LogoutOutlined />} onClick={handleLogout}>退出</Button>
        </Header>
        <Content style={{ margin: 16, padding: 24, background: '#fff', borderRadius: 8 }}>
          <Outlet />
        </Content>
      </AntLayout>
    </AntLayout>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add web/src/components/Layout.tsx
git commit -m "feat: add layout with sidebar navigation and auth guard"
```

---

### Task 9: Dashboard Page

**Files:**
- Create: `web/src/pages/Dashboard.tsx`
- Create: `web/src/components/StatCard.tsx`

- [ ] **Step 1: Write components/StatCard.tsx**

```tsx
import { Card, Statistic } from 'antd'

interface Props {
  title: string
  value: number | string
  prefix?: string
  suffix?: string
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
```

- [ ] **Step 2: Write pages/Dashboard.tsx**

```tsx
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
```

- [ ] **Step 3: Commit**

```bash
git add web/src/pages/Dashboard.tsx web/src/components/StatCard.tsx
git commit -m "feat: add dashboard page with stat cards and trend chart"
```

---

### Task 10: Ad Management Page

**Files:**
- Create: `web/src/pages/Ads.tsx`

- [ ] **Step 1: Write pages/Ads.tsx**

```tsx
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
```

- [ ] **Step 2: Commit**

```bash
git add web/src/pages/Ads.tsx
git commit -m "feat: add ad management page with filters and status actions"
```

---

### Task 11: Integration Verification — Full Stack Smoke Test

**Files:**
- Modify: `web/src/pages/Login.tsx` (minor fix if needed)

- [ ] **Step 1: Start Go backend**

```bash
cd server && go build -o qianchuan-saas . && DATABASE_URL=postgres://qianchuan:qianchuan_dev@localhost:5432/qianchuan?sslmode=disable JWT_SECRET=test-secret ./qianchuan-saas &
```

- [ ] **Step 2: Start frontend dev server**

```bash
cd web && npm run dev &
```

- [ ] **Step 3: Verify key flows**

1. Open http://localhost:5173 → should redirect to /login
2. Register a new user → should succeed
3. Login → should redirect to /dashboard
4. Dashboard page loads stat cards and trend chart (data may be empty initially)
5. Navigate to /ads → shows ad management table
6. Account filter dropdown populated
7. Logout → redirects to /login

- [ ] **Step 4: Run backend tests**

```bash
cd server && go test ./handler/ -v -run TestRegisterAndLogin
```

Expected: PASS

- [ ] **Step 5: Commit final adjustments (if any)**

```bash
git add -A
git commit -m "chore: final Phase 2 adjustments and verification"
```

---

## Phase 3-4 概览

| Phase | 内容 | 关键交付 |
|-------|------|----------|
| **Phase 3** | Python 策略服务 + gRPC + 规则引擎 + 素材库页面 | 规则 CRUD、自动执行、素材管理 |
| **Phase 4** | AI 推荐 + 规则管理页面 + AI推荐页面 + E2E 测试 | 异常检测、ROI 预测、预算优化、Playwright |
