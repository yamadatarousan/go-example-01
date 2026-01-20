import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'

// Inter フォントの設定
// Google Fontsから自動的に読み込まれ、最適化される
const inter = Inter({ subsets: ['latin'] })

// ページのメタデータ（SEO用）
export const metadata: Metadata = {
  title: 'TODO App',
  description: 'シンプルで使いやすいTODO管理アプリケーション',
}

// ルートレイアウト
// 全ページで共通のHTML構造を定義する
// Server Componentとして実行される（"use client"がないため）
export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="ja">
      <body className={inter.className}>
        {/* children には各ページのコンテンツが入る */}
        {children}
      </body>
    </html>
  )
}
