package util

// TruncateForLog はログ出力用にテキストをトランケートする。
// ユーザー入力がログに直接出力されることを防ぐ。
func TruncateForLog(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) > maxRunes {
		return string(runes[:maxRunes]) + "..."
	}
	return s
}
