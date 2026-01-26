// ログインフォームコンポーネント
// React Hook Form と Zod を組み合わせた型安全なフォーム実装

"use client"; // クライアントコンポーネント（useStateやイベントハンドラを使うため）

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
// ----------------------------------------------------------------------------
// Server Action のインポート
// ----------------------------------------------------------------------------
// "use server" が付いたファイルから関数をインポート
// 見た目は普通の関数だが、呼び出すとサーバー側で実行される
import { login } from "./actions";

// ============================================================================
// Zod スキーマ定義
// ============================================================================
// Zodはバリデーションルールと型定義を同時に行えるライブラリ
// このスキーマから TypeScript の型も自動生成される

const loginSchema = z.object({
  // email: 文字列で、メール形式であること
  email: z
    .string()
    .min(1, "メールアドレスを入力してください")  // 空文字チェック
    .email("有効なメールアドレスを入力してください"), // メール形式チェック

  // password: 文字列で、8文字以上であること
  password: z
    .string()
    .min(1, "パスワードを入力してください")
    .min(8, "パスワードは8文字以上で入力してください"),
});

// ============================================================================
// 型定義
// ============================================================================
// z.infer<typeof スキーマ> でスキーマから TypeScript 型を自動生成
// これにより、スキーマと型定義を二重に書く必要がなくなる

type LoginFormData = z.infer<typeof loginSchema>;
// ↑ これは以下と同等:
// type LoginFormData = {
//   email: string;
//   password: string;
// }

// ============================================================================
// コンポーネント
// ============================================================================

export function LoginForm() {
  // --------------------------------------------------------------------------
  // ルーターとエラー状態
  // --------------------------------------------------------------------------
  // useRouter: Next.jsのクライアント側ナビゲーション用フック
  // App Routerでは "next/navigation" からインポート（Pages Routerは "next/router"）
  const router = useRouter();

  // サーバーから返されたエラーメッセージを保持
  // React Hook Formのerrorsはバリデーションエラー用、これはAPIエラー用
  const [serverError, setServerError] = useState<string | null>(null);

  // --------------------------------------------------------------------------
  // React Hook Form の初期化
  // --------------------------------------------------------------------------
  // useForm<T> にジェネリクスで型を渡すと、フォームデータが型安全になる
  // zodResolver(スキーマ) で Zod のバリデーションを React Hook Form に統合

  const {
    register,      // input要素にフォームを紐付ける関数
    handleSubmit,  // フォーム送信時のラッパー（バリデーション実行後にコールバックを呼ぶ）
    formState: {
      errors,        // 各フィールドのバリデーションエラー
      isSubmitting,  // 送信中かどうか（非同期送信時のローディング表示に使う）
    },
  } = useForm<LoginFormData>({
    // resolver: バリデーションロジックを外部ライブラリ（Zod）に委譲
    resolver: zodResolver(loginSchema),

    // defaultValues: フォームの初期値
    defaultValues: {
      email: "",
      password: "",
    },
  });

  // --------------------------------------------------------------------------
  // 送信処理
  // --------------------------------------------------------------------------
  // handleSubmit でラップされるため、この関数が呼ばれる時点で
  // バリデーションは通過済み → data は LoginFormData 型として安全に使える

  const onSubmit = async (data: LoginFormData) => {
    // 送信前にサーバーエラーをクリア
    setServerError(null);

    // ------------------------------------------------------------------------
    // Server Action の呼び出し
    // ------------------------------------------------------------------------
    // login(data) は見た目は普通の関数呼び出しだが、実際には:
    // 1. Next.jsが内部的にHTTP POSTリクエストを生成
    // 2. サーバー側で login() 関数が実行される
    // 3. 戻り値がシリアライズされてクライアントに返る
    //
    // awaitで待機している間、isSubmittingがtrueになる
    const result = await login(data);

    // ------------------------------------------------------------------------
    // 結果の処理
    // ------------------------------------------------------------------------
    // Server Actionは ActionResult<User> 型を返す
    // { success: true, data: User } または { success: false, error: string }
    if (result.success) {
      // ログイン成功 → TODO一覧ページへリダイレクト
      // router.push()はクライアント側でのナビゲーション（ページ全体のリロードなし）
      router.push("/todos");
    } else {
      // ログイン失敗 → エラーメッセージを表示
      setServerError(result.error);
    }
  };

  // --------------------------------------------------------------------------
  // レンダリング
  // --------------------------------------------------------------------------

  return (
    // handleSubmit(onSubmit) は:
    // 1. フォーム送信をpreventDefault
    // 2. Zodスキーマでバリデーション実行
    // 3. エラーがあれば errors に格納、なければ onSubmit を呼び出し
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      {/* メールアドレス入力欄 */}
      <div className="space-y-2">
        <Label htmlFor="email">メールアドレス</Label>
        <Input
          id="email"
          type="email"
          placeholder="example@example.com"
          // register("フィールド名") は以下を返す:
          // { name, onChange, onBlur, ref }
          // これをスプレッドすることで input を React Hook Form に登録
          {...register("email")}
        />
        {/* エラーメッセージの表示 */}
        {/* errors.email は email フィールドにエラーがある場合のみ存在 */}
        {errors.email && (
          <p className="text-sm text-red-500">{errors.email.message}</p>
        )}
      </div>

      {/* パスワード入力欄 */}
      <div className="space-y-2">
        <Label htmlFor="password">パスワード</Label>
        <Input
          id="password"
          type="password"
          placeholder="8文字以上"
          {...register("password")}
        />
        {errors.password && (
          <p className="text-sm text-red-500">{errors.password.message}</p>
        )}
      </div>

      {/* 送信ボタン */}
      {/* isSubmitting が true の間はボタンを無効化してローディング表示 */}
      <Button type="submit" className="w-full" disabled={isSubmitting}>
        {isSubmitting ? "ログイン中..." : "ログイン"}
      </Button>
    </form>
  );
}
