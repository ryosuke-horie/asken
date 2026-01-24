import { describe, it, expect } from 'vitest'
import { formatTargetDate } from './WeightSection'

describe('formatTargetDate', () => {
  describe('正常系', () => {
    it('日付文字列を「X月Y日」形式にフォーマットすべき', () => {
      expect(formatTargetDate('2026-02-23')).toBe('2月23日')
    })

    it('1月の日付を正しくフォーマットすべき', () => {
      expect(formatTargetDate('2026-01-01')).toBe('1月1日')
    })

    it('12月の日付を正しくフォーマットすべき', () => {
      expect(formatTargetDate('2026-12-31')).toBe('12月31日')
    })
  })

  describe('異常系', () => {
    it('無効な日付文字列の場合、「日付不明」を返すべき', () => {
      expect(formatTargetDate('invalid-date')).toBe('日付不明')
    })

    it('空文字列の場合、「日付不明」を返すべき', () => {
      expect(formatTargetDate('')).toBe('日付不明')
    })
  })
})
