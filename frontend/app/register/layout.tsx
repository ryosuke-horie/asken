import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: '新規登録',
}

export default function RegisterLayout({ children }: { children: React.ReactNode }) {
  return <>{children}</>
}
