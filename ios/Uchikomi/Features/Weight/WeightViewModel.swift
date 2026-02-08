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
    var errorMessage: String?

    private let repository: WeightRepositoryProtocol

    private static let iso8601Formatter: ISO8601DateFormatter = {
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return f
    }()

    private static let iso8601FallbackFormatter: ISO8601DateFormatter = {
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime]
        return f
    }()

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
            async let chartData = loadChartRecords()
            async let goalData = repository.getGoal()

            let (chart, fetchedGoal) = try await (chartData, goalData)
            chartRecords = chart
            goal = fetchedGoal
            todayRecords = filterTodayRecords(from: chart)
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
        return records.filter { record in
            guard let date = Self.parseISO8601(record.recordedAt) else { return false }
            return calendar.isDate(date, inSameDayAs: today)
        }
    }

    static func parseISO8601(_ string: String) -> Date? {
        iso8601Formatter.date(from: string) ?? iso8601FallbackFormatter.date(from: string)
    }
}
