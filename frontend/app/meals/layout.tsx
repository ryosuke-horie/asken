import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: '食事記録',
}

export default function MealsLayout({ children }: { children: React.ReactNode }) {
  return <>{children}</>
}
