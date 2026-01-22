// ログインフォームコンポーネント
// React Hook Form と Zod を組み合わせた型安全なフォーム実装

"use client"; // クライアントコンポーネント（useStateやイベントハンドラを使うため）

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

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
    // TODO: 次のタスクでServer Actionを呼び出す
    console.log("ログイン試行:", data);
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
