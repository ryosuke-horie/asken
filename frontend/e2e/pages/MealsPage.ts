import { type Page, type Locator, errors } from '@playwright/test'

export type MealType = 'breakfast' | 'lunch' | 'dinner' | 'snack'
export type InputType = 'mylist' | 'image' | 'text' | 'skipped'

export class MealsPage {
  readonly page: Page

  // ヘッダー要素
  readonly backButton: Locator
  readonly dateInfo: Locator
  readonly title: Locator

  // 入力タブ
  readonly mylistTab: Locator
  readonly imageTab: Locator
  readonly textTab: Locator
  readonly skippedTab: Locator

  // マイリスト関連
  readonly mylistLoading: Locator
  readonly mylistEmpty: Locator
  readonly mylistItems: Locator
  readonly quantitySection: Locator
  readonly mylistSubmitButton: Locator
  readonly mylistBackButton: Locator

  // テキスト入力関連
  readonly textInputs: Locator
  readonly addInputButton: Locator
  readonly submitTextButton: Locator
  readonly removeInputButtons: Locator

  // 食べなかった関連
  readonly skippedButton: Locator
  readonly skippedDescription: Locator

  // 共通
  readonly loadingIndicator: Locator
  readonly errorMessage: Locator
  readonly successMessage: Locator
  readonly syncError: Locator

  // 登録済み食事一覧
  readonly mealsSection: Locator
  readonly mealsList: Locator
  readonly mealItems: Locator

  constructor(page: Page) {
    this.page = page

    // ヘッダー
    this.backButton = page.getByRole('button', { name: '← 戻る' })
    this.dateInfo = page.locator('[class*="dateInfo"]')
    this.title = page.locator('h1[class*="title"]')

    // 入力タブ
    this.mylistTab = page.getByRole('button', { name: 'マイリスト' })
    this.imageTab = page.getByRole('button', { name: '画像' })
    this.textTab = page.getByRole('button', { name: 'テキスト' })
    this.skippedTab = page.getByRole('button', { name: '食べなかった' })

    // マイリスト
    this.mylistLoading = page.getByText('読み込み中...')
    this.mylistEmpty = page.getByText('マイリストにアイテムがありません')
    this.mylistItems = page.locator('[class*="itemCard"]')
    this.quantitySection = page.locator('[class*="quantitySection"]')
    this.mylistSubmitButton = page.getByRole('button', { name: '記録する' })
    this.mylistBackButton = page.getByRole('button', { name: '← 一覧に戻る' })

    // テキスト入力
    this.textInputs = page.locator('[class*="textInput"]')
    this.addInputButton = page.getByRole('button', { name: '追加' })
    this.submitTextButton = page.getByRole('button', { name: '解析' })
    this.removeInputButtons = page.locator('[aria-label="削除"]')

    // 食べなかった
    this.skippedButton = page.getByRole('button', { name: /食べませんでした|記録中/ })
    this.skippedDescription = page.getByText('この食事を「食べなかった」として記録します')

    // 共通
    this.loadingIndicator = page.getByText(/分析中|記録中|送信中/)
    this.errorMessage = page.locator('[class*="error"]')
    this.successMessage = page.locator('[class*="successMessage"]')
    this.syncError = page.locator('[class*="syncError"]')

    // 登録済み食事一覧
    this.mealsSection = page.locator('[class*="mealsSection"]')
    this.mealsList = page.locator('[class*="mealsList"]')
    this.mealItems = page.locator('[class*="mealItem"]')
  }

  async goto(mealType: MealType, date?: string) {
    const dateParam = date || new Date().toISOString().split('T')[0]
    await this.page.goto(`/meals/${mealType}?date=${dateParam}`)
  }

  async waitForLoad() {
    try {
      await this.mylistLoading.waitFor({ state: 'hidden', timeout: 10000 })
    } catch (error) {
      if (!(error instanceof errors.TimeoutError)) {
        throw error
      }
    }
    await this.page.waitForLoadState('networkidle')
  }

  async selectInputType(inputType: InputType) {
    switch (inputType) {
      case 'mylist':
        await this.mylistTab.click()
        break
      case 'image':
        await this.imageTab.click()
        break
      case 'text':
        await this.textTab.click()
        break
      case 'skipped':
        await this.skippedTab.click()
        break
    }
  }

  async selectMylistItem(itemName: string) {
    const item = this.page.locator('[class*="itemCard"]').filter({ hasText: itemName })
    const count = await item.count()
    if (count === 0) {
      const allItems = await this.mylistItems.allTextContents()
      throw new Error(
        `Mylist item "${itemName}" not found. Available items: ${allItems.join(', ') || 'none'}`
      )
    }
    await item.click()
  }

