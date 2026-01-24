'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useAuth } from '@/contexts/AuthContext'
import styles from './Navigation.module.css'

export default function Navigation() {
  const pathname = usePathname()
  const { isAuthenticated, logout } = useAuth()

  if (!isAuthenticated) {
    return null
  }

  const navItems = [
    { href: '/', label: 'ホーム' },
    { href: '/mylist', label: 'マイリスト' },
  ]

  return (
    <nav className={styles.nav}>
      <ul className={styles.list}>
        {navItems.map((item) => (
          <li key={item.href}>
            <Link
              href={item.href}
              className={`${styles.link} ${pathname === item.href ? styles.active : ''}`}
            >
              {item.label}
            </Link>
          </li>
        ))}
      </ul>
      <button type="button" onClick={logout} className={styles.logoutButton}>
        ログアウト
      </button>
    </nav>
  )
}
