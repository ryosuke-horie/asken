import { type Page, type Locator } from '@playwright/test'

export class EquipmentPage {
  readonly page: Page
  readonly title: Locator
  readonly subtitle: Locator
  readonly backButton: Locator
  readonly newNameInput: Locator
  readonly createButton: Locator
  readonly normalizeButton: Locator
  readonly loading: Locator
  readonly emptyMessage: Locator
  readonly equipmentList: Locator

  constructor(page: Page) {
    this.page = page
    this.title = page.getByRole('heading', { name: '器具設定' })
    this.subtitle = page.locator('[class*="subtitle"]')
    this.backButton = page.getByRole('link', { name: '← 戻る' })
    this.newNameInput = page.getByPlaceholder('新しい器具の名前')
    this.createButton = page.getByRole('button', { name: '追加', exact: true })
    this.normalizeButton = page.getByRole('button', { name: 'AI正規化して追加' })
    this.loading = page.getByText('読み込み中...')
    this.emptyMessage = page.getByText('登録されている器具はありません')
    this.equipmentList = page.locator('[class*="list"]')
  }

  async goto(locationId: string) {
    await this.page.goto(`/training/locations/${locationId}/equipment`)
  }

  async waitForLoad() {
    await this.title.waitFor({ state: 'visible' })
    await this.loading.waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {})
  }

  async createEquipment(name: string) {
    await this.newNameInput.fill(name)
    await this.createButton.click()
    // 201（成功）または500（重複エラー）を待つ
    await this.page.waitForResponse(
      (res) => res.url().includes('/equipment') &&
               (res.status() === 201 || res.status() === 500)
    )
  }

  getEquipmentCard(name: string) {
    return this.page.locator('[class*="equipmentCard"]').filter({
      has: this.page.locator(`[class*="equipmentName"]:text-is("${name}")`)
    })
  }

  async editEquipment(currentName: string, newName: string) {
    const card = this.getEquipmentCard(currentName)
    await card.getByRole('button', { name: '編集' }).click()
    // 編集モードになるとカード内にinputが表示される（名前は非表示になる）
    // ページ全体から編集フォーム内のinputを探す
    const editForm = this.page.locator('[class*="editForm"]')
    await editForm.waitFor({ state: 'visible' })
    const editInput = editForm.locator('input[type="text"]')
    await editInput.fill(newName)
    await editForm.getByRole('button', { name: '保存' }).click()
    await this.page.waitForResponse((res) => res.url().includes('/api/training/equipment/') && res.status() === 200)
  }

  async deleteEquipment(name: string) {
    const card = this.getEquipmentCard(name)
    this.page.once('dialog', (dialog) => dialog.accept())
    await card.getByRole('button', { name: '削除' }).click()
    await this.page.waitForResponse((res) => res.url().includes('/api/training/equipment/') && res.status() === 204)
  }

  async cancelDeleteEquipment(name: string) {
    const card = this.getEquipmentCard(name)
    this.page.once('dialog', (dialog) => dialog.dismiss())
    await card.getByRole('button', { name: '削除' }).click()
  }

  async goBack() {
    await this.backButton.click()
  }
}
