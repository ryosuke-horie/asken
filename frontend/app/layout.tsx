import Image from 'next/image'
import { AuthProvider } from '@/contexts/AuthContext'
import './globals.css'

import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: {
    default: 'ウチコミ',
    template: '%s | ウチコミ',
  },
  description: '格闘技の減量・体重コントロールを支援するアプリ',
  robots: {
    index: false,
    follow: false,
    googleBot: {
      index: false,
      follow: false,
    },
  },
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ja">
      <body>
        <AuthProvider>
          <header>
            <Image src="/logo-cropped.png" alt="ウチコミ" width={150} height={50} priority />
          </header>
          <main>{children}</main>
          <footer>
            <p>&copy; 2025 ウチコミ.</p>
          </footer>
        </AuthProvider>
      </body>
    </html>
  )
}
