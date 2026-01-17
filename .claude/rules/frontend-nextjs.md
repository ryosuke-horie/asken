---
paths:
  - "frontend/**/*.{ts,tsx,js,jsx}"
  - "frontend/**/*.json"
---

# フロントエンド開発規約（Next.js / TypeScript）

## 命名規則

- **コンポーネント**: PascalCase（例: `ImageUpload.tsx`）
- **関数・変数**: camelCase（例: `analyzeImage`, `foodData`）
- **定数**: UPPER_SNAKE_CASE（例: `API_BASE_URL`）
- **型・インターフェース**: PascalCase、接頭辞`I`は使用しない（例: `Food`, `NutritionInfo`）

## コンポーネント設計

- **関数コンポーネント**を使用し、クラスコンポーネントは使用しない
- **Server Components**をデフォルトとし、クライアントサイドの状態管理が必要な場合のみ`'use client'`を使用
- コンポーネントは**単一責任の原則**に従い、1つの責務のみを持つ
- 100行を超えるコンポーネントは分割を検討する

```typescript
// ✅ 良い例
export default async function FoodList() {
  // Server Component（デフォルト）
  const foods = await fetchFoods();
  return <FoodTable foods={foods} />;
}

// ❌ 悪い例 - クライアントコンポーネントを不必要に使用
'use client'
export default function FoodList() {
  // Server Componentで十分な場合
}
```

## TypeScript使用ルール

- **`any`型の使用は禁止** - `unknown`または適切な型を定義する
- **暗黙的な型推論**を活用し、不要な型注釈は省略する
- すべての関数は**戻り値の型を明示**する（短い関数を除く）
- **null合体演算子**（`??`）と**オプショナルチェイニング**（`?.`）を積極的に使用

```typescript
// ✅ 良い例
function calculateCalories(food: Food): number {
  return food.calories ?? 0;
}

const foodName = food?.name ?? '不明';

// ❌ 悪い例
function calculateCalories(food: any) {  // anyは禁止
  return food.calories;
}
```

## ファイル構成

- 1ファイル1エクスポート（コンポーネント、フック、ユーティリティ）
- 関連する型定義は同じファイルに配置
- 共有される型は`src/types/`に配置
- Server ComponentsとClient Componentsは明確に分離（`components/server/`と`components/client/`）

## インポート順序とベストプラクティス

```typescript
// 1. React関連の外部ライブラリ
import { useState } from 'react';
import { useRouter } from 'next/navigation';

// 2. 外部ライブラリ（直接インポートを使用 - Barrel Importを避ける）
// ❌ 悪い例: import { Check, X } from 'lucide-react'
// ✅ 良い例:
import Check from 'lucide-react/dist/esm/icons/check';
import X from 'lucide-react/dist/esm/icons/x';

// 3. 内部ライブラリ（絶対パス）
import { Button } from '@/components/client/Button';
import { useFood } from '@/hooks/useFood';

// 4. 型定義
import type { Food, NutritionInfo } from '@/types';

// 5. 相対パス
import { NutritionCard } from './NutritionCard';
```

### Barrel Importの回避

アイコンライブラリやUIライブラリからインポートする際は、直接インポートを使用してバンドルサイズを削減する：

```typescript
// ❌ 悪い例 - barrel importは全モジュールをロード（1,000+ modules）
import { Button, TextField } from '@mui/material';

// ✅ 良い例 - 必要なモジュールのみをロード
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
```

### Next.js 13.5+の場合（推奨）

`next.config.js`で`optimizePackageImports`を設定すれば、barrel importを使用可能：

```javascript
// next.config.js
module.exports = {
  experimental: {
    optimizePackageImports: ['lucide-react', '@mui/material']
  }
}
```

## パフォーマンス最適化（Vercel推奨）

### 1. バンドルサイズの最適化（CRITICAL）

**Dynamic Importsで重いコンポーネントを遅延ロード**

```typescript
// ❌ 悪い例 - 重いコンポーネントがメインバンドルに含まれる
import { MonacoEditor } from './monaco-editor';

function CodePanel({ code }: { code: string }) {
  return <MonacoEditor value={code} />;
}

// ✅ 良い例 - 必要になったときにロード
import dynamic from 'next/dynamic';

const MonacoEditor = dynamic(
  () => import('./monaco-editor').then(m => m.MonacoEditor),
  { ssr: false }
);

function CodePanel({ code }: { code: string }) {
  return <MonacoEditor value={code} />;
}
```

**非クリティカルなサードパーティライブラリの遅延ロード**

```typescript
// ❌ 悪い例 - アナリティクスがメインバンドルをブロック
import { Analytics } from '@vercel/analytics/react';

export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        {children}
        <Analytics />
      </body>
    </html>
  );
}

// ✅ 良い例 - ハイドレーション後にロード
import dynamic from 'next/dynamic';

const Analytics = dynamic(
  () => import('@vercel/analytics/react').then(m => m.Analytics),
  { ssr: false }
);

export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        {children}
        <Analytics />
      </body>
    </html>
  );
}
```

### 2. Server Componentsのパフォーマンス（HIGH）

**React.cache()によるリクエスト内重複排除**

