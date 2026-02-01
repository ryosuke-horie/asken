import SwiftUI

// MARK: - LoginView

struct LoginView: View {
    @Environment(AuthManager.self) private var authManager
    @State private var viewModel: LoginViewModel?

    var body: some View {
        NavigationStack {
            VStack(spacing: 24) {
                Spacer()

                // Logo & Title
                VStack(spacing: 8) {
                    Image(systemName: "figure.martial.arts")
                        .font(.system(size: 60))
                        .foregroundStyle(Theme.primary)

                    Text("ウチコミ")
                        .font(.largeTitle)
                        .fontWeight(.bold)

                    Text("減量・体重管理アプリ")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }

                Spacer()

                // Form
                if let viewModel {
                    LoginForm(viewModel: viewModel)
                }

                Spacer()
            }
            .padding()
            .navigationTitle("")
        }
        .onAppear {
            if viewModel == nil {
                viewModel = LoginViewModel(authManager: authManager)
            }
        }
    }
}

// MARK: - LoginForm

private struct LoginForm: View {
    @Bindable var viewModel: LoginViewModel

    var body: some View {
        VStack(spacing: 16) {
            // Email Field
            TextField("メールアドレス", text: $viewModel.email)
                .textContentType(.emailAddress)
                .keyboardType(.emailAddress)
                .autocapitalization(.none)
                .textFieldStyle(.roundedBorder)

            // Password Field
            SecureField("パスワード", text: $viewModel.password)
                .textContentType(.password)
                .textFieldStyle(.roundedBorder)

            // Error Message
            if let errorMessage = viewModel.errorMessage {
                Text(errorMessage)
                    .font(.caption)
                    .foregroundStyle(.red)
                    .multilineTextAlignment(.center)
            }

            // Login Button
            Button {
                Task {
                    await viewModel.login()
                }
            } label: {
                if viewModel.isLoading {
                    ProgressView()
                        .progressViewStyle(.circular)
                        .tint(.white)
                } else {
                    Text("ログイン")
                }
            }
            .frame(maxWidth: .infinity)
            .padding()
            .background(viewModel.isValid ? Theme.primary : Color.gray)
            .foregroundStyle(.white)
            .clipShape(RoundedRectangle(cornerRadius: 10))
            .disabled(!viewModel.isValid || viewModel.isLoading)
        }
        .padding(.horizontal)
    }
}

#Preview {
    LoginView()
        .environment(AuthManager())
}
