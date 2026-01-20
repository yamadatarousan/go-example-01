// 設定ページ
// Server Component として実装

import Link from 'next/link'
import { getProfileAction } from './actions'
import { ProfileForm } from './profile-form'
import { LogoutButton } from './logout-button'

export default async function SettingsPage() {
  const result = await getProfileAction()
  const profile = result.data

  return (
    <div className="min-h-screen bg-gray-50">
      {/* ヘッダー */}
      <header className="bg-white shadow-sm">
        <div className="max-w-4xl mx-auto px-4 py-4 flex justify-between items-center">
          <h1 className="text-2xl font-bold text-gray-900">設定</h1>
          <nav className="flex items-center space-x-4">
            <Link href="/dashboard" className="text-gray-600 hover:text-gray-900">
              ダッシュボード
            </Link>
            <Link href="/todos" className="text-gray-600 hover:text-gray-900">
              TODO一覧
            </Link>
            <Link href="/projects" className="text-gray-600 hover:text-gray-900">
              プロジェクト
            </Link>
          </nav>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8">
        {/* エラー表示 */}
        {result.error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
            プロフィールの取得に失敗しました: {result.error}
          </div>
        )}

        <div className="space-y-6">
          {/* プロフィールセクション */}
          <div className="bg-white rounded-lg shadow">
            <div className="p-6 border-b border-gray-200">
              <h2 className="text-lg font-semibold text-gray-900">
                プロフィール
              </h2>
              <p className="text-sm text-gray-500 mt-1">
                あなたのアカウント情報を管理します
              </p>
            </div>
            <div className="p-6">
              {profile ? (
                <ProfileForm profile={profile} />
              ) : (
                <p className="text-gray-500">プロフィールを読み込めませんでした</p>
              )}
            </div>
          </div>

          {/* アカウント情報セクション */}
          {profile && (
            <div className="bg-white rounded-lg shadow">
              <div className="p-6 border-b border-gray-200">
                <h2 className="text-lg font-semibold text-gray-900">
                  アカウント情報
                </h2>
              </div>
              <div className="p-6">
                <dl className="space-y-4">
                  <div>
                    <dt className="text-sm font-medium text-gray-500">メールアドレス</dt>
                    <dd className="mt-1 text-sm text-gray-900">{profile.email}</dd>
                  </div>
                  <div>
                    <dt className="text-sm font-medium text-gray-500">役割</dt>
                    <dd className="mt-1">
                      <span
                        className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                          profile.role === 'admin'
                            ? 'bg-purple-100 text-purple-800'
                            : 'bg-gray-100 text-gray-800'
                        }`}
                      >
                        {profile.role === 'admin' ? '管理者' : 'ユーザー'}
                      </span>
                    </dd>
                  </div>
                  <div>
                    <dt className="text-sm font-medium text-gray-500">登録日</dt>
                    <dd className="mt-1 text-sm text-gray-900">
                      {new Date(profile.created_at).toLocaleDateString('ja-JP', {
                        year: 'numeric',
                        month: 'long',
                        day: 'numeric',
                      })}
                    </dd>
                  </div>
                </dl>
              </div>
            </div>
          )}

          {/* ログアウトセクション */}
          <div className="bg-white rounded-lg shadow">
            <div className="p-6 border-b border-gray-200">
              <h2 className="text-lg font-semibold text-gray-900">
                セッション
              </h2>
            </div>
            <div className="p-6">
              <p className="text-sm text-gray-600 mb-4">
                ログアウトすると、再度ログインが必要になります
              </p>
              <LogoutButton />
            </div>
          </div>

          {/* 開発中の注意書き */}
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
            <div className="flex">
              <svg
                className="h-5 w-5 text-yellow-400 mr-2"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                />
              </svg>
              <div>
                <h3 className="text-sm font-medium text-yellow-800">開発中</h3>
                <p className="mt-1 text-sm text-yellow-700">
                  プロフィール更新機能は現在モックデータを使用しています。
                  バックエンドの実装完了後に有効になります。
                </p>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  )
}