  getMylistItemCard(itemName: string): Locator {
    return this.page.locator('[class*="itemCard"]').filter({ hasText: itemName })
  }

  async submitMylistItem() {
    await this.mylistSubmitButton.click()
    await this.waitForApiResponse('/api/meals/from-mylist')
  }

  async fillTextInput(index: number, text: string) {
    const inputs = this.page.locator('[class*="textInput"]')
    await inputs.nth(index).fill(text)
  }

  async addTextInput() {
    await this.addInputButton.click()
  }

  async removeTextInput(index: number) {
    const removeButtons = this.page.locator('[aria-label="削除"]')
    await removeButtons.nth(index).click()
  }

  async submitTextInput() {
    await this.submitTextButton.click()
  }

  async submitSkipped() {
    await this.skippedButton.click()
    await this.waitForApiResponse('/api/meals/skip')
  }

  getMealItem(index: number): Locator {
    return this.mealItems.nth(index)
  }

  getSkippedMealItem(): Locator {
    return this.mealItems.filter({ hasText: '食べませんでした' })
  }

  async deleteMeal(index: number) {
    const mealItem = this.getMealItem(index)
    const deleteButton = mealItem.getByRole('button', { name: '削除' })
    // confirmダイアログを自動的に受け入れる
    this.page.once('dialog', (dialog) => dialog.accept())
    await deleteButton.click()
    await this.waitForApiResponse('/api/history')
  }

  async waitForApiResponse(urlPattern: string, options?: { timeout?: number }) {
    const timeout = options?.timeout ?? 15000

    try {
      const response = await this.page.waitForResponse(
        (res) => res.url().includes(urlPattern),
        { timeout }
      )

      if (!response.ok()) {
        const body = await response.text().catch(() => 'Unable to read body')
        throw new Error(`API ${urlPattern} returned error status ${response.status()}: ${body}`)
      }

      return response
    } catch (error) {
      if (error instanceof errors.TimeoutError) {
        throw new Error(`Timeout waiting for API response matching "${urlPattern}" after ${timeout}ms`)
      }
      throw error
    }
  }

  async waitForAnalysisComplete(context?: string) {
    const description = context ? `analysis (${context})` : 'analysis'
    const timeout = 60000

    try {
      await this.loadingIndicator.waitFor({ state: 'hidden', timeout })
    } catch (error) {
      if (error instanceof errors.TimeoutError) {
        const isStillLoading = await this.loadingIndicator.isVisible().catch(() => 'unknown')
        const errorVisible = await this.errorMessage.isVisible().catch(() => 'unknown')
        throw new Error(
          `Timeout waiting for ${description} to complete after ${timeout}ms. ` +
            `Loading indicator still visible: ${isStillLoading}, Error message visible: ${errorVisible}`
        )
      }
      throw error
    }
  }

  async getErrorMessage(): Promise<string> {
    try {
      await this.errorMessage.waitFor({ state: 'visible', timeout: 3000 })
      const text = await this.errorMessage.textContent()
      if (!text) {
        throw new Error('Error message element is visible but has no text content')
      }
      return text
    } catch (error) {
      if (error instanceof errors.TimeoutError) {
        throw new Error(
          'No error message visible within 3000ms. If expecting no error, use isErrorMessageVisible() instead.'
        )
      }
      throw error
    }
  }

  async isErrorMessageVisible(): Promise<boolean> {
    try {
      await this.errorMessage.waitFor({ state: 'visible', timeout: 3000 })
      return true
    } catch (error) {
      if (error instanceof errors.TimeoutError) {
        return false
      }
      throw error
    }
  }

  async getMealsCount(): Promise<number> {
    return await this.mealItems.count()
  }

  async waitForSuccessMessage(options?: { timeout?: number }): Promise<void> {
    const timeout = options?.timeout ?? 5000

    try {
      await this.successMessage.waitFor({ state: 'visible', timeout })
    } catch (error) {
      if (error instanceof errors.TimeoutError) {
        const errorVisible = await this.errorMessage.isVisible().catch(() => 'unknown')
        const loadingVisible = await this.loadingIndicator.isVisible().catch(() => 'unknown')
        throw new Error(
          `Success message not visible within ${timeout}ms. ` +
            `Error message visible: ${errorVisible}, Loading visible: ${loadingVisible}`
        )
      }
      throw error
    }
  }
}
