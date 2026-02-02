import GoogleSignIn
import SwiftUI
import UchikomiCore

// MARK: - UchikomiApp

@main
struct UchikomiApp: App {
    @UIApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate
    @State private var authManager = AuthManager(
        firebaseAuthService: FirebaseAuthService.shared,
        appleSignInManager: AppleSignInManager()
    )

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environment(authManager)
                .onOpenURL { url in
                    GIDSignIn.sharedInstance.handle(url)
                }
        }
    }
}

// MARK: - ContentView

struct ContentView: View {
    @Environment(AuthManager.self) private var authManager

    var body: some View {
        Group {
            if authManager.isAuthenticated {
                MainTabView()
            } else {
                LoginView()
            }
        }
    }
}

// MARK: - MainTabView

struct MainTabView: View {
    var body: some View {
        MealsView()
            .tint(Theme.primary)
    }
}
