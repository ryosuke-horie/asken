import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { fetcher } from './fetcher'

describe('fetcher', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('正常なレスポンスをJSONとして返すべき', async () => {
    const mockData = { id: 1, name: 'test' }
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockData),
      })
    )

    const result = await fetcher('http://example.com/api')

    expect(result).toEqual(mockData)
  })

  it('レスポンスがokでない場合はエラーをスローすべき', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        status: 404,
      })
    )

    await expect(fetcher('http://example.com/api')).rejects.toThrow('API error: 404')
  })

  it('500エラー時にステータスコードを含むエラーをスローすべき', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
      })
    )

    await expect(fetcher('http://example.com/api')).rejects.toThrow('API error: 500')
  })

  it('ネットワークエラー時にエラーを伝播すべき', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Network error')))

    await expect(fetcher('http://example.com/api')).rejects.toThrow('Network error')
  })
})
