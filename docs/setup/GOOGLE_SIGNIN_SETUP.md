# Google Sign-In セットアップ手順

## 概要

Firebase AuthenticationでGoogle Sign-Inを有効にするには、以下の手順が必要です：

1. Google Cloud ConsoleでWeb OAuthクライアントを作成（一度だけ手動）
2. Firebase AuthenticationでGoogle Sign-Inプロバイダーを有効化（CLI/APIで可能）
3. iOSアプリの設定を確認

## 開発環境でのテスト

iOSシミュレータではパスキー認証が動作しないため、開発用のモック認証を使用します。
詳細は[CONTRIB.md](../CONTRIB.md)の「開発用モック認証の設定」を参照してください。

## 制約事項

- 個人プロジェクト（GCP組織に属さないプロジェクト）では、OAuthクライアントのプログラム作成に制限があります
- IAP Brand APIは組織所属プロジェクトでのみ利用可能
- そのため、Web OAuthクライアントの作成は一度だけ手動で行う必要があります

## 手順1: Web OAuthクライアントの作成（手動・一度だけ）

### 1.1 Google Cloud Consoleにアクセス

```
https://console.cloud.google.com/apis/credentials?project=utikomi-dev
```

### 1.2 OAuth 2.0クライアントIDを作成

1. 「+ CREATE CREDENTIALS」をクリック
2. 「OAuth client ID」を選択
3. アプリケーションの種類: Web application を選択
4. 名前: `Firebase Auth Web Client` など任意の名前
5. 承認済みのリダイレクトURI:
   - `https://utikomi-dev.firebaseapp.com/__/auth/handler`
6. 「CREATE」をクリック

### 1.3 クライアントIDとシークレットを記録

作成後に表示される以下の情報を安全に保存：
- Client ID: `xxx.apps.googleusercontent.com`
- Client Secret: `GOCSPX-xxx`

## 手順2: Firebase AuthでGoogle Sign-Inを有効化

### 方法A: Terraform（推奨）

環境変数を設定してTerraformを実行：

```bash
cd infrastructure/environments/dev

# 環境変数を設定
export TF_VAR_google_oauth_client_id="your-web-client-id.apps.googleusercontent.com"
export TF_VAR_google_oauth_client_secret="your-web-client-secret"

# Terraformを実行
terraform plan
terraform apply
```

### 方法B: REST API（CLI）

Web OAuthクライアントのIDとシークレットを使用して、REST APIでGoogle Sign-Inを有効化：

```bash
# 環境変数を設定
export WEB_CLIENT_ID="your-web-client-id.apps.googleusercontent.com"
export WEB_CLIENT_SECRET="your-web-client-secret"

# Google Sign-Inプロバイダーを有効化
curl -X POST \
  -H "Authorization: Bearer $(gcloud auth print-access-token)" \
  -H "Content-Type: application/json" \
  -H "x-goog-user-project: utikomi-dev" \
  "https://identitytoolkit.googleapis.com/admin/v2/projects/utikomi-dev/defaultSupportedIdpConfigs?idpId=google.com" \
  -d "{
    \"name\": \"projects/utikomi-dev/defaultSupportedIdpConfigs/google.com\",
    \"enabled\": true,
    \"clientId\": \"$WEB_CLIENT_ID\",
    \"clientSecret\": \"$WEB_CLIENT_SECRET\"
  }"
```

## 手順3: iOSアプリの設定確認

### 3.1 GoogleService-Info.plist

以下のキーが正しく設定されていることを確認：

| キー | 説明 |
|:---|:---|
| `CLIENT_ID` | iOS OAuthクライアントID |
| `REVERSED_CLIENT_ID` | iOS OAuthクライアントIDの逆順（URLスキーム用） |
| `API_KEY` | Firebase API Key |
| `GOOGLE_APP_ID` | Firebase App ID |

### 3.2 project.yml（XcodeGen）

URL Schemeが正しく設定されていることを確認：

```yaml
info:
  properties:
    CFBundleURLTypes:
      - CFBundleURLSchemes:
          - com.googleusercontent.apps.YOUR-IOS-CLIENT-ID
        CFBundleURLName: "Google Sign-In"
    GIDClientID: "YOUR-IOS-CLIENT-ID.apps.googleusercontent.com"
```

### 3.3 Uchikomi.entitlements

Sign in with Apple用のエンタイトルメント：

```xml
<key>com.apple.developer.applesignin</key>
<array>
    <string>Default</string>
</array>
```

## トラブルシューティング

### エラー: "Firebaseの設定に問題があります"

原因: Firebase AuthenticationでGoogle Sign-Inプロバイダーが有効になっていない

解決: 手順2を実行してGoogle Sign-Inを有効化

### エラー: "client_secret cannot be empty"

原因: REST APIにWeb OAuthクライアントのシークレットが必要

解決: 手順1でWeb OAuthクライアントを作成し、そのシークレットを使用

### エラー: "Project must belong to an organization"

原因: 個人プロジェクトではIAP Brand APIが使用不可

解決: 手動でOAuthクライアントを作成（CLI/Terraformでは作成不可）

## 参考リンク

- [Firebase Authentication - Google Sign-In (iOS)](https://firebase.google.com/docs/auth/ios/google-signin)
- [Google Cloud - OAuth 2.0 credentials](https://console.cloud.google.com/apis/credentials)
- [Identity Platform - Enable Google provider](https://cloud.google.com/identity-platform/docs/how-to-enable-application-for-google)
