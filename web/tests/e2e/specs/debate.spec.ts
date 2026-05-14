import { test, expect } from '../fixtures/base'
import { expectTextVisible } from '../helpers/common'

test.describe('Debate', () => {
  test('browse debates home page', async ({ page }) => {
    await page.goto('/debate')
    await expect(page.getByText('辩论')).toBeVisible()
    await expect(page.getByText('结构化讨论')).toBeVisible()
  })

  test('debate page shows filter controls', async ({ page }) => {
    await page.goto('/debate')
    await expect(page.locator('select').first()).toBeVisible()
    await expect(page.getByPlaceholder('标签筛选')).toBeVisible()
  })

  test('debate shows empty state or debate cards', async ({ page }) => {
    await page.goto('/debate')
    await page.waitForTimeout(2000)
    const emptyState = page.getByText('暂无辩论')
    const debateCards = page.locator('.a-card-hover').or(page.locator('[class*="cursor-pointer"]').filter({ hasText: /进行中|已结题|已归档/ }))
    const hasContent = (await emptyState.isVisible().catch(() => false)) || (await debateCards.first().isVisible().catch(() => false))
    expect(hasContent).toBeTruthy()
  })

  test('create new debate as authenticated user', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/debate')
    await authenticatedPage.waitForTimeout(1000)

    const createBtn = authenticatedPage.getByRole('button', { name: '发起辩论' })
    if (await createBtn.isVisible().catch(() => false)) {
      await createBtn.click()
      await expect(authenticatedPage.getByText('发起辩论')).toBeVisible()

      await authenticatedPage.getByPlaceholder('辩论主题').fill(`E2E Test Debate ${Date.now()}`)
      await authenticatedPage.getByPlaceholder('简短描述').fill('This is a test debate from E2E.')
      await authenticatedPage.locator('textarea[placeholder*="详细说明"]').fill('Detailed background content for the E2E test debate.')
      await authenticatedPage.getByPlaceholder('例如：科技，社会，哲学').fill('测试,E2E')

      const submitBtn = authenticatedPage.getByRole('button', { name: '创建' })
      await submitBtn.click()
      await authenticatedPage.waitForTimeout(3000)
    }
  })

  test('view debate detail page', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/debate')
    await authenticatedPage.waitForTimeout(2000)

    const firstDebate = authenticatedPage.locator('.a-card-hover.cursor-pointer, [class*="cursor-pointer"]').first()
    if (await firstDebate.isVisible().catch(() => false)) {
      await firstDebate.click()
      await authenticatedPage.waitForTimeout(2000)
      await expect(authenticatedPage).toHaveURL(/\/debate\//)
    }
  })

  test('add argument to debate as authenticated user', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/debate')
    await authenticatedPage.waitForTimeout(2000)

    const firstDebate = authenticatedPage.locator('.a-card-hover.cursor-pointer, [class*="cursor-pointer"]').first()
    if (await firstDebate.isVisible().catch(() => false)) {
      await firstDebate.click()
      await authenticatedPage.waitForTimeout(2000)

      const addArgBtn = authenticatedPage.getByRole('button', { name: '添加论点' })
      if (await addArgBtn.isVisible().catch(() => false)) {
        await addArgBtn.click()
        await expect(authenticatedPage.getByText(/添加根论点|回复论点/)).toBeVisible()

        await authenticatedPage.locator('select').first().selectOption('support')

        const textarea = authenticatedPage.locator('textarea[placeholder*="阐述"]').first()
        if (await textarea.isVisible().catch(() => false)) {
          await textarea.fill('E2E test argument content')
          const submitBtn = authenticatedPage.getByRole('button', { name: '添加' })
          await submitBtn.click()
          await authenticatedPage.waitForTimeout(2000)
        }
      }
    }
  })

  test('vote on argument as authenticated user', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/debate')
    await authenticatedPage.waitForTimeout(2000)

    const firstDebate = authenticatedPage.locator('.a-card-hover.cursor-pointer, [class*="cursor-pointer"]').first()
    if (await firstDebate.isVisible().catch(() => false)) {
      await firstDebate.click()
      await authenticatedPage.waitForTimeout(2000)

      const voteBtn = authenticatedPage.locator('button', { hasText: /赞成|反对/ }).first()
      if (await voteBtn.isVisible().catch(() => false)) {
        await voteBtn.click()
        await authenticatedPage.waitForTimeout(1000)
      }
    }
  })

  test('filter debates by status', async ({ page }) => {
    await page.goto('/debate')
    await page.waitForTimeout(1000)

    const statusSelect = page.locator('select').first()
    if (await statusSelect.isVisible().catch(() => false)) {
      await statusSelect.selectOption('open')
      await page.getByRole('button', { name: '筛选' }).click()
      await page.waitForTimeout(1000)
    }
  })
})
