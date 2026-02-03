import AuthenticationServices
import GoogleSignInSwift
import SwiftUI
import UchikomiCore

// MARK: - LoginView

struct LoginView: View {
    @Environment(AuthManager.self) private var authManager
    @State private var viewModel: LoginViewModel?

    var body: some View {
        NavigationStack {
            VStack(spacing: 32) {
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

                // Sign-In Buttons
                if let viewModel {
                    SignInButtons(viewModel: viewModel)
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

// MARK: - SignInButtons

private struct SignInButtons: View {
    @Bindable var viewModel: LoginViewModel

    var body: some View {
        VStack(spacing: 16) {
            // Error Message
            if let errorMessage = viewModel.errorMessage {
                Text(errorMessage)
                    .font(.caption)
                    .foregroundStyle(.red)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal)
            }

            if viewModel.isLoading {
                ProgressView()
                    .progressViewStyle(.circular)
                    .padding()
            } else {
                #if DEBUG && targetEnvironment(simulator)
                // シミュレータ専用: 開発用ログインボタン
                Button {
                    Task {
                        await viewModel.signInWithMock()
                    }
                } label: {
                    HStack {
                        Image(systemName: "hammer.fill")
                        Text("開発用ログイン")
                    }
                    .frame(maxWidth: .infinity)
                    .frame(height: 50)
                }
                .buttonStyle(.borderedProminent)
                .tint(.orange)
                .padding(.horizontal)
                #else
                // Google Sign-In Button
                GoogleSignInButton(
                    viewModel: GoogleSignInButtonViewModel(
                        scheme: .dark,
                        style: .wide,
                        state: .normal
                    )
                ) {
                    Task {
                        await viewModel.signInWithGoogle()
                    }
                }
                .frame(height: 50)
                .padding(.horizontal)

                // Apple Sign-In Button (Personal Teamでは使用不可のため一時的に無効化)
                // TODO: Apple Developer Program登録後に有効化
                // SignInWithAppleButton(
                //     .signIn,
                //     onRequest: { request in
                //         request.requestedScopes = [.fullName, .email]
                //     },
                //     onCompletion: { result in
                //         switch result {
                //         case .success:
                //             Task {
                //                 await viewModel.signInWithApple()
                //             }
                //         case let .failure(error):
                //             if let authError = error as? ASAuthorizationError,
                //                authError.code == .canceled {
                //                 // ユーザーキャンセルは無視
                //                 return
                //             }
                //             viewModel.errorMessage = error.localizedDescription
                //         }
                //     }
                // )
                // .signInWithAppleButtonStyle(.black)
                // .frame(height: 50)
                // .padding(.horizontal)
                #endif
            }
        }
    }
}

#Preview {
    LoginView()
        .environment(
            AuthManager(
                firebaseAuthService: FirebaseAuthService.shared,
                appleSignInManager: AppleSignInManager()
            )
        )
}
