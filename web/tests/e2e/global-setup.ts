import { chromium, type FullConfig } from '@playwright/test'

const ADMIN_EMAIL = 'admin@atoman.com'
const ADMIN_PASSWORD = 'admin123'
const AUTH_FILE = './tests/e2e/.auth/admin.json'

async function globalSetup(config: FullConfig) {
  const browser = await chromium.launch()
  const context = await browser.newContext()
  const page = await context.newPage()

  await page.goto('http://localhost:5173/login')
  await page.getByPlaceholder('输入用户名或邮箱').fill(ADMIN_EMAIL)
  await page.getByLabel('通行密码').fill(ADMIN_PASSWORD)
  await page.getByRole('button', { name: '登 录' }).click()
  await page.waitForURL(/^(?!\/login)/, { timeout: 10000 })

  await context.storageState({ path: AUTH_FILE })
  await browser.close()
}

export default globalSetup
