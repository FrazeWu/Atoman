<template>
  <div class="artist-history">
    <div class="page-header">
      <RouterLink :to="`/music/artists/${artistId}`" class="back-link">← 返回艺术家</RouterLink>
      <h1 class="page-title">编辑历史</h1>
    </div>

    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else>
      <p class="total">共 {{ total }} 次编辑</p>

      <!-- Diff viewer -->
      <div v-if="showDiff && diff" class="diff-panel">
        <h3 class="diff-title">版本对比</h3>
        <pre class="diff-content">{{ JSON.stringify(diff, null, 2) }}</pre>
        <button @click="showDiff = false" class="btn-close-diff">关闭</button>
      </div>

      <!-- Revision list -->
      <div class="revision-list">
        <div
          v-for="rev in revisions"
          :key="rev.id"
          class="revision-item"
          :class="{ selected: compareFrom === rev.version_number || compareTo === rev.version_number }"
        >
          <div class="rev-meta">
            <span class="rev-version">v{{ rev.version_number }}</span>
            <span class="rev-editor">{{ rev.editor?.username || '未知' }}</span>
            <span class="rev-date">{{ formatDate(rev.created_at) }}</span>
            <span class="rev-status" :class="`rev-status-${rev.status}`">{{ rev.status }}</span>
          </div>
          <p class="rev-summary">{{ rev.edit_summary || '（无摘要）' }}</p>
          <div class="rev-actions">
            <button @click="selectForCompare(rev.version_number)" class="btn-compare">
              {{ compareFrom === rev.version_number ? '已选（起）' : compareTo === rev.version_number ? '已选（止）' : '选择对比' }}
            </button>
            <button
              v-if="authStore.isAuthenticated"
              @click="revert(rev.version_number)"
              class="btn-revert"
            >还原</button>
          </div>
        </div>
      </div>

      <div v-if="compareFrom && compareTo" class="compare-bar">
        <span>对比 v{{ Math.min(compareFrom, compareTo) }} → v{{ Math.max(compareFrom, compareTo) }}</span>
        <button @click="fetchDiff" class="btn-run-diff">查看差异</button>
        <button @click="compareFrom = null; compareTo = null" class="btn-clear">清除</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const authStore = useAuthStore()
const API_URL = import.meta.env.VITE_API_URL || '/api'

const artistId = computed(() => route.params.artistId as string)
const revisions = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const total = ref(0)

const compareFrom = ref<number | null>(null)
const compareTo = ref<number | null>(null)
const diff = ref<any>(null)
const showDiff = ref(false)

const formatDate = (d: string) => new Date(d).toLocaleString('zh-CN')

const fetchRevisions = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await fetch(`${API_URL}/artists/${artistId.value}/revisions?limit=50`, {
      headers: authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {},
    })
    const data = await res.json()
    revisions.value = data.data || []
    total.value = data.total || 0
  } catch (e: any) {
    error.value = e.message || 'Failed to fetch revisions'
  } finally {
    loading.value = false
  }
}

const fetchDiff = async () => {
  if (!compareFrom.value || !compareTo.value) return
  try {
    const v1 = Math.min(compareFrom.value, compareTo.value)
    const v2 = Math.max(compareFrom.value, compareTo.value)
    const res = await fetch(`${API_URL}/albums/dummy/revisions/diff?v1=${v1}&v2=${v2}`)
    const data = await res.json()
    diff.value = data.data || {}
    showDiff.value = true
  } catch (e) {
    console.error('Failed to fetch diff:', e)
  }
}

const selectForCompare = (version: number) => {
  if (!compareFrom.value) {
    compareFrom.value = version
  } else if (!compareTo.value) {
    compareTo.value = version
  } else {
    compareFrom.value = version
    compareTo.value = null
  }
}

const revert = async (version: number) => {
  const summary = prompt('还原摘要（可选）：')
  try {
    await fetch(`${API_URL}/artists/${artistId.value}/revert/${version}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`,
      },
      body: JSON.stringify({ edit_summary: summary || `Revert to v${version}` }),
    })
    await fetchRevisions()
  } catch (e) {
    console.error('Failed to revert:', e)
  }
}

