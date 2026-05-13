<template>
  <RouterLink :to="`/post/${post.id}`" class="post-card-link">
    <div class="post-card">
      <!-- Cover Image -->
      <div v-if="post.cover_url" class="post-cover">
        <img :src="post.cover_url" :alt="post.title" class="post-cover-img" />
      </div>

      <div class="post-body">
        <!-- Title -->
        <h2 class="post-title line-clamp-2">{{ post.title }}</h2>

        <!-- Summary -->
        <p v-if="post.summary" class="post-summary line-clamp-3">{{ post.summary }}</p>
        <div v-else class="flex-spacer" />

        <!-- Meta row -->
        <div class="post-meta">
          <div class="post-author">
            <div class="author-avatar">
              {{ (post.user?.display_name || post.user?.username || '?').charAt(0).toUpperCase() }}
            </div>
            <span class="author-name">
              <template v-if="showChannel && post.channel">
                <RouterLink
                  :to="`/channel/${post.channel.slug || post.channel.id}`"
                  class="channel-link"
                  @click.stop
                >{{ post.channel.name }}</RouterLink>
              </template>
              <template v-else>{{ post.user?.display_name || post.user?.username }}</template>
            </span>
          </div>
          <div class="post-stats">
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

defineProps<{ post: Post; showChannel?: boolean }>()

const formatDate = (dateStr: string) => {
  const d = new Date(dateStr)
  return `${d.getFullYear()}.${String(d.getMonth() + 1).padStart(2, '0')}.${String(d.getDate()).padStart(2, '0')}`
}
</script>

<style scoped>
.post-card-link { display: block; text-decoration: none; color: inherit; }
.post-card {
  background: var(--a-color-bg);
  border: 2px solid var(--a-color-fg);
  display: flex;
  flex-direction: column;
  height: 100%;
  transition: box-shadow 0.3s;
}
.post-card:hover { box-shadow: 10px 10px 0px 0px rgba(0,0,0,1); }
.post-cover {
  border-bottom: 2px solid var(--a-color-fg);
  overflow: hidden;
}
.post-cover-img {
  width: 100%;
  aspect-ratio: 16/9;
  object-fit: cover;
  display: block;
  transition: transform 0.5s;
}
.post-card:hover .post-cover-img { transform: scale(1.05); }
.post-body {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  flex: 1;
}
.post-title {
  font-size: 1.125rem;
  font-weight: 900;
  letter-spacing: -0.02em;
  line-height: 1.3;
  margin: 0 0 0.5rem;
}
.post-card:hover .post-title { text-decoration: underline; }
.post-summary {
  font-size: 0.875rem;
  color: var(--a-color-muted);
  font-weight: 500;
  line-height: 1.6;
  margin: 0 0 1rem;
  flex: 1;
}
.flex-spacer { flex: 1; }
.post-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--a-color-disabled-border);
}
.post-author { display: flex; align-items: center; gap: 0.5rem; }
.author-avatar {
  width: 24px;
  height: 24px;
  border-radius: 9999px;
  background: var(--a-color-fg);
  color: var(--a-color-bg);
  font-size: 0.7rem;
  font-weight: 900;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.author-name { font-size: 0.75rem; font-weight: 700; color: var(--a-color-muted); }
.channel-link { color: var(--a-color-muted); text-decoration: none; font-weight: 700; }
.channel-link:hover { text-decoration: underline; }
.post-stats {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #9ca3af;
}
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
