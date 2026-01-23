import Link from 'next/link';
import './globals.css'

export const metadata = {
  title: 'ウチコミ - 格闘技向け体重管理アプリ',
  description: '柔術/キックボクシングなど格闘技の減量・体重コントロールを支援するアプリ',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="ja">
      <body>
        <header>
          <h1>ウチコミ - 格闘技向け体重管理アプリ</h1>
          <nav>
            <Link href="/">ホーム</Link>
          </nav>
        </header>
        <main>{children}</main>
        <footer>
          <p>&copy; 2024 ウチコミ. Powered by Gemini API.</p>
        </footer>
      </body>
    </html>
  )
}
