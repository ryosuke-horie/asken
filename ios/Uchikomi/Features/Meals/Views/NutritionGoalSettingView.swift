import os
import SwiftUI

private let logger = Logger(subsystem: Bundle.main.bundleIdentifier ?? "Uchikomi", category: "NutritionGoalSettingView")

// MARK: - NutritionGoalSettingView

struct NutritionGoalSettingView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var targetCaloriesText: String
    @State private var isSaving = false
    @State private var errorMessage: String?

    // ユーザー属性入力用
    @State private var showUserProfileInput = false
    @State private var selectedGender: Gender = .male
    @State private var age: Double = 30
    @State private var heightCm: String = "170"
    @State private var weightKg: String = "70"
    @State private var selectedActivityLevel: ActivityLevel = .moderatelyActive

    // 推奨カロリー計算結果
    @State private var recommendedCalories: Double?
    @State private var isCalculating = false

    let currentGoal: NutritionGoal?
    let currentWeight: Double?
    let goalWeight: Double?
    let repository: NutritionGoalRepositoryProtocol
    let onSaved: () -> Void

    init(
        currentGoal: NutritionGoal? = nil,
        currentWeight: Double? = nil,
        goalWeight: Double? = nil,
        repository: NutritionGoalRepositoryProtocol = NutritionGoalRepository(),
        onSaved: @escaping () -> Void = {}
    ) {
        self.currentGoal = currentGoal
        self.currentWeight = currentWeight
        self.goalWeight = goalWeight
        self.repository = repository
        self.onSaved = onSaved
        _targetCaloriesText = State(
            initialValue: currentGoal.map { String(format: "%.0f", $0.targetCalories) } ?? ""
        )
        // 体重の初期値を設定
        _weightKg = State(
            initialValue: currentWeight.map { String(format: "%.1f", $0) } ?? "70"
        )
    }

    private var currentPhase: NutritionPhase {
        NutritionGoalCalculator.calculatePhase(
            currentWeight: currentWeight,
            goalWeight: goalWeight
        )
    }

    private var calculatedPFC: PFCRatios {
        guard let calories = Double(targetCaloriesText) else {
            return PFCRatios(protein: 0, fat: 0, carbs: 0)
        }
        return NutritionGoalCalculator.calculatePFCTargets(
            calories: calories,
            phase: currentPhase
        )
    }

    private var isValid: Bool {
        guard let calories = Double(targetCaloriesText) else { return false }
        return calories >= 1_000 && calories <= 5_000
    }

    var body: some View {
        NavigationStack {
            Form {
                Section("現在のフェーズ") {
                    HStack {
                        Text("フェーズ")
                            .foregroundStyle(.secondary)
                        Spacer()
                        Text(currentPhase.displayName)
                            .fontWeight(.semibold)
                    }

                    Text(currentPhaseDescription)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }

                // 推奨カロリー表示（計算済みの場合）
                if let recommended = recommendedCalories {
                    Section("推奨カロリー") {
                        HStack {
                            VStack(alignment: .leading, spacing: 4) {
                                Text("あなたの推奨値")
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                                Text("\(Int(recommended)) kcal")
                                    .font(.title2)
                                    .fontWeight(.bold)
                            }
                            Spacer()
                            Button("採用") {
                                targetCaloriesText = String(format: "%.0f", recommended)
                            }
                            .buttonStyle(.bordered)
                        }
                    }
                }

                // プリセットクイック選択（選手層向け）
                if let presets = selectedActivityLevel.athletePresetCalories {
                    Section("クイック選択") {
                        ForEach(presets, id: \.self) { preset in
                            Button {
                                targetCaloriesText = String(format: "%.0f", preset)
                            } label: {
                                HStack {
                                    Text("\(Int(preset)) kcal")
                                    Spacer()
                                }
                                .contentShape(Rectangle())
                            }
                        }
                    }
                }

                Section("目標カロリー") {
                    HStack {
                        TextField("2000", text: $targetCaloriesText)
                            .keyboardType(.numberPad)
                            .font(.title2)

                        Text("kcal")
                            .foregroundStyle(.secondary)
                    }
                }

                // ユーザー属性入力（推奨値計算用）
                Section {
                    Button {
                        withAnimation {
                            showUserProfileInput.toggle()
                        }
                    } label: {
                        HStack {
                            Text("推奨値を計算")
                            Spacer()
                            Image(systemName: "chevron.right")
                                .rotationEffect(.degrees(showUserProfileInput ? 90 : 0))
                                .foregroundStyle(.secondary)
                        }
                        .contentShape(Rectangle())
                    }
                    .buttonStyle(.plain)

                    if showUserProfileInput {
                        VStack(spacing: 16) {
                            // 性別選択
                            HStack {
                                Text("性別")
                                    .frame(width: 80, alignment: .leading)
                                Picker("性別", selection: $selectedGender) {
                                    ForEach(Gender.allCases, id: \.self) { gender in
                                        Text(gender.displayName).tag(gender)
                                    }
                                }
                                .pickerStyle(.segmented)
                            }

                            // 年齢入力
                            HStack {
                                Text("年齢")
                                    .frame(width: 80, alignment: .leading)
                                Stepper("\(Int(age)) 歳", value: $age, in: 15...80)
                                    .frame(maxWidth: .infinity, alignment: .leading)
                            }

                            // 身長入力
                            HStack {
                                Text("身長")
                                    .frame(width: 80, alignment: .leading)
                                TextField("170", text: $heightCm)
                                    .keyboardType(.numberPad)
                                    .textFieldStyle(.roundedBorder)
                                    .frame(width: 80)
                                Text("cm")
                                    .foregroundStyle(.secondary)
                            }

                            // 体重入力
                            HStack {
                                Text("体重")
                                    .frame(width: 80, alignment: .leading)
                                TextField("70", text: $weightKg)
                                    .keyboardType(.numberPad)
                                    .textFieldStyle(.roundedBorder)
                                    .frame(width: 80)
                                Text("kg")
                                    .foregroundStyle(.secondary)
                            }

                            // 活動レベル選択
                            VStack(alignment: .leading, spacing: 8) {
                                Text("活動レベル")
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                                Picker("活動レベル", selection: $selectedActivityLevel) {
                                    ForEach(ActivityLevel.allCases, id: \.self) { level in
                                        Text(level.displayName).tag(level)
                                    }
                                }
                                .pickerStyle(.menu)
                            }

                            // 計算ボタン
                            Button {
                                calculateRecommendedCalories()
                            } label: {
                                HStack {
                                    if isCalculating {
                                        ProgressView()
                                            .scaleEffect(0.8)
                                    } else {
                                        Text("推奨カロリーを計算")
                                            .fontWeight(.semibold)
                                    }
                                }
                                .frame(maxWidth: .infinity)
                                .padding()
                                .background(isUserProfileValid ? Color.accentColor : Color(.systemGray5))
                                .foregroundStyle(.white)
                                .cornerRadius(10)
                            }
                            .disabled(isCalculating || !isUserProfileValid)
                        }
                        .padding(.vertical, 8)
                    }
                }

                if isValid {
                    Section("計算されたPFCバランス") {
                        PFCBreakdownRow(
                            name: "たんぱく質",
                            value: calculatedPFC.protein,
                            color: .red
                        )

                        PFCBreakdownRow(
                            name: "脂質",
                            value: calculatedPFC.fat,
                            color: .yellow
                        )

                        PFCBreakdownRow(
                            name: "炭水化物",
                            value: calculatedPFC.carbs,
                            color: .blue
                        )
                    }
                }

                if let errorMessage {
                    Section {
                        Text(errorMessage)
                            .foregroundStyle(.red)
                            .font(.caption)
                    }
                }
            }
            .navigationTitle("栄養目標の設定")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("キャンセル") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("保存") {
                        Task { await save() }
                    }
                    .disabled(!isValid || isSaving)
                }
            }
        }
    }

    private var currentPhaseDescription: String {
        switch currentPhase {
        case .weightLoss:
            "減量中です。高たんぱく質・低脂質のバランスで筋肉を維持しながら体重を減らします。"
        case .maintenance:
            "維持期です。バランスの良いPFC比率で現在の体重を維持します。"
        case .weightGain:
            "増量中です。高脂質のバランスで効率的に体重を増やします。"
        }
    }

    private func save() async {
        guard let calories = Double(targetCaloriesText) else { return }

        isSaving = true
        errorMessage = nil
        defer { isSaving = false }

        do {
            _ = try await repository.setGoal(targetCalories: calories)
            onSaved()
            dismiss()
        } catch let error as APIError {
            logger.error("栄養目標保存でAPIエラー: \(error.localizedDescription)")
            errorMessage = error.localizedDescription
        } catch {
            logger.error("栄養目標保存で予期しないエラー: \(error.localizedDescription)")
            errorMessage = "保存に失敗しました"
        }
    }

    // MARK: - 推奨カロリー計算

    private var isUserProfileValid: Bool {
        guard let height = Double(heightCm),
              let weight = Double(weightKg) else { return false }
        return height >= 100 && height <= 250 && age >= 15 && age <= 80 && weight >= 30 && weight <= 200
    }

    private func calculateRecommendedCalories() {
        // 前回のエラーをクリア
        errorMessage = nil

        guard let height = Double(heightCm),
              let weight = Double(weightKg) else {
            errorMessage = "身長と体重を入力してください"
            return
        }

        // 引数で渡されたcurrentWeightがある場合はそちらを優先
        let weightForCalculation = currentWeight ?? weight

        isCalculating = true
        defer { isCalculating = false }

        // Harris-Benedict式で推奨カロリーを計算
        let recommended = RecommendedCaloriesCalculator.calculate(
            gender: selectedGender,
            weightKg: weightForCalculation,
            heightCm: height,
            age: Int(age),
            activityLevel: selectedActivityLevel
        )

        recommendedCalories = recommended

        let weightSource = currentWeight != nil ? "登録済み" : "入力値"
        logger.debug("推奨カロリー計算: gender=\(selectedGender.displayName), weight=\(weightForCalculation)kg(\(weightSource)), height=\(height)cm, age=\(Int(age))歳, activity=\(selectedActivityLevel.displayName) -> \(Int(recommended))kcal")
    }
}

// MARK: - PFCBreakdownRow

private struct PFCBreakdownRow: View {
    let name: String
    let value: Double
    let color: Color

    var body: some View {
        HStack {
            Circle()
                .fill(color)
                .frame(width: 8, height: 8)

            Text(name)
                .foregroundStyle(.primary)

            Spacer()

            Text(String(format: "%.1fg", value))
                .foregroundStyle(.primary)
        }
    }
}

#Preview {
    NutritionGoalSettingView(
        currentGoal: nil,
        currentWeight: 70.0,
        goalWeight: 65.0,
        onSaved: {}
    )
}

#Preview("目標あり") {
    NutritionGoalSettingView(
        currentGoal: NutritionGoal(
            calories: 2_000,
            protein: 100,
            fat: 60,
            carbohydrates: 250
        ),
        currentWeight: 68.0,
        goalWeight: 65.0,
        onSaved: {}
    )
}

#Preview("体重なし") {
    NutritionGoalSettingView(
        currentGoal: nil,
        currentWeight: nil,
        goalWeight: 65.0,
        onSaved: {}
    )
}
