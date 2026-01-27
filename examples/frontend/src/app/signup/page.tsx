// サインアップページ
// 新規ユーザー登録のためのページ

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { SignupForm } from "./signup-form";

export default function SignupPage() {
  return (
    <main className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>サインアップ</CardTitle>
          <CardDescription>
            メールアドレスとパスワードを入力してください
          </CardDescription>
        </CardHeader>
        <CardContent>
          <SignupForm />
        </CardContent>
      </Card>
    </main>
  );
}
