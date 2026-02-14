import Foundation

// MARK: - Gender

enum Gender: String, Codable, CaseIterable {
    case male
    case female

    var displayName: String {
        switch self {
        case .male: "男性"
        case .female: "女性"
        }
    }
}

// MARK: - ActivityLevel

enum ActivityLevel: String, Codable, CaseIterable {
    case sedentary // 座業的（ほとんど運動しない）
    case lightlyActive // 軽い活動（軽い運動週1-3日）
    case moderatelyActive // 中程度活動（中等度の運動週3-5日）
    case veryActive // 高活動（激しい運動週6-7日）
    case athlete // 選手層（毎日の高強度トレーニング）

    var displayName: String {
        switch self {
        case .sedentary: "ほとんど運動しない"
        case .lightlyActive: "軽い運動（週1-3日）"
        case .moderatelyActive: "中程度の運動（週3-5日）"
        case .veryActive: "激しい運動（週6-7日）"
        case .athlete: "選手層（毎日のトレーニング）"
        }
    }

    /// 活動係数を返す（Harris-Benedict式で使用）
    var activityMultiplier: Double {
        switch self {
        case .sedentary: 1.2
        case .lightlyActive: 1.375
        case .moderatelyActive: 1.55
        case .veryActive: 1.725
        case .athlete: 1.9
        }
    }
}

// MARK: - RecommendedCaloriesCalculator

enum RecommendedCaloriesCalculator {
    /// Harris-Benedict式で推奨カロリーを計算
    static func calculate(
        gender: Gender,
        weightKg: Double,
        heightCm: Double,
        age: Int,
        activityLevel: ActivityLevel
    ) -> Double {
        // 基礎代謝（BMR）- 改訂Harris-Benedict式
        let bmr = switch gender {
        case .male:
            // 男性: 88.362 + 13.397 × 体重(kg) + 4.799 × 身長(cm) - 5.677 × 年齢
            88.362 + 13.397 * weightKg + 4.799 * heightCm - 5.677 * Double(age)
        case .female:
            // 女性: 447.593 + 9.247 × 体重(kg) + 3.098 × 身長(cm) - 4.330 × 年齢
            447.593 + 9.247 * weightKg + 3.098 * heightCm - 4.330 * Double(age)
        }

        // 活動レベルで補正
        return bmr * activityLevel.activityMultiplier
    }
}
