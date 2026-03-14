<script setup lang="ts">
import { RouterLink, useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useFeedStore } from '@/stores/feed'
import { ref, computed, onMounted, onUnmounted } from 'vue'

const authStore = useAuthStore()
const feedStore = useFeedStore()
const router = useRouter()
const route = useRoute()
const isDropdownOpen = ref(false)
const isNotifOpen = ref(false)

const handleLogout = () => {
  feedStore.stopPolling()
  authStore.logout()
  router.push('/login')
  isDropdownOpen.value = false
}

const currentModule = computed(() => {
  if (route.path.startsWith('/blog')) return 'blog'
  if (route.path.startsWith('/music')) return 'music'
  return 'orbit' // Default to orbit since it's the home page
})

// Close dropdowns when clicking outside
const handleClickOutside = (e: MouseEvent) => {
  const target = e.target as HTMLElement
  if (!target.closest('[data-dropdown]')) {
    isDropdownOpen.value = false
    isNotifOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  if (authStore.isAuthenticated) {
    feedStore.startPolling()
  }
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})

const openNotif = async () => {
  isNotifOpen.value = !isNotifOpen.value
  isDropdownOpen.value = false
  if (isNotifOpen.value) {
    await feedStore.fetchNotifications()
  }
}

const handleNotifClick = async (notif: any) => {
  isNotifOpen.value = false
  if (!notif.read_at) await feedStore.markRead(notif.id)
  if (notif.target_type === 'post' && notif.target_id) {
    router.push(`/blog/posts/${notif.target_id}`)
  }
}

