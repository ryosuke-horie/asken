import { describe, it, expect } from 'vitest'
import { formatDateShort, formatDateFull } from './date'

describe('formatDateShort', () => {
  it('有効な日付をM/D形式にフォーマットすべき', () => {
    expect(formatDateShort('2024-01-15')).toBe('1/15')
    expect(formatDateShort('2024-12-31')).toBe('12/31')
    expect(formatDateShort('2024-06-01')).toBe('6/1')
  })

  it('先頭のゼロを除去すべき', () => {
    expect(formatDateShort('2024-01-05')).toBe('1/5')
    expect(formatDateShort('2024-09-09')).toBe('9/9')
  })

  it('不正な形式の場合は「日付不明」を返すべき', () => {
    expect(formatDateShort('invalid')).toBe('日付不明')
    expect(formatDateShort('')).toBe('日付不明')
    expect(formatDateShort('2024/01/15')).toBe('日付不明')
    expect(formatDateShort('2024-01')).toBe('日付不明')
  })

  it('パースできない数値の場合は「日付不明」を返すべき', () => {
    expect(formatDateShort('2024-XX-15')).toBe('日付不明')
    expect(formatDateShort('2024-01-XX')).toBe('日付不明')
  })
})

describe('formatDateFull', () => {
  it('有効な日付をYYYY/M/D形式にフォーマットすべき', () => {
    expect(formatDateFull('2024-01-15')).toBe('2024/1/15')
    expect(formatDateFull('2024-12-31')).toBe('2024/12/31')
    expect(formatDateFull('2025-06-01')).toBe('2025/6/1')
  })

  it('先頭のゼロを除去すべき', () => {
    expect(formatDateFull('2024-01-05')).toBe('2024/1/5')
    expect(formatDateFull('2024-09-09')).toBe('2024/9/9')
  })

  it('不正な形式の場合は「日付不明」を返すべき', () => {
    expect(formatDateFull('invalid')).toBe('日付不明')
    expect(formatDateFull('')).toBe('日付不明')
    expect(formatDateFull('2024/01/15')).toBe('日付不明')
    expect(formatDateFull('2024-01')).toBe('日付不明')
  })

  it('パースできない数値の場合は「日付不明」を返すべき', () => {
    expect(formatDateFull('XXXX-01-15')).toBe('日付不明')
    expect(formatDateFull('2024-XX-15')).toBe('日付不明')
    expect(formatDateFull('2024-01-XX')).toBe('日付不明')
  })
})
