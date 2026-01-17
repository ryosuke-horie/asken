/**
 * localStorage管理ユーティリティ
 * 分析IDの保存・取得・削除を型安全に実装
 */

const ANALYSIS_ID_KEY = 'asken_analysis_id';

/**
 * 分析IDをlocalStorageに保存
 * @param analysisId - 保存する分析ID
 */
export function saveAnalysisId(analysisId: string): void {
  try {
    localStorage.setItem(ANALYSIS_ID_KEY, analysisId);
  } catch (error) {
    console.error('Failed to save analysis ID to localStorage:', error);
  }
}

/**
 * localStorageから分析IDを取得
 * @returns 保存されている分析ID、存在しない場合はnull
 */
export function getAnalysisId(): string | null {
  try {
    return localStorage.getItem(ANALYSIS_ID_KEY);
  } catch (error) {
    console.error('Failed to get analysis ID from localStorage:', error);
    return null;
  }
}

/**
 * localStorageから分析IDを削除
 */
export function clearAnalysisId(): void {
  try {
    localStorage.removeItem(ANALYSIS_ID_KEY);
  } catch (error) {
    console.error('Failed to clear analysis ID from localStorage:', error);
  }
}
