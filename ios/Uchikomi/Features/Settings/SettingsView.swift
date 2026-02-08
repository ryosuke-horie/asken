import SwiftUI
import UchikomiCore

// MARK: - SettingsView

struct SettingsView: View {
    @Environment(AuthManager.self) private var authManager
    @State private var viewModel: SettingsViewModel?

    var body: some View {
        NavigationStack {
            List {
                if let viewModel {
                    accountSection(viewModel: viewModel)
                    notificationSection
                    logoutSection(viewModel: viewModel)
                }
            }
            .navigationTitle("設定")
            .alert(
                "エラー",
                isPresented: Binding(
                    get: { viewModel?.showLogoutError ?? false },
                    set: { viewModel?.showLogoutError = $0 }
                )
            ) {
                Button("OK", role: .cancel) {}
            } message: {
                Text(viewModel?.logoutErrorMessage ?? "")
            }
        }
        .onAppear {
            if viewModel == nil {
                viewModel = SettingsViewModel(authManager: authManager)
            }
        }
    }

    // MARK: - Sections

    private func accountSection(viewModel: SettingsViewModel) -> some View {
        Section("アカウント") {
            if let name = viewModel.userName {
                LabeledContent("名前", value: name)
            }
            LabeledContent("メール", value: viewModel.userEmail)
        }
    }

    private var notificationSection: some View {
        Section {
            NavigationLink("通知設定") {
                NotificationSettingsView()
            }
        }
    }

    private func logoutSection(viewModel: SettingsViewModel) -> some View {
        Section {
            Button("ログアウト", role: .destructive) {
                viewModel.logout()
            }
        }
    }
}
