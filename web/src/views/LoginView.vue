<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const email = ref('')
const password = ref('')
const username = ref('')
const errorMsg = ref('')
const loading = ref(false)
const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const isRegister = computed(() => route.path === '/register')

const handleSubmit = async () => {
  errorMsg.value = ''
  loading.value = true
  try {
    if (isRegister.value) {
      await authStore.register(username.value, email.value, password.value)
    } else {
      await authStore.loginWithPassword(email.value, password.value)
    }
    const redirect = route.query.redirect as string
    router.push(redirect || '/')
  } catch (error: any) {
    errorMsg.value = error.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <h1 class="login-title">{{ isRegister ? '加入我们' : '欢迎回来' }}</h1>
      <p class="login-sub">进入 Atoman 的数字领域</p>

      <form @submit.prevent="handleSubmit" class="login-form">
        <div v-if="isRegister" class="field">
          <label class="field-label">用户名</label>
          <input type="text" required class="field-input" v-model="username" />
        </div>

        <div class="field">
          <label class="field-label">邮箱地址</label>
          <input type="email" required class="field-input" v-model="email" />
        </div>

        <div class="field">
          <label class="field-label">通行密码</label>
          <input type="password" required class="field-input" v-model="password" />
        </div>

        <button type="submit" class="submit-btn" :disabled="loading">
          {{ loading ? '请稍候...' : (isRegister ? '注册账号' : '登 录') }}
        </button>
      </form>

      <div v-if="errorMsg" class="error-msg">{{ errorMsg }}</div>

      <div class="login-footer">
        <span v-if="isRegister">
          已有账号？ <RouterLink to="/login" class="toggle-link">立即登录</RouterLink>
        </span>
        <span v-else>
          还没有账号？ <RouterLink to="/register" class="toggle-link">立即加入档案室</RouterLink>
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: calc(100vh - 64px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
}
.login-card {
  max-width: 28rem;
  width: 100%;
  background: #fff;
  border: 2px solid #000;
  padding: 3rem;
  box-shadow: 20px 20px 0px 0px rgba(0,0,0,1);
}
.login-title { font-size: 2.25rem; font-weight: 900; letter-spacing: -0.05em; margin: 0 0 0.5rem; }
.login-sub { color: #9ca3af; font-weight: 500; margin: 0 0 2rem; }
.login-form { display: flex; flex-direction: column; gap: 1.5rem; }
.field { display: flex; flex-direction: column; gap: 0.5rem; }
.field-label { font-size: 0.75rem; font-weight: 900; text-transform: uppercase; letter-spacing: 0.1em; }
.field-input {
  border: 2px solid #000;
  padding: 0.75rem;
  font-size: 0.875rem;
  outline: none;
  transition: background 0.15s;
  width: 100%;
  box-sizing: border-box;
}
.field-input:focus { background: #f9fafb; }
.submit-btn {
  width: 100%;
  background: #000;
  color: #fff;
  padding: 1rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  border: 2px solid #000;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.875rem;
}
.submit-btn:hover { background: #fff; color: #000; }
.submit-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.error-msg {
  margin-top: 1rem;
  padding: 0.75rem 1rem;
  border: 2px solid #000;
  background: #fef2f2;
  font-size: 0.875rem;
  font-weight: 700;
  color: #dc2626;
}
.login-footer {
  margin-top: 2rem;
  padding-top: 2rem;
  border-top: 1px solid #f3f4f6;
  text-align: center;
  font-size: 0.875rem;
  font-weight: 500;
}
.toggle-link { font-weight: 900; text-decoration: underline; color: #000; }
</style>
