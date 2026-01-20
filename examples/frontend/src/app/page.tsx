// ホームページ（ルート "/"）
// Server Component として実行される
// 認証状態に応じてリダイレクトを行う

import { redirect } from 'next/navigation'
import { cookies } from 'next/headers'

// Server Component: サーバーサイドで実行され、
// 認証状態をチェックしてリダイレクト処理を行う
export default async function HomePage() {
  // サーバーサイドでCookieを取得
  const cookieStore = await cookies()
  const token = cookieStore.get('token')

  // トークンがあれば TODO 一覧へ、なければログインページへ
  if (token) {
    redirect('/todos')
  } else {
    redirect('/login')
  }
}
