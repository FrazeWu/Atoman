<template>
  <div class="max-w-5xl mx-auto px-8 py-12 pb-48">
    <div class="flex items-center justify-between mb-8">
      <h1 class="text-4xl font-black tracking-tighter border-l-8 border-black pl-6">我的收藏</h1>
      <button
        @click="showNewFolder = true"
        class="text-xs font-black uppercase tracking-widest border-2 border-black px-4 py-2 hover:bg-black hover:text-white transition-all"
      >
        + 新建收藏夹
      </button>
    </div>

    <div class="flex gap-0 border-2 border-black min-h-[60vh]">
      <!-- Left: Folder list -->
      <div class="w-56 flex-shrink-0 border-r-2 border-black">
        <button
          @click="activeFolder = null"
          class="w-full text-left px-5 py-4 font-black text-sm border-b-2 border-black transition-all"
          :class="activeFolder === null ? 'bg-black text-white' : 'hover:bg-gray-50'"
        >
          全部收藏
        </button>
        <button
          v-for="folder in folders"
          :key="folder.id"
          @click="activeFolder = folder.id"
          class="w-full text-left px-5 py-4 font-medium text-sm border-b border-gray-100 transition-all flex items-center justify-between group"
          :class="activeFolder === folder.id ? 'bg-black text-white' : 'hover:bg-gray-50'"
        >
          <span>{{ folder.name }}</span>
          <button
            @click.stop="deleteFolder(folder.id)"
            class="text-xs opacity-0 group-hover:opacity-100 transition-opacity font-bold"
            :class="activeFolder === folder.id ? 'text-white' : 'text-red-500'"
          >
            ✕
          </button>
        </button>
      </div>

      <!-- Right: Bookmarked posts -->
      <div class="flex-1 p-6">
        <div v-if="loadingPosts" class="grid grid-cols-1 md:grid-cols-2 gap-5">
          <div v-for="i in 4" :key="i" class="h-36 bg-gray-100 border border-gray-200 animate-pulse" />
        </div>
        <div v-else-if="!filteredBookmarks.length" class="h-full flex items-center justify-center text-gray-400 font-medium">
          暂无收藏
        </div>
        <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-5">
          <PostCard v-for="bm in filteredBookmarks" :key="bm.id" :post="bm.post!" />
        </div>
      </div>
    </div>

    <!-- New folder modal -->
    <div v-if="showNewFolder" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div class="bg-white border-2 border-black p-8 w-full max-w-sm shadow-[20px_20px_0px_0px_rgba(0,0,0,1)]">
        <h3 class="text-xl font-black tracking-tight mb-5">新建收藏夹</h3>
        <input
          v-model="newFolderName"
          placeholder="收藏夹名称"
          class="w-full border-2 border-black p-3 font-medium focus:outline-none mb-4"
          @keyup.enter="createFolder"
        />
        <div class="flex gap-2">
          <button @click="createFolder" class="flex-1 bg-black text-white py-2 font-black uppercase tracking-widest border-2 border-black hover:bg-white hover:text-black transition-all">
            创建
          </button>
          <button @click="showNewFolder = false" class="px-5 py-2 font-black border-2 border-black hover:bg-black hover:text-white transition-all">
            取消
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import PostCard from '@/components/blog/PostCard.vue'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/composables/useApi'
import type { Bookmark, BookmarkFolder } from '@/types'

const authStore = useAuthStore()
const api = useApi()

const folders = ref<BookmarkFolder[]>([])
const bookmarks = ref<Bookmark[]>([])
const activeFolder = ref<number | null>(null)
const loadingPosts = ref(true)
const showNewFolder = ref(false)
const newFolderName = ref('')

const filteredBookmarks = computed(() => {
  if (activeFolder.value === null) return bookmarks.value.filter(b => b.post)
  return bookmarks.value.filter(b => b.bookmark_folder_id === activeFolder.value && b.post)
})

const authHeader = computed(() => ({ Authorization: `Bearer ${authStore.token}` }))

const fetchAll = async () => {
  loadingPosts.value = true
  try {
    const [fRes, bRes] = await Promise.all([
      fetch(api.blog.bookmarkFolders, { headers: authHeader.value }),
      fetch(api.blog.bookmarks, { headers: authHeader.value })
    ])
    if (fRes.ok) folders.value = (await fRes.json()).data || []
    if (bRes.ok) bookmarks.value = (await bRes.json()).data || []
  } catch (e) {
    console.error(e)
  } finally {
    loadingPosts.value = false
  }
}

const createFolder = async () => {
  if (!newFolderName.value.trim()) return
  try {
    const res = await fetch(api.blog.bookmarkFolders, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeader.value },
      body: JSON.stringify({ name: newFolderName.value })
    })
    if (res.ok) {
      showNewFolder.value = false
      newFolderName.value = ''
      await fetchAll()
    }
  } catch (e) {
    console.error(e)
  }
}

const deleteFolder = async (id: number) => {
  try {
    await fetch(api.blog.bookmarkFolder(id), { method: 'DELETE', headers: authHeader.value })
    if (activeFolder.value === id) activeFolder.value = null
    await fetchAll()
  } catch (e) {
    console.error(e)
  }
}

onMounted(fetchAll)
</script>
