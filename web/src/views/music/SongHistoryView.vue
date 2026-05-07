<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const authStore = useAuthStore()
const API_URL = import.meta.env.VITE_API_URL || '/api'

const songId = computed(() => route.params.songId as string)
const revisions = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const total = ref(0)
const compareFrom = ref<number | null>(null)
const compareTo = ref<number | null>(null)
const diff = ref<Record<string, any> | null>(null)

const canRevert = computed(() => authStore.user?.role === 'admin')

const authHeaders = () => authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {}

const fetchRevisions = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await fetch(`${API_URL}/songs/${songId.value}/revisions?limit=50`, {
      headers: authHeaders(),
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || '加载修订历史失败')
    revisions.value = data.data || []
    total.value = data.total || 0
  } catch (e: any) {
    error.value = e.message || '加载修订历史失败'
  } finally {
    loading.value = false
  }
}

const fetchDiff = async () => {
  if (!compareFrom.value || !compareTo.value) return
  const v1 = Math.min(compareFrom.value, compareTo.value)
  const v2 = Math.max(compareFrom.value, compareTo.value)
  const res = await fetch(`${API_URL}/songs/${songId.value}/revisions/diff?v1=${v1}&v2=${v2}`, {
    headers: authHeaders(),
  })
  const data = await res.json()
  if (!res.ok) {
    error.value = data.error || '加载版本差异失败'
    return
  }
  diff.value = data.data || {}
}

const revertTo = async (version: number) => {
  if (!confirm(`回退到版本 #${version}？`)) return
  const res = await fetch(`${API_URL}/songs/${songId.value}/revert/${version}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${authStore.token}`,
    },
    body: JSON.stringify({ edit_summary: `回退到歌曲版本 ${version}` }),
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    alert(data.error || '回退失败')
    return
  }
  await fetchRevisions()
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleString('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const statusClass = (status: string) => {
  const map: Record<string, string> = {
    approved: 'status-approved',
    pending: 'status-pending',
    rejected: 'status-rejected',
    superseded: 'status-superseded',
  }
  return map[status] || ''
}

onMounted(fetchRevisions)
</script>

<template>
  <div class="page-container">
    <RouterLink to="/music" class="back-link">← 返回音乐库</RouterLink>

    <header class="header">
      <h1 class="title">歌曲修订历史</h1>
      <p class="subtitle">共 {{ total }} 次修订</p>
    </header>

    <section v-if="revisions.length > 1" class="diff-selector">
      <select v-model="compareFrom" class="select">
        <option :value="null">选择版本...</option>
        <option v-for="rev in revisions" :key="`from-${rev.version_number}`" :value="rev.version_number">
          版本 #{{ rev.version_number }}
        </option>
      </select>
      <span class="diff-label">对比</span>
      <select v-model="compareTo" class="select">
        <option :value="null">选择版本...</option>
        <option v-for="rev in revisions" :key="`to-${rev.version_number}`" :value="rev.version_number">
          版本 #{{ rev.version_number }}
        </option>
      </select>
      <button @click="fetchDiff" class="btn" :disabled="!compareFrom || !compareTo">对比</button>
    </section>

    <section v-if="diff" class="diff-view">
      <div class="diff-header">
        <h2>版本差异</h2>
        <button @click="diff = null" class="btn-close">×</button>
      </div>
      <div v-if="Object.keys(diff).length === 0" class="empty">两个版本无差异</div>
      <div v-for="(changes, field) in diff" :key="field" class="diff-field">
        <div class="diff-field-name">{{ field }}</div>
        <div class="diff-values">
          <div><span class="label">旧值</span>{{ changes.from }}</div>
          <div><span class="label">新值</span>{{ changes.to }}</div>
        </div>
      </div>
    </section>

    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error">{{ error }}</div>

    <section v-else class="revision-list">
      <article
        v-for="rev in revisions"
        :key="rev.id"
        class="revision-item"
        :class="{ 'is-current': rev.is_current }"
      >
        <div class="revision-header">
          <span class="version-badge" :class="statusClass(rev.status)">
            #{{ rev.version_number }}
            <span v-if="rev.is_current" class="current-badge">当前</span>
          </span>
          <span class="status" :class="statusClass(rev.status)">{{ rev.status }}</span>
        </div>
        <div class="revision-meta">
          <span>编辑者：{{ rev.editor?.username || 'Unknown' }}</span>
          <span>时间：{{ formatDate(rev.created_at) }}</span>
        </div>
        <div class="summary">摘要：{{ rev.edit_summary || '无' }}</div>
        <div class="type">类型：{{ rev.edit_type }}</div>
        <button
          v-if="canRevert && !rev.is_current"
          @click="revertTo(rev.version_number)"
          class="btn btn-revert"
        >
          回退
        </button>
      </article>

      <div v-if="revisions.length === 0" class="empty">暂无修订历史</div>
    </section>
  </div>
