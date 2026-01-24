import { type Page, type Locator } from '@playwright/test'

export class LoginPage {
  readonly page: Page
  readonly emailInput: Locator
  readonly passwordInput: Locator
  readonly submitButton: Locator
  readonly errorMessage: Locator
  readonly registerLink: Locator
  readonly loadingText: Locator

  constructor(page: Page) {
    this.page = page
    this.emailInput = page.locator('#email')
    this.passwordInput = page.locator('#password')
    this.submitButton = page.locator('button[type="submit"]')
    this.errorMessage = page.locator('[class*="error"]')
    this.registerLink = page.locator('a[href="/register"]')
    this.loadingText = page.getByText('読み込み中...')
  }

  async goto() {
    await this.page.goto('/login')
  }

  async login(email: string, password: string) {
    await this.emailInput.fill(email)
    await this.passwordInput.fill(password)
    await this.submitButton.click()
  }

  async waitForLoad() {
    await this.loadingText.waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {
      // ローディングが表示されない場合は無視
    })
  }

  async getErrorMessage(): Promise<string | null> {
    try {
      await this.errorMessage.waitFor({ state: 'visible', timeout: 5000 })
      return await this.errorMessage.textContent()
    } catch {
      return null
    }
  }
}
