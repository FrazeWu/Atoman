<template>
  <div class="login-page">
    <div class="login-card">
      <h1 class="login-title">{{ isRegister ? '加入我们' : '欢迎回来' }}</h1>
      <p class="login-sub">进入 Atoman 的数字领域</p>

      <form @submit.prevent="handleSubmit" class="login-form">
        <AInput
          v-if="isRegister"
          v-model="username"
          label="用户名"
          placeholder="输入用户名"
          :error="fieldErrors.username"
        />

        <div class="a-field">
          <label class="a-field-label">{{ isRegister ? '邮箱地址' : '用户名或邮箱' }}</label>
          <div v-if="isRegister" class="code-row">
            <input
              type="email"
              required
              class="a-input"
              v-model="email"
              placeholder="请输入邮箱地址"
            />
            <button
              type="button"
              class="a-btn a-btn--secondary a-btn--sm code-btn"
              @click="sendVerificationCode"
              :disabled="countdown > 0"
            >
              {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
            </button>
          </div>
          <input
            v-else
            type="text"
            required
            class="a-input"
            v-model="email"
            placeholder="输入用户名或邮箱"
          />
        </div>

        <AInput
          v-if="isRegister"
          v-model="verificationCode"
          label="验证码"
          placeholder="6 位数字验证码"
          maxlength="6"
          :error="fieldErrors.code"
        />

        <AInput
          v-model="password"
          label="通行密码"
          type="password"
          placeholder="输入密码"
          :error="fieldErrors.password"
        />

        <AInput
          v-if="isRegister"
          v-model="passwordConfirm"
          label="确认密码"
          type="password"
          placeholder="再次输入密码"
          :error="fieldErrors.passwordConfirm"
        />

        <ABtn
          type="submit"
          variant="primary"
          size="lg"
          block
          :loading="loading"
          loading-text="请稍候..."
        >
          {{ isRegister ? '注册账号' : '登 录' }}
        </ABtn>
      </form>

      <div v-if="errorMsg" class="a-error" style="margin-top:1rem">{{ errorMsg }}</div>

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

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/composables/useApi'
import AInput from '@/components/ui/AInput.vue'
import ABtn from '@/components/ui/ABtn.vue'

const email = ref('')
const password = ref('')
const passwordConfirm = ref('')
const username = ref('')
const verificationCode = ref('')
const codeSent = ref(false)
const countdown = ref(0)
const errorMsg = ref('')
const loading = ref(false)
const fieldErrors = ref<Record<string, string>>({})
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
      codeSent.value = false
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
    if (!response.ok) throw new Error(data.details || data.error || '发送验证码失败')
    codeSent.value = true
    startCountdown()
  } catch (error: any) {
    errorMsg.value = error.message || '发送验证码失败'
  }
}

const handleSubmit = async () => {
  errorMsg.value = ''
  fieldErrors.value = {}
  loading.value = true
  try {
    if (isRegister.value) {
      if (!verificationCode.value) {
        fieldErrors.value.code = '请输入验证码'
        loading.value = false
        return
      }
      if (password.value !== passwordConfirm.value) {
        fieldErrors.value.passwordConfirm = '两次输入的密码不一致'
        loading.value = false
        return
      }
      if (password.value.length < 6) {
        fieldErrors.value.password = '密码长度至少为 6 位'
        loading.value = false
        return
      }
      await authStore.register(username.value, email.value, password.value, passwordConfirm.value, verificationCode.value)
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
  background: var(--a-color-bg);
  border: var(--a-border);
  padding: 3rem;
  box-shadow: var(--a-shadow-modal);
}
.login-title {
  font-size: 2.25rem;
  font-weight: var(--a-font-weight-black);
  letter-spacing: var(--a-letter-spacing-tight);
  margin: 0 0 0.5rem;
}
.login-sub {
  color: var(--a-color-muted-soft);
  font-weight: var(--a-font-weight-normal);
  margin: 0 0 2rem;
}
.login-form {
  display: flex;
  flex-direction: column;
  gap: var(--a-space-5);
}
.code-row {
  display: flex;
  gap: var(--a-space-2);
}
.code-row .a-input { flex: 1; }
.code-btn { flex-shrink: 0; white-space: nowrap; }
.login-footer {
  margin-top: 2rem;
  padding-top: 2rem;
  border-top: 1px solid #f3f4f6;
  text-align: center;
  font-size: var(--a-text-sm);
  font-weight: var(--a-font-weight-normal);
}
.toggle-link {
  font-weight: var(--a-font-weight-black);
  text-decoration: underline;
  color: var(--a-color-fg);
}
@media (max-width: 480px) {
  .login-page { padding: 1rem; align-items: flex-start; padding-top: 3rem; }
  .login-card { padding: 2rem 1.5rem; }
  .login-title { font-size: 1.75rem; }
}
</style>
