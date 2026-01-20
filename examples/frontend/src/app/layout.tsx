// ルートレイアウト
// 全ページで共通のHTML構造を定義する

import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";

// Inter フォントの設定
const inter = Inter({ subsets: ["latin"] });

// ページのメタデータ
export const metadata: Metadata = {
  title: "TODO App",
  description: "シンプルで使いやすいTODO管理アプリケーション",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ja">
      <body className={inter.className}>{children}</body>
    </html>
  );
}
