<template>
  <div class="a-page-sm" style="padding-bottom:12rem">
    <div class="a-section-header" style="margin-bottom:2.5rem">
      <h1 class="a-title a-accent-l">编辑资料</h1>
      <RouterLink :to="`/user/${authStore.user?.username}`" class="a-link">← 我的主页</RouterLink>
    </div>

    <!-- Avatar preview -->
    <div class="a-card" style="display:flex;align-items:center;gap:1.5rem;margin-bottom:2rem">
      <div style="width:5rem;height:5rem;border-radius:9999px;background:#000;display:flex;align-items:center;justify-content:center;color:#fff;font-size:1.875rem;font-weight:900;flex-shrink:0;overflow:hidden">
        <img v-if="form.avatar_url" :src="form.avatar_url" alt="avatar" style="width:100%;height:100%;object-fit:cover" />
        <span v-else>{{ (form.display_name || authStore.user?.username || '?').charAt(0).toUpperCase() }}</span>
      </div>
      <div>
        <p style="font-weight:900;font-size:1.125rem">{{ form.display_name || authStore.user?.username }}</p>
        <p class="a-muted" style="font-size:.875rem">@{{ authStore.user?.username }}</p>
      </div>
    </div>

    <form @submit.prevent="save" style="display:flex;flex-direction:column;gap:1.5rem">
      <div class="a-field">
        <label class="a-field-label">显示名称</label>
        <input v-model="form.display_name" placeholder="用于展示的名称" class="a-input" />
      </div>
      <div class="a-field">
        <label class="a-field-label">个人简介</label>
        <textarea v-model="form.bio" placeholder="介绍一下自己..." rows="4" class="a-textarea" />
      </div>
      <div class="a-field">
        <label class="a-field-label">个人网站</label>
        <input v-model="form.website" placeholder="https://yoursite.com" type="url" class="a-input" />
      </div>
      <div class="a-field">
        <label class="a-field-label">所在地</label>
        <input v-model="form.location" placeholder="城市或地区" class="a-input" />
      </div>
      <div class="a-field">
        <label class="a-field-label">头像 URL</label>
        <input v-model="form.avatar_url" placeholder="https://example.com/avatar.jpg" class="a-input" />
      </div>

      <div style="border-top:2px solid #000;padding-top:1.5rem">
        <h2 class="a-subtitle" style="margin-bottom:1rem">修改密码</h2>
        <div style="display:flex;flex-direction:column;gap:1rem">
          <div class="a-field">
            <label class="a-field-label">新密码</label>
            <input v-model="newPassword" type="password" placeholder="留空表示不修改" class="a-input" />
          </div>
          <div class="a-field">
            <label class="a-field-label">确认新密码</label>
            <input v-model="confirmPassword" type="password" placeholder="再次输入新密码" class="a-input" />
          </div>
        </div>
      </div>

      <div v-if="error" class="a-error">{{ error }}</div>
      <div v-if="success" class="a-success">✓ 保存成功</div>

      <ABtn block type="submit" :loading="saving" loadingText="保存中...">保存更改</ABtn>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import ABtn from '@/components/ui/ABtn.vue'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/composables/useApi'

const authStore = useAuthStore()
const api = useApi()

const form = ref({
  display_name: '',
  bio: '',
  website: '',
  location: '',
  avatar_url: '',
})

const newPassword = ref('')
const confirmPassword = ref('')
const saving = ref(false)
const error = ref('')
const success = ref(false)

const loadProfile = async () => {
  try {
    const res = await fetch(api.users.me, {
      headers: { Authorization: `Bearer ${authStore.token}` }
    })
    if (res.ok) {
      const d = await res.json()
      const u = d.data || d
      form.value = {
        display_name: u.display_name || '',
        bio: u.bio || '',
        website: u.website || '',
        location: u.location || '',
        avatar_url: u.avatar_url || '',
      }
    }
  } catch (e) {
    console.error(e)
  }
}

const save = async () => {
  error.value = ''
  success.value = false

  if (newPassword.value && newPassword.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return
  }

  saving.value = true
  try {
    const payload: Record<string, string> = { ...form.value }
    if (newPassword.value) payload.password = newPassword.value

    const res = await fetch(api.users.settings, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`
      },
      body: JSON.stringify(payload)
    })

    if (res.ok) {
      const d = await res.json()
      const updated = d.data || d
      // Update local user state
      if (authStore.user) {
        authStore.user.display_name = updated.display_name
        authStore.user.avatar_url = updated.avatar_url
        authStore.user.bio = updated.bio
        localStorage.setItem('user', JSON.stringify(authStore.user))
      }
      newPassword.value = ''
      confirmPassword.value = ''
      success.value = true
      setTimeout(() => { success.value = false }, 3000)
    } else {
      const err = await res.json()
      error.value = err.error || '保存失败'
    }
  } catch (e) {
    error.value = '网络错误，请重试'
  } finally {
    saving.value = false
  }
}

onMounted(loadProfile)
</script>
