<template>
  <header class="topbar">
    <div class="topbar-inner">
      <!-- Logo -->
      <RouterLink to="/" class="brand">ATOMAN</RouterLink>

      <!-- Nav -->
      <nav class="nav">
        <RouterLink to="/feed" class="nav-link" :class="{ active: $route.path === '/' || $route.path.startsWith('/feed') }">订阅</RouterLink>
        <RouterLink to="/music" class="nav-link" :class="{ active: $route.path.startsWith('/music') || $route.path.startsWith('/artist=') }">音乐</RouterLink>
        <RouterLink to="/blog" class="nav-link" :class="{ active: isBlogContext }">博客</RouterLink>
        <RouterLink to="/debate" class="nav-link" :class="{ active: $route.path.startsWith('/debate') }">辩论</RouterLink>
        <RouterLink to="/timeline" class="nav-link" :class="{ active: $route.path.startsWith('/timeline') }">时间线</RouterLink>

        <!-- Blog sub-links when in blog context -->
        <template v-if="isBlogContext">
          <span class="nav-sep">|</span>
          <RouterLink v-if="authStore.isAuthenticated" to="/post/new" class="nav-link-sm">写文章</RouterLink>
          <RouterLink to="/blog" class="nav-link-sm">发现</RouterLink>
        </template>
      </nav>

      <!-- Right side -->
      <div class="nav-right">
        <!-- Notifications dropdown -->
        <div v-if="authStore.isAuthenticated" class="dropdown-wrap" data-dropdown="notif">
          <button class="notif-btn" @click="toggleDropdown('notif')">
            通知<span v-if="unreadCount > 0" class="notif-count">{{ unreadCount > 9 ? '9+' : unreadCount }}</span>
          </button>
          <div v-if="activeDropdown === 'notif'" class="dropdown notif-dropdown">
            <div v-if="notifications.length === 0" class="notif-drop-empty">暂无通知</div>
            <div
              v-for="n in recentNotifications"
              :key="n.id"
              class="notif-drop-item"
              :class="{ unread: !n.read_at }"
              @click="openNotification(n)"
            >
              <span class="notif-drop-dot" v-if="!n.read_at" />
              <span class="notif-drop-text">{{ n.content }}</span>
            </div>
            <div class="dropdown-divider" />
            <RouterLink to="/blog/notifications" class="dropdown-item notif-drop-all" @click="closeDropdown">
              查看全部通知
            </RouterLink>
          </div>
        </div>

        <!-- User menu -->
        <div v-if="authStore.isAuthenticated" class="dropdown-wrap" data-dropdown="user">
          <button class="user-btn" @click="toggleDropdown('user')">
            <span class="user-avatar">{{ userInitial }}</span>
            <span class="user-name">{{ authStore.user?.username }}</span>
            <span class="chevron" :style="activeDropdown === 'user' ? 'transform:rotate(180deg)' : ''">▾</span>
          </button>
          <div v-if="activeDropdown === 'user'" class="dropdown user-dropdown">
            <RouterLink :to="`/user/${authStore.user?.username}`" class="dropdown-item" @click="closeDropdown">我的主页</RouterLink>
            <RouterLink to="/blog/bookmarks" class="dropdown-item" @click="closeDropdown">我的收藏</RouterLink>
            <RouterLink to="/blog/settings" class="dropdown-item" @click="closeDropdown">编辑资料</RouterLink>
            <div class="dropdown-divider" />
            <button class="dropdown-item dropdown-item-danger" @click="logout">退出登录</button>
          </div>
        </div>

        <!-- Login -->
        <RouterLink v-else to="/login" class="login-btn">登录</RouterLink>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { RouterLink, useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/composables/useApi'
import type { Notification } from '@/types'

const authStore = useAuthStore()
const router = useRouter()
const api = useApi()

const activeDropdown = ref<string | null>(null)
const notifications = ref<Notification[]>([])

const unreadCount = computed(() => notifications.value.filter(n => !n.read_at).length)
const recentNotifications = computed(() => notifications.value.slice(0, 5))
const userInitial = computed(() => (authStore.user?.username || '?').charAt(0).toUpperCase())

const route = useRoute()
const isBlogContext = computed(() => {
  const p = route.path
  return (
    p.startsWith('/blog') ||
    p.startsWith('/channel') ||
    p.startsWith('/collection') ||
    p.startsWith('/post') ||
    p.startsWith('/user')
  )
})

const toggleDropdown = (name: string) => {
  activeDropdown.value = activeDropdown.value === name ? null : name
}

const closeDropdown = () => { activeDropdown.value = null }

const handleClickOutside = (e: MouseEvent) => {
  const target = e.target as HTMLElement
  if (!target.closest('[data-dropdown]')) closeDropdown()
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  if (authStore.isAuthenticated) fetchNotifications()
})

onUnmounted(() => document.removeEventListener('click', handleClickOutside))

const fetchNotifications = async () => {
  if (!authStore.token) return
  try {
    const res = await fetch('/api/notifications', {
      headers: { Authorization: `Bearer ${authStore.token}` }
    })
    if (res.ok) {
      const d = await res.json()
      notifications.value = d.data || []
    }
  } catch (e) { console.error(e) }
}

const logout = () => {
  authStore.logout()
  closeDropdown()
  router.push('/login')
}

