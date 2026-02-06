# 共通パターン

## APIレスポンス形式（Go）

```go
type APIResponse[T any] struct {
    Success bool   `json:"success"`
    Data    T      `json:"data,omitempty"`
    Error   string `json:"error,omitempty"`
}

func respondJSON[T any](w http.ResponseWriter, status int, data T) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(APIResponse[T]{
        Success: status < 400,
        Data:    data,
    })
}

func respondError(w http.ResponseWriter, status int, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(APIResponse[any]{
        Success: false,
        Error:   message,
    })
}
```

## エラーハンドリングパターン（Swift）

async/awaitではthrowsでエラーを伝播する:

```swift
enum AppError: Error {
    case networkError(underlying: Error)
    case parseError(message: String)
    case notFound
}

func analyzeImage(_ imageData: Data) async throws -> NutritionData {
    do {
        return try await geminiService.analyze(imageData)
    } catch {
        throw AppError.networkError(underlying: error)
    }
}
```

## Repositoryパターン（Go）

```go
type FoodRepository interface {
    GetByID(ctx context.Context, id string) (*Food, error)
    Search(ctx context.Context, query string) ([]*Food, error)
    Create(ctx context.Context, food *Food) error
}
```

## Repositoryパターン（Swift）

```swift
protocol MealRepositoryProtocol {
    func fetchDailyMeals(date: Date) async throws -> [Meal]
    func saveMeal(_ meal: Meal) async throws
}
```
