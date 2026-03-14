<template>
  <RouterLink :to="`/blog/posts/${post.id}`" class="block group">
    <div
      class="bg-white border-2 border-black hover:shadow-[10px_10px_0px_0px_rgba(0,0,0,1)] transition-all duration-300 flex flex-col h-full"
    >
      <!-- Cover Image -->
      <div v-if="post.cover_url" class="border-b-2 border-black overflow-hidden">
        <img
          :src="post.cover_url"
          :alt="post.title"
          class="w-full aspect-video object-cover group-hover:scale-105 transition-transform duration-500"
        />
      </div>

      <div class="p-6 flex flex-col flex-1">
        <!-- Title -->
        <h2 class="text-xl font-black tracking-tight leading-tight mb-2 group-hover:underline line-clamp-2">
          {{ post.title }}
        </h2>

        <!-- Summary -->
        <p v-if="post.summary" class="text-sm text-gray-600 font-medium leading-relaxed mb-4 line-clamp-3 flex-1">
          {{ post.summary }}
        </p>
        <div v-else class="flex-1" />

        <!-- Meta row -->
        <div class="flex items-center justify-between mt-4 pt-4 border-t border-gray-200">
          <!-- Author -->
          <div class="flex items-center gap-2">
            <div class="w-6 h-6 rounded-full bg-black flex items-center justify-center text-white text-xs font-black">
              {{ (post.user?.display_name || post.user?.username || '?').charAt(0).toUpperCase() }}
            </div>
            <span class="text-xs font-bold text-gray-700">
              {{ post.user?.display_name || post.user?.username }}
            </span>
          </div>

          <!-- Stats -->
          <div class="flex items-center gap-3 text-xs font-black uppercase tracking-widest text-gray-400">
            <span v-if="post.likes_count !== undefined">♥ {{ post.likes_count }}</span>
            <span v-if="post.comments_count !== undefined">💬 {{ post.comments_count }}</span>
            <span>{{ formatDate(post.created_at) }}</span>
          </div>
        </div>
      </div>
    </div>
  </RouterLink>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
import type { Post } from '@/types'

defineProps<{ post: Post }>()

const formatDate = (dateStr: string) => {
  const d = new Date(dateStr)
  return `${d.getFullYear()}.${String(d.getMonth() + 1).padStart(2, '0')}.${String(d.getDate()).padStart(2, '0')}`
}
</script>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.line-clamp-3 {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
