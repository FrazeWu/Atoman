<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import ASelect from '@/components/ui/ASelect.vue'

const route = useRoute()
const authStore = useAuthStore()
const API_URL = import.meta.env.VITE_API_URL || '/api'

const albumId = computed(() => route.params.albumId as string)
const revisions = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const total = ref(0)

// Comparison
const compareFrom = ref<number | null>(null)
const compareTo = ref<number | null>(null)
const diff = ref<any>(null)
const showDiff = ref(false)

const revisionOptions = computed(() => [
  { label: '选择版本...', value: null },
  ...revisions.value.map(rev => ({ label: `版本 #${rev.version_number}`, value: rev.version_number }))
])

const fetchRevisions = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await fetch(`${API_URL}/albums/${albumId.value}/revisions?limit=50`, {
      headers: authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {}
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
    const res = await fetch(`${API_URL}/albums/${albumId.value}/revisions/diff?v1=${v1}&v2=${v2}`, {
      headers: authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {}
    })
    const data = await res.json()
    diff.value = data.data || {}
    showDiff.value = true
  } catch (e) {
    console.error('Failed to fetch diff', e)
  }
}

const revertTo = async (version: number) => {
  if (!confirm(`回退到版本 #${version}？`)) return
  try {
    await fetch(`${API_URL}/albums/${albumId.value}/revert/${version}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`
      },
      body: JSON.stringify({ edit_summary: `回退到版本 ${version}` })
    })
    alert('回退成功')
    fetchRevisions()
  } catch (e: any) {
    alert(e.message || '回退失败')
  }
}

const canRevert = computed(() => authStore.user?.role === 'admin')

const formatDate = (date: string) => {
  return new Date(date).toLocaleString('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const statusClass = (status: string) => {
  const map: Record<string, string> = {
    approved: 'status-approved',
    pending: 'status-pending',
    rejected: 'status-rejected',
    superseded: 'status-superseded'
  }
  return map[status] || ''
}

onMounted(() => {
  fetchRevisions()
})
</script>

<template>
  <div class="page-container">
    <RouterLink :to="`/music/albums/${albumId}`" class="back-link">
      ← 返回专辑
    </RouterLink>

    <div class="header">
      <h1 class="title">修订历史</h1>
      <p class="subtitle">共 {{ total }} 次修订</p>
    </div>

    <!-- Diff Selector -->
    <div class="diff-selector" v-if="revisions.length > 1">
      <ASelect v-model="compareFrom" :options="revisionOptions" class="select" />
      <span class="diff-label">对比</span>
      <ASelect v-model="compareTo" :options="revisionOptions" class="select" />
      <button @click="fetchDiff" class="btn-compare" :disabled="!compareFrom || !compareTo">
        对比
      </button>
    </div>

    <!-- Diff View -->
    <div v-if="showDiff && diff" class="diff-view">
      <div class="diff-header">
        <h3>版本差异</h3>
        <button @click="showDiff = false" class="btn-close">×</button>
      </div>
      <div class="diff-content">
        <div v-for="(changes, field) in diff" :key="field" class="diff-field">
          <div class="diff-field-name">{{ field }}</div>
          <div class="diff-values">
            <div class="diff-value diff-from">
              <span class="diff-label-small">旧值:</span>
              <span class="diff-text">{{ changes.from }}</span>
            </div>
            <div class="diff-value diff-to">
              <span class="diff-label-small">新值:</span>
              <span class="diff-text">{{ changes.to }}</span>
            </div>
          </div>
        </div>
        <div v-if="Object.keys(diff).length === 0" class="no-diff">
          两个版本无差异
        </div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading">加载中...</div>

    <!-- Error -->
    <div v-else-if="error" class="error">{{ error }}</div>

    <!-- Revision List -->
    <div v-else class="revision-list">
      <div
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
          <div class="editor">
            <span class="label">编辑者:</span>
            <span class="value">{{ rev.editor?.username || 'Unknown' }}</span>
          </div>
          <div class="date">
            <span class="label">时间:</span>
            <span class="value">{{ formatDate(rev.created_at) }}</span>
          </div>
        </div>

        <div class="revision-summary">
          <span class="label">摘要:</span>
          <span class="value">{{ rev.edit_summary || '无' }}</span>
        </div>

        <div class="revision-type">
          <span class="label">类型:</span>
          <span class="value type-badge">{{ rev.edit_type }}</span>
        </div>

        <div v-if="rev.review_notes" class="review-notes">
          <span class="label">审核备注:</span>
          <span class="value">{{ rev.review_notes }}</span>
        </div>

        <div class="revision-actions">
          <RouterLink
            :to="`/music/albums/${albumId}/history/${rev.version_number}`"
            class="btn-action"
          >
            查看
          </RouterLink>
          <button
            v-if="canRevert && !rev.is_current"
            @click="revertTo(rev.version_number)"
            class="btn-action btn-revert"
          >
            回退
          </button>
        </div>
      </div>

      <div v-if="revisions.length === 0" class="empty">
        暂无修订历史
      </div>
    </div>
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
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 2rem;
  font-weight: 700;
  text-decoration: none;
  color: #000;
  transition: opacity 0.2s;
}
.back-link:hover { text-decoration: underline; }

