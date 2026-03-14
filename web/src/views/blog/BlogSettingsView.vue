<template>
  <div class="max-w-2xl mx-auto px-8 py-12 pb-48">
    <div class="flex items-center justify-between mb-10">
      <h1 class="text-4xl font-black tracking-tighter border-l-8 border-black pl-6">编辑资料</h1>
      <RouterLink
        :to="`/blog/@${authStore.user?.username}`"
        class="text-xs font-black uppercase tracking-widest border-b-2 border-black hover:opacity-60"
      >
        ← 我的主页
      </RouterLink>
    </div>

    <!-- Avatar preview -->
    <div class="border-2 border-black p-6 mb-8 flex items-center gap-6">
      <div class="w-20 h-20 rounded-full bg-black flex items-center justify-center text-white text-3xl font-black flex-shrink-0 overflow-hidden">
        <img v-if="form.avatar_url" :src="form.avatar_url" alt="avatar" class="w-full h-full object-cover" />
        <span v-else>{{ (form.display_name || authStore.user?.username || '?').charAt(0).toUpperCase() }}</span>
      </div>
      <div class="flex-1">
        <p class="font-black text-lg">{{ form.display_name || authStore.user?.username }}</p>
        <p class="text-sm text-gray-400 font-medium">@{{ authStore.user?.username }}</p>
      </div>
    </div>

    <form @submit.prevent="save" class="flex flex-col gap-6">
      <!-- Display name -->
      <div>
        <label class="text-xs font-black uppercase tracking-widest text-gray-500 block mb-2">显示名称</label>
        <input
          v-model="form.display_name"
          placeholder="用于展示的名称"
          class="w-full border-2 border-black p-4 font-medium focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] outline-none transition-all"
        />
      </div>

      <!-- Bio -->
      <div>
        <label class="text-xs font-black uppercase tracking-widest text-gray-500 block mb-2">个人简介</label>
        <textarea
          v-model="form.bio"
          placeholder="介绍一下自己..."
          rows="4"
          class="w-full border-2 border-black p-4 font-medium focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] outline-none transition-all resize-none"
        />
      </div>

      <!-- Website -->
      <div>
        <label class="text-xs font-black uppercase tracking-widest text-gray-500 block mb-2">个人网站</label>
        <input
          v-model="form.website"
          placeholder="https://yoursite.com"
          type="url"
          class="w-full border-2 border-black p-4 font-medium focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] outline-none transition-all"
        />
      </div>

      <!-- Location -->
      <div>
        <label class="text-xs font-black uppercase tracking-widest text-gray-500 block mb-2">所在地</label>
        <input
          v-model="form.location"
          placeholder="城市或地区"
          class="w-full border-2 border-black p-4 font-medium focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] outline-none transition-all"
        />
      </div>

      <!-- Avatar URL -->
      <div>
        <label class="text-xs font-black uppercase tracking-widest text-gray-500 block mb-2">头像 URL</label>
        <input
          v-model="form.avatar_url"
          placeholder="https://example.com/avatar.jpg"
          class="w-full border-2 border-black p-4 font-medium focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] outline-none transition-all"
        />
      </div>

      <!-- Divider: Password change -->
      <div class="border-t-2 border-black pt-6">
        <h2 class="text-lg font-black tracking-tight mb-4">修改密码</h2>
        <div class="flex flex-col gap-4">
          <div>
            <label class="text-xs font-black uppercase tracking-widest text-gray-500 block mb-2">新密码</label>
            <input
              v-model="newPassword"
              type="password"
              placeholder="留空表示不修改"
              class="w-full border-2 border-black p-4 font-medium focus:outline-none"
            />
          </div>
          <div>
            <label class="text-xs font-black uppercase tracking-widest text-gray-500 block mb-2">确认新密码</label>
            <input
              v-model="confirmPassword"
              type="password"
              placeholder="再次输入新密码"
              class="w-full border-2 border-black p-4 font-medium focus:outline-none"
            />
          </div>
        </div>
      </div>

      <!-- Feedback -->
      <div v-if="error" class="border-2 border-red-500 bg-red-50 p-4 text-red-700 font-bold text-sm">
        {{ error }}
      </div>
      <div v-if="success" class="border-2 border-green-600 bg-green-50 p-4 text-green-700 font-bold text-sm">
        ✓ 保存成功
      </div>

      <!-- Submit -->
      <button
        type="submit"
        :disabled="saving"
        class="w-full py-4 bg-black text-white font-black uppercase tracking-widest border-2 border-black hover:bg-white hover:text-black transition-all disabled:opacity-40"
      >
        {{ saving ? '保存中...' : '保存更改' }}
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
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
