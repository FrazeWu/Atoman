import { type Page, expect } from '@playwright/test'

export async function waitForNetworkIdle(page: Page, timeout = 10000): Promise<void> {
  await page.waitForLoadState('networkidle', { timeout }).catch(() => {})
}

export async function expectTextVisible(page: Page, text: string): Promise<void> {
  await expect(page.getByText(text)).toBeVisible()
}

export async function expectUrlContains(page: Page, fragment: string): Promise<void> {
  await expect(page).toHaveURL(new RegExp(fragment))
}

export async function fillAndSubmit(page: Page, selector: string, value: string): Promise<void> {
  await page.locator(selector).fill(value)
}

export async function clickAndWaitForNavigation(page: Page, locator: ReturnType<Page['locator']> | ReturnType<Page['getByRole']> | ReturnType<Page['getByText']>): Promise<void> {
  await Promise.all([
    page.waitForNavigation({ timeout: 15000 }).catch(() => {}),
    locator.click(),
  ])
}
