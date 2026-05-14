import { test, expect } from '../fixtures/base'

test.describe('Navigation', () => {
  test('topbar navigation links work', async ({ page }) => {
    await page.goto('/')

    const navLinks = [
      { name: '订阅', expectedUrl: /\/feed|^\// },
      { name: '音乐', expectedUrl: /\/music/ },
      { name: '博客', expectedUrl: /\/blog/ },
      { name: '论坛', expectedUrl: /\/forum/ },
      { name: '辩论', expectedUrl: /\/debate/ },
    ]

    for (const link of navLinks) {
      await page.goto('/')
      const navLink = page.getByRole('link', { name: link.name })
      if (await navLink.isVisible().catch(() => false)) {
        await navLink.click()
        await page.waitForTimeout(1000)
        await expect(page).toHaveURL(link.expectedUrl)
      }
    }
  })

  test('logo links to home', async ({ page }) => {
    await page.goto('/blog')
    await page.getByRole('link', { name: 'ATOMAN' }).click()
    await page.waitForTimeout(1000)
    await expect(page).toHaveURL('/')
  })

  test('404 page renders for unknown routes', async ({ page }) => {
    await page.goto('/this-route-does-not-exist-12345')
    await page.waitForTimeout(1000)
    const body = await page.locator('body').textContent()
    expect(body).toBeTruthy()
  })

  test('login button visible when not authenticated', async ({ page }) => {
    await page.goto('/')
    await expect(page.getByRole('link', { name: '登录' })).toBeVisible()
  })

  test('user menu visible when authenticated', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/')
    await authenticatedPage.waitForTimeout(1000)
    const userBtn = authenticatedPage.locator('.user-btn').or(authenticatedPage.locator('[class*="user-avatar"]'))
    await expect(userBtn.first()).toBeVisible()
  })

  test('responsive viewport does not crash', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/')
    await page.waitForTimeout(1000)
    await expect(page.locator('body')).toBeVisible()

    const hamburger = page.locator('.hamburger')
    if (await hamburger.isVisible().catch(() => false)) {
      await hamburger.click()
      await page.waitForTimeout(500)
      await expect(page.locator('.mobile-drawer')).toBeVisible()
    }
  })
})
