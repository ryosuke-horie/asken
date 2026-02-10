import Foundation
import os

private let logger = Logger(subsystem: Bundle.main.bundleIdentifier ?? "Uchikomi", category: "WeightViewModel")

// MARK: - WeightViewModel

@Observable
final class WeightViewModel {
    var todayRecords: [WeightRecord] = []
    var chartRecords: [WeightRecord] = []
    var goal: WeightGoal?
    var selectedPeriod: ChartPeriod = .week
    var isLoading = false
    var isInitialLoading = true
    var errorMessage: String?

    private let repository: WeightRepositoryProtocol

    init(repository: WeightRepositoryProtocol = WeightRepository()) {
        self.repository = repository
    }

    var latestWeight: Double? {
        todayRecords.last?.weightKg
    }

    var weightDifferenceFromGoal: Double? {
        guard let latest = latestWeight, let goal else { return nil }
        return latest - goal.targetWeightKg
    }

    func loadData() async {
        isLoading = true
        isInitialLoading = true
        errorMessage = nil

        do {
            chartRecords = try await loadChartRecords()
            todayRecords = filterTodayRecords(from: chartRecords)
        } catch let error as APIError {
            logger.error("体重データ取得でAPIエラー: \(error.localizedDescription)")
            errorMessage = error.localizedDescription
            isLoading = false
            isInitialLoading = false
            return
        } catch {
            logger.error("体重データ取得で予期しないエラー: \(error.localizedDescription)")
            errorMessage = "体重データの取得に失敗しました"
            isLoading = false
            isInitialLoading = false
            return
        }

        do {
            goal = try await repository.getGoal()
        } catch {
            logger.warning("目標体重の取得に失敗、データなしで続行: \(error.localizedDescription)")
            goal = nil
        }

        isLoading = false
        isInitialLoading = false
    }

    func loadChartData() async {
        // 初期ロード中は期間変更を無視
        guard !isInitialLoading else { return }

        isLoading = true
        errorMessage = nil

        do {
            chartRecords = try await loadChartRecords()
        } catch let error as APIError {
            logger.error("チャートデータ更新でAPIエラー: \(error.localizedDescription)")
            errorMessage = error.localizedDescription
            isLoading = false
            return
        } catch {
            logger.error("チャートデータ更新で予期しないエラー: \(error.localizedDescription)")
            errorMessage = "チャートデータの更新に失敗しました"
            isLoading = false
            return
        }

        isLoading = false
    }

    func deleteRecord(id: String) async {
        errorMessage = nil

        do {
            try await repository.deleteRecord(id: id)
            await loadData()
        } catch let error as APIError {
            logger.error("体重記録削除でAPIエラー: \(error.localizedDescription)")
            errorMessage = error.localizedDescription
        } catch {
            logger.error("体重記録削除で予期しないエラー: \(error.localizedDescription)")
            errorMessage = "削除に失敗しました"
        }
    }

    private func loadChartRecords() async throws -> [WeightRecord] {
        let to = Date()
        guard let from = Calendar.current.date(byAdding: .day, value: -selectedPeriod.days, to: to) else {
            logger.error("チャート期間の日付計算に失敗: days=\(self.selectedPeriod.days)")
            throw NSError(
                domain: "WeightViewModel",
                code: 0,
                userInfo: [NSLocalizedDescriptionKey: "チャート期間の計算に失敗しました"]
            )
        }
        let response = try await repository.getRecords(from: from, to: to)
        return response.records
    }

    private func filterTodayRecords(from records: [WeightRecord]) -> [WeightRecord] {
        let calendar = Calendar.current
        let today = Date()
        var parseFailureCount = 0

        let filtered = records.filter { record in
            guard let date = WeightRecord.parseISO8601(record.recordedAt) else {
                parseFailureCount += 1
                return false
            }
            return calendar.isDate(date, inSameDayAs: today)
        }

        if parseFailureCount > 0 {
            logger.warning("体重記録の日付パースに失敗: \(parseFailureCount)件")
        }

        return filtered
    }
}
