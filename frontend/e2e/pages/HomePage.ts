import { type Page, type Locator } from '@playwright/test'

export class HomePage {
  readonly page: Page
  readonly weightSection: Locator
  readonly loadingText: Locator

  constructor(page: Page) {
    this.page = page
    this.weightSection = page.locator('section').filter({ hasText: '体重管理' })
    this.loadingText = page.getByText('読み込み中...')
  }

  async goto() {
    await this.page.goto('/')
  }

  async waitForLoad() {
    await this.loadingText.waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {
      // ローディングが表示されない場合は無視
    })
  }

  async isWeightSectionVisible(): Promise<boolean> {
    try {
      await this.weightSection.waitFor({ state: 'visible', timeout: 5000 })
      return true
    } catch {
      return false
    }
  }

  async logout() {
    // UIにログアウトボタンがないため、localStorageをクリアしてリダイレクト
    await this.page.evaluate(() => {
      localStorage.removeItem('asken_auth_token')
      localStorage.removeItem('asken_user')
    })
    await this.page.reload()
  }
}