</template>

<style scoped>
.page-container {
  max-width: 900px;
  margin: 0 auto;
  padding: 2rem;
}
.back-link {
  display: inline-flex;
  margin-bottom: 2rem;
  font-weight: 800;
  color: #000;
  text-decoration: none;
}
.back-link:hover { text-decoration: underline; }
.header { margin-bottom: 2rem; }
.title {
  font-size: 2rem;
  font-weight: 900;
  letter-spacing: 0;
  margin: 0;
}
.subtitle { color: #666; margin-top: .5rem; }
.diff-selector {
  display: flex;
  gap: .75rem;
  align-items: center;
  padding: 1rem;
  border: 2px solid #000;
  margin-bottom: 1.5rem;
  flex-wrap: wrap;
}
.select {
  border: 2px solid #000;
  padding: .5rem;
  background: #fff;
  font-weight: 700;
}
.btn {
  border: 2px solid #000;
  background: #000;
  color: #fff;
  padding: .5rem 1rem;
  font-weight: 900;
  cursor: pointer;
}
.btn:disabled { opacity: .45; cursor: not-allowed; }
.diff-view {
  border: 2px solid #000;
  padding: 1rem;
  margin-bottom: 1.5rem;
  background: #fafafa;
}
.diff-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 2px solid #000;
  padding-bottom: .75rem;
  margin-bottom: 1rem;
}
.diff-header h2 { font-size: 1.1rem; margin: 0; }
.btn-close {
  border: 0;
  background: none;
  font-size: 1.5rem;
  cursor: pointer;
}
.diff-field {
  border-bottom: 1px solid #ddd;
  padding: .75rem 0;
}
.diff-field-name { font-weight: 900; margin-bottom: .5rem; }
.diff-values {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}
.label {
  display: block;
  font-size: .75rem;
  color: #666;
  font-weight: 900;
  margin-bottom: .25rem;
}
.revision-list { display: flex; flex-direction: column; gap: 1rem; }
.revision-item {
  border: 2px solid #000;
  padding: 1rem;
  background: #fff;
}
.revision-item.is-current { box-shadow: 5px 5px 0 #000; }
.revision-header,
.revision-meta {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
}
.version-badge,
.status {
  border: 2px solid #000;
  padding: .25rem .5rem;
  font-weight: 900;
}
.current-badge {
  margin-left: .5rem;
  color: #059669;
}
.status-approved { background: #dcfce7; }
.status-pending { background: #fef3c7; }
.status-rejected { background: #fee2e2; }
.status-superseded { background: #e5e7eb; }
.summary,
.type {
  margin-top: .75rem;
}
.btn-revert { margin-top: 1rem; }
.loading,
.error,
.empty {
  padding: 2rem;
  text-align: center;
  border: 2px solid #000;
}
.error { color: #dc2626; }
@media (max-width: 720px) {
  .diff-values { grid-template-columns: 1fr; }
}
</style>