const openNotification = (n: Notification) => {
  closeDropdown()
  if (!n.read_at) {
    fetch(`/api/notifications/${n.id}/read`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authStore.token}` },
    }).catch(() => {})
    const idx = notifications.value.findIndex(x => x.id === n.id)
    if (idx !== -1) notifications.value[idx] = { ...notifications.value[idx], read_at: new Date().toISOString() }
  }
  if (n.target_type === 'post' && n.target_id) {
    router.push(`/post/${n.target_id}`)
  } else {
    router.push('/blog/notifications')
  }
}
</script>

<style scoped>
.topbar {
  position: sticky;
  top: 0;
  z-index: 50;
  background: #fff;
  border-bottom: 2px solid #000;
  height: 64px;
}
.topbar-inner {
  max-width: 1152px;
  margin: 0 auto;
  padding: 0 2rem;
  height: 100%;
  display: flex;
  align-items: center;
  gap: 2rem;
}
.brand {
  font-size: 1.25rem;
  font-weight: 900;
  letter-spacing: -0.05em;
  color: #000;
  text-decoration: none;
  flex-shrink: 0;
}
.nav {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  flex: 1;
}
.nav-link {
  font-size: 0.875rem;
  font-weight: 700;
  color: #6b7280;
  text-decoration: none;
  transition: color 0.2s;
}
.nav-link:hover { color: #000; text-decoration: underline; }
.nav-link.active { color: #000; }
.nav-sep { color: #d1d5db; }
.nav-link-sm {
  font-size: 0.75rem;
  font-weight: 700;
  color: #9ca3af;
  text-decoration: none;
  transition: color 0.2s;
}
.nav-link-sm:hover { color: #000; text-decoration: underline; }
.nav-right {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-left: auto;
}
.dropdown-wrap { position: relative; }
.notif-btn {
  font-size: var(--a-text-sm);
  font-weight: var(--a-font-weight-strong);
  color: var(--a-color-muted);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  position: relative;
  transition: color 0.2s;
}
.notif-btn:hover { color: var(--a-color-fg); text-decoration: underline; }
.notif-count {
  display: inline-block;
  margin-left: 3px;
  background: var(--a-color-fg);
  color: var(--a-color-bg);
  font-size: 0.6rem;
  font-weight: var(--a-font-weight-black);
  border-radius: 9999px;
  padding: 1px 5px;
  line-height: 1;
  vertical-align: middle;
}
.notif-dropdown { width: 280px; right: 0; }
.notif-drop-empty {
  padding: var(--a-space-4);
  font-size: 0.8rem;
  color: var(--a-color-muted-soft);
  text-align: center;
}
.notif-drop-item {
  display: flex;
  align-items: flex-start;
  gap: var(--a-space-2);
  padding: 0.625rem var(--a-space-4);
  cursor: pointer;
  transition: background 0.1s;
}
.notif-drop-item:hover { background: var(--a-color-surface); }
.notif-drop-item:hover .notif-drop-text { text-decoration: underline; }
.notif-drop-dot {
  width: 6px;
  height: 6px;
  border-radius: 9999px;
  background: var(--a-color-fg);
  flex-shrink: 0;
  margin-top: 0.3rem;
}
.notif-drop-text {
  font-size: 0.8rem;
  font-weight: var(--a-font-weight-normal);
  line-height: 1.4;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
.notif-drop-item.unread .notif-drop-text { font-weight: var(--a-font-weight-strong); }
.notif-drop-all {
  font-size: var(--a-text-xs);
  font-weight: var(--a-font-weight-strong);
  text-align: center;
  color: var(--a-color-muted);
}
.user-btn {
  display: flex;
  align-items: center;
  gap: var(--a-space-2);
  background: none;
  border: var(--a-border);
  cursor: pointer;
  padding: 0.375rem 0.75rem;
  font-weight: var(--a-font-weight-strong);
  font-size: var(--a-text-sm);
  transition: all 0.2s;
}
.user-btn:hover { text-decoration: underline; }
.user-avatar {
  width: 24px;
  height: 24px;
  border-radius: 9999px;
  background: var(--a-color-fg);
  color: var(--a-color-bg);
  font-weight: var(--a-font-weight-black);
  font-size: var(--a-text-xs);
  display: flex;
  align-items: center;
  justify-content: center;
}
.user-name { font-weight: var(--a-font-weight-strong); }
.chevron { font-size: var(--a-text-xs); transition: transform 0.2s; }
.dropdown {
  position: absolute;
  right: 0;
  top: calc(100% + 4px);
  background: var(--a-color-bg);
  border: var(--a-border);
  box-shadow: var(--a-shadow-dropdown);
  z-index: var(--a-z-dropdown);
  min-width: 140px;
}
.user-dropdown { width: 144px; }
.dropdown-item {
  display: block;
  width: 100%;
  text-align: left;
  padding: 0.625rem var(--a-space-4);
  font-size: var(--a-text-sm);
  font-weight: var(--a-font-weight-strong);
  color: var(--a-color-fg);
  text-decoration: none;
  background: none;
  border: none;
  border-bottom: 1px solid #f3f4f6;
  cursor: pointer;
}
.dropdown-item:last-child { border-bottom: none; }
.dropdown-item:hover { text-decoration: underline; background: none; }
.dropdown-item-danger { color: var(--a-color-danger); }
.dropdown-item-danger:hover { background: none; text-decoration: underline; }
.dropdown-divider { height: 1px; background: #f3f4f6; margin: 0.25rem 0; }
.login-btn {
  font-size: var(--a-text-sm);
  font-weight: var(--a-font-weight-black);
  text-decoration: none;
  color: var(--a-color-bg);
  background: var(--a-color-fg);
  border: var(--a-border);
  padding: 0.375rem var(--a-space-4);
  transition: all 0.2s;
  text-transform: uppercase;
  letter-spacing: var(--a-letter-spacing-widest);
}
.login-btn:hover { text-decoration: underline; }
</style>
