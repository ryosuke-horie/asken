'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useState } from 'react'
import { useAuth } from '@/contexts/AuthContext'
import styles from './Navigation.module.css'

export default function Navigation() {
  const pathname = usePathname()
  const { isAuthenticated, logout } = useAuth()
  const [isMenuOpen, setIsMenuOpen] = useState(false)

  if (!isAuthenticated) {
    return null
  }

  const navItems = [
    { href: '/', label: 'ホーム' },
    { href: '/mylist', label: 'マイリスト' },
    { href: '/training', label: 'トレーニング' },
    { href: '/settings', label: '設定' },
  ]

  const toggleMenu = () => setIsMenuOpen(!isMenuOpen)
  const closeMenu = () => setIsMenuOpen(false)

  return (
    <nav className={styles.nav}>
      <button
        type="button"
        className={styles.hamburger}
        onClick={toggleMenu}
        aria-label="メニューを開く"
        aria-expanded={isMenuOpen}
      >
        <span className={`${styles.hamburgerLine} ${isMenuOpen ? styles.open : ''}`} />
      </button>
      <ul className={`${styles.list} ${isMenuOpen ? styles.listOpen : ''}`}>
        {navItems.map((item) => (
          <li key={item.href}>
            <Link
              href={item.href}
              className={`${styles.link} ${pathname === item.href ? styles.active : ''}`}
              onClick={closeMenu}
            >
              {item.label}
            </Link>
          </li>
        ))}
        <li className={styles.logoutItem}>
          <button type="button" onClick={logout} className={styles.logoutButton}>
            ログアウト
          </button>
        </li>
      </ul>
    </nav>
  )
}
