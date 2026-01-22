import { HistoryItem } from '@/types/nutrition';
import HistoryListItem from '@/components/client/HistoryListItem';
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
        <HistoryListItem key={item.id} item={item} />
      ))}
    </div>
  );
}
