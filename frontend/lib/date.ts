/**
 * 日付文字列を「M/D」形式にフォーマット（グラフ用）
 * @param dateStr YYYY-MM-DD形式の日付文字列
 */
export function formatDateShort(dateStr: string): string {
  const parts = dateStr.split('-')
  if (parts.length !== 3) {
    return '日付不明'
  }
  const month = parseInt(parts[1], 10)
  const day = parseInt(parts[2], 10)
  if (isNaN(month) || isNaN(day)) {
    return '日付不明'
  }
  return `${month}/${day}`
}

/**
 * 日付文字列を「YYYY/M/D」形式にフォーマット（表示用）
 * @param dateStr YYYY-MM-DD形式の日付文字列
 */
export function formatDateFull(dateStr: string): string {
  const parts = dateStr.split('-')
  if (parts.length !== 3) {
    return '日付不明'
  }
  const year = parseInt(parts[0], 10)
  const month = parseInt(parts[1], 10)
  const day = parseInt(parts[2], 10)
  if (isNaN(year) || isNaN(month) || isNaN(day)) {
    return '日付不明'
  }
  return `${year}/${month}/${day}`
}
