package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPostgresDB_InvalidURL(t *testing.T) {
	// 無効な接続URLでエラーが返されることを確認
	config := Config{
		DatabaseURL: "invalid://url",
	}

	db, err := NewPostgresDB(config)
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestNewPostgresDB_ConnectionRefused(t *testing.T) {
	// 存在しないホストへの接続でエラーが返されることを確認
	config := Config{
		DatabaseURL: "postgres://user:password@localhost:9999/testdb?sslmode=disable",
	}

	db, err := NewPostgresDB(config)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "データベース接続の確認に失敗")
}

func TestNewPostgresDB_DefaultPoolSettings(t *testing.T) {
	// NOTE: このテストは実際のPostgreSQLインスタンスが必要です
	// CI/CDではスキップするか、testcontainersを使用してください
	t.Skip("実際のPostgreSQLインスタンスが必要なため、統合テスト環境でのみ実行")

	config := Config{
		DatabaseURL: "postgres://asken:asken@localhost:5432/asken?sslmode=disable",
	}

	db, err := NewPostgresDB(config)
	if err != nil {
		t.Skipf("PostgreSQLに接続できません: %v", err)
	}
	defer db.Close()

	assert.NoError(t, err)
	assert.NotNil(t, db)

	// 接続プールのデフォルト設定を確認
	stats := db.Stats()
	assert.Equal(t, 25, stats.MaxOpenConnections)
}

func TestNewPostgresDB_CustomPoolSettings(t *testing.T) {
	// NOTE: このテストは実際のPostgreSQLインスタンスが必要です
	t.Skip("実際のPostgreSQLインスタンスが必要なため、統合テスト環境でのみ実行")

	config := Config{
		DatabaseURL:     "postgres://asken:asken@localhost:5432/asken?sslmode=disable",
		MaxOpenConns:    10,
		MaxIdleConns:    3,
		ConnMaxLifetime: 10 * time.Minute,
	}

	db, err := NewPostgresDB(config)
	if err != nil {
		t.Skipf("PostgreSQLに接続できません: %v", err)
	}
	defer db.Close()

	assert.NoError(t, err)
	assert.NotNil(t, db)

	// カスタム接続プール設定を確認
	stats := db.Stats()
	assert.Equal(t, 10, stats.MaxOpenConnections)
}
