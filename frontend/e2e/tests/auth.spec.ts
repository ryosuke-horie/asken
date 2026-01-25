import { test, expect } from '../fixtures'

// テストユーザー（db-seedで作成されるデータ）
const TEST_USER = {
  email: 'test@example.com',
  password: 'Pass0123',
}

const INVALID_USER = {
  email: 'invalid@example.com',
  password: 'wrongpassword',
}

test.describe('認証フロー', () => {
  test.beforeEach(async ({ page }) => {
    // 各テスト前に認証状態をクリア
    await page.goto('/login')
    await page.evaluate(() => {
      localStorage.removeItem('uchikomi_auth_token')
      localStorage.removeItem('uchikomi_user')
      document.cookie = 'uchikomi_auth_token=; path=/; max-age=0'
    })
  })

  test.describe('ログイン', () => {
    test('正しい認証情報でログインできるべき', async ({ loginPage, homePage }) => {
      await loginPage.goto()
      await loginPage.waitForLoad()

      await loginPage.login(TEST_USER.email, TEST_USER.password)

      // ホームページにリダイレクトされることを確認
      await expect(loginPage.page).toHaveURL('/')

      // 体重管理セクションが表示されることを確認
      const isVisible = await homePage.isWeightSectionVisible()
      expect(isVisible).toBe(true)
    })

    test('誤った認証情報でエラーを表示すべき', async ({ loginPage }) => {
      await loginPage.goto()
      await loginPage.waitForLoad()

      await loginPage.login(INVALID_USER.email, INVALID_USER.password)

      // エラーメッセージが表示されることを確認
      const errorMessage = await loginPage.getErrorMessage()
      expect(errorMessage).not.toBeNull()
      expect(errorMessage).toContain('メールアドレスまたはパスワード')

      // ログインページに留まることを確認
      await expect(loginPage.page).toHaveURL('/login')
    })

    test('空のフォームで送信できないべき', async ({ loginPage }) => {
      await loginPage.goto()
      await loginPage.waitForLoad()

      // 空のまま送信ボタンをクリック
      await loginPage.submitButton.click()

      // ログインページに留まることを確認（HTML5バリデーションで送信されない）
      await expect(loginPage.page).toHaveURL('/login')
    })

    test('新規登録ページへのリンクが機能すべき', async ({ loginPage }) => {
      await loginPage.goto()
      await loginPage.waitForLoad()

      await loginPage.registerLink.click()

      await expect(loginPage.page).toHaveURL('/register')
    })
  })

  test.describe('ログアウト', () => {
    test('ログアウト後にログインページにリダイレクトすべき', async ({
      loginPage,
      homePage,
    }) => {
      // まずログイン
      await loginPage.goto()
      await loginPage.waitForLoad()
      await loginPage.login(TEST_USER.email, TEST_USER.password)
      await expect(loginPage.page).toHaveURL('/')

      // ログアウト
      await homePage.logout()

      // ログインページにリダイレクトされることを確認
      await expect(homePage.page).toHaveURL('/login')
    })
  })

  test.describe('未認証アクセス', () => {
    test('未認証でホームページにアクセスするとログインページにリダイレクトすべき', async ({
      homePage,
    }) => {
      await homePage.goto()

      // ログインページにリダイレクトされることを確認
      await expect(homePage.page).toHaveURL('/login')
    })
  })
})
