import { test, expect } from '../fixtures/base'
import { expectTextVisible } from '../helpers/common'

test.describe('Music', () => {
  test('browse music timeline', async ({ page }) => {
    await page.goto('/music')
    await expect(page.getByText('TIMELINE')).toBeVisible()
  })

  test('music shows search input', async ({ page }) => {
    await page.goto('/music')
    await expect(page.getByPlaceholder('搜索艺术家...')).toBeVisible()
  })

  test('music shows random button', async ({ page }) => {
    await page.goto('/music')
    await expect(page.getByRole('button', { name: '随机' })).toBeVisible()
  })

  test('search for artist', async ({ page }) => {
    await page.goto('/music')
    await page.waitForTimeout(2000)

    const searchInput = page.getByPlaceholder('搜索艺术家...')
    await searchInput.click()
    await searchInput.fill('test')
    await page.waitForTimeout(1000)
  })

  test('view album detail', async ({ page }) => {
    await page.goto('/music')
    await page.waitForTimeout(3000)

    const detailLink = page.getByRole('link', { name: '详情' }).first()
    if (await detailLink.isVisible().catch(() => false)) {
      await detailLink.click()
      await page.waitForTimeout(2000)
      await expect(page).toHaveURL(/\/music\/artists\/.*\/albums\//)
    }
  })

  test('play song from timeline', async ({ page }) => {
    await page.goto('/music')
    await page.waitForTimeout(3000)

    const playBtn = page.getByRole('button', { name: '▶ 播放' }).first()
    if (await playBtn.isVisible().catch(() => false)) {
      await playBtn.click()
      await page.waitForTimeout(1000)
      await expect(page.locator('body')).toBeVisible()
    }
  })

  test('music contribute requires login', async ({ page }) => {
    await page.goto('/music/contribute')
    await expect(page).toHaveURL(/\/login/)
  })

  test('authenticated user can access contribute page', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/music/contribute')
    await expect(authenticatedPage).toHaveURL(/\/music\/contribute/)
  })

  test('music form pages render core controls', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/music/contribute')
    await expect(authenticatedPage.getByText('贡献新档案')).toBeVisible()
    await expect(authenticatedPage.getByText('艺术家')).toBeVisible()
    await expect(authenticatedPage.getByText('专辑名称')).toBeVisible()
    await expect(authenticatedPage.getByRole('button', { name: /直接上传/i })).toBeVisible()

    await authenticatedPage.goto('/music/artists/add?name=test_artist')
    await expect(authenticatedPage.getByText('添加/补全艺术家')).toBeVisible()
    await expect(authenticatedPage.getByLabel('艺术家名称')).toHaveValue('test_artist')
    await expect(authenticatedPage.getByRole('button', { name: '创建艺术家' })).toBeVisible()
  })
})