onMounted(fetchRevisions)
</script>

<style scoped>
.artist-history {
  max-width: 48rem;
  margin: 0 auto;
  padding: 2rem;
  padding-bottom: 12rem;
}
.back-link {
  font-size: 0.75rem;
  font-weight: 900;
  text-decoration: none;
  color: #6b7280;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  display: block;
  margin-bottom: 0.5rem;
}
.back-link:hover { color: #000; }
.page-title {
  font-size: 2rem;
  font-weight: 900;
  letter-spacing: -0.04em;
  margin: 0 0 1rem;
}
.total {
  font-size: 0.75rem;
  color: #6b7280;
  margin-bottom: 1rem;
}
.loading, .error { color: #6b7280; padding: 2rem 0; }
.diff-panel {
  border: 2px solid #000;
  padding: 1rem;
  margin-bottom: 1.5rem;
  background: #f9fafb;
}
.diff-title { font-size: 0.875rem; font-weight: 900; margin: 0 0 0.5rem; }
.diff-content {
  font-size: 0.75rem;
  overflow: auto;
  max-height: 300px;
  background: #fff;
  padding: 0.75rem;
  border: 1px solid #e5e7eb;
}
.btn-close-diff {
  margin-top: 0.75rem;
  border: 2px solid #000;
  padding: 0.25rem 0.75rem;
  font-size: 0.625rem;
  font-weight: 900;
  text-transform: uppercase;
  cursor: pointer;
  background: #fff;
}
.btn-close-diff:hover { background: #000; color: #fff; }
.revision-list { display: flex; flex-direction: column; gap: 0.75rem; }
.revision-item {
  border: 2px solid #000;
  padding: 1rem;
  transition: box-shadow 0.15s;
}
.revision-item:hover { box-shadow: 4px 4px 0 0 rgba(0,0,0,1); }
.revision-item.selected { border-color: #000; background: #f3f4f6; }
.rev-meta {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.375rem;
  flex-wrap: wrap;
}
.rev-version {
  font-size: 0.75rem;
  font-weight: 900;
  background: #000;
  color: #fff;
  padding: 0.125rem 0.5rem;
}
.rev-editor { font-size: 0.875rem; font-weight: 700; }
.rev-date { font-size: 0.75rem; color: #6b7280; }
.rev-status {
  font-size: 0.5rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  padding: 0.125rem 0.375rem;
  border: 1px solid;
}
.rev-status-approved { border-color: #166534; color: #166534; }
.rev-status-pending { border-color: #92400e; color: #92400e; }
.rev-status-rejected { border-color: #991b1b; color: #991b1b; }
.rev-summary { font-size: 0.875rem; color: #374151; margin: 0 0 0.75rem; }
.rev-actions { display: flex; gap: 0.5rem; }
.btn-compare, .btn-revert {
  border: 1px solid #000;
  padding: 0.25rem 0.625rem;
  font-size: 0.625rem;
  font-weight: 900;
  text-transform: uppercase;
  cursor: pointer;
  background: #fff;
  transition: all 0.15s;
}
.btn-compare:hover, .btn-revert:hover { background: #000; color: #fff; }
.compare-bar {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-top: 1.5rem;
  padding: 0.75rem;
  border: 2px solid #000;
  background: #f9fafb;
  font-size: 0.875rem;
  font-weight: 700;
}
.btn-run-diff {
  border: 2px solid #000;
  padding: 0.25rem 0.75rem;
  font-size: 0.625rem;
  font-weight: 900;
  text-transform: uppercase;
  cursor: pointer;
  background: #000;
  color: #fff;
}
.btn-clear {
  border: 1px solid #6b7280;
  padding: 0.25rem 0.625rem;
  font-size: 0.625rem;
  font-weight: 900;
  text-transform: uppercase;
  cursor: pointer;
  background: #fff;
  color: #6b7280;
}
</style>
