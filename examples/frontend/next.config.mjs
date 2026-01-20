/** @type {import('next').NextConfig} */
const nextConfig = {
  // 実験的機能: Server Actionsを有効化（Next.js 14ではデフォルトで有効）
  experimental: {
    serverActions: {
      bodySizeLimit: '2mb',
    },
  },
  // 環境変数の設定
  env: {
    // バックエンドAPIのベースURL
    API_BASE_URL: process.env.API_BASE_URL || 'http://localhost:8080',
  },
}

export default nextConfig
