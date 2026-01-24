import { type Page, type Locator, errors } from '@playwright/test'

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
    try {
      await this.loadingText.waitFor({ state: 'hidden', timeout: 10000 })
    } catch (error) {
      if (error instanceof errors.TimeoutError) {
        // ローディングが表示されないか既に非表示の場合は正常
        return
      }
      throw error
    }
  }

  async isWeightSectionVisible(): Promise<boolean> {
    try {
      await this.weightSection.waitFor({ state: 'visible', timeout: 5000 })
      return true
    } catch (error) {
      if (error instanceof errors.TimeoutError) {
        return false
      }
      throw error
    }
  }

  async logout() {
    // UIにログアウトボタンがないため、localStorageをクリアしてリロード
    // リロード後、認証ガードによりログインページにリダイレクトされる
    await this.page.evaluate(() => {
      localStorage.removeItem('asken_auth_token')
      localStorage.removeItem('asken_user')
    })
    await this.page.reload()
  }
}
