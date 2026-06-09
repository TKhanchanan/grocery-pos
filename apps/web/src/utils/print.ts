const RECEIPT_PRINT_SCALE = '--receipt-print-scale'
const A4_PRINTABLE_WIDTH_MM = 194
const A4_PRINTABLE_HEIGHT_MM = 281
const RECEIPT_WIDTH_MM = 80
const RECEIPT_PADDING_MM = 4
const PX_PER_MM = 96 / 25.4

export function prepareReceiptPrintArea(elementID = 'receipt-print-area') {
  const receipt = document.getElementById(elementID)
  if (!receipt) return

  receipt.style.removeProperty(RECEIPT_PRINT_SCALE)

  const clone = receipt.cloneNode(true) as HTMLElement
  clone.id = `${elementID}-measure`
  clone.style.position = 'fixed'
  clone.style.left = '-10000px'
  clone.style.top = '0'
  clone.style.width = `${RECEIPT_WIDTH_MM}mm`
  clone.style.maxWidth = `${RECEIPT_WIDTH_MM}mm`
  clone.style.padding = `${RECEIPT_PADDING_MM}mm`
  clone.style.boxSizing = 'border-box'
  clone.style.visibility = 'hidden'
  clone.style.pointerEvents = 'none'
  clone.style.transform = 'none'
  clone.style.boxShadow = 'none'
  clone.style.border = '0'
  clone.style.borderRadius = '0'
  clone.style.background = '#ffffff'

  document.body.appendChild(clone)
  const rect = clone.getBoundingClientRect()
  const width = Math.max(clone.scrollWidth, rect.width)
  const height = Math.max(clone.scrollHeight, rect.height)
  clone.remove()

  const maxWidth = A4_PRINTABLE_WIDTH_MM * PX_PER_MM
  const maxHeight = A4_PRINTABLE_HEIGHT_MM * PX_PER_MM
  const scale = Math.min(1, maxWidth / width, maxHeight / height)
  receipt.style.setProperty(RECEIPT_PRINT_SCALE, String(Math.max(0.1, Math.floor(scale * 1000) / 1000)))
}

export function resetReceiptPrintArea(elementID = 'receipt-print-area') {
  document.getElementById(elementID)?.style.removeProperty(RECEIPT_PRINT_SCALE)
}
