import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { decodeJwt } from 'jose'
import { TOKEN_COOKIE_NAME, PUBLIC_ROUTES } from '@/lib/constants/auth'

function isTokenExpired(token: string): boolean {
  try {
    const payload = decodeJwt(token)
    if (!payload.exp) {
      return true
    }
    const now = Math.floor(Date.now() / 1000)
    return payload.exp < now
  } catch {
    return true
  }
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // 静的ファイル、API、_nextはスキップ
  if (
    pathname.startsWith('/_next') ||
    pathname.startsWith('/api') ||
    pathname.includes('.') // 静的ファイル
  ) {
    return NextResponse.next()
  }

  // 公開ルートはスキップ
  if (PUBLIC_ROUTES.includes(pathname)) {
    return NextResponse.next()
  }

  // Cookieからトークンを取得
  const token = request.cookies.get(TOKEN_COOKIE_NAME)?.value

  // トークンがない、または期限切れの場合はログインにリダイレクト
  if (!token || isTokenExpired(token)) {
    const loginUrl = new URL('/login', request.url)
    return NextResponse.redirect(loginUrl)
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico).*)'],
}
