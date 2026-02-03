# Privacy Manifest (PrivacyInfo.xcprivacy)

## 概要
- App Store Connectでは新規アプリや更新アプリに対してPrivacy manifestの提出が求められる。
- SDKのプライバシー署名も要求されるため、サードパーティSDKの更新状況に注意すること。

## レビュー観点
- `ios/**/PrivacyInfo.xcprivacy` の有無
- plistとして妥当な形式か
- 収集データ・追跡・第三者SDKの宣言と実装の整合

## 実務メモ
- plistとして読み取りできない場合は審査で不備になる可能性がある。
- 形式確認のため `plutil -lint` を使うのが安全。
