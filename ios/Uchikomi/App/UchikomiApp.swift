import SwiftUI

// MARK: - UchikomiApp

@main
struct UchikomiApp: App {
    @State private var authManager = AuthManager()

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environment(authManager)
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
