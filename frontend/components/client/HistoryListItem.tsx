'use client';

import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { HistoryItem } from '@/types/nutrition';
import HistoryItemImage from '@/components/client/HistoryItemImage';
import DeleteHistoryButton from '@/components/client/DeleteHistoryButton';
import styles from './HistoryListItem.module.css';

interface HistoryListItemProps {
  item: HistoryItem;
}

export default function HistoryListItem({ item }: HistoryListItemProps) {
  const router = useRouter();

  const handleDeleteSuccess = () => {
    router.refresh();
  };

  return (
    <div className={styles.wrapper}>
      <Link href={`/history/${item.id}`} className={styles.item}>
        <div className={styles.imageContainer}>
          <HistoryItemImage imagePath={item.image_path} alt="食事画像" />
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
      <div className={styles.actions}>
        <DeleteHistoryButton
          historyId={item.id}
          iconOnly
          onSuccess={handleDeleteSuccess}
        />
      </div>
    </div>
  );
}
