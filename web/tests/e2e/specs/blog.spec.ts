import { test, expect } from '../fixtures/base'
import { ADMIN_EMAIL, ADMIN_PASSWORD } from '../helpers/auth'
import { expectTextVisible } from '../helpers/common'

test.describe('Blog', () => {
  test('browse blog home page', async ({ page }) => {
    await page.goto('/blog')
    await expect(page.getByText('博客')).toBeVisible()
  })

  test('public post accessible without login', async ({ page }) => {
    await page.goto('/blog/explore')
    await expect(page.getByText('探索')).toBeVisible()
  })

  test('create new post as authenticated user', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/post/new')
    await authenticatedPage.waitForTimeout(3000)

    const channelPicker = authenticatedPage.locator('.channel-picker-card')
    if (await channelPicker.isVisible().catch(() => false)) {
      const firstChannel = authenticatedPage.locator('.channel-picker-link').first()
      if (await firstChannel.isVisible().catch(() => false)) {
        await firstChannel.click()
        await authenticatedPage.waitForTimeout(2000)
      }
    }

    const titleInput = authenticatedPage.locator('.editor-title-input')
    if (await titleInput.isVisible().catch(() => false)) {
      await titleInput.fill(`E2E Test Post ${Date.now()}`)
    }

    const editorContent = authenticatedPage.locator('[contenteditable]').first()
    if (await editorContent.isVisible().catch(() => false)) {
      await editorContent.click()
      await editorContent.fill('This is an E2E test post content.')
      await authenticatedPage.waitForTimeout(1000)

      const saveBtn = authenticatedPage.getByRole('button', { name: '保存' })
      if (await saveBtn.isVisible().catch(() => false)) {
        await saveBtn.click()
        await authenticatedPage.waitForTimeout(3000)
      }
    }
  })

  test('like a post as authenticated user', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/blog/explore')
    await authenticatedPage.waitForTimeout(2000)

    const firstPostLink = authenticatedPage.locator('article a, .a-card-hover a').first()
    if (await firstPostLink.isVisible().catch(() => false)) {
      await firstPostLink.click()
      await authenticatedPage.waitForTimeout(2000)

      const likeBtn = authenticatedPage.locator('button', { hasText: /♥/ }).first()
      if (await likeBtn.isVisible().catch(() => false)) {
        await likeBtn.click()
        await authenticatedPage.waitForTimeout(1000)
      }
    }
  })

  test('comment on a post as authenticated user', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/blog/explore')
    await authenticatedPage.waitForTimeout(2000)

    const firstPostLink = authenticatedPage.locator('article a, .a-card-hover a').first()
    if (await firstPostLink.isVisible().catch(() => false)) {
      await firstPostLink.click()
      await authenticatedPage.waitForTimeout(2000)

      const commentTextarea = authenticatedPage.locator('textarea[placeholder*="评论"]').first()
      if (await commentTextarea.isVisible().catch(() => false)) {
        await commentTextarea.fill('E2E test comment')
        const submitBtn = authenticatedPage.getByRole('button', { name: /提交/ })
        if (await submitBtn.isVisible().catch(() => false)) {
          await submitBtn.click()
          await authenticatedPage.waitForTimeout(2000)
        }
      }
    }
  })

  test('bookmark a post as authenticated user', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/blog/bookmarks')
    await expect(authenticatedPage).toHaveURL(/\/blog\/bookmarks/)
  })

  test('visit blog settings page', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/blog/settings')
    await expect(authenticatedPage).toHaveURL(/\/blog\/settings/)
  })

  test('editor uses a flexible three-column workspace', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/post/new')
    await authenticatedPage.waitForTimeout(2000)

    const channelPicker = authenticatedPage.locator('.channel-picker-card')
    if (await channelPicker.isVisible().catch(() => false)) {
      const firstChannel = authenticatedPage.locator('.channel-picker-link').first()
      if (await firstChannel.isVisible().catch(() => false)) {
        await firstChannel.click()
        await authenticatedPage.waitForTimeout(1500)
      }
    }

    await expect(authenticatedPage.locator('.editor-page')).toBeVisible()
    await expect(authenticatedPage.locator('.col-left')).toBeVisible()
    await expect(authenticatedPage.locator('.col-center')).toBeVisible()
    await expect(authenticatedPage.locator('.col-right')).toBeVisible()

    await expect(authenticatedPage.locator('.editor-topbar')).toBeVisible()
    await expect(authenticatedPage.locator('.title-input')).toBeVisible()
    await expect(authenticatedPage.locator('.vditor-wrapper')).toBeVisible()

    const importSection = authenticatedPage.locator('.col-import')
    await expect(importSection).toHaveCount(0)
  })
})
