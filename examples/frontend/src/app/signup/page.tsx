// サインアップページ
// Server Component として実行される

import Link from 'next/link'
import { SignupForm } from './signup-form'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

export default function SignupPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl font-bold text-center">
            アカウント作成
          </CardTitle>
          <CardDescription className="text-center">
            新しいアカウントを作成します
          </CardDescription>
        </CardHeader>
        <CardContent>
          {/* クライアントコンポーネントのフォーム */}
          <SignupForm />

          {/* ログインへのリンク */}
          <div className="mt-4 text-center text-sm">
            すでにアカウントをお持ちの方は{' '}
            <Link
              href="/login"
              className="text-primary underline underline-offset-4 hover:text-primary/80"
            >
              ログイン
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
