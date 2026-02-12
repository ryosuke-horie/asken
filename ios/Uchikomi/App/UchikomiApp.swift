import GoogleSignIn
import os
import SwiftUI
import UchikomiCore

private let logger = Logger(
    subsystem: Bundle.main.bundleIdentifier ?? "Uchikomi",
    category: "MainTabView"
)

// MARK: - UchikomiApp

@main
struct UchikomiApp: App {
    @UIApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate
    @State private var authManager: AuthManager

    init() {
        // AuthServiceProvider.shared を使用して、APIClient と同じサービスを共有
        _authManager = State(initialValue: AuthManager(
            firebaseAuthService: AuthServiceProvider.shared,
            appleSignInManager: AppleSignInManager()
        ))
    }

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
    @Environment(\.scenePhase) private var scenePhase

    var body: some View {
        TabView {
            MealsView()
                .tabItem { Label("食事", systemImage: "fork.knife") }
            WeightView()
                .tabItem { Label("体重", systemImage: "scalemass") }
            MyMenuListView()
                .tabItem { Label("マイメニュー", systemImage: "star") }
            SettingsView()
                .tabItem { Label("設定", systemImage: "gearshape") }
        }
        .tint(Theme.primary)
        .onChange(of: scenePhase) { _, newPhase in
            if newPhase == .active {
                Task {
                    await refreshTodayNotifications()
                }
            }
        }
    }

    private func refreshTodayNotifications() async {
        let store = NotificationSettingsStore()
        let settings = store.load()
        guard settings.isGlobalEnabled else { return }

        let repository = MealRepository()
        let notificationManager = NotificationManager()
        do {
            let dailyMeals = try await repository.getDailyMeals(date: Date())
            var recordedMealTypes: Set<MealType> = []

            for mealType in MealType.reminderTargets {
                let meals = dailyMeals.meals.meals(for: mealType)
                if !meals.isEmpty {
                    recordedMealTypes.insert(mealType)
                    await notificationManager.cancelDeliveredNotification(for: mealType)
                }
            }

            if !recordedMealTypes.isEmpty {
                await notificationManager.refreshMealNotifications(
                    settings: settings,
                    recordedMealTypes: recordedMealTypes
                )
            }
        } catch {
            logger.error("当日の食事記録取得に失敗（通知はそのまま維持）: \(error.localizedDescription)")
        }
    }
}
