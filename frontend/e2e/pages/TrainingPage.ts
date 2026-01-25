import { type Page, type Locator } from '@playwright/test'

export class TrainingPage {
  readonly page: Page
  readonly title: Locator
  readonly settingsButton: Locator
  readonly suggestButton: Locator
  readonly calendarContainer: Locator
  readonly monthLabel: Locator
  readonly prevMonthButton: Locator
  readonly nextMonthButton: Locator
  readonly locationSelect: Locator
  readonly durationInput: Locator
  readonly notesInput: Locator
  readonly saveButton: Locator
  readonly cancelButton: Locator
  readonly formTitle: Locator
  readonly loading: Locator
  readonly exerciseList: Locator
  readonly ratingGroup: Locator
  readonly recordList: Locator
  readonly noEquipmentMessage: Locator

  constructor(page: Page) {
    this.page = page
    this.title = page.getByRole('heading', { name: 'トレーニング記録' })
    this.settingsButton = page.getByRole('link', { name: '場所設定' })
    this.suggestButton = page.getByRole('link', { name: 'メニュー提案' })
    this.calendarContainer = page.locator('[class*="calendarContainer"]')
    this.monthLabel = page.locator('[class*="monthLabel"]')
    this.prevMonthButton = page.getByRole('button', { name: '←' })
    this.nextMonthButton = page.getByRole('button', { name: '→' })
    this.locationSelect = page.locator('select')
    this.durationInput = page.getByPlaceholder('例: 60')
    this.notesInput = page.getByPlaceholder('練習内容や気づきなど')
    this.saveButton = page.getByRole('button', { name: '保存' })
    this.cancelButton = page.getByRole('button', { name: 'キャンセル' })
    this.formTitle = page.locator('[class*="formTitle"]')
    this.loading = page.getByText('読み込み中...')
    this.exerciseList = page.locator('[class*="exerciseList"]')
    this.ratingGroup = page.locator('[class*="ratingGroup"]')
    this.recordList = page.locator('[class*="recordList"]')
    this.noEquipmentMessage = page.getByText('この場所には器具が登録されていません')
  }

  async goto() {
    await this.page.goto('/training')
  }