```typescript
// lib/cache.ts
import { cache } from 'react';

// ✅ 認証情報の取得を重複排除
export const getCurrentUser = cache(async () => {
  const session = await auth();
  if (!session?.user?.id) return null;
  return await db.user.findUnique({
    where: { id: session.user.id }
  });
});

// ✅ 食品データの取得を重複排除
export const getFoodById = cache(async (id: string) => {
  return await db.food.findUnique({ where: { id } });
});
```

**注意**: `React.cache()`は**プリミティブ型の引数**を使用すること（オブジェクトは参照比較のため、キャッシュミスが発生する）：

```typescript
// ❌ 悪い例 - オブジェクトは毎回新しい参照
const getUser = cache(async (params: { uid: number }) => {
  return await db.user.findUnique({ where: { id: params.uid } });
});
getUser({ uid: 1 }); // キャッシュミス

// ✅ 良い例 - プリミティブ型を使用
const getUser = cache(async (uid: number) => {
  return await db.user.findUnique({ where: { id: uid } });
});
getUser(1); // キャッシュヒット
```

**並列データフェッチング（コンポーネント構成による）**

```typescript
// ❌ 悪い例 - 逐次的にデータを取得（遅い）
export default async function Page() {
  const header = await fetchHeader(); // 待機
  const sidebar = await fetchSidebar(); // 待機
  return (
    <div>
      <div>{header}</div>
      <div>{sidebar}</div>
    </div>
  );
}

// ✅ 良い例 - 並列にデータを取得（速い）
async function Header() {
  const data = await fetchHeader();
  return <div>{data}</div>;
}

async function Sidebar() {
  const data = await fetchSidebar();
  return <div>{data}</div>;
}

export default function Page() {
  // HeaderとSidebarのfetchは並列実行される
  return (
    <div>
      <Header />
      <Sidebar />
    </div>
  );
}
```

**Promise.all()で独立した操作を並列化**

```typescript
// ❌ 悪い例 - 逐次実行（3回のラウンドトリップ）
const user = await fetchUser();
const posts = await fetchPosts();
const comments = await fetchComments();

// ✅ 良い例 - 並列実行（1回のラウンドトリップ）
const [user, posts, comments] = await Promise.all([
  fetchUser(),
  fetchPosts(),
  fetchComments()
]);
```

**Suspense境界の戦略的配置**

```typescript
// ❌ 悪い例 - レイアウト全体がデータ取得を待つ
async function Page() {
  const data = await fetchData(); // ページ全体をブロック
  return (
    <div>
      <div>Header</div>
      <div><DataDisplay data={data} /></div>
      <div>Footer</div>
    </div>
  );
}

// ✅ 良い例 - レイアウトは即座に表示、データはストリーミング
function Page() {
  return (
    <div>
      <div>Header</div>
      <Suspense fallback={<Skeleton />}>
        <DataDisplay />
      </Suspense>
      <div>Footer</div>
    </div>
  );
}

async function DataDisplay() {
  const data = await fetchData(); // このコンポーネントのみブロック
  return <div>{data.content}</div>;
}
```

### 3. クライアントサイドのデータフェッチング（MEDIUM-HIGH）

**SWRによる自動重複排除**

```typescript
// ❌ 悪い例 - 各インスタンスが個別にfetch
'use client';

function UserList() {
  const [users, setUsers] = useState([]);
  useEffect(() => {
    fetch('/api/users')
      .then(r => r.json())
      .then(setUsers);
  }, []);
  return <div>{users.map(renderUser)}</div>;
}

// ✅ 良い例 - SWRが自動的に重複排除・キャッシュ
'use client';
import useSWR from 'swr';

const fetcher = (url: string) => fetch(url).then(r => r.json());

function UserList() {
  const { data: users } = useSWR('/api/users', fetcher);
  if (!users) return <div>Loading...</div>;
  return <div>{users.map(renderUser)}</div>;
}
```

### 4. 再レンダリングの最適化（MEDIUM）

**memoによる高コストコンポーネントの最適化**

```typescript
// ❌ 悪い例 - loadingでも計算が実行される
function Profile({ user, loading }: Props) {
  const avatar = useMemo(() => {
    const id = computeAvatarId(user); // loadingでも実行
    return <Avatar id={id} />;
  }, [user]);

  if (loading) return <Skeleton />;
  return <div>{avatar}</div>;
}

// ✅ 良い例 - loadingの場合は計算をスキップ
const UserAvatar = memo(function UserAvatar({ user }: { user: User }) {
  const id = useMemo(() => computeAvatarId(user), [user]);
  return <Avatar id={id} />;
});

function Profile({ user, loading }: Props) {
  if (loading) return <Skeleton />;
  return (
    <div>
      <UserAvatar user={user} />
    </div>
  );
}
```

**静的JSXのホイスト**

```typescript
// ❌ 悪い例 - 毎レンダリングでJSXを再生成
function Container() {
  return (
    <div>
      {loading && <div className="animate-pulse h-20 bg-gray-200" />}
    </div>
  );
}

// ✅ 良い例 - JSXを再利用
const loadingSkeleton = (
  <div className="animate-pulse h-20 bg-gray-200" />
);

function Container() {
  return (
    <div>
      {loading && loadingSkeleton}
    </div>
  );
}
```

### 5. その他のパフォーマンス最適化

- **画像**: `next/image`を使用（自動最適化、遅延ロード）
- **フォント**: `next/font`を使用（フォント最適化）
