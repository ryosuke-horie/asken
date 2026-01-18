import { Suspense } from 'react';
import HistoryList from '@/components/server/HistoryList';
import Pagination from '@/components/server/Pagination';
import { HistoryListResponse } from '@/types/nutrition';
import styles from './page.module.css';

async function fetchHistory(page: number): Promise<HistoryListResponse> {
  const response = await fetch(
    `http://localhost:8080/api/history?page=${page}&limit=20`,
    {
      cache: 'no-store',
    }
  );

  if (!response.ok) {
    throw new Error('履歴の取得に失敗しました');
  }

  return response.json();
}

export default async function HistoryPage({
  searchParams,
}: {
  searchParams: { page?: string };
}) {
  const page = parseInt(searchParams.page || '1', 10);
  const data = await fetchHistory(page);

  return (
    <div className={styles.container}>
      <h1>食事履歴</h1>

      <Suspense fallback={<div>読み込み中...</div>}>
        <HistoryList items={data.items} />
      </Suspense>

      <Pagination
        currentPage={page}
        totalItems={data.total}
        itemsPerPage={20}
      />
    </div>
  );
}