.header {
  margin-bottom: 2rem;
}

.title {
  font-size: 2rem;
  font-weight: 900;
  letter-spacing: -0.05em;
  margin: 0 0 0.5rem 0;
}

.subtitle {
  color: #6b7280;
  margin: 0;
}

/* Diff Selector */
.diff-selector {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 2rem;
  padding: 1rem;
  background: #f9fafb;
  border: 2px solid #000;
}

.select {
  padding: 0.5rem 1rem;
  border: 2px solid #000;
  font-size: 0.875rem;
  font-weight: 700;
  background: #fff;
}

.diff-label {
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.btn-compare {
  padding: 0.5rem 1.5rem;
  background: #000;
  color: #fff;
  border: none;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  cursor: pointer;
}
.btn-compare:disabled {
  background: #9ca3af;
  cursor: not-allowed;
}

/* Diff View */
.diff-view {
  margin-bottom: 2rem;
  border: 2px solid #000;
  background: #fff;
}

.diff-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: #f9fafb;
  border-bottom: 2px solid #000;
}

.diff-header h3 {
  margin: 0;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  font-size: 0.875rem;
}

.btn-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
}

.diff-content {
  padding: 1rem;
}

.diff-field {
  margin-bottom: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #e5e7eb;
}

.diff-field:last-child {
  border-bottom: none;
}

.diff-field-name {
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-size: 0.75rem;
  margin-bottom: 0.5rem;
  color: #6b7280;
}

.diff-values {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.diff-value {
  padding: 0.75rem;
  border: 2px solid;
}

.diff-from {
  border-color: #dc2626;
  background: #fef2f2;
}

.diff-to {
  border-color: #16a34a;
  background: #f0fdf4;
}

.diff-label-small {
  display: block;
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  margin-bottom: 0.25rem;
}

.diff-text {
  word-break: break-all;
}

.no-diff {
  text-align: center;
  padding: 2rem;
  color: #6b7280;
}

/* Revision List */
.revision-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.revision-item {
  padding: 1.5rem;
  border: 2px solid #000;
  background: #fff;
  transition: box-shadow 0.2s;
}

.revision-item.is-current {
  border-color: #16a34a;
  background: #f0fdf4;
}

.revision-item:hover {
  box-shadow: 8px 8px 0px 0px rgba(0,0,0,1);
}

.revision-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
}

.version-badge {
  font-size: 1.25rem;
  font-weight: 900;
}

.current-badge {
  font-size: 0.75rem;
  background: #16a34a;
  color: #fff;
  padding: 0.125rem 0.5rem;
  margin-left: 0.5rem;
  text-transform: uppercase;
}

.status {
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  padding: 0.25rem 0.75rem;
  border: 2px solid;
}

.status-approved { background: #dcfce7; border-color: #16a34a; color: #16a34a; }
.status-pending { background: #fef3c7; border-color: #ca8a04; color: #ca8a04; }
.status-rejected { background: #fee2e2; border-color: #dc2626; color: #dc2626; }
.status-superseded { background: #f3f4f6; border-color: #6b7280; color: #6b7280; }

.revision-meta {
  display: flex;
  gap: 2rem;
  margin-bottom: 0.75rem;
  font-size: 0.875rem;
}

.revision-meta .label {
  font-weight: 700;
  color: #6b7280;
  margin-right: 0.5rem;
}

.revision-summary,
.revision-type,
.review-notes {
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
}

.revision-summary .label,
.revision-type .label,
.review-notes .label {
  font-weight: 700;
  color: #6b7280;
  margin-right: 0.5rem;
}

.type-badge {
  display: inline-block;
  padding: 0.125rem 0.5rem;
  background: #000;
  color: #fff;
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
}

.review-notes {
  padding: 0.75rem;
  background: #f9fafb;
  border-left: 4px solid #000;
}

.revision-actions {
  display: flex;
  gap: 1rem;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 2px solid #f3f4f6;
}

.btn-action {
  padding: 0.5rem 1rem;
  border: 2px solid #000;
  background: #fff;
  font-weight: 700;
  font-size: 0.875rem;
  text-decoration: none;
  color: #000;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-action:hover {
  background: #000;
  color: #fff;
}

.btn-revert {
  background: #fef3c7;
}

.btn-revert:hover {
  background: #ca8a04;
  border-color: #ca8a04;
  color: #000;
}

.loading,
.error,
.empty {
  text-align: center;
  padding: 3rem;
  color: #6b7280;
}

.error {
  color: #dc2626;
}
</style>