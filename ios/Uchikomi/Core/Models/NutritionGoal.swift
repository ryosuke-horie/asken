import Foundation

// MARK: - NutritionGoal

struct NutritionGoal: Codable {
    let targetCalories: Double
    let targetProtein: Double
    let targetFat: Double
    let targetCarbohydrates: Double
    let phase: NutritionPhase
    let updatedAt: String

    enum CodingKeys: String, CodingKey {
        case targetCalories = "target_calories"
        case targetProtein = "target_protein"
        case targetFat = "target_fat"
        case targetCarbohydrates = "target_carbohydrates"
        case phase
        case updatedAt = "updated_at"
    }

    /// NutritionSummaryCardとの互換性のためのエイリアス
    var calories: Double {
        targetCalories
    }

    var protein: Double {
        targetProtein
    }

    var fat: Double {
        targetFat
    }

    var carbohydrates: Double {
        targetCarbohydrates
    }

    /// UIプレビュー用の簡易イニシャライザ
    init(
        calories: Double,
        protein: Double,
        fat: Double,
        carbohydrates: Double,
        phase: NutritionPhase = .maintenance,
        updatedAt: String = ISO8601DateFormatter().string(from: Date())
    ) {
        self.targetCalories = calories
        self.targetProtein = protein
        self.targetFat = fat
        self.targetCarbohydrates = carbohydrates
        self.phase = phase
        self.updatedAt = updatedAt
    }
}

// MARK: - NutritionGoalNullableResponse

struct NutritionGoalNullableResponse: Codable {
    let goal: NutritionGoal?
}

// MARK: - PFCRatios

struct PFCRatios {
    let protein: Double
    let fat: Double
    let carbs: Double
}

// MARK: - NutritionPhase

enum NutritionPhase: String, Codable {
    case weightLoss = "weight_loss"
    case maintenance
    case weightGain = "weight_gain"

    var displayName: String {
        switch self {
        case .weightLoss: "減量期"
        case .maintenance: "維持期"
        case .weightGain: "増量期"
        }
    }

    var pfcRatios: PFCRatios {
        switch self {
        case .weightLoss:
            PFCRatios(protein: 0.30, fat: 0.20, carbs: 0.50)
        case .maintenance:
            PFCRatios(protein: 0.20, fat: 0.25, carbs: 0.55)
        case .weightGain:
            PFCRatios(protein: 0.20, fat: 0.30, carbs: 0.50)
        }
    }
}

// MARK: - SetNutritionGoalRequest

struct SetNutritionGoalRequest: Encodable {
    let targetCalories: Double

    enum CodingKeys: String, CodingKey {
        case targetCalories = "target_calories"
    }
}

// MARK: - NutritionGoalCalculator

enum NutritionGoalCalculator {
    /// 目標体重と現在体重から栄養フェーズを判定
    static func calculatePhase(currentWeight: Double?, goalWeight: Double?) -> NutritionPhase {
        guard let current = currentWeight, let goal = goalWeight else {
            return .maintenance
        }

        let diff = current - goal

        if diff > 1.0 {
            return .weightLoss
        } else if diff < -1.0 {
            return .weightGain
        } else {
            return .maintenance
        }
    }

    /// 目標カロリーとフェーズからPFC目標値を計算
    static func calculatePFCTargets(
        calories: Double,
        phase: NutritionPhase
    ) -> PFCRatios {
        let ratios = phase.pfcRatios

        // たんぱく質: 1g = 4kcal
        let protein = calories * ratios.protein / 4

        // 脂質: 1g = 9kcal
        let fat = calories * ratios.fat / 9

        // 炭水化物: 1g = 4kcal
        let carbs = calories * ratios.carbs / 4

        return PFCRatios(protein: protein, fat: fat, carbs: carbs)
    }
}
