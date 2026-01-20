// ヘッダーコンポーネント
// Client Component: ログアウト処理を行う

'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { logoutAction } from './actions'

export function Header() {
  const router = useRouter()
  const [isLoading, setIsLoading] = useState(false)

  // ログアウト処理
  const handleLogout = async () => {
    setIsLoading(true)
    try {
      await logoutAction()
      router.push('/login')
      router.refresh()
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <header className="border-b">
      <div className="container mx-auto px-4 py-4 flex justify-between items-center">
        <h2 className="text-xl font-semibold">TODO App</h2>
        <Button
          variant="outline"
          onClick={handleLogout}
          disabled={isLoading}
        >
          {isLoading ? 'ログアウト中...' : 'ログアウト'}
        </Button>
      </div>
    </header>
  )
}
