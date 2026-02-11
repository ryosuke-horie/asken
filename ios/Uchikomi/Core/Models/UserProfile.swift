import Foundation

// MARK: - UserProfile

struct UserProfile: Codable, Equatable {
    let gender: Gender
    let birthDate: Date
    let heightCm: Double
    let activityLevel: ActivityLevel

    enum CodingKeys: String, CodingKey {
        case gender
        case birthDate
        case heightCm = "height_cm"
        case activityLevel
    }
}

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
    case sedentary       // 座業的（ほと運動しない）
    case lightlyActive    // 軽い活動（軽い運動週1-3日）
    case moderatelyActive // 中程度活動（中等度の運動週3-5日）
    case veryActive       // 高活動（激しい運動週6-7日）
    case athlete          // 選手層（毎日の高強度トレーニング）

    var displayName: String {
        switch self {
        case .sedentary: "ほと運動しない"
        case .lightlyActive: "軽い運動（週1-3日）"
        case .moderatelyActive: "中程度の運動（週3-5日）"
        case .veryActive: "激しい運動（週6-7日）"
        case .athlete: "選手層（毎日のトレーニング）"
        }
    }

    /// 活動係数を返す（Harris-Benedict式で使用）
    var activityMultiplier: Double {
        switch self {
        case .sedentary: return 1.2
        case .lightlyActive: return 1.375
        case .moderatelyActive: return 1.55
        case .veryActive: return 1.725
        case .athlete: return 1.9
        }
    }

    /// 選手層向けのプリセットカロリーを返す
    var athletePresetCalories: [Double]? {
        switch self {
        case .athlete:
            return [3000.0, 3500.0, 4000.0, 4500.0]
        case .veryActive:
            return [2500.0, 3000.0, 3500.0, 4000.0]
        default:
            return nil
        }
    }
}

// MARK: - RecommendedCaloriesCalculator

struct RecommendedCaloriesCalculator {

    /// Harris-Benedict式で推奨カロリーを計算
    static func calculate(
        gender: Gender,
        weightKg: Double,
        heightCm: Double,
        age: Int,
        activityLevel: ActivityLevel
    ) -> Double {

        // 基礎代謝（BMR）
        let bmr: Double
        switch gender {
        case .male:
            // 男性: 13.397 × 体重(kg) + 4.799 × 身長 - 年齢 × 0.694 - 55
            bmr = 13.397 * weightKg + 4.799 * heightCm - Double(age) * 0.694 - 55
        case .female:
            // 女性: 9.247 × 体重(kg) + 3.098 × 身長 - 年齢 × 0.428 - 161
            bmr = 9.247 * weightKg + 3.098 * heightCm - Double(age) * 0.428 - 161
        }

        // 活動レベルで補正
        let tdee = bmr * activityLevel.activityMultiplier

        return tdee
    }

    /// 年齢を計算
    static func calculateAge(from birthDate: Date) -> Int {
        let calendar = Calendar.current
        let now = Date()
        let ageComponents = calendar.dateComponents([.year], from: birthDate, to: now)
        return ageComponents.year ?? 0
    }
}
