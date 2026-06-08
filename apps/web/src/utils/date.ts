export function thaiLocale() {
  return 'th-TH-u-ca-buddhist'
}

export function appLocale(language: 'th' | 'en') {
  return language === 'th' ? thaiLocale() : 'en-US'
}

export function formatThaiDate(value: string | number | Date | null | undefined) {
  return formatAppDate(value, 'th')
}

export function formatAppDate(value: string | number | Date | null | undefined, language: 'th' | 'en') {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return typeof value === 'string' ? value : '-'
  return new Intl.DateTimeFormat(appLocale(language), {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  }).format(date)
}

export function formatThaiDateTime(value: string | number | Date | null | undefined) {
  return formatAppDateTime(value, 'th')
}

export function formatAppDateTime(value: string | number | Date | null | undefined, language: 'th' | 'en') {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return typeof value === 'string' ? value : '-'
  return new Intl.DateTimeFormat(appLocale(language), {
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
  return formatAppNumericDate(value, 'th')
}

export function formatAppNumericDate(value: string | number | Date | null | undefined, language: 'th' | 'en') {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return typeof value === 'string' ? value : '-'
  return `${twoDigit(date.getDate())}/${twoDigit(date.getMonth() + 1)}/${date.getFullYear() + (language === 'th' ? 543 : 0)}`
}

export function formatThaiNumericDateTime(value: string | number | Date | null | undefined) {
  return formatAppNumericDateTime(value, 'th')
}

export function formatAppNumericDateTime(value: string | number | Date | null | undefined, language: 'th' | 'en') {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return typeof value === 'string' ? value : '-'
  return `${formatAppNumericDate(date, language)} ${twoDigit(date.getHours())}:${twoDigit(date.getMinutes())}:${twoDigit(date.getSeconds())}`
}
