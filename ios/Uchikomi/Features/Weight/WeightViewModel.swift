import Foundation

@Observable
final class WeightViewModel {
    var selectedPeriod: WeightPeriod = .week
    var records: [WeightRecord] = []
    var stats: WeightStats?
    var latestRecord: WeightRecord?
    var goal: WeightGoal?
    var isLoading = false
    var errorMessage: String?

    // Input fields
    var inputWeight = ""
    var inputDate = Date()
    var isSaving = false

    // Goal input
    var goalTargetWeight = ""
    var goalTargetDate = Date()
    var isSettingGoal = false

    private let repository: WeightRepositoryProtocol

    init(repository: WeightRepositoryProtocol = WeightRepository()) {
        self.repository = repository
    }

    var latestWeightText: String {
        guard let latest = latestRecord else { return "-" }
        return String(format: "%.1f", latest.weight)
    }

    var goalDifferenceText: String? {
        guard let latest = latestRecord, let goal = goal else { return nil }
        let diff = latest.weight - goal.targetWeight
        if diff > 0 {
            return String(format: "目標まで %.1f kg", diff)
        } else if diff < 0 {
            return String(format: "目標達成！ %.1f kg 超過", abs(diff))
        } else {
            return "目標達成！"
        }
    }

    func loadData() async {
        isLoading = true
        errorMessage = nil

        do {
            async let recordsTask = repository.getRecords(period: selectedPeriod)
            async let goalTask = repository.getGoal()

            let response = try await recordsTask
            records = response.records
            stats = response.stats
            latestRecord = response.latest

            goal = try await goalTask
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "データの取得に失敗しました"
        }

        isLoading = false
    }

    func saveWeight() async {
        guard let weight = Double(inputWeight) else {
            errorMessage = "体重を正しく入力してください"
            return
        }

        guard weight > 0 && weight < 500 else {
            errorMessage = "有効な体重を入力してください"
            return
        }

        isSaving = true
        errorMessage = nil

        do {
            let newRecord = try await repository.createRecord(weight: weight, recordedAt: inputDate)
            records.insert(newRecord, at: 0)
            latestRecord = newRecord
            inputWeight = ""
            inputDate = Date()
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "保存に失敗しました"
        }

        isSaving = false
    }

    func setGoal() async {
        guard let weight = Double(goalTargetWeight) else {
            errorMessage = "目標体重を正しく入力してください"
            return
        }

        guard weight > 0 && weight < 500 else {
            errorMessage = "有効な体重を入力してください"
            return
        }

        isSettingGoal = true
        errorMessage = nil

        do {
            try await repository.setGoal(targetWeight: weight, targetDate: goalTargetDate)
            goal = try await repository.getGoal()
            goalTargetWeight = ""
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "目標の設定に失敗しました"
        }

        isSettingGoal = false
    }

    func changePeriod(to period: WeightPeriod) {
        selectedPeriod = period
        Task {
            await loadData()
        }
    }
}
