import { createPinia, setActivePinia } from 'pinia'

import router from '@/router'
import { useAuthStore } from '@/stores/auth'

const makeToken = (expSecondsFromNow: number) => {
  const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }))
  const payload = btoa(JSON.stringify({ exp: Math.floor(Date.now() / 1000) + expSecondsFromNow }))
  return `${header}.${payload}.signature`
}

describe('router auth guards', () => {
  beforeEach(async () => {
    localStorage.clear()
    setActivePinia(createPinia())
    await router.replace('/')
  })

  it('redirects unauthenticated user to login for protected routes', async () => {
    const auth = useAuthStore()
    auth.logout()

    await router.push('/post/new')

    expect(router.currentRoute.value.path).toBe('/login')
    expect(router.currentRoute.value.query.redirect).toBe('/post/new')
  })

  it('redirects non-admin user away from admin routes', async () => {
    const auth = useAuthStore()
    auth.token = makeToken(3600)
    auth.user = { username: 'member', role: 'user' } as never
    auth.isAuthenticated = true

    await router.push('/music/admin/review')

    expect(router.currentRoute.value.path).toBe('/feed')
  })
})
