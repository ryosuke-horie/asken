import Foundation

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
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "体重データの取得に失敗しました"
        }

        isLoading = false
    }

    func loadChartData() async {
        do {
            chartRecords = try await loadChartRecords()
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "チャートデータの更新に失敗しました"
        }
    }

    func deleteRecord(id: String) async {
        errorMessage = nil

        do {
            try await repository.deleteRecord(id: id)
            await loadData()
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
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
        let from = Calendar.current.date(byAdding: .day, value: -selectedPeriod.days, to: to) ?? to
        let response = try await repository.getRecords(from: from, to: to)
        return response.records
    }
}
