import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { fetcher, createAuthFetcher } from './fetcher'

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
        text: () => Promise.resolve(''),
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
        text: () => Promise.resolve(''),
      })
    )

    await expect(fetcher('http://example.com/api')).rejects.toThrow('API error: 500')
  })

  it('サーバーエラーメッセージがある場合はそれをスローすべき', async () => {
    const serverErrorMessage = 'データベースエラーが発生しました'
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        text: () => Promise.resolve(serverErrorMessage),
      })
    )

    await expect(fetcher('http://example.com/api')).rejects.toThrow(serverErrorMessage)
  })

  it('ネットワークエラー時にエラーを伝播すべき', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Network error')))

    await expect(fetcher('http://example.com/api')).rejects.toThrow('Network error')
  })
})

describe('createAuthFetcher', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('トークンがある場合はAuthorizationヘッダーを設定すべき', async () => {
    const mockData = { id: 1 }
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockData),
    })
    vi.stubGlobal('fetch', mockFetch)

    const authFetcher = createAuthFetcher('test-token')
    await authFetcher('http://example.com/api')

    expect(mockFetch).toHaveBeenCalledWith('http://example.com/api', {
      headers: { Authorization: 'Bearer test-token' },
    })
  })

  it('トークンがnullの場合はAuthorizationヘッダーを設定しないべき', async () => {
    const mockData = { id: 1 }
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockData),
    })
    vi.stubGlobal('fetch', mockFetch)

    const authFetcher = createAuthFetcher(null)
    await authFetcher('http://example.com/api')

    expect(mockFetch).toHaveBeenCalledWith('http://example.com/api', {
      headers: {},
    })
  })

  it('正常なレスポンスをJSONとして返すべき', async () => {
    const mockData = { user: 'authenticated' }
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockData),
      })
    )

    const authFetcher = createAuthFetcher('test-token')
    const result = await authFetcher('http://example.com/api')

    expect(result).toEqual(mockData)
  })

  it('認証エラー(401)時にエラーをスローすべき', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        text: () => Promise.resolve('Unauthorized'),
      })
    )

    const authFetcher = createAuthFetcher('invalid-token')
    await expect(authFetcher('http://example.com/api')).rejects.toThrow('Unauthorized')
  })

  it('サーバーエラーメッセージがある場合はそれをスローすべき', async () => {
    const serverErrorMessage = '認証トークンが無効です'
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        status: 403,
        text: () => Promise.resolve(serverErrorMessage),
      })
    )

    const authFetcher = createAuthFetcher('test-token')
    await expect(authFetcher('http://example.com/api')).rejects.toThrow(serverErrorMessage)
  })

  it('エラーメッセージが空の場合はステータスコードを含むエラーをスローすべき', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        text: () => Promise.resolve(''),
      })
    )

    const authFetcher = createAuthFetcher('test-token')
    await expect(authFetcher('http://example.com/api')).rejects.toThrow('API error: 500')
  })
})
