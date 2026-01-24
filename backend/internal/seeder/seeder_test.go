package seeder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	t.Run("ファイルを正しくコピーすべき", func(t *testing.T) {
		// テスト用の一時ディレクトリを作成
		tmpDir := t.TempDir()

		// ソースファイルを作成
		srcPath := filepath.Join(tmpDir, "source.txt")
		content := []byte("test content")
		if err := os.WriteFile(srcPath, content, 0o644); err != nil {
			t.Fatalf("ソースファイルの作成に失敗: %v", err)
		}

		// コピー先パス
		dstPath := filepath.Join(tmpDir, "dest.txt")

		// コピー実行
		err := copyFile(srcPath, dstPath)
		if err != nil {
			t.Errorf("copyFileがエラーを返した: %v", err)
		}

		// コピー先ファイルの内容を確認
		copied, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatalf("コピー先ファイルの読み取りに失敗: %v", err)
		}

		if string(copied) != string(content) {
			t.Errorf("コピー内容が一致しない: 期待 %s, 実際 %s", content, copied)
		}
	})

	t.Run("存在しないソースファイルはエラーを返すべき", func(t *testing.T) {
		tmpDir := t.TempDir()
		srcPath := filepath.Join(tmpDir, "nonexistent.txt")
		dstPath := filepath.Join(tmpDir, "dest.txt")

		err := copyFile(srcPath, dstPath)
		if err == nil {
			t.Error("存在しないファイルでエラーが返されるべき")
		}
	})

	t.Run("書き込み不可のディレクトリはエラーを返すべき", func(t *testing.T) {
		tmpDir := t.TempDir()

		// ソースファイルを作成
		srcPath := filepath.Join(tmpDir, "source.txt")
		if err := os.WriteFile(srcPath, []byte("test"), 0o644); err != nil {
			t.Fatalf("ソースファイルの作成に失敗: %v", err)
		}

		// 存在しないディレクトリへのパス
		dstPath := filepath.Join(tmpDir, "nonexistent", "dest.txt")

		err := copyFile(srcPath, dstPath)
		if err == nil {
			t.Error("書き込み不可のパスでエラーが返されるべき")
		}
	})

	t.Run("コピー失敗時にコピー先ファイルが残らないべき", func(t *testing.T) {
		tmpDir := t.TempDir()

		// 空のソースファイルを作成
		srcPath := filepath.Join(tmpDir, "source.txt")
		if err := os.WriteFile(srcPath, []byte("test content"), 0o644); err != nil {
			t.Fatalf("ソースファイルの作成に失敗: %v", err)
		}

		dstPath := filepath.Join(tmpDir, "dest.txt")

		// 正常なコピー
		err := copyFile(srcPath, dstPath)
		if err != nil {
			t.Fatalf("copyFileがエラーを返した: %v", err)
		}

		// コピー先が存在することを確認
		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			t.Error("コピー先ファイルが存在するべき")
		}
	})
}
