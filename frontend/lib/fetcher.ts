export const fetcher = async (url: string) => {
  const res = await fetch(url)
  if (!res.ok) {
    const errorText = await res.text()
    throw new Error(errorText || `API error: ${res.status}`)
  }
  return res.json()
}

export const createAuthFetcher = (token: string | null) => async (url: string) => {
  const headers: HeadersInit = {}
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(url, { headers })
  if (!res.ok) {
    const errorText = await res.text()
    throw new Error(errorText || `API error: ${res.status}`)
  }
  return res.json()
}
