import { test, expect } from '../fixtures/base'
import { loginViaUI, logoutViaUI, ADMIN_EMAIL, ADMIN_PASSWORD } from '../helpers/auth'
import { expectTextVisible, expectUrlContains } from '../helpers/common'

test.describe('Authentication', () => {
  test('login with valid credentials redirects to home', async ({ page }) => {
    await loginViaUI(page, ADMIN_EMAIL, ADMIN_PASSWORD)
    await expect(page).toHaveURL('/')
    await expectTextVisible(page, 'ATOMAN')
  })

  test('login with invalid credentials shows error', async ({ page }) => {
    await page.goto('/login')
    await page.getByPlaceholder('输入用户名或邮箱').fill('wrong@email.com')
    await page.getByLabel('通行密码').fill('wrongpassword')
    await page.getByRole('button', { name: '登 录' }).click()
    await expect(page.locator('.error-msg')).toBeVisible()
  })

  test('logout returns to login page', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/')
    await authenticatedPage.getByRole('button', { name: /▾/ }).first().click()
    await authenticatedPage.getByRole('button', { name: '退出登录' }).click()
    await expect(authenticatedPage).toHaveURL(/\/login/)
  })

  test('unauthenticated user redirected to login when accessing protected route', async ({ page }) => {
    await page.goto('/post/new')
    await expect(page).toHaveURL(/\/login.*redirect/)
  })

  test('unauthenticated user redirected to login when accessing forum new topic', async ({ page }) => {
    await page.goto('/forum/new')
    await expect(page).toHaveURL(/\/login.*redirect/)
  })

  test('unauthenticated user redirected to login when accessing music contribute', async ({ page }) => {
    await page.goto('/music/contribute')
    await expect(page).toHaveURL(/\/login.*redirect/)
  })

  test('authenticated user visiting login redirects away', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/login')
    await expect(authenticatedPage).not.toHaveURL(/\/login$/)
  })

  test('register page shows registration form', async ({ page }) => {
    await page.goto('/register')
    await expect(page.getByText('加入我们')).toBeVisible()
    await expect(page.getByPlaceholder('请输入邮箱地址')).toBeVisible()
    await expect(page.getByText('获取验证码')).toBeVisible()
  })

  test('login page shows login form', async ({ page }) => {
    await page.goto('/login')
    await expect(page.getByText('欢迎回来')).toBeVisible()
    await expect(page.getByPlaceholder('输入用户名或邮箱')).toBeVisible()
  })

  test('switch between login and register pages', async ({ page }) => {
    await page.goto('/login')
    await page.getByRole('link', { name: '立即加入档案室' }).click()
    await expect(page).toHaveURL('/register')
    await expect(page.getByText('加入我们')).toBeVisible()

    await page.getByRole('link', { name: '立即登录' }).click()
    await expect(page).toHaveURL('/login')
    await expect(page.getByText('欢迎回来')).toBeVisible()
  })
})
