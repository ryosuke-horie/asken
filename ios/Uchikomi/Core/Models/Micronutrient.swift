import SwiftUI

// MARK: - MicronutrientKey

/// バックエンド側の定義: backend/pkg/gemini/micronutrients.go (AllMicronutrients) と同期を保つこと
enum MicronutrientKey: String, CaseIterable, Identifiable {
    var id: String {
        rawValue
    }

    case ironMg = "iron_mg"
    case calciumMg = "calcium_mg"
    case zincMg = "zinc_mg"
    case fiberG = "fiber_g"
    case vitaminAUg = "vitamin_a_ug"
    case vitaminB1Mg = "vitamin_b1_mg"
    case vitaminB2Mg = "vitamin_b2_mg"
    case vitaminB6Mg = "vitamin_b6_mg"
    case vitaminB12Ug = "vitamin_b12_ug"
    case vitaminCMg = "vitamin_c_mg"
    case vitaminDUg = "vitamin_d_ug"
    case vitaminEMg = "vitamin_e_mg"

    var displayName: String {
        switch self {
        case .ironMg: "鉄分"
        case .calciumMg: "カルシウム"
        case .zincMg: "亜鉛"
        case .fiberG: "食物繊維"
        case .vitaminAUg: "ビタミンA"
        case .vitaminB1Mg: "ビタミンB1"
        case .vitaminB2Mg: "ビタミンB2"
        case .vitaminB6Mg: "ビタミンB6"
        case .vitaminB12Ug: "ビタミンB12"
        case .vitaminCMg: "ビタミンC"
        case .vitaminDUg: "ビタミンD"
        case .vitaminEMg: "ビタミンE"
        }
    }

    var unit: String {
        switch self {
        case .ironMg, .calciumMg, .zincMg,
             .vitaminB1Mg, .vitaminB2Mg, .vitaminB6Mg,
             .vitaminCMg, .vitaminEMg:
            "mg"
        case .vitaminAUg, .vitaminB12Ug, .vitaminDUg:
            "μg"
        case .fiberG:
            "g"
        }
    }

    var color: Color {
        switch self {
        case .ironMg: .red
        case .calciumMg: .cyan
        case .zincMg: .brown
        case .fiberG: .green
        case .vitaminAUg: .orange
        case .vitaminB1Mg: .yellow
        case .vitaminB2Mg: .mint
        case .vitaminB6Mg: .teal
        case .vitaminB12Ug: .pink
        case .vitaminCMg: .orange.opacity(0.7)
        case .vitaminDUg: .indigo
        case .vitaminEMg: .purple
        }
    }

    var defaultTarget: Double {
        switch self {
        case .ironMg: 7.5
        case .calciumMg: 700
        case .zincMg: 10
        case .fiberG: 21
        case .vitaminAUg: 800
        case .vitaminB1Mg: 1.3
        case .vitaminB2Mg: 1.5
        case .vitaminB6Mg: 1.3
        case .vitaminB12Ug: 2.4
        case .vitaminCMg: 100
        case .vitaminDUg: 8.5
        case .vitaminEMg: 6.5
        }
    }

    /// ミネラル・食物繊維のグループ
    static let minerals: [MicronutrientKey] = [.ironMg, .calciumMg, .zincMg, .fiberG]

    /// ビタミンのグループ
    static let vitamins: [MicronutrientKey] = [
        .vitaminAUg, .vitaminB1Mg, .vitaminB2Mg, .vitaminB6Mg,
        .vitaminB12Ug, .vitaminCMg, .vitaminDUg, .vitaminEMg,
    ]

    /// 全栄養素のデフォルト目標値の辞書
    static let defaultTargets: [String: Double] = allCases.reduce(into: [:]) { $0[$1.rawValue] = $1.defaultTarget }
}
