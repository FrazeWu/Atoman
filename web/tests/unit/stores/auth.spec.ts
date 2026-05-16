import { createPinia, setActivePinia } from 'pinia'

import { useAuthStore } from '@/stores/auth'

const makeToken = (expSecondsFromNow: number) => {
  const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }))
  const payload = btoa(JSON.stringify({ exp: Math.floor(Date.now() / 1000) + expSecondsFromNow }))
  return `${header}.${payload}.signature`
}

describe('auth store', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
  })

  it('loads non-expired token from localStorage', () => {
    localStorage.setItem('token', makeToken(3600))
    localStorage.setItem('user', JSON.stringify({ username: 'demo', role: 'user' }))

    const auth = useAuthStore()

    expect(auth.isAuthenticated).toBe(true)
    expect(auth.user?.username).toBe('demo')
  })

  it('clears expired token on initialization', () => {
    localStorage.setItem('token', makeToken(-10))
    localStorage.setItem('user', JSON.stringify({ username: 'expired', role: 'user' }))

    const auth = useAuthStore()

    expect(auth.isAuthenticated).toBe(false)
    expect(auth.token).toBeNull()
    expect(localStorage.getItem('token')).toBeNull()
    expect(localStorage.getItem('user')).toBeNull()
  })
})
