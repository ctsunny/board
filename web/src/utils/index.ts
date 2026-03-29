import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'

dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

export function formatDate(date: string | Date | null | undefined, fmt = 'YYYY-MM-DD HH:mm'): string {
  if (!date) return '-'
  return dayjs(date).format(fmt)
}

export function fromNow(date: string | Date | null | undefined): string {
  if (!date) return '-'
  return dayjs(date).fromNow()
}

export function isExpired(date: string | Date | null | undefined): boolean {
  if (!date) return false
  return dayjs(date).isBefore(dayjs())
}

export function isExpiringSoon(date: string | Date | null | undefined, days = 7): boolean {
  if (!date) return false
  const d = dayjs(date)
  return d.isAfter(dayjs()) && d.isBefore(dayjs().add(days, 'day'))
}

export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}

export function formatMoney(amount: number | string): string {
  const n = Number(amount)
  if (isNaN(n)) return '-'
  return `¥${n.toFixed(2)}`
}

export function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

export function getListData(data: unknown): Record<string, unknown>[] {
  if (Array.isArray(data)) return data as Record<string, unknown>[]
  if (data && typeof data === 'object') {
    const page = data as { data?: unknown[]; list?: unknown[]; items?: unknown[] }
    if (Array.isArray(page.data)) return page.data as Record<string, unknown>[]
    if (Array.isArray(page.list)) return page.list as Record<string, unknown>[]
    if (Array.isArray(page.items)) return page.items as Record<string, unknown>[]
  }
  return []
}
