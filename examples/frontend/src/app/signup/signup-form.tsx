// サインアップフォームコンポーネント
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
import { signup } from "./actions";

// ============================================================================
// Zod スキーマ定義
// ============================================================================

const signupSchema = z.object({
  email: z
    .string()
    .min(1, "メールアドレスを入力してください")
    .email("有効なメールアドレスを入力してください"),
  password: z
    .string()
    .min(1, "パスワードを入力してください")
    .min(8, "パスワードは8文字以上で入力してください"),
});

// ============================================================================
// 型定義
// ============================================================================

type SignupFormData = z.infer<typeof signupSchema>;

// ============================================================================
// コンポーネント
// ============================================================================

export function SignupForm() {
  const router = useRouter();
  const [serverError, setServerError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<SignupFormData>({
    resolver: zodResolver(signupSchema),
    defaultValues: {
      email: "",
      password: "",
    },
  });

  const onSubmit = async (data: SignupFormData) => {
    setServerError(null);

    const result = await signup(data);

    if (result.success) {
      // サインアップ成功 → TODO一覧ページへ
      router.push("/todos");
    } else {
      setServerError(result.error);
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="email">メールアドレス</Label>
        <Input
          id="email"
          type="email"
          placeholder="example@example.com"
          {...register("email")}
        />
        {errors.email && (
          <p className="text-sm text-red-500">{errors.email.message}</p>
        )}
      </div>

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

      <Button type="submit" className="w-full" disabled={isSubmitting}>
        {isSubmitting ? "登録中..." : "サインアップ"}
      </Button>

      {serverError && (
        <p className="text-sm text-red-500">{serverError}</p>
      )}
    </form>
  );
}
