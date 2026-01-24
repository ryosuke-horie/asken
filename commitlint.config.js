module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'type-enum': [
      2,
      'always',
      [
        'feat',     // 新機能
        'fix',      // バグ修正
        'docs',     // ドキュメント
        'style',    // コードスタイル（フォーマット）
        'refactor', // リファクタリング
        'perf',     // パフォーマンス改善
        'test',     // テスト
        'chore',    // ビルド、設定等
        'ci',       // CI/CD関連
        'revert',   // リバート
      ],
    ],
    'subject-case': [0], // 日本語対応のため無効化
    'header-max-length': [2, 'always', 100],
  },
};
