'use client';

import { useRouter } from 'next/navigation';
import { useState } from 'react';
import styles from './DeleteHistoryButton.module.css';

interface DeleteHistoryButtonProps {
  historyId: string;
}

export default function DeleteHistoryButton({ historyId }: DeleteHistoryButtonProps) {
  const router = useRouter();
  const [isDeleting, setIsDeleting] = useState(false);

  const handleDelete = async () => {
    if (!confirm('この履歴を削除してもよろしいですか?')) {
      return;
    }

    setIsDeleting(true);

    try {
      const response = await fetch(`http://localhost:8080/api/history/${historyId}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        throw new Error('削除に失敗しました');
      }

      router.push('/history');
      router.refresh();
    } catch (error) {
      console.error('削除エラー:', error);
      alert('削除に失敗しました');
      setIsDeleting(false);
    }
  };

  return (
    <button
      onClick={handleDelete}
      disabled={isDeleting}
      className={styles.button}
    >
      {isDeleting ? '削除中...' : '削除'}
    </button>
  );
}
