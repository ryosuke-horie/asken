import { test, expect } from '../fixtures'

const TEST_USER = {
  email: 'test@example.com',
  password: 'Pass0123',
}

const MEAL_TYPE_LABELS = {
  breakfast: '朝食',
  lunch: '昼食',
  dinner: '夕食',
  snack: '間食',
} as const

const SEED_MYLIST_ITEMS = [
  { name: '鶏むね肉定食', calories: 434 },
  { name: 'プロテインシェイク', calories: 212 },
  { name: 'サラダチキン', calories: 127 },
  { name: 'オートミール朝食', calories: 306 },
  { name: '焼き鮭定食', calories: 440 },
]

test.describe('食事記録', () => {
  test.beforeEach(async ({ loginPage }) => {
    await loginPage.goto()
    await loginPage.login(TEST_USER.email, TEST_USER.password)
    await expect(loginPage.page).toHaveURL('/')
  })

  test.describe('ページ表示', () => {
    test('朝食ページが正しく表示されるべき', async ({ mealsPage }) => {
      await mealsPage.goto('breakfast')
      await mealsPage.waitForLoad()

      await expect(mealsPage.title).toContainText(MEAL_TYPE_LABELS.breakfast)
      await expect(mealsPage.mylistTab).toBeVisible()
      await expect(mealsPage.imageTab).toBeVisible()
      await expect(mealsPage.textTab).toBeVisible()
      await expect(mealsPage.skippedTab).toBeVisible()
    })

    test('昼食ページが正しく表示されるべき', async ({ mealsPage }) => {
      await mealsPage.goto('lunch')
      await mealsPage.waitForLoad()

      await expect(mealsPage.title).toContainText(MEAL_TYPE_LABELS.lunch)
    })

    test('夕食ページが正しく表示されるべき', async ({ mealsPage }) => {
      await mealsPage.goto('dinner')
      await mealsPage.waitForLoad()

      await expect(mealsPage.title).toContainText(MEAL_TYPE_LABELS.dinner)
    })

    test('間食ページが正しく表示されるべき', async ({ mealsPage }) => {
      await mealsPage.goto('snack')
      await mealsPage.waitForLoad()

      await expect(mealsPage.title).toContainText(MEAL_TYPE_LABELS.snack)
    })

    test('マイリストタブがデフォルトで選択されているべき', async ({ mealsPage }) => {
      await mealsPage.goto('lunch')
      await mealsPage.waitForLoad()

      const firstItem = mealsPage.mylistItems.first()
      await expect(firstItem).toBeVisible()
    })

    test('日付が正しく表示されるべき', async ({ mealsPage }) => {
      const testDate = '2026-01-20'
      await mealsPage.goto('breakfast', testDate)
      await mealsPage.waitForLoad()

      await expect(mealsPage.dateInfo).toContainText('1月20日')
    })
  })

  test.describe('入力タブ切り替え', () => {
    test('画像タブに切り替えられるべき', async ({ mealsPage }) => {
      await mealsPage.goto('lunch')
      await mealsPage.waitForLoad()

      await mealsPage.selectInputType('image')

      const fileInput = mealsPage.page.locator('input[type="file"]')
      await expect(fileInput).toBeAttached()
    })

    test('テキストタブに切り替えられるべき', async ({ mealsPage }) => {
      await mealsPage.goto('lunch')
      await mealsPage.waitForLoad()

      await mealsPage.selectInputType('text')

      await expect(mealsPage.textInputs.first()).toBeVisible()
      await expect(mealsPage.addInputButton).toBeVisible()
      await expect(mealsPage.submitTextButton).toBeVisible()
    })

    test('食べなかったタブに切り替えられるべき', async ({ mealsPage }) => {
      await mealsPage.goto('lunch')
      await mealsPage.waitForLoad()

      await mealsPage.selectInputType('skipped')

      await expect(mealsPage.skippedDescription).toBeVisible()
      await expect(mealsPage.skippedButton).toBeVisible()
    })
  })

  test.describe('マイリストからの記録', () => {
    test('マイリストアイテムを選択して詳細画面が表示されるべき', async ({ mealsPage }) => {
      await mealsPage.goto('lunch')
      await mealsPage.waitForLoad()

      await mealsPage.selectMylistItem(SEED_MYLIST_ITEMS[0].name)

      await expect(mealsPage.quantitySection).toBeVisible()
      await expect(mealsPage.mylistSubmitButton).toBeVisible()
      await expect(mealsPage.mylistBackButton).toBeVisible()
    })

    test('マイリストアイテムを選択して記録できるべき', async ({ mealsPage }) => {
      await mealsPage.goto('lunch')
      await mealsPage.waitForLoad()

      const initialCount = await mealsPage.getMealsCount()

      await mealsPage.selectMylistItem(SEED_MYLIST_ITEMS[0].name)
      await mealsPage.submitMylistItem()

      await expect(mealsPage.mealItems).toHaveCount(initialCount + 1, { timeout: 10000 })
    })

    test('一覧に戻るボタンで選択をキャンセルできるべき', async ({ mealsPage }) => {
      await mealsPage.goto('lunch')
      await mealsPage.waitForLoad()

      await mealsPage.selectMylistItem(SEED_MYLIST_ITEMS[0].name)
      await expect(mealsPage.quantitySection).toBeVisible()

      await mealsPage.mylistBackButton.click()

      await expect(mealsPage.quantitySection).not.toBeVisible()
      await expect(mealsPage.mylistItems.first()).toBeVisible()
    })
  })

  test.describe('食べなかった記録', () => {
    test('食事をスキップとして記録できるべき', async ({ mealsPage }) => {
      await mealsPage.goto('snack')
      await mealsPage.waitForLoad()

      await mealsPage.selectInputType('skipped')

      await expect(mealsPage.skippedDescription).toBeVisible()
      await expect(mealsPage.skippedButton).toBeVisible()

      await mealsPage.submitSkipped()

      const skippedItem = mealsPage.getSkippedMealItem()
      await expect(skippedItem).toBeVisible({ timeout: 10000 })
    })
  })

  test.describe('テキスト入力', () => {
    test('テキスト入力フィールドに入力できるべき', async ({ mealsPage }) => {
      await mealsPage.goto('lunch')
      await mealsPage.waitForLoad()

      await mealsPage.selectInputType('text')

      await mealsPage.fillTextInput(0, 'ご飯一杯')

      const input = mealsPage.textInputs.first()
      await expect(input).toHaveValue('ご飯一杯')
    })

    test('テキスト入力フィールドを追加できるべき', async ({ mealsPage }) => {
      await mealsPage.goto('lunch')
      await mealsPage.waitForLoad()

      await mealsPage.selectInputType('text')

      const initialCount = await mealsPage.textInputs.count()
      await mealsPage.addTextInput()

      const afterCount = await mealsPage.textInputs.count()
      expect(afterCount).toBe(initialCount + 1)
    })

    test('空欄では解析ボタンが無効になるべき', async ({ mealsPage }) => {
      await mealsPage.goto('lunch')
      await mealsPage.waitForLoad()

      await mealsPage.selectInputType('text')

      await expect(mealsPage.submitTextButton).toBeDisabled()
    })
  })

  test.describe('既存記録の操作', () => {
    test('登録済み食事を削除できるべき', async ({ mealsPage }) => {
      await mealsPage.goto('lunch')
      await mealsPage.waitForLoad()

      await mealsPage.selectMylistItem(SEED_MYLIST_ITEMS[1].name)
      await mealsPage.submitMylistItem()
      await expect(mealsPage.mealItems.first()).toBeVisible({ timeout: 10000 })

      const initialCount = await mealsPage.getMealsCount()

      await mealsPage.deleteMeal(0)

      await expect(mealsPage.mealItems).toHaveCount(initialCount - 1, { timeout: 10000 })
    })
  })

  test.describe('ナビゲーション', () => {
    test('戻るボタンでホーム画面に戻れるべき', async ({ mealsPage }) => {
      const today = new Date().toISOString().split('T')[0]
      await mealsPage.goto('dinner', today)
      await mealsPage.waitForLoad()

      await mealsPage.backButton.click()

      await expect(mealsPage.page).toHaveURL(`/?date=${today}`)
    })
  })
})
