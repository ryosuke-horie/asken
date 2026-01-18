import Link from 'next/link';
import { Suspense } from 'react';
import HistoryDetailView from '@/components/server/HistoryDetailView';
import DeleteHistoryButton from '@/components/client/DeleteHistoryButton';
import { HistoryDetail } from '@/types/nutrition';
import styles from './page.module.css';

async function fetchHistoryDetail(id: string): Promise<HistoryDetail> {
  const response = await fetch(`http://localhost:8080/api/history/${id}`, {
    cache: 'no-store',
  });

  if (!response.ok) {
    throw new Error('履歴詳細の取得に失敗しました');
  }

  return response.json();
}

export default async function HistoryDetailPage({
  params,
}: {
  params: { id: string };
}) {
  const detail = await fetchHistoryDetail(params.id);

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <Link href="/history" className={styles.backLink}>
          ← 履歴一覧に戻る
        </Link>
        <h1>履歴詳細</h1>
        <p className={styles.date}>
          {new Date(detail.created_at).toLocaleString('ja-JP', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
          })}
        </p>
      </div>

      <Suspense fallback={<div>読み込み中...</div>}>
        <HistoryDetailView detail={detail} />
      </Suspense>

      <div className={styles.actions}>
        <DeleteHistoryButton historyId={params.id} />
      </div>
    </div>
  );
}
