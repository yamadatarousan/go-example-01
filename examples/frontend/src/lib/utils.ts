// ユーティリティ関数
// shadcn/ui で使用される cn 関数を定義

import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

/**
 * クラス名を結合するユーティリティ関数
 * clsx でクラス名を結合し、tailwind-merge で重複を解決する
 *
 * @example
 * cn('px-2 py-1', 'px-4') // => 'py-1 px-4' (px-2 は px-4 で上書き)
 * cn('text-red-500', condition && 'text-blue-500')
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
