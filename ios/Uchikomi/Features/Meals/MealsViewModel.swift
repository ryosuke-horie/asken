import Foundation

@Observable
final class MealsViewModel {
    var selectedDate = Date()
    var dailyMeals: DailyMeals?
    var isLoading = false
    var errorMessage: String?
    var isDeleting = false
    var isRecalculating = false

    private let repository: MealRepositoryProtocol
    private var autoReloadTask: Task<Void, Never>?

    init(repository: MealRepositoryProtocol = MealRepository()) {
        self.repository = repository
    }

    deinit {
        autoReloadTask?.cancel()
    }

    /// 保存後の自動リロードをスケジュール
    /// バックエンドでの非同期栄養素再計算の完了を待つため、10秒後と20秒後にリロード
    func scheduleAutoReload() {
        isRecalculating = true
        autoReloadTask?.cancel()
        autoReloadTask = Task { [weak self] in
            guard let self else { return }

            // deferで確実にisRecalculatingをリセット
            defer {
                Task { @MainActor in
                    self.isRecalculating = false
                }
            }

            for delay in [10, 20] {
                do {
                    try await Task.sleep(nanoseconds: UInt64(delay) * 1_000_000_000)
                } catch {
                    // キャンセルされた場合は早期終了
                    return
                }

                if Task.isCancelled { return }
                await loadMeals()
            }
        }
    }

    var formattedDate: String {
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "ja_JP")
        formatter.dateFormat = "M月d日(E)"
        return formatter.string(from: selectedDate)
    }

    var isToday: Bool {
        Calendar.current.isDateInToday(selectedDate)
    }

    func loadMeals() async {
        isLoading = true
        errorMessage = nil

        do {
            dailyMeals = try await repository.getDailyMeals(date: selectedDate)
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "食事データの取得に失敗しました"
        }

        isLoading = false
    }

    func goToPreviousDay() {
        selectedDate = Calendar.current.date(byAdding: .day, value: -1, to: selectedDate) ?? selectedDate
        Task {
            await loadMeals()
        }
    }

    func goToNextDay() {
        guard !isToday else { return }
        selectedDate = Calendar.current.date(byAdding: .day, value: 1, to: selectedDate) ?? selectedDate
        Task {
            await loadMeals()
        }
    }

    func goToToday() {
        selectedDate = Date()
        Task {
            await loadMeals()
        }
    }

    func deleteHistory(id: String) async {
        isDeleting = true
        errorMessage = nil

        do {
            try await repository.deleteHistory(historyId: id)
            await loadMeals()
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "削除に失敗しました"
        }

        isDeleting = false
    }
}
