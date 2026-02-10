import SwiftUI

// MARK: - NotificationSettingsView

struct NotificationSettingsView: View {
    @State private var viewModel = NotificationSettingsViewModel()

    var body: some View {
        List {
            globalToggleSection
            if viewModel.settings.isGlobalEnabled {
                weightNotificationSection
                mealNotificationsSection
            }
        }
        .navigationTitle("通知設定")
        .task {
            await viewModel.checkPermission()
        }
        .alert("通知の許可が必要です", isPresented: $viewModel.showPermissionAlert) {
            Button("設定を開く") {
                if let url = URL(string: UIApplication.openSettingsURLString) {
                    UIApplication.shared.open(url)
                }
            }
            Button("キャンセル", role: .cancel) {}
        } message: {
            Text("通知リマインダーを使用するには、設定アプリで通知を許可してください")
        }
    }

    // MARK: - Sections

    private var globalToggleSection: some View {
        Section {
            Toggle("リマインダー通知", isOn: Binding(
                get: { viewModel.settings.isGlobalEnabled },
                set: { _ in
                    Task { await viewModel.toggleGlobalEnabled() }
                }
            ))
        } footer: {
            Text("食事と体重の時間帯に記録のリマインドを通知します")
        }
    }

    private var mealNotificationsSection: some View {
        Section("通知時間") {
            ForEach(
                viewModel.settings.meals.filter { $0.mealType != .snack },
                id: \.mealType
            ) { setting in
                MealNotificationRow(
                    setting: setting,
                    onToggle: {
                        Task { await viewModel.toggleMealEnabled(for: setting.mealType) }
                    },
                    onTimeChange: { hour, minute in
                        Task { await viewModel.updateTime(for: setting.mealType, hour: hour, minute: minute) }
                    }
                )
            }
        }
    }

    private var weightNotificationSection: some View {
        Section("体重リマインダー") {
            WeightNotificationRow(
                setting: viewModel.settings.weight,
                onToggle: {
                    Task { await viewModel.toggleWeightEnabled() }
                },
                onTimeChange: { hour, minute in
                    Task { await viewModel.updateWeightTime(hour: hour, minute: minute) }
                }
            )
        }
    }
}

// MARK: - MealNotificationRow

private struct MealNotificationRow: View {
    let setting: MealNotificationSetting
    let onToggle: () -> Void
    let onTimeChange: (Int, Int) -> Void

    @State private var selectedTime: Date

    init(
        setting: MealNotificationSetting,
        onToggle: @escaping () -> Void,
        onTimeChange: @escaping (Int, Int) -> Void
    ) {
        self.setting = setting
        self.onToggle = onToggle
        self.onTimeChange = onTimeChange

        var components = DateComponents()
        components.hour = setting.hour
        components.minute = setting.minute
        _selectedTime = State(initialValue: Calendar.current.date(from: components) ?? Date())
    }

    var body: some View {
        HStack {
            Image(systemName: setting.mealType.icon)
                .foregroundStyle(setting.isEnabled ? Theme.primary : .secondary)
                .frame(width: 24)

            Toggle(isOn: Binding(
                get: { setting.isEnabled },
                set: { _ in onToggle() }
            )) {
                HStack {
                    Text(setting.mealType.displayName)
                    Spacer()
                    if setting.isEnabled {
                        DatePicker(
                            "",
                            selection: $selectedTime,
                            displayedComponents: .hourAndMinute
                        )
                        .labelsHidden()
                        .onChange(of: selectedTime) { _, newValue in
                            let components = Calendar.current.dateComponents([.hour, .minute], from: newValue)
                            onTimeChange(components.hour ?? 0, components.minute ?? 0)
                        }
                    }
                }
            }
        }
    }
}

// MARK: - WeightNotificationRow

private struct WeightNotificationRow: View {
    let setting: WeightNotificationSetting
    let onToggle: () -> Void
    let onTimeChange: (Int, Int) -> Void

    @State private var selectedTime: Date

    init(
        setting: WeightNotificationSetting,
        onToggle: @escaping () -> Void,
        onTimeChange: @escaping (Int, Int) -> Void
    ) {
        self.setting = setting
        self.onToggle = onToggle
        self.onTimeChange = onTimeChange

        var components = DateComponents()
        components.hour = setting.hour
        components.minute = setting.minute
        _selectedTime = State(initialValue: Calendar.current.date(from: components) ?? Date())
    }

    var body: some View {
        HStack {
            Image(systemName: "scalemass")
                .foregroundStyle(setting.isEnabled ? Theme.primary : .secondary)
                .frame(width: 24)

            Toggle(isOn: Binding(
                get: { setting.isEnabled },
                set: { _ in onToggle() }
            )) {
                HStack {
                    Text("体重")
                    Spacer()
                    if setting.isEnabled {
                        DatePicker(
                            "",
                            selection: $selectedTime,
                            displayedComponents: .hourAndMinute
                        )
                        .labelsHidden()
                        .onChange(of: selectedTime) { _, newValue in
                            let components = Calendar.current.dateComponents([.hour, .minute], from: newValue)
                            onTimeChange(components.hour ?? 0, components.minute ?? 0)
                        }
                    }
                }
            }
        }
    }
}
