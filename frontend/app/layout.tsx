import './globals.css'

export const metadata = {
  title: 'asken - カロリー計算アプリ',
  description: '画像から食事内容を判定し、カロリーと栄養素を計算するアプリ',
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
          <h1>asken - カロリー計算アプリ</h1>
        </header>
        <main>{children}</main>
        <footer>
          <p>&copy; 2024 asken. Powered by Gemini API.</p>
        </footer>
      </body>
    </html>
  )
}
