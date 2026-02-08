import Foundation
import Testing

@testable import Uchikomi

@Suite
struct WeightInputViewModelTests {
    private func makeTestRecord(
        weightKg: Double = 65.3,
        note: String? = "朝食前"
    ) -> WeightRecord {
        WeightRecord(
            id: "record-1",
            weightKg: weightKg,
            recordedAt: "2026-02-08T07:30:00Z",
            note: note,
            createdAt: "2026-02-08T07:35:00Z",
            updatedAt: "2026-02-08T07:35:00Z"
        )
    }

    @Test
    @MainActor
    func 新規作成モードで初期値が空であるべき() {
        let mockRepo = WeightRepositoryProtocolMock()
        let viewModel = WeightInputViewModel(repository: mockRepo)

        #expect(viewModel.weightText == "")
        #expect(viewModel.memo == "")
        #expect(viewModel.isEditing == false)
        #expect(viewModel.isValid == false)
    }

    @Test
    @MainActor
    func 編集モードで既存値が設定されるべき() {
        let mockRepo = WeightRepositoryProtocolMock()
        let record = makeTestRecord()
        let viewModel = WeightInputViewModel(editingRecord: record, repository: mockRepo)

        #expect(viewModel.weightText == "65.3")
        #expect(viewModel.memo == "朝食前")
        #expect(viewModel.isEditing == true)
    }

    @Test
    @MainActor
    func 有効な体重値でisValidがtrueであるべき() {
        let mockRepo = WeightRepositoryProtocolMock()
        let viewModel = WeightInputViewModel(repository: mockRepo)
        viewModel.weightText = "65.0"

        #expect(viewModel.isValid == true)
    }

    @Test
    @MainActor
    func 範囲外の体重値でisValidがfalseであるべき() {
        let mockRepo = WeightRepositoryProtocolMock()
        let viewModel = WeightInputViewModel(repository: mockRepo)

        viewModel.weightText = "10.0"
        #expect(viewModel.isValid == false)

        viewModel.weightText = "500.0"
        #expect(viewModel.isValid == false)

        viewModel.weightText = "abc"
        #expect(viewModel.isValid == false)
    }

    @Test
    @MainActor
    func incrementWeightで0点1kg増加すべき() {
        let mockRepo = WeightRepositoryProtocolMock()
        let viewModel = WeightInputViewModel(repository: mockRepo)
        viewModel.weightText = "65.0"

        viewModel.incrementWeight()

        #expect(viewModel.weightText == "65.1")
    }

    @Test
    @MainActor
    func decrementWeightで0点1kg減少すべき() {
        let mockRepo = WeightRepositoryProtocolMock()
        let viewModel = WeightInputViewModel(repository: mockRepo)
        viewModel.weightText = "65.0"

        viewModel.decrementWeight()

        #expect(viewModel.weightText == "64.9")
    }

    @Test
    @MainActor
    func decrementWeightで20kg未満にならないべき() {
        let mockRepo = WeightRepositoryProtocolMock()
        let viewModel = WeightInputViewModel(repository: mockRepo)
        viewModel.weightText = "20.0"

        viewModel.decrementWeight()

        #expect(viewModel.weightText == "20.0")
    }

    @Test
    @MainActor
    func 新規保存成功時にdidSaveがtrueになるべき() async {
        let mockRepo = WeightRepositoryProtocolMock()
        mockRepo.createRecordHandler = { weightKg, _, note in
            WeightRecord(
                id: "new-record",
                weightKg: weightKg,
                recordedAt: "2026-02-08T07:30:00Z",
                note: note,
                createdAt: "2026-02-08T07:35:00Z",
                updatedAt: "2026-02-08T07:35:00Z"
            )
        }

        let viewModel = WeightInputViewModel(repository: mockRepo)
        viewModel.weightText = "65.3"
        viewModel.memo = "朝食前"

        await viewModel.save()

        #expect(viewModel.didSave == true)
        #expect(viewModel.isSaving == false)
        #expect(viewModel.errorMessage == nil)
        #expect(mockRepo.createRecordCallCount == 1)
    }

    @Test
    @MainActor
    func 更新保存成功時にdidSaveがtrueになるべき() async {
        let mockRepo = WeightRepositoryProtocolMock()
        let record = makeTestRecord()
        mockRepo.updateRecordHandler = { _, weightKg, note in
            WeightRecord(
                id: "record-1",
                weightKg: weightKg,
                recordedAt: "2026-02-08T07:30:00Z",
                note: note,
                createdAt: "2026-02-08T07:35:00Z",
                updatedAt: "2026-02-08T07:35:00Z"
            )
        }

        let viewModel = WeightInputViewModel(editingRecord: record, repository: mockRepo)
        viewModel.weightText = "64.8"

        await viewModel.save()

        #expect(viewModel.didSave == true)
        #expect(mockRepo.updateRecordCallCount == 1)
    }

    @Test
    @MainActor
    func 保存失敗時にエラーメッセージが設定されるべき() async {
        let mockRepo = WeightRepositoryProtocolMock()
        mockRepo.createRecordHandler = { _, _, _ in
            throw NSError(domain: "test", code: 500)
        }

        let viewModel = WeightInputViewModel(repository: mockRepo)
        viewModel.weightText = "65.0"

        await viewModel.save()

        #expect(viewModel.didSave == false)
        #expect(viewModel.errorMessage != nil)
    }

    @Test
    @MainActor
    func クイックノート設定が反映されるべき() {
        let mockRepo = WeightRepositoryProtocolMock()
        let viewModel = WeightInputViewModel(repository: mockRepo)

        viewModel.setQuickNote("起床時")

        #expect(viewModel.memo == "起床時")
    }

    @Test
    @MainActor
    func 削除成功時にdidSaveがtrueになるべき() async {
        let mockRepo = WeightRepositoryProtocolMock()
        let record = makeTestRecord()
        mockRepo.deleteRecordHandler = { _ in }

        let viewModel = WeightInputViewModel(editingRecord: record, repository: mockRepo)

        await viewModel.delete()

        #expect(viewModel.didSave == true)
        #expect(mockRepo.deleteRecordCallCount == 1)
    }

    @Test
    @MainActor
    func 無効な体重値で保存がスキップされるべき() async {
        let mockRepo = WeightRepositoryProtocolMock()
        let viewModel = WeightInputViewModel(repository: mockRepo)
        viewModel.weightText = "abc"

        await viewModel.save()

        #expect(viewModel.didSave == false)
        #expect(mockRepo.createRecordCallCount == 0)
    }
}
