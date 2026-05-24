import { test, expect } from '@playwright/test'

test.describe('关键用户流程', () => {
  test('登录 → 看板加载', async ({ page }) => {
    await page.goto('/login')
    await page.fill('input[placeholder="邮箱"]', 'test@example.com')
    await page.fill('input[placeholder="密码"]', 'test123')
    await page.click('button[type="submit"]')
    await expect(page).toHaveURL('/dashboard')
    await expect(page.locator('.ant-statistic')).toHaveCount(4)
  })

  test('导航到全域计划', async ({ page }) => {
    await page.goto('/login')
    await page.fill('input[placeholder="邮箱"]', 'test@example.com')
    await page.fill('input[placeholder="密码"]', 'test123')
    await page.click('button[type="submit"]')
    await page.click('text=全域计划')
    await expect(page).toHaveURL('/ads')
    await expect(page.locator('.ant-table')).toBeVisible()
  })

  test('导航到规则管理', async ({ page }) => {
    await page.goto('/login')
    await page.fill('input[placeholder="邮箱"]', 'test@example.com')
    await page.fill('input[placeholder="密码"]', 'test123')
    await page.click('button[type="submit"]')
    await page.click('text=规则管理')
    await expect(page).toHaveURL('/rules')
    await expect(page.locator('text=新建规则')).toBeVisible()
  })
})
