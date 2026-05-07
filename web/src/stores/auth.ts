import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { User } from '@/types';

const API_URL = import.meta.env.VITE_API_URL || '/api';

function parseJwtPayload(token: string): { exp?: number } | null {
  try {
    const parts = token.split('.')
    if (parts.length !== 3) return null
    const payload = JSON.parse(atob(parts[1]))
    return payload
  } catch {
    return null
  }
}

function isTokenExpired(token: string): boolean {
  const payload = parseJwtPayload(token)
  if (!payload?.exp) return true
  return Date.now() >= payload.exp * 1000
}

function loadStoredUser(): User | null {
  const rawUser = localStorage.getItem('user')
  if (!rawUser) return null

  try {
    return JSON.parse(rawUser) as User
  } catch {
    localStorage.removeItem('user')
    return null
  }
}

function checkAndClearExpiredToken() {
  const storedToken = localStorage.getItem('token')
  if (storedToken && isTokenExpired(storedToken)) {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    return false
  }
  return true
}

async function parseApiResponse(response: Response) {
  const text = await response.text()

  if (!text) {
    return {}
  }

  try {
    return JSON.parse(text)
  } catch {
    return { error: `服务返回非 JSON 响应 (${response.status})` }
  }
}

export const useAuthStore = defineStore('auth', () => {
  checkAndClearExpiredToken()
  const storedToken = localStorage.getItem('token')
  const tokenNotExpired = storedToken && !isTokenExpired(storedToken)
  const token = ref<string | null>(tokenNotExpired ? storedToken : null)
  const user = ref<User | null>(tokenNotExpired ? loadStoredUser() : null)
  const isAuthenticated = ref(!!token.value)

  const loginWithPassword = async (email: string, password: string) => {
    const response = await fetch(`${API_URL}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: email, password }),
    })

    const data = await parseApiResponse(response)

    if (!response.ok) {
      throw new Error(data.error || `登录失败 (${response.status})`)
    }

    token.value = data.token
    user.value = data.user
    isAuthenticated.value = true
    localStorage.setItem('token', data.token)
    localStorage.setItem('user', JSON.stringify(data.user))
  }

  const register = async (username: string, email: string, password: string, passwordConfirm: string, verificationCode: string) => {
    const response = await fetch(`${API_URL}/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ 
        username, 
        email, 
        password, 
        password_confirm: passwordConfirm,
        verification_code: verificationCode 
      }),
    })

    const data = await parseApiResponse(response)

    if (!response.ok) {
      throw new Error(data.error || `注册失败 (${response.status})`)
    }

    token.value = data.token
    user.value = data.user
    isAuthenticated.value = true
    localStorage.setItem('token', data.token)
    localStorage.setItem('user', JSON.stringify(data.user))
  }

  const logout = () => {
    token.value = null;
    user.value = null;
    isAuthenticated.value = false;
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  };

  const validateSession = () => {
    if (token.value && isTokenExpired(token.value)) {
      logout()
      return false
    }
    return !!token.value
  }

  return {
    token,
    user,
    isAuthenticated,
    loginWithPassword,
    register,
    validateSession,
    logout
  };
});
