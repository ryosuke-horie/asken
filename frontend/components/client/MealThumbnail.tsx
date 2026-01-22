'use client'

import { useState } from 'react'
import styles from './MealThumbnail.module.css'

interface MealThumbnailProps {
  src: string
  alt?: string
  className?: string
}

export default function MealThumbnail({ src, alt = '食事', className }: MealThumbnailProps) {
  const [hasError, setHasError] = useState(false)

  if (hasError) {
    return (
      <div className={`${styles.placeholder} ${className || ''}`}>
        <span className={styles.placeholderIcon}>🍽️</span>
      </div>
    )
  }

  return (
    <img
      src={src}
      alt={alt}
      className={className}
      onError={() => setHasError(true)}
    />
  )
}