const relativeTime = (d: string) => {
  const diff = Date.now() - new Date(d).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return '刚刚'
  if (mins < 60) return `${mins}分钟前`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}小时前`
  return `${Math.floor(hours / 24)}天前`
}
</script>

<template>
  <header class="sticky top-0 z-50 bg-white border-b-2 border-black h-16 flex items-center px-8 shadow-sm">
    <div class="max-w-6xl w-full mx-auto flex justify-between items-center">
      <div class="flex items-center gap-8">
        <RouterLink to="/" class="text-2xl font-black tracking-tighter hover:opacity-70 transition-opacity">
          ATOMAN
        </RouterLink>

        <!-- Module Switcher -->
        <nav class="flex gap-4 font-bold text-lg">
          <RouterLink to="/"
            class="px-2 py-1 transition-all"
            :class="currentModule === 'orbit' ? 'underline decoration-2 underline-offset-8' : 'hover:underline hover:decoration-2 hover:underline-offset-8 text-gray-500 hover:text-black'">
            订阅
          </RouterLink>
          <RouterLink to="/music"
            class="px-2 py-1 transition-all"
            :class="currentModule === 'music' ? 'underline decoration-2 underline-offset-8' : 'hover:underline hover:decoration-2 hover:underline-offset-8 text-gray-500 hover:text-black'">
            音乐
          </RouterLink>
          <RouterLink to="/blog"
            class="px-2 py-1 transition-all"
            :class="currentModule === 'blog' ? 'underline decoration-2 underline-offset-8' : 'hover:underline hover:decoration-2 hover:underline-offset-8 text-gray-500 hover:text-black'">
            博客
          </RouterLink>
        </nav>
      </div>

      <nav class="flex gap-6 font-medium items-center">
        <!-- Module specific links -->
        <template v-if="currentModule === 'music'">
          <RouterLink v-if="authStore.isAuthenticated" to="/upload" class="hover:underline">上传</RouterLink>
          <RouterLink v-if="authStore.user?.role === 'admin'" to="/admin/review"
            class="text-red-600 hover:text-red-800 font-bold hover:underline">
            审核队列
          </RouterLink>
        </template>

        <template v-if="currentModule === 'blog'">
          <RouterLink to="/blog/explore" class="hover:underline">发现</RouterLink>
          <RouterLink v-if="authStore.isAuthenticated" to="/blog/posts/new"
            class="hover:underline font-bold text-blue-600">写文章</RouterLink>
        </template>

        <!-- Notification Bell -->
        <div v-if="authStore.isAuthenticated" class="relative" data-dropdown>
          <button
            @click.stop="openNotif"
            class="relative hover:opacity-70 transition-opacity flex items-center"
          >
            <span class="text-sm font-black uppercase tracking-widest hover:underline">消息</span>
            <span
              v-if="feedStore.unreadCount > 0"
              class="absolute -top-1 -right-1 bg-black text-white text-xs font-black rounded-full h-4 w-4 flex items-center justify-center leading-none"
            >
              {{ feedStore.unreadCount > 9 ? '9+' : feedStore.unreadCount }}
            </span>
          </button>

          <!-- Notification dropdown -->
          <div
            v-show="isNotifOpen"
            class="absolute right-0 top-full mt-2 w-80 bg-white border-2 border-black shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] animate-in fade-in slide-in-from-top-1 duration-200 z-50"
          >
            <div class="flex items-center justify-between px-4 py-3 border-b-2 border-black">
              <span class="font-black text-sm uppercase tracking-widest">通知</span>
              <button
                v-if="feedStore.unreadCount > 0"
                @click="feedStore.markAllRead()"
                class="text-xs font-black text-gray-400 hover:text-black transition-colors"
              >
                全部已读
              </button>
            </div>
            <div class="max-h-80 overflow-y-auto">
              <div v-if="!feedStore.notifications.length" class="py-10 text-center text-gray-400 text-sm font-medium">
                暂无通知
              </div>
              <button
                v-for="notif in feedStore.notifications.slice(0, 10)"
                :key="notif.id"
                @click="handleNotifClick(notif)"
                class="w-full text-left px-4 py-3 border-b border-gray-100 hover:bg-gray-50 transition-colors flex gap-3 items-start"
                :class="!notif.read_at ? 'bg-gray-50' : ''"
              >
                <span class="w-2 h-2 rounded-full flex-shrink-0 mt-2"
                  :class="!notif.read_at ? 'bg-black' : 'bg-gray-200'" />
                <div class="flex-1 min-w-0">
                  <p class="text-sm font-medium leading-snug line-clamp-2">{{ notif.content }}</p>
                  <p class="text-xs text-gray-400 font-black mt-1">{{ relativeTime(notif.created_at) }}</p>
                </div>
              </button>
            </div>
          </div>
        </div>

        <!-- User dropdown -->
        <div v-if="authStore.isAuthenticated" class="relative" data-dropdown>
          <button
            @click.stop="isDropdownOpen = !isDropdownOpen; isNotifOpen = false"
            class="font-black hover:underline flex items-center gap-1 py-2"
          >
            {{ authStore.user?.display_name || authStore.user?.username }}
            <span class="text-xs transition-transform duration-200" :class="{ 'rotate-180': isDropdownOpen }">▼</span>
          </button>

          <div
            v-show="isDropdownOpen"
            class="absolute right-0 top-full mt-1 w-36 animate-in fade-in slide-in-from-top-1 duration-200"
          >
            <div class="bg-white border-2 border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] flex flex-col">
              <RouterLink :to="`/blog/@${authStore.user?.username}`"
                class="w-full text-left px-4 py-3 hover:bg-black hover:text-white font-bold transition-colors"
                @click="isDropdownOpen = false">
                我的主页
              </RouterLink>
              <RouterLink to="/blog/bookmarks"
                class="w-full text-left px-4 py-3 hover:bg-black hover:text-white font-bold transition-colors border-t-2 border-black"
                @click="isDropdownOpen = false">
                我的收藏
              </RouterLink>
              <RouterLink to="/blog/settings"
                class="w-full text-left px-4 py-3 hover:bg-black hover:text-white font-bold transition-colors border-t-2 border-black"
                @click="isDropdownOpen = false">
                设置
              </RouterLink>
              <button @click="handleLogout"
                class="w-full text-left px-4 py-3 hover:bg-black hover:text-white font-bold transition-colors border-t-2 border-black">
                登出
              </button>
            </div>
          </div>
        </div>

        <RouterLink v-if="!authStore.isAuthenticated" to="/login" class="hover:underline">登录</RouterLink>
        <RouterLink to="/about" class="hover:underline">关于</RouterLink>
      </nav>
    </div>
  </header>
</template>
