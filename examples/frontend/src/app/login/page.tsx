// ログインページ
// ユーザーがメールアドレスとパスワードでログインするためのページ

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

export default function LoginPage() {
  return (
    <main className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>ログイン</CardTitle>
          <CardDescription>
            メールアドレスとパスワードを入力してください
          </CardDescription>
        </CardHeader>
        <CardContent>
          {/* ログインフォームは次のタスクで実装 */}
          <p className="text-muted-foreground">フォームは次のタスクで実装予定</p>
        </CardContent>
      </Card>
    </main>
  );
}
