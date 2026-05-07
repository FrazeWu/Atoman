<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/composables/useApi'

const email = ref('')
const password = ref('')
const passwordConfirm = ref('')
const username = ref('')
const verificationCode = ref('')
const codeSent = ref(false)
const countdown = ref(0)
const errorMsg = ref('')
const loading = ref(false)
const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const api = useApi()

const isRegister = computed(() => route.path === '/register')

const startCountdown = () => {
  countdown.value = 60
  const timer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearInterval(timer)
      codeSent.value = false // Allow resending after countdown finishes
    }
  }, 1000)
}

const sendVerificationCode = async () => {
  if (!email.value || !email.value.includes('@')) {
    errorMsg.value = '请输入有效的邮箱地址'
    return
  }
  
  try {
    const response = await fetch(api.auth.sendVerification, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: email.value }),
    })
    
    const data = await response.json()
    
    if (!response.ok) {
      throw new Error(data.details || data.error || '发送验证码失败')
    }
    
    codeSent.value = true
    startCountdown()
  } catch (error: any) {
    console.error('发送验证码失败:', error)
    errorMsg.value = error.message || '发送验证码失败'
  }
}

const handleSubmit = async () => {
  errorMsg.value = ''
  loading.value = true
  try {
    if (isRegister.value) {
      if (!verificationCode.value) {
        errorMsg.value = '请输入验证码'
        loading.value = false
        return
      }
      if (password.value !== passwordConfirm.value) {
        errorMsg.value = '两次输入的密码不一致'
        loading.value = false
        return
      }
      if (password.value.length < 6) {
        errorMsg.value = '密码长度至少为 6 位'
        loading.value = false
        return
      }
      await authStore.register(username.value, email.value, password.value, passwordConfirm.value, verificationCode.value)
    } else {
      // Login: email field can be username or email
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
          <label class="field-label">{{ isRegister ? '邮箱地址' : '用户名或邮箱' }}</label>
          <input 
            :type="isRegister ? 'email' : 'text'" 
            required 
            class="field-input" 
            v-model="email" 
            :placeholder="isRegister ? '请输入邮箱地址' : '输入用户名或邮箱'"
          />
        </div>

        <div v-if="isRegister" class="field">
          <label class="field-label">验证码</label>
          <div style="display:flex;gap:0.5rem">
            <input 
              type="text" 
              required 
              class="field-input" 
              v-model="verificationCode" 
              maxlength="6" 
              placeholder="6 位数字验证码"
              style="flex:1"
            />
            <button 
              type="button" 
              class="code-btn" 
              @click="sendVerificationCode" 
              :disabled="countdown > 0"
              style="white-space:nowrap;min-width:120px"
            >
              {{ countdown > 0 ? `${countdown}秒后重发` : '获取验证码' }}
            </button>
          </div>
        </div>

        <div class="field">
          <label class="field-label">通行密码</label>
          <input type="password" required class="field-input" v-model="password" />
        </div>

        <div v-if="isRegister" class="field">
          <label class="field-label">确认密码</label>
          <input 
            type="password" 
            required 
            class="field-input" 
            v-model="passwordConfirm" 
            placeholder="再次输入密码"
          />
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
.code-btn {
  background: #fff;
  color: #000;
  padding: 0.75rem 1rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border: 2px solid #000;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.75rem;
  flex-shrink: 0;
}
.code-btn:hover:not(:disabled) { background: #000; color: #fff; }
.code-btn:disabled { opacity: 0.5; cursor: not-allowed; background: #f3f4f6; border-color: #d1d5db; }
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

@media (max-width: 480px) {
  .login-page { padding: 1rem; align-items: flex-start; padding-top: 3rem; }
  .login-card { padding: 2rem 1.5rem; box-shadow: 8px 8px 0px 0px rgba(0,0,0,1); }
  .login-title { font-size: 1.75rem; }
}
</style>
