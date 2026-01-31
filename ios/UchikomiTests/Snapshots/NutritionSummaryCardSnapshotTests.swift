import SnapshotTesting
import SwiftUI
import Testing
@testable import Uchikomi

@Suite @MainActor struct NutritionSummaryCardSnapshotTests {

    @Test func 標準データでの表示() {
        let card = NutritionSummaryCard(
            calories: 650,
            protein: 25.5,
            fat: 22.3,
            carbohydrates: 78.0
        )

        let hostingController = UIHostingController(
            rootView: card
                .padding()
                .background(Color(.systemGroupedBackground))
        )
        hostingController.view.frame = CGRect(x: 0, y: 0, width: 390, height: 200)

        assertSnapshot(of: hostingController, as: .image)
    }

    @Test func ゼロ値での表示() {
        let card = NutritionSummaryCard(
            calories: 0,
            protein: 0,
            fat: 0,
            carbohydrates: 0
        )

        let hostingController = UIHostingController(
            rootView: card
                .padding()
                .background(Color(.systemGroupedBackground))
        )
        hostingController.view.frame = CGRect(x: 0, y: 0, width: 390, height: 200)

        assertSnapshot(of: hostingController, as: .image)
    }

    @Test func 大きな値での表示() {
        let card = NutritionSummaryCard(
            calories: 9999,
            protein: 999.9,
            fat: 999.9,
            carbohydrates: 999.9
        )

        let hostingController = UIHostingController(
            rootView: card
                .padding()
                .background(Color(.systemGroupedBackground))
        )
        hostingController.view.frame = CGRect(x: 0, y: 0, width: 390, height: 200)

        assertSnapshot(of: hostingController, as: .image)
    }
}
