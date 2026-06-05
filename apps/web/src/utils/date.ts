export function thaiLocale() {
  return 'th-TH-u-ca-buddhist'
}

export function formatThaiDate(value: string | number | Date | null | undefined) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return typeof value === 'string' ? value : '-'
  return new Intl.DateTimeFormat(thaiLocale(), {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  }).format(date)
}

export function formatThaiDateTime(value: string | number | Date | null | undefined) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return typeof value === 'string' ? value : '-'
  return new Intl.DateTimeFormat(thaiLocale(), {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

function twoDigit(value: number) {
  return String(value).padStart(2, '0')
}

export function formatThaiNumericDate(value: string | number | Date | null | undefined) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return typeof value === 'string' ? value : '-'
  return `${twoDigit(date.getDate())}/${twoDigit(date.getMonth() + 1)}/${date.getFullYear() + 543}`
}

export function formatThaiNumericDateTime(value: string | number | Date | null | undefined) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return typeof value === 'string' ? value : '-'
  return `${formatThaiNumericDate(date)} ${twoDigit(date.getHours())}:${twoDigit(date.getMinutes())}:${twoDigit(date.getSeconds())}`
}
