import Link from 'next/link';
import { HistoryItem } from '@/types/nutrition';
import styles from './HistoryList.module.css';

interface HistoryListProps {
  items: HistoryItem[];
}

export default function HistoryList({ items }: HistoryListProps) {
  if (items.length === 0) {
    return (
      <div className={styles.empty}>
        <p>履歴がありません</p>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      {items.map((item) => (
        <Link href={`/history/${item.id}`} key={item.id} className={styles.item}>
          <div className={styles.imageContainer}>
            <img
              src={`http://localhost:8080/api/images/${item.image_path.split('/').pop()}`}
              alt="食事画像"
              className={styles.image}
              onError={(e) => {
                e.currentTarget.src = '/placeholder.png';
              }}
            />
          </div>
          <div className={styles.details}>
            <p className={styles.date}>
              {new Date(item.created_at).toLocaleString('ja-JP', {
                year: 'numeric',
                month: '2-digit',
                day: '2-digit',
                hour: '2-digit',
                minute: '2-digit',
              })}
            </p>
            <div className={styles.nutrients}>
              <span className={styles.calories}>
                {Math.round(item.total_calories)} kcal
              </span>
              <span className={styles.protein}>
                P: {item.total_protein.toFixed(1)}g
              </span>
              <span className={styles.fat}>
                F: {item.total_fat.toFixed(1)}g
              </span>
              <span className={styles.carbs}>
                C: {item.total_carbohydrates.toFixed(1)}g
              </span>
            </div>
          </div>
        </Link>
      ))}
    </div>
  );
}
