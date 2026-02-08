import Foundation
import Testing

@testable import Uchikomi

@Suite
struct WeightViewModelTests {
    private func makeTestRecord(
        id: String = "record-1",
        weightKg: Double = 65.3,
        note: String? = nil
    ) -> WeightRecord {
        WeightRecord(
            id: id,
            weightKg: weightKg,
            recordedAt: "2026-02-08T07:30:00Z",
            note: note,
            createdAt: "2026-02-08T07:35:00Z",
            updatedAt: "2026-02-08T07:35:00Z"
        )
    }

    private func makeTestGoal(targetWeightKg: Double = 63.0) -> WeightGoal {
        WeightGoal(targetWeightKg: targetWeightKg, updatedAt: "2026-01-15T10:00:00Z")
    }

    private func makeEmptyResponse() -> WeightRecordsListResponse {
        WeightRecordsListResponse(records: [], dailySummary: [:], goal: nil)
    }

    @Test
    @MainActor
    func データ読み込み成功時に記録と目標が設定されるべき() async {
        let mockRepo = WeightRepositoryProtocolMock()
        let testGoal = makeTestGoal()

        // 今日の日付でレコードを作成（todayRecordsフィルタリングに必要）
        let todayStr = ISO8601DateFormatter().string(from: Date())
        let todayRecord = WeightRecord(
            id: "record-1",
            weightKg: 65.3,
            recordedAt: todayStr,
            note: nil,
            createdAt: todayStr,
            updatedAt: todayStr
        )

        mockRepo.getRecordsHandler = { _, _ in
            WeightRecordsListResponse(
                records: [todayRecord],
                dailySummary: ["2026-02-08": DailySummary(latestWeight: 65.3, count: 1)],
                goal: testGoal
            )
        }
        mockRepo.getGoalHandler = { testGoal }

        let viewModel = WeightViewModel(repository: mockRepo)
        await viewModel.loadData()

        // chartRecordsから今日のレコードがフィルタリングされる
        #expect(viewModel.todayRecords.count == 1)
        #expect(viewModel.todayRecords.first?.weightKg == 65.3)
        #expect(viewModel.goal?.targetWeightKg == 63.0)
        #expect(viewModel.isLoading == false)
        #expect(viewModel.errorMessage == nil)
        // loadDataでgetRecordsは1回のみ呼ばれる（todayRecordsはchartRecordsからフィルタリング）
        #expect(mockRepo.getRecordsCallCount == 1)
    }

    @Test
    @MainActor
    func 最新体重が正しく計算されるべき() async {
        let mockRepo = WeightRepositoryProtocolMock()
        let todayStr = ISO8601DateFormatter().string(from: Date())

        mockRepo.getRecordsHandler = { _, _ in
            WeightRecordsListResponse(
                records: [
                    WeightRecord(
                        id: "r1", weightKg: 65.5, recordedAt: todayStr,
                        note: nil, createdAt: todayStr, updatedAt: todayStr
                    ),
                    WeightRecord(
                        id: "r2", weightKg: 65.1, recordedAt: todayStr,
                        note: nil, createdAt: todayStr, updatedAt: todayStr
                    ),
                ],
                dailySummary: [:],
                goal: nil
            )
        }
        mockRepo.getGoalHandler = { nil }

        let viewModel = WeightViewModel(repository: mockRepo)
        await viewModel.loadData()

        #expect(viewModel.latestWeight == 65.1)
    }

    @Test
    @MainActor
    func 目標との差分が正しく計算されるべき() async {
        let mockRepo = WeightRepositoryProtocolMock()
        let todayStr = ISO8601DateFormatter().string(from: Date())

        mockRepo.getRecordsHandler = { _, _ in
            WeightRecordsListResponse(
                records: [WeightRecord(
                    id: "record-1", weightKg: 65.0, recordedAt: todayStr,
                    note: nil, createdAt: todayStr, updatedAt: todayStr
                )],
                dailySummary: [:],
                goal: self.makeTestGoal(targetWeightKg: 63.0)
            )
        }
        mockRepo.getGoalHandler = { self.makeTestGoal(targetWeightKg: 63.0) }

        let viewModel = WeightViewModel(repository: mockRepo)
        await viewModel.loadData()

        #expect(viewModel.weightDifferenceFromGoal == 2.0)
    }

    @Test
    @MainActor
    func 目標未設定時は差分がnilであるべき() async {
        let mockRepo = WeightRepositoryProtocolMock()

        mockRepo.getRecordsHandler = { _, _ in self.makeEmptyResponse() }
        mockRepo.getGoalHandler = { nil }

        let viewModel = WeightViewModel(repository: mockRepo)
        await viewModel.loadData()

        #expect(viewModel.weightDifferenceFromGoal == nil)
    }

    @Test
    @MainActor
    func データ読み込み失敗時にエラーメッセージが設定されるべき() async {
        let mockRepo = WeightRepositoryProtocolMock()
        mockRepo.getRecordsHandler = { _, _ in
            throw NSError(domain: "test", code: 500)
        }
        mockRepo.getGoalHandler = { nil }

        let viewModel = WeightViewModel(repository: mockRepo)
        await viewModel.loadData()

        #expect(viewModel.errorMessage != nil)
        #expect(viewModel.isLoading == false)
    }

    @Test
    @MainActor
    func 記録削除後にデータが再読み込みされるべき() async {
        let mockRepo = WeightRepositoryProtocolMock()
        mockRepo.deleteRecordHandler = { _ in }
        mockRepo.getRecordsHandler = { _, _ in self.makeEmptyResponse() }
        mockRepo.getGoalHandler = { nil }

        let viewModel = WeightViewModel(repository: mockRepo)
        await viewModel.deleteRecord(id: "record-1")

        #expect(mockRepo.deleteRecordCallCount == 1)
        // loadDataが呼ばれるためgetRecordsも呼ばれる
        #expect(mockRepo.getRecordsCallCount >= 1)
    }

    @Test
    @MainActor
    func 期間変更後にチャートデータが更新されるべき() async {
        let mockRepo = WeightRepositoryProtocolMock()
        mockRepo.getRecordsHandler = { _, _ in self.makeEmptyResponse() }
        mockRepo.getGoalHandler = { nil }

        let viewModel = WeightViewModel(repository: mockRepo)
        viewModel.selectedPeriod = .month
        await viewModel.loadChartData()

        #expect(viewModel.selectedPeriod == .month)
        // getRecordsが呼ばれていることを確認
        #expect(mockRepo.getRecordsCallCount >= 1)
    }

    @Test
    @MainActor
    func 記録削除失敗時にエラーメッセージが設定されるべき() async {
        let mockRepo = WeightRepositoryProtocolMock()
        mockRepo.deleteRecordHandler = { _ in
            throw NSError(domain: "test", code: 500)
        }

        let viewModel = WeightViewModel(repository: mockRepo)
        await viewModel.deleteRecord(id: "record-1")

        #expect(viewModel.errorMessage != nil)
        #expect(mockRepo.deleteRecordCallCount == 1)
    }
}
