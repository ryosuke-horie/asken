---
paths:
  - "**/*test*.{ts,tsx,go}"
  - "**/tests/**/*"
  - "**/__tests__/**/*"
---

# テスト駆動開発（TDD）

## テストの原則

1. **新機能実装時は必ずテストを先に書く**（Red → Green → Refactor）
2. テストカバレッジは**80%以上**を目標とする
3. テストは**独立性**を保ち、実行順序に依存しない
4. テストコードも**リファクタリング対象**とする

## フロントエンドテスト

### ツール

- **Jest**: ユニットテスト
- **React Testing Library**: コンポーネントテスト
- **Playwright**: E2Eテスト（オプション、MVP後）

### テスト方針

- コンポーネントは**ユーザーの視点**でテストする
- **モック**は最小限に留め、実際の動作に近い環境でテストする
- API呼び出しはモックする（MSWなどを使用）

```typescript
// ✅ 良い例
test('画像をアップロードすると分析結果が表示される', async () => {
  const file = new File(['dummy'], 'food.jpg', { type: 'image/jpeg' });

  render(<ImageUpload />);

  const input = screen.getByLabelText(/画像をアップロード/i);
  await userEvent.upload(input, file);

  // 分析結果が表示されるまで待機
  expect(await screen.findByText(/カロリー:/i)).toBeInTheDocument();
});
```

## バックエンドテスト

### ツール

- **標準ライブラリ**: `testing`パッケージ
- **testify**: アサーションとモック
- **sqlmock**: データベースモック

### テスト方針

- **テーブル駆動テスト**を使用
- **依存性注入**でモックを容易にする
- Gemini CLI実行は**モック**する（実際のAPI呼び出しはテストしない）

```go
// ✅ 良い例 - Gemini CLI実行のモック
type mockGeminiCaller struct {
    mock.Mock
}

func (m *mockGeminiCaller) AnalyzeImage(ctx context.Context, imagePath string) (*AnalysisResult, error) {
    args := m.Called(ctx, imagePath)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*AnalysisResult), args.Error(1)
}

func TestFoodService_AnalyzeFoodImage(t *testing.T) {
    mockGemini := new(mockGeminiCaller)
    mockGemini.On("AnalyzeImage", mock.Anything, "/path/to/image.jpg").
        Return(&AnalysisResult{
            Foods: []string{"ごはん", "味噌汁"},
            Calories: 450,
        }, nil)

    service := NewFoodService(mockGemini, nil)
    result, err := service.AnalyzeFoodImage(context.Background(), "/path/to/image.jpg")

    require.NoError(t, err)
    assert.Equal(t, 450, result.Calories)
    mockGemini.AssertExpectations(t)
}
```

## テスト実行

```bash
# フロントエンド
cd frontend
npm test              # ユニットテスト
npm run test:watch    # ウォッチモード

# バックエンド
cd backend
go test ./...         # すべてのテスト
go test -cover ./...  # カバレッジ付き
go test -v ./...      # 詳細モード
```
