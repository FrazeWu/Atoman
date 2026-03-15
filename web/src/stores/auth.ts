import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { User } from '@/types';

const API_URL = import.meta.env.VITE_API_URL || '/api';

function loadStoredUser(): User | null {
  const rawUser = localStorage.getItem('user')
  if (!rawUser) return null

  try {
    return JSON.parse(rawUser) as User
  } catch {
    // Clear invalid persisted state from older/broken responses.
    localStorage.removeItem('user')
    return null
  }
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
  const token = ref<string | null>(localStorage.getItem('token'));
  const user = ref<User | null>(loadStoredUser());
  const isAuthenticated = ref(!!token.value);

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

  const register = async (username: string, email: string, password: string) => {
    const response = await fetch(`${API_URL}/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, email, password }),
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

  return {
    token,
    user,
    isAuthenticated,
    loginWithPassword,
    register,
    logout
  };
});
