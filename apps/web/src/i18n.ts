export type AppLanguage = 'th' | 'en'
export type AppTheme = 'light' | 'dark'

export const messages = {
  en: {
    'app.name': 'Grocery POS',
    'app.subtitle': 'Inventory System',
    'topbar.kicker': 'Small grocery operations',
    'topbar.workspace': 'Foundation workspace',
    'topbar.alerts': 'Alerts',
    'topbar.menu': 'Menu',
    'topbar.logout': 'Logout',
    'topbar.navigation': 'Navigation',
    'topbar.close': 'Close',
    'settings.title': 'Settings',
    'settings.eyebrow': 'System',
    'settings.description': 'Choose language, theme, and display preferences for this workstation.',
    'settings.appearance': 'Appearance',
    'settings.appearanceDescription': 'These settings are saved on this browser.',
    'settings.language': 'Language',
    'settings.theme': 'Theme',
    'settings.light': 'Light',
    'settings.dark': 'Dark',
    'settings.english': 'English',
    'settings.thai': 'Thai',
    'settings.preview': 'Preview',
    'settings.previewText': 'The app shell, navigation, forms, tables, and cards will follow the selected theme.',
    'login.secureAccess': 'Secure access',
    'login.title': 'Login',
    'login.demoUsers': 'Demo users: admin, manager, cashier. Password: password.',
    'login.username': 'Username',
    'login.password': 'Password',
    'login.loading': 'Logging in...',
    'login.submit': 'Login',
    'nav.dashboard': 'Dashboard',
    'nav.pos': 'POS',
    'nav.products': 'Products',
    'nav.categories': 'Categories',
    'nav.restock': 'Restock',
    'nav.stockMovements': 'Stock Movements',
    'nav.locations': 'Locations',
    'nav.transfers': 'Transfers',
    'nav.salesHistory': 'Sales History',
    'nav.receiptDetail': 'Receipt Detail',
    'nav.alerts': 'Alerts',
    'nav.reports': 'Reports',
    'nav.exports': 'Exports',
    'nav.imports': 'Imports',
    'nav.purchaseOrders': 'Purchase Orders',
    'nav.suppliers': 'Suppliers',
    'nav.users': 'Users',
    'nav.settings': 'Settings',
  },
  th: {
    'app.name': 'ระบบขายหน้าร้าน',
    'app.subtitle': 'จัดการสต็อกสินค้า',
    'topbar.kicker': 'งานร้านขายของชำ',
    'topbar.workspace': 'พื้นที่ทำงานหลัก',
    'topbar.alerts': 'แจ้งเตือน',
    'topbar.menu': 'เมนู',
    'topbar.logout': 'ออกจากระบบ',
    'topbar.navigation': 'เมนูนำทาง',
    'topbar.close': 'ปิด',
    'settings.title': 'ตั้งค่า',
    'settings.eyebrow': 'ระบบ',
    'settings.description': 'เลือกภาษา ธีม และการแสดงผลสำหรับเครื่องนี้',
    'settings.appearance': 'การแสดงผล',
    'settings.appearanceDescription': 'การตั้งค่านี้จะบันทึกไว้ในเบราว์เซอร์นี้',
    'settings.language': 'ภาษา',
    'settings.theme': 'ธีม',
    'settings.light': 'สว่าง',
    'settings.dark': 'มืด',
    'settings.english': 'อังกฤษ',
    'settings.thai': 'ไทย',
    'settings.preview': 'ตัวอย่าง',
    'settings.previewText': 'โครงแอป เมนู ฟอร์ม ตาราง และการ์ดจะใช้ธีมที่เลือก',
    'login.secureAccess': 'เข้าสู่ระบบอย่างปลอดภัย',
    'login.title': 'เข้าสู่ระบบ',
    'login.demoUsers': 'ผู้ใช้ตัวอย่าง: admin, manager, cashier รหัสผ่าน: password',
    'login.username': 'ชื่อผู้ใช้',
    'login.password': 'รหัสผ่าน',
    'login.loading': 'กำลังเข้าสู่ระบบ...',
    'login.submit': 'เข้าสู่ระบบ',
    'nav.dashboard': 'แดชบอร์ด',
    'nav.pos': 'ขายหน้าร้าน',
    'nav.products': 'สินค้า',
    'nav.categories': 'หมวดหมู่',
    'nav.restock': 'เติมสต็อก',
    'nav.stockMovements': 'ประวัติสต็อก',
    'nav.locations': 'สถานที่เก็บ',
    'nav.transfers': 'โอนย้ายสต็อก',
    'nav.salesHistory': 'ประวัติการขาย',
    'nav.receiptDetail': 'รายละเอียดใบเสร็จ',
    'nav.alerts': 'แจ้งเตือน',
    'nav.reports': 'รายงาน',
    'nav.exports': 'ส่งออกข้อมูล',
    'nav.imports': 'นำเข้าข้อมูล',
    'nav.purchaseOrders': 'ใบสั่งซื้อ',
    'nav.suppliers': 'ซัพพลายเออร์',
    'nav.users': 'ผู้ใช้',
    'nav.settings': 'ตั้งค่า',
  },
} as const

export type TranslationKey = keyof typeof messages.en

export function translateMessage(language: AppLanguage, key: TranslationKey) {
  return messages[language][key] ?? messages.en[key] ?? key
}

export function readStoredLanguage(): AppLanguage {
  const value = localStorage.getItem('app_language')
  return value === 'th' || value === 'en' ? value : 'th'
}

export function readStoredTheme(): AppTheme {
  const value = localStorage.getItem('app_theme')
  return value === 'light' || value === 'dark' ? value : 'light'
}

export function applyDocumentPreferences(language: AppLanguage, theme: AppTheme) {
  document.documentElement.lang = language
  document.documentElement.classList.toggle('dark', theme === 'dark')
}
