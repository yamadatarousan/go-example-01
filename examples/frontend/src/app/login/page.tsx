// ログインページ
// Server Component として実行される

import Link from 'next/link'
import { LoginForm } from './login-form'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

export default function LoginPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl font-bold text-center">
            ログイン
          </CardTitle>
          <CardDescription className="text-center">
            メールアドレスとパスワードを入力してください
          </CardDescription>
        </CardHeader>
        <CardContent>
          {/* クライアントコンポーネントのフォーム */}
          <LoginForm />

          {/* サインアップへのリンク */}
          <div className="mt-4 text-center text-sm">
            アカウントをお持ちでない方は{' '}
            <Link
              href="/signup"
              className="text-primary underline underline-offset-4 hover:text-primary/80"
            >
              サインアップ
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
