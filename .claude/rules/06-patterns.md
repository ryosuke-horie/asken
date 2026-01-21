# 共通パターン

## APIレスポンス形式

```typescript
interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: string
  meta?: {
    total: number
    page: number
    limit: number
  }
}
```

## カスタムフックパターン

```typescript
export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value)

  useEffect(() => {
    const handler = setTimeout(() => setDebouncedValue(value), delay)
    return () => clearTimeout(handler)
  }, [value, delay])

  return debouncedValue
}
```

## Result型パターン

エラーハンドリングにはResult型を使用:

```typescript
type Result<T, E = Error> =
  | { success: true; data: T }
  | { success: false; error: E }

async function analyzeImage(imagePath: string): Promise<Result<NutritionData>> {
  try {
    const data = await geminiAnalyze(imagePath)
    return { success: true, data }
  } catch (error) {
    return { success: false, error: error as Error }
  }
}
```