  async waitForLoad() {
    await this.title.waitFor({ state: 'visible' })
    await this.loading.waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {})
  }

  async selectDate(day: number) {
    const dayButton = this.page.locator(`[class*="calendarDay"]:not([class*="otherMonth"])`).filter({
      has: this.page.locator(`[class*="dayNumber"]:text-is("${day}")`)
    })
    await dayButton.click()
  }

  async selectLocation(locationName: string) {
    await this.locationSelect.selectOption({ label: locationName })
  }

  async setDuration(minutes: number) {
    await this.durationInput.fill(minutes.toString())
  }

  async selectMenu(menuName: string) {
    const checkbox = this.getMenuCheckbox(menuName)
    await checkbox.check()
  }

  async deselectMenu(menuName: string) {
    const checkbox = this.getMenuCheckbox(menuName)
    await checkbox.uncheck()
  }

  getMenuCheckbox(menuName: string): Locator {
    return this.page.locator('[class*="exerciseItem"]').filter({
      hasText: menuName
    }).locator('input[type="checkbox"]')
  }

  async setMenuSetsReps(menuName: string, sets: number, reps: number) {
    const exerciseItem = this.page.locator('[class*="exerciseItem"]').filter({
      hasText: menuName
    })
    const setsInput = exerciseItem.locator('[class*="exerciseInputs"] input').first()
    const repsInput = exerciseItem.locator('[class*="exerciseInputs"] input').nth(1)

    await setsInput.fill(sets.toString())
    await repsInput.fill(reps.toString())
  }

  async selectEquipment(equipmentName: string) {
    const checkbox = this.getEquipmentCheckbox(equipmentName)
    await checkbox.check()
  }

  getEquipmentCheckbox(equipmentName: string): Locator {
    return this.page.locator('[class*="exerciseItem"]').filter({
      hasText: equipmentName
    }).locator('input[type="checkbox"]')
  }

  async setEquipmentSetsReps(equipmentName: string, sets: number, reps: number) {
    const exerciseItem = this.page.locator('[class*="exerciseItem"]').filter({
      hasText: equipmentName
    })
    const setsInput = exerciseItem.locator('[class*="exerciseInputs"] input').first()
    const repsInput = exerciseItem.locator('[class*="exerciseInputs"] input').nth(1)

    await setsInput.fill(sets.toString())
    await repsInput.fill(reps.toString())
  }

  async setIntensity(level: number) {
    // 強度は最初のratingGroupに対応
    const intensityGroup = this.page.locator('[class*="formGroup"]').filter({
      has: this.page.locator('label:has-text("強度")')
    })
    const starButton = intensityGroup.locator('[class*="ratingButton"]').nth(level - 1)
    await starButton.click()
  }

  async setSatisfaction(level: number) {
    // 満足度は2つ目のratingGroupに対応
    const satisfactionGroup = this.page.locator('[class*="formGroup"]').filter({
      has: this.page.locator('label:has-text("満足度")')
    })
    const starButton = satisfactionGroup.locator('[class*="ratingButton"]').nth(level - 1)
    await starButton.click()
  }

  async setNotes(notes: string) {
    await this.notesInput.fill(notes)
  }

  async save() {
    await this.saveButton.click()
    await this.page.waitForResponse(
      (res) => res.url().includes('/api/training/records') && (res.status() === 200 || res.status() === 201)
    )
  }

  async cancel() {
    await this.cancelButton.click()
  }

  async goToLocationsSettings() {
    await this.settingsButton.click()
  }

  getRecordCard(index: number): Locator {
    return this.recordList.locator('[class*="recordCard"]').nth(index)
  }

  getRecordCardByLocation(locationName: string): Locator {
    // 最初にマッチする記録カード（最新の記録）を返す
    return this.recordList.locator('[class*="recordCard"]').filter({
      hasText: locationName
    }).first()
  }

  async editRecord(index: number) {
    const card = this.getRecordCard(index)
    await card.getByRole('button', { name: '編集' }).click()
  }

  async deleteRecord(index: number) {
    const card = this.getRecordCard(index)
    // dialogイベントを先に設定
    const dialogPromise = this.page.waitForEvent('dialog')
    await card.getByRole('button', { name: '削除' }).click()
    const dialog = await dialogPromise
    await dialog.accept()
  }

  async deleteRecordWithConfirm(index: number) {
    const card = this.getRecordCard(index)
    // dialogイベントを先に設定
    const dialogPromise = this.page.waitForEvent('dialog')
    await card.getByRole('button', { name: '削除' }).click()
    const dialog = await dialogPromise
    await dialog.accept()
    // 削除APIのレスポンスを待つ
    await this.page.waitForResponse(
      (res) => res.url().includes('/api/training/records') && res.status() === 204
    )
  }

  async cancelDeleteRecord(index: number) {
    const card = this.getRecordCard(index)
    // dialogイベントを先に設定
    const dialogPromise = this.page.waitForEvent('dialog')
    await card.getByRole('button', { name: '削除' }).click()
    const dialog = await dialogPromise
    await dialog.dismiss()
  }

  async isEquipmentSectionVisible(): Promise<boolean> {
    // 器具セクションは場所選択後に表示される
    const equipmentLabel = this.page.locator('label:has-text("器具")')
    return equipmentLabel.isVisible()
  }

  async isExerciseInputsVisible(exerciseName: string): Promise<boolean> {
    const exerciseItem = this.page.locator('[class*="exerciseItem"]').filter({
      hasText: exerciseName
    })
    const inputs = exerciseItem.locator('[class*="exerciseInputs"]')
    return inputs.isVisible()
  }

  getRecordIndicator(day: number) {
    return this.page.locator(`[class*="calendarDay"]:not([class*="otherMonth"])`).filter({
      has: this.page.locator(`[class*="dayNumber"]:text-is("${day}")`)
    }).locator('[class*="recordCount"]')
  }

  getExercisePreviewItem(recordIndex: number, exerciseName: string): Locator {
    const card = this.getRecordCard(recordIndex)
    return card.locator('[class*="exercisePreviewItem"]').filter({
      hasText: exerciseName
    })
  }

  async getExercisePreviewDetail(recordIndex: number, exerciseName: string): Promise<string | null> {
    const item = this.getExercisePreviewItem(recordIndex, exerciseName)
    const detail = item.locator('[class*="exercisePreviewDetail"]')
    if (await detail.isVisible()) {
      return detail.textContent()
    }
    return null
  }
}
