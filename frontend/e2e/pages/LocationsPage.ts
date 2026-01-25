import { type Page, type Locator } from '@playwright/test'

export class LocationsPage {
  readonly page: Page
  readonly title: Locator
  readonly backButton: Locator
  readonly newNameInput: Locator
  readonly createButton: Locator
  readonly loading: Locator
  readonly emptyMessage: Locator
  readonly locationList: Locator

  constructor(page: Page) {
    this.page = page
    this.title = page.getByRole('heading', { name: 'トレーニング場所設定' })
    this.backButton = page.getByRole('link', { name: '← 戻る' })
    this.newNameInput = page.getByPlaceholder('新しい場所の名前')
    this.createButton = page.getByRole('button', { name: '追加' })
    this.loading = page.getByText('読み込み中...')
    this.emptyMessage = page.getByText('登録されている場所はありません')
    this.locationList = page.locator('[class*="list"]')
  }

  async goto() {
    await this.page.goto('/training/locations')
  }

  async waitForLoad() {
    await this.title.waitFor({ state: 'visible' })
    await this.loading.waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {})
  }

  async createLocation(name: string) {
    await this.newNameInput.fill(name)
    await this.createButton.click()
    await this.page.waitForResponse((res) => res.url().includes('/api/training/locations') && res.status() === 201)
  }

  getLocationCard(name: string) {
    return this.page.locator('[class*="locationCard"]').filter({
      has: this.page.locator(`[class*="locationName"]:text-is("${name}")`)
    })
  }

  async editLocation(currentName: string, newName: string) {
    const card = this.getLocationCard(currentName)
    await card.getByRole('button', { name: '編集' }).click()
    // 編集モードになるとカード内にinputが表示される（名前は非表示になる）
    // ページ全体から編集フォーム内のinputを探す
    const editForm = this.page.locator('[class*="editForm"]')
    await editForm.waitFor({ state: 'visible' })
    const editInput = editForm.locator('input[type="text"]')
    await editInput.fill(newName)
    await editForm.getByRole('button', { name: '保存' }).click()
    await this.page.waitForResponse((res) => res.url().includes('/api/training/locations/') && res.status() === 200)
  }

  async deleteLocation(name: string) {
    const card = this.getLocationCard(name)
    this.page.once('dialog', (dialog) => dialog.accept())
    await card.getByRole('button', { name: '削除' }).click()
    await this.page.waitForResponse((res) => res.url().includes('/api/training/locations/') && res.status() === 204)
  }

  async cancelDeleteLocation(name: string) {
    const card = this.getLocationCard(name)
    this.page.once('dialog', (dialog) => dialog.dismiss())
    await card.getByRole('button', { name: '削除' }).click()
  }

  async goToEquipmentSettings(locationName: string) {
    const card = this.getLocationCard(locationName)
    await card.getByRole('link', { name: '器具設定' }).click()
  }

  async goBack() {
    await this.backButton.click()
  }
}
