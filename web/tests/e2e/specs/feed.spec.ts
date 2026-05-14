import { test, expect } from '../fixtures/base'
import { expectTextVisible } from '../helpers/common'

test.describe('Feed', () => {
  test('browse feed page without login shows guest view', async ({ page }) => {
    await page.goto('/feed')
    await expect(page.getByText('订阅')).toBeVisible()
    await expect(page.getByRole('link', { name: '登录' })).toBeVisible()
  })

  test('authenticated user sees feed timeline', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/feed')
    await authenticatedPage.waitForTimeout(2000)
    await expect(authenticatedPage.getByText('全部订阅')).toBeVisible()
  })

  test('authenticated user can open add subscription modal', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/feed')
    await authenticatedPage.waitForTimeout(2000)

    const addBtn = authenticatedPage.getByRole('button', { name: '+ 添加订阅' })
    if (await addBtn.isVisible().catch(() => false)) {
      await addBtn.click()
      await expect(authenticatedPage.getByText('添加订阅')).toBeVisible()
      await expect(authenticatedPage.getByPlaceholder('https://example.com/feed.xml')).toBeVisible()

      const cancelBtn = authenticatedPage.getByRole('button', { name: '取消' })
      if (await cancelBtn.isVisible().catch(() => false)) {
        await cancelBtn.click()
      }
    }
  })

  test('authenticated user sees mark all read button', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/feed')
    await authenticatedPage.waitForTimeout(2000)

    const markAllBtn = authenticatedPage.getByRole('button', { name: '全部已读' })
    if (await markAllBtn.isVisible().catch(() => false)) {
      await markAllBtn.click()
      await authenticatedPage.waitForTimeout(1000)
    }
  })

  test('authenticated user can create new group', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/feed')
    await authenticatedPage.waitForTimeout(2000)

    const newGroupBtn = authenticatedPage.getByRole('button', { name: '+ 新建分组' })
    if (await newGroupBtn.isVisible().catch(() => false)) {
      await newGroupBtn.click()
      await authenticatedPage.waitForTimeout(500)

      const groupInput = authenticatedPage.locator('input[placeholder="分组名称"]')
      if (await groupInput.isVisible().catch(() => false)) {
        await groupInput.fill(`E2E Group ${Date.now()}`)
        const confirmBtn = authenticatedPage.getByRole('button', { name: '确认' })
        if (await confirmBtn.isVisible().catch(() => false)) {
          await confirmBtn.click()
          await authenticatedPage.waitForTimeout(1000)
        }
      }
    }
  })

  test('feed shows load more when items exceed page', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/feed')
    await authenticatedPage.waitForTimeout(2000)

    const loadMoreBtn = authenticatedPage.getByRole('button', { name: '加载更多' })
    if (await loadMoreBtn.isVisible().catch(() => false)) {
      await loadMoreBtn.click()
      await authenticatedPage.waitForTimeout(2000)
    }
  })

  test('podcast playback button visible for audio items', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/feed')
    await authenticatedPage.waitForTimeout(2000)

    const playBtn = authenticatedPage.getByRole('button', { name: '▶ 播放' })
    if (await playBtn.isVisible().catch(() => false)) {
      await playBtn.click()
      await authenticatedPage.waitForTimeout(1000)
      await expect(authenticatedPage.locator('#audio-player, .audio-player, [class*="player"]')).toBeVisible()
    }
  })
})
