import { test, expect } from '../fixtures'

const TEST_USER = {
  email: 'test@example.com',
  password: 'Pass0123',
}

test.describe('トレーニング管理', () => {
  test.beforeEach(async ({ loginPage }) => {
    await loginPage.goto()
    await loginPage.login(TEST_USER.email, TEST_USER.password)
    await expect(loginPage.page).toHaveURL('/')
  })

  test.describe('場所管理', () => {
    test('場所を新規作成できるべき', async ({ locationsPage }) => {
      await locationsPage.goto()
      await locationsPage.waitForLoad()

      const locationName = `テスト場所_${Date.now()}`
      await locationsPage.createLocation(locationName)

      await expect(locationsPage.getLocationCard(locationName)).toBeVisible()
    })

    test('場所の名前を編集できるべき', async ({ locationsPage }) => {
      await locationsPage.goto()
      await locationsPage.waitForLoad()

      // テスト用の場所を作成
      const originalName = `編集前_${Date.now()}`
      await locationsPage.createLocation(originalName)
      await expect(locationsPage.getLocationCard(originalName)).toBeVisible()

      // 名前を編集
      const newName = `編集後_${Date.now()}`
      await locationsPage.editLocation(originalName, newName)

      await expect(locationsPage.getLocationCard(newName)).toBeVisible()
      await expect(locationsPage.getLocationCard(originalName)).not.toBeVisible()
    })

    test('場所を削除できるべき', async ({ locationsPage }) => {
      await locationsPage.goto()
      await locationsPage.waitForLoad()

      // テスト用の場所を作成
      const locationName = `削除対象_${Date.now()}`
      await locationsPage.createLocation(locationName)
      await expect(locationsPage.getLocationCard(locationName)).toBeVisible()

      // 削除
      await locationsPage.deleteLocation(locationName)

      await expect(locationsPage.getLocationCard(locationName)).not.toBeVisible()
    })

    test('削除確認ダイアログでキャンセルすると場所が残るべき', async ({ locationsPage }) => {
      await locationsPage.goto()
      await locationsPage.waitForLoad()

      // テスト用の場所を作成
      const locationName = `キャンセル対象_${Date.now()}`
      await locationsPage.createLocation(locationName)
      await expect(locationsPage.getLocationCard(locationName)).toBeVisible()

      // 削除キャンセル
      await locationsPage.cancelDeleteLocation(locationName)

      // 場所がまだ存在することを確認
      await expect(locationsPage.getLocationCard(locationName)).toBeVisible()
    })
  })

  test.describe('器具管理', () => {
    test.beforeEach(async ({ locationsPage, page }) => {
      // テスト用の場所を作成
      await locationsPage.goto()
      await locationsPage.waitForLoad()

      const locationName = `器具テスト場所_${Date.now()}`
      await locationsPage.createLocation(locationName)

      // 作成した場所の器具設定画面に遷移
      await locationsPage.goToEquipmentSettings(locationName)
      await page.waitForURL(/\/training\/locations\/.*\/equipment/)
    })

    test('器具を新規作成できるべき', async ({ equipmentPage }) => {
      await equipmentPage.waitForLoad()

      const equipmentName = `テスト器具_${Date.now()}`
      await equipmentPage.createEquipment(equipmentName)

      await expect(equipmentPage.getEquipmentCard(equipmentName)).toBeVisible()
    })

    test('器具の名前を編集できるべき', async ({ equipmentPage }) => {
      await equipmentPage.waitForLoad()

      // テスト用の器具を作成
      const originalName = `編集前器具_${Date.now()}`
      await equipmentPage.createEquipment(originalName)
      await expect(equipmentPage.getEquipmentCard(originalName)).toBeVisible()

      // 名前を編集
      const newName = `編集後器具_${Date.now()}`
      await equipmentPage.editEquipment(originalName, newName)

      await expect(equipmentPage.getEquipmentCard(newName)).toBeVisible()
      await expect(equipmentPage.getEquipmentCard(originalName)).not.toBeVisible()
    })

    test('器具を削除できるべき', async ({ equipmentPage }) => {
      await equipmentPage.waitForLoad()

      // テスト用の器具を作成
      const equipmentName = `削除対象器具_${Date.now()}`
      await equipmentPage.createEquipment(equipmentName)
      await expect(equipmentPage.getEquipmentCard(equipmentName)).toBeVisible()

      // 削除
      await equipmentPage.deleteEquipment(equipmentName)

      await expect(equipmentPage.getEquipmentCard(equipmentName)).not.toBeVisible()
    })

    test('削除確認ダイアログでキャンセルすると器具が残るべき', async ({ equipmentPage }) => {
      await equipmentPage.waitForLoad()

      // テスト用の器具を作成
      const equipmentName = `キャンセル対象器具_${Date.now()}`
      await equipmentPage.createEquipment(equipmentName)
      await expect(equipmentPage.getEquipmentCard(equipmentName)).toBeVisible()

      // 削除キャンセル
      await equipmentPage.cancelDeleteEquipment(equipmentName)

      // 器具がまだ存在することを確認
      await expect(equipmentPage.getEquipmentCard(equipmentName)).toBeVisible()
    })
  })

  test.describe('練習記録', () => {
    // 並行実行で同じリソースを作成しようとして重複エラーになるのを防ぐ
    test.describe.configure({ mode: 'serial' })

    const RECORD_TEST_LOCATION = 'E2Eテスト用ジム'
    const RECORD_TEST_EQUIPMENT = 'E2Eテスト用器具'

    test.beforeEach(async ({ locationsPage, page }) => {
      // テスト用の場所を作成（既存でなければ）
      await locationsPage.goto()
      await locationsPage.waitForLoad()

      // 場所一覧のロードを待つ（空の場合はemptyMessageが表示される）
      await Promise.race([
        locationsPage.getLocationCard(RECORD_TEST_LOCATION).waitFor({ state: 'visible', timeout: 5000 }),
        locationsPage.emptyMessage.waitFor({ state: 'visible', timeout: 5000 }),
      ]).catch(() => {})

      const locationCard = locationsPage.getLocationCard(RECORD_TEST_LOCATION)
      const isLocationVisible = await locationCard.isVisible().catch(() => false)
      if (!isLocationVisible) {
        await locationsPage.newNameInput.fill(RECORD_TEST_LOCATION)
        await locationsPage.createButton.click()
        // 201（成功）または500（重複エラー）を待つ
        await page.waitForResponse(
          (res) => res.url().includes('/api/training/locations') &&
                   !res.url().includes('/equipment') &&
                   (res.status() === 201 || res.status() === 500)
        )
        // 場所カードが表示されるまで待つ
        await locationCard.waitFor({ state: 'visible', timeout: 5000 })
      }

      // テスト用の器具を作成（場所に紐づく）
      await locationsPage.goToEquipmentSettings(RECORD_TEST_LOCATION)
      await page.waitForURL(/\/training\/locations\/.*\/equipment/)

      // 器具ページのロードを待つ
      await page.getByRole('heading', { name: '器具設定' }).waitFor({ state: 'visible' })
      await page.getByText('読み込み中...').waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {})

      // 器具一覧のロードを待つ
      await Promise.race([
        page.locator('[class*="equipmentCard"]').first().waitFor({ state: 'visible', timeout: 5000 }),
        page.getByText('登録されている器具はありません').waitFor({ state: 'visible', timeout: 5000 }),
      ]).catch(() => {})

      const equipmentCard = page.locator('[class*="equipmentCard"]').filter({
        hasText: RECORD_TEST_EQUIPMENT
      })
      const equipmentVisible = await equipmentCard.isVisible().catch(() => false)
      if (!equipmentVisible) {
        await page.getByPlaceholder('新しい器具の名前').fill(RECORD_TEST_EQUIPMENT)
        await page.getByRole('button', { name: '追加', exact: true }).click()
        // 201（成功）または500（重複エラー）を待つ
        await page.waitForResponse(
          (res) => res.url().includes('/api/training/equipment') &&
                   (res.status() === 201 || res.status() === 500)
        )
        // 器具カードが表示されるまで待つ
        await equipmentCard.waitFor({ state: 'visible', timeout: 5000 }).catch(() => {})
      }
    })

    test('記録一覧を表示できるべき', async ({ trainingPage }) => {
      await trainingPage.goto()
      await trainingPage.waitForLoad()

      await expect(trainingPage.title).toBeVisible()
      await expect(trainingPage.calendarContainer).toBeVisible()
      await expect(trainingPage.monthLabel).toBeVisible()
    })

    test('場所を選択すると器具セクションが表示されるべき', async ({ trainingPage }) => {
      await trainingPage.goto()
      await trainingPage.waitForLoad()

      // 今日の日付を選択
      const today = new Date()
      await trainingPage.selectDate(today.getDate())

      // 場所を選択
      await trainingPage.selectLocation(RECORD_TEST_LOCATION)

      // 器具セクションが表示されることを確認
      const isVisible = await trainingPage.isEquipmentSectionVisible()
      expect(isVisible).toBe(true)

      // 器具のチェックボックスが表示されることを確認
      const equipmentCheckbox = trainingPage.getEquipmentCheckbox(RECORD_TEST_EQUIPMENT)
      await expect(equipmentCheckbox).toBeVisible()
    })

    test('練習メニューを選択するとセット・回数入力欄が表示されるべき', async ({ trainingPage }) => {
      await trainingPage.goto()
      await trainingPage.waitForLoad()

      // 今日の日付を選択
      const today = new Date()
      await trainingPage.selectDate(today.getDate())

      // 固定メニュー「スパーリング」を選択
      const menuName = 'スパーリング'
      await trainingPage.selectMenu(menuName)

      // セット・回数入力欄が表示されることを確認
      const isInputsVisible = await trainingPage.isExerciseInputsVisible(menuName)
      expect(isInputsVisible).toBe(true)
    })

    test('器具を選択するとセット・回数入力欄が表示されるべき', async ({ trainingPage }) => {
      await trainingPage.goto()
      await trainingPage.waitForLoad()

      // 今日の日付を選択
      const today = new Date()
      await trainingPage.selectDate(today.getDate())

      // 場所を選択（器具表示のため）
      await trainingPage.selectLocation(RECORD_TEST_LOCATION)

      // 器具を選択
      await trainingPage.selectEquipment(RECORD_TEST_EQUIPMENT)

      // セット・回数入力欄が表示されることを確認
      const isInputsVisible = await trainingPage.isExerciseInputsVisible(RECORD_TEST_EQUIPMENT)
      expect(isInputsVisible).toBe(true)
    })

    test('練習記録を詳細に作成できるべき', async ({ trainingPage }) => {
      await trainingPage.goto()
      await trainingPage.waitForLoad()

      // 今日の日付を選択
      const today = new Date()
      await trainingPage.selectDate(today.getDate())

      // 場所を選択
      await trainingPage.selectLocation(RECORD_TEST_LOCATION)

      // 練習時間を設定
      await trainingPage.setDuration(60)

      // メニューを選択してセット・回数を入力
      await trainingPage.selectMenu('スパーリング')
      await trainingPage.setMenuSetsReps('スパーリング', 3, 5)

      // 器具を選択してセット・回数を入力
      await trainingPage.selectEquipment(RECORD_TEST_EQUIPMENT)
      await trainingPage.setEquipmentSetsReps(RECORD_TEST_EQUIPMENT, 2, 10)

      // 強度と満足度を設定
      await trainingPage.setIntensity(4)
      await trainingPage.setSatisfaction(5)

      // メモを入力
      await trainingPage.setNotes('E2Eテスト用メモ')

      // 保存
      await trainingPage.save()

      // 記録カードが表示されることを確認
      const recordCard = trainingPage.getRecordCardByLocation(RECORD_TEST_LOCATION)
      await expect(recordCard).toBeVisible()

      // 練習時間が表示されることを確認
      await expect(recordCard.getByText('60分')).toBeVisible()
    })

    test('練習記録のプレビューにセット・回数が表示されるべき', async ({ trainingPage }) => {
      await trainingPage.goto()
      await trainingPage.waitForLoad()

      // 今日の日付を選択
      const today = new Date()
      await trainingPage.selectDate(today.getDate())

      // 場所を選択
      await trainingPage.selectLocation(RECORD_TEST_LOCATION)

      // 練習時間を設定
      await trainingPage.setDuration(45)

      // メニューを選択してセット・回数を入力
      await trainingPage.selectMenu('打ち込み・ミット')
      await trainingPage.setMenuSetsReps('打ち込み・ミット', 4, 8)

      // 強度と満足度を設定
      await trainingPage.setIntensity(3)
      await trainingPage.setSatisfaction(4)

      // 保存
      await trainingPage.save()

      // 記録カードのプレビューにセット・回数が表示されることを確認
      const recordCard = trainingPage.getRecordCardByLocation(RECORD_TEST_LOCATION)
      await expect(recordCard).toBeVisible()

      // エクササイズプレビューにセット・回数が表示される
      const exercisePreview = recordCard.locator('[class*="exercisePreviewItem"]').filter({
        hasText: '打ち込み・ミット'
      })
      await expect(exercisePreview).toBeVisible()
      await expect(exercisePreview.locator('[class*="exercisePreviewDetail"]')).toContainText('4セット')
      await expect(exercisePreview.locator('[class*="exercisePreviewDetail"]')).toContainText('8回')
    })

    test('練習記録を更新できるべき', async ({ trainingPage }) => {
      await trainingPage.goto()
      await trainingPage.waitForLoad()

      // 今日の日付を選択
      const today = new Date()
      await trainingPage.selectDate(today.getDate())

      // まず記録を作成
      await trainingPage.selectLocation(RECORD_TEST_LOCATION)
      await trainingPage.setDuration(30)
      await trainingPage.selectMenu('スパーリング')
      await trainingPage.setIntensity(3)
      await trainingPage.save()

      // 記録カードが表示されることを確認
      let recordCard = trainingPage.getRecordCardByLocation(RECORD_TEST_LOCATION)
      await expect(recordCard).toBeVisible()
      await expect(recordCard.getByText('30分')).toBeVisible()

      // 編集ボタンをクリック
      await recordCard.getByRole('button', { name: '編集' }).click()

      // 練習時間を変更
      await trainingPage.setDuration(45)

      // 更新ボタンをクリック
      await trainingPage.page.getByRole('button', { name: '更新' }).click()
      await trainingPage.page.waitForResponse(
        (res) => res.url().includes('/api/training/records') && res.status() === 200
      )

      // 更新後の値が反映されていることを確認
      recordCard = trainingPage.getRecordCardByLocation(RECORD_TEST_LOCATION)
      await expect(recordCard.getByText('45分')).toBeVisible()
    })

    test('練習記録を削除できるべき', async ({ trainingPage, page }) => {
      await trainingPage.goto()
      await trainingPage.waitForLoad()

      // 今日の日付を選択
      const today = new Date()
      await trainingPage.selectDate(today.getDate())

      // 削除前の記録数を保存
      const recordCards = trainingPage.recordList.locator('[class*="recordCard"]:not([class*="recordCardHeader"])')
      const countBefore = await recordCards.count()

      // 記録がない場合は作成
      if (countBefore === 0) {
        await trainingPage.selectLocation(RECORD_TEST_LOCATION)
        await trainingPage.setDuration(99)
        await trainingPage.selectMenu('技練習（ドリル）')
        await trainingPage.save()
      }

      // 記録カードが表示されていることを確認
      const recordCardsAfterSetup = trainingPage.recordList.locator('[class*="recordCard"]:not([class*="recordCardHeader"])')
      const countAfterSetup = await recordCardsAfterSetup.count()
      expect(countAfterSetup).toBeGreaterThan(0)

      // 最初の記録カードを削除
      const firstRecordCard = recordCardsAfterSetup.first()
      const deleteButton = firstRecordCard.getByRole('button', { name: '削除' })

      // 削除されるレコードのIDを取得（DELETEリクエストのURLから）
      let deletedRecordId = ''
      await Promise.all([
        page.waitForEvent('dialog').then(dialog => dialog.accept()),
        page.waitForResponse((res) => {
          const isDelete = res.url().includes('/api/training/records') && res.request().method() === 'DELETE'
          if (isDelete && res.status() === 204) {
            // URLからIDを抽出: /api/training/records/{id}
            const urlParts = res.url().split('/')
            deletedRecordId = urlParts[urlParts.length - 1]
            return true
          }
          return false
        }),
        deleteButton.click(),
      ])

      // レコードIDが取得できたことを確認
      expect(deletedRecordId).not.toBe('')

      // ページをリロードして最新データを取得
      await page.reload()
      await trainingPage.waitForLoad()

      // 今日の日付を再選択
      await trainingPage.selectDate(today.getDate())

      // 記録の数が1つ減っていることを確認
      const countAfterDelete = await recordCardsAfterSetup.count()
      expect(countAfterDelete).toBe(countAfterSetup - 1)
    })

    test('削除確認ダイアログでキャンセルすると記録が残るべき', async ({ trainingPage, page }) => {
      await trainingPage.goto()
      await trainingPage.waitForLoad()

      // 今日の日付を選択
      const today = new Date()
      await trainingPage.selectDate(today.getDate())

      // 記録カードを取得
      const recordCards = trainingPage.recordList.locator('[class*="recordCard"]:not([class*="recordCardHeader"])')

      // 記録がない場合は作成
      const countBefore = await recordCards.count()
      if (countBefore === 0) {
        await trainingPage.selectLocation(RECORD_TEST_LOCATION)
        await trainingPage.setDuration(88)
        await trainingPage.selectMenu('対人練習')
        await trainingPage.save()
      }

      // 記録カードが表示されていることを確認
      const countAfterSetup = await recordCards.count()
      expect(countAfterSetup).toBeGreaterThan(0)

      // 最初の記録カードの削除をキャンセル
      const firstRecordCard = recordCards.first()
      const deleteButton = firstRecordCard.getByRole('button', { name: '削除' })

      // ダイアログハンドラーを設定（confirmダイアログを拒否）
      page.once('dialog', async (dialog) => {
        await dialog.dismiss()
      })

      // 削除ボタンをクリック
      await deleteButton.click()

      // 記録数が変わっていないことを確認
      await expect(recordCards).toHaveCount(countAfterSetup)
    })
  })
})
