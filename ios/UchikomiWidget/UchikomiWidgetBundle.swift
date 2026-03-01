import SwiftUI
import WidgetKit

// MARK: - UchikomiWidgetBundle

@main
struct UchikomiWidgetBundle: WidgetBundle {
    var body: some Widget {
        WeightWidget()
        MealWidget()
    }
}
