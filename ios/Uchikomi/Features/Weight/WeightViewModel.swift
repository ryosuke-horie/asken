import Foundation
import os

private let logger = Logger(subsystem: Bundle.main.bundleIdentifier ?? "Uchikomi", category: "WeightViewModel")

@Observable
final class WeightViewModel {
    var todayRecords: [WeightRecord] = []
    var chartRecords: [WeightRecord] = []
    var goal: WeightGoal?
    var selectedPeriod: ChartPeriod = .week
    var isLoading = false
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
        errorMessage = nil

        do {
            async let todayData = loadTodayRecords()
            async let chartData = loadChartRecords()
            async let goalData = repository.getGoal()

            let (today, chart, fetchedGoal) = try await (todayData, chartData, goalData)
            todayRecords = today
            chartRecords = chart
            goal = fetchedGoal
        } catch let error as APIError {
            logger.error("体重データ取得でAPIエラー: \(error.localizedDescription)")
            errorMessage = error.localizedDescription
        } catch {
            logger.error("体重データ取得で予期しないエラー: \(error.localizedDescription)")
            errorMessage = "体重データの取得に失敗しました"
        }

        isLoading = false
    }

    func loadChartData() async {
        do {
            chartRecords = try await loadChartRecords()
        } catch let error as APIError {
            logger.error("チャートデータ更新でAPIエラー: \(error.localizedDescription)")
            errorMessage = error.localizedDescription
        } catch {
            logger.error("チャートデータ更新で予期しないエラー: \(error.localizedDescription)")
            errorMessage = "チャートデータの更新に失敗しました"
        }
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

    private func loadTodayRecords() async throws -> [WeightRecord] {
        let today = Date()
        let response = try await repository.getRecords(from: today, to: today)
        return response.records
    }

    private func loadChartRecords() async throws -> [WeightRecord] {
        let to = Date()
        guard let from = Calendar.current.date(byAdding: .day, value: -selectedPeriod.days, to: to) else {
            logger.error("チャート期間の日付計算に失敗: days=\(self.selectedPeriod.days)")
            return []
        }
        let response = try await repository.getRecords(from: from, to: to)
        return response.records
    }
}
