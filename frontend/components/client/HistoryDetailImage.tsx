'use client';

import { useState } from 'react';
import styles from './HistoryDetailImage.module.css';

interface HistoryDetailImageProps {
  imagePath: string;
  alt: string;
}

export default function HistoryDetailImage({ imagePath, alt }: HistoryDetailImageProps) {
  const [imageError, setImageError] = useState(false);

  const imageUrl = imageError
    ? '/placeholder.png'
    : `http://localhost:8080/api/images/${imagePath.split('/').pop()}`;

  return (
    <img
      src={imageUrl}
      alt={alt}
      className={styles.image}
      onError={() => setImageError(true)}
    />
  );
}
