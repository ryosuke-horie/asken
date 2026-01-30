import SwiftUI

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

struct MainTabView: View {
    var body: some View {
        TabView {
            MealsView()
                .tabItem {
                    Label("食事", systemImage: "fork.knife")
                }

            WeightView()
                .tabItem {
                    Label("体重", systemImage: "scalemass")
                }
        }
        .tint(Theme.primary)
    }
}
