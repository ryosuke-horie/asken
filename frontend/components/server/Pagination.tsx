import Link from 'next/link'
import styles from './Pagination.module.css'

interface PaginationProps {
  currentPage: number
  totalItems: number
  itemsPerPage: number
}

export default function Pagination({ currentPage, totalItems, itemsPerPage }: PaginationProps) {
  const totalPages = Math.ceil(totalItems / itemsPerPage)

  if (totalPages <= 1) {
    return null
  }

  const pages: number[] = []
  for (let i = 1; i <= totalPages; i++) {
    pages.push(i)
  }

  return (
    <div className={styles.container}>
      {currentPage > 1 && (
        <Link href={`/history?page=${currentPage - 1}`} className={styles.button}>
          前へ
        </Link>
      )}

      <div className={styles.pages}>
        {pages.map((page) => (
          <Link
            key={page}
            href={`/history?page=${page}`}
            className={
              page === currentPage ? `${styles.pageButton} ${styles.active}` : styles.pageButton
            }
          >
            {page}
          </Link>
        ))}
      </div>

      {currentPage < totalPages && (
        <Link href={`/history?page=${currentPage + 1}`} className={styles.button}>
          次へ
        </Link>
      )}
    </div>
  )
}
