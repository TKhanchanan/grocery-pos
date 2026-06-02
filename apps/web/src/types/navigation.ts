export interface NavigationItem {
  label: string
  to: string
  roles?: Role[]
}

export type Role = 'ADMIN' | 'MANAGER' | 'CASHIER'

export interface User {
  id: number
  username: string
  fullName: string
  role: Role
  active: boolean
  createdAt: string
}
