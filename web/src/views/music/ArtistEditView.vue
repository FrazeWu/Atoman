<template>
  <div class="artist-edit">
    <div class="page-header">
      <RouterLink :to="`/music/artists/${artistId}`" class="back-link">← 返回</RouterLink>
      <h1 class="page-title">编辑艺术家</h1>
    </div>

    <div v-if="loading" class="loading">加载中...</div>
    <form v-else @submit.prevent="submitEdit" class="edit-form">
      <div class="field">
        <label class="field-label">艺术家名称</label>
        <input v-model="form.name" type="text" required class="field-input" />
      </div>

      <div class="field">
        <label class="field-label">简介</label>
        <textarea v-model="form.bio" rows="4" class="field-textarea" />
      </div>

      <div class="field-row">
        <div class="field">
          <label class="field-label">国籍</label>
          <input v-model="form.nationality" type="text" class="field-input" />
        </div>
        <div class="field">
          <label class="field-label">出生年份</label>
          <input v-model.number="form.birth_year" type="number" class="field-input" />
        </div>
        <div class="field">
          <label class="field-label">逝世年份</label>
          <input v-model.number="form.death_year" type="number" class="field-input" />
        </div>
      </div>

      <div class="field">
        <label class="field-label">成员（逗号分隔）</label>
        <input v-model="form.members" type="text" class="field-input" />
      </div>

      <div class="field">
        <label class="field-label">封面图片 URL</label>
        <input v-model="form.image_url" type="text" class="field-input" />
      </div>

      <!-- Aliases -->
      <div class="field">
        <label class="field-label">别名管理</label>
        <div v-if="aliases.length" class="aliases-list">
          <div v-for="alias in aliases" :key="alias.id" class="alias-item">
            <span class="alias-name">{{ alias.alias }}</span>
            <span v-if="alias.is_main_name" class="alias-main">(主名)</span>
            <button type="button" @click="deleteAlias(alias.id)" class="alias-delete">删除</button>
          </div>
        </div>
        <div class="alias-add">
          <input v-model="newAlias" type="text" placeholder="添加别名..." class="field-input alias-input" />
          <label class="alias-main-label">
            <input type="checkbox" v-model="newAliasIsMain" /> 主名
          </label>
          <button type="button" @click="addAlias" class="btn-add-alias">添加</button>
        </div>
      </div>

      <div class="field">
        <label class="field-label">编辑摘要 <span class="required">*</span></label>
        <input
          v-model="form.edit_summary"
          type="text"
          required
          placeholder="简要说明此次编辑内容..."
          class="field-input"
        />
      </div>

      <div v-if="submitError" class="submit-error">{{ submitError }}</div>

      <div class="form-actions">
        <button type="submit" :disabled="submitting" class="btn-submit">
          {{ submitting ? '提交中...' : '提交编辑' }}
        </button>
        <RouterLink :to="`/music/artists/${artistId}`" class="btn-cancel">取消</RouterLink>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import type { ArtistAlias } from '@/types'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const API_URL = import.meta.env.VITE_API_URL || '/api'

const artistId = route.params.artistId as string

const loading = ref(true)
const submitting = ref(false)
const submitError = ref('')

const form = ref({
  name: '',
  bio: '',
  nationality: '',
  birth_year: 0,
  death_year: 0,
  members: '',
  image_url: '',
  edit_summary: '',
})

const aliases = ref<ArtistAlias[]>([])
const newAlias = ref('')
const newAliasIsMain = ref(false)

const fetchArtist = async () => {
  loading.value = true
  try {
    const res = await fetch(`${API_URL}/artists/${artistId}`, {
      headers: { Authorization: `Bearer ${authStore.token}` },
    })
    const data = await res.json()
    const artist = data.data
    form.value.name = artist.name || ''
    form.value.bio = artist.bio || ''
    form.value.nationality = artist.nationality || ''
    form.value.birth_year = artist.birth_year || 0
    form.value.death_year = artist.death_year || 0
    form.value.members = artist.members || ''
    form.value.image_url = artist.image_url || ''
    aliases.value = artist.aliases || []
  } catch (e) {
    console.error('Failed to load artist:', e)
  } finally {
    loading.value = false
  }
}

const addAlias = async () => {
  if (!newAlias.value.trim()) return
  try {
    const res = await fetch(`${API_URL}/artists/${artistId}/aliases`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`,
      },
      body: JSON.stringify({ alias: newAlias.value.trim(), is_main_name: newAliasIsMain.value }),
    })
    const data = await res.json()
    aliases.value.push(data.data)
    newAlias.value = ''
    newAliasIsMain.value = false
  } catch (e) {
    console.error('Failed to add alias:', e)
  }
}

const deleteAlias = async (aliasId: string) => {
  try {
    await fetch(`${API_URL}/artists/${artistId}/aliases/${aliasId}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${authStore.token}` },
    })
    aliases.value = aliases.value.filter((a) => a.id !== aliasId)
  } catch (e) {
    console.error('Failed to delete alias:', e)
  }
}

const submitEdit = async () => {
  if (!form.value.edit_summary.trim()) {
    submitError.value = '编辑摘要不能为空'
    return
  }
  submitting.value = true
  submitError.value = ''
  try {
    // Use revision-based edit
    const latestRes = await fetch(`${API_URL}/artists/${artistId}/revisions?limit=1`, {
      headers: { Authorization: `Bearer ${authStore.token}` },
    })
    const latestData = await latestRes.json()
    const baseRevision = latestData.data?.[0]?.version_number || 0

    const res = await fetch(`${API_URL}/artists/${artistId}/edit`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`,
      },
      body: JSON.stringify({
        base_revision: baseRevision,
        changes: {
          name: form.value.name,
          bio: form.value.bio,
          nationality: form.value.nationality,
          birth_year: form.value.birth_year,
          death_year: form.value.death_year,
          members: form.value.members,
          image_url: form.value.image_url,
        },
        edit_summary: form.value.edit_summary,
      }),
    })
    if (!res.ok) {
      const err = await res.json()
      throw new Error(err.error || 'Failed to submit edit')
    }
    router.push(`/music/artists/${artistId}`)
  } catch (e: any) {
    submitError.value = e.message || 'Submission failed'
  } finally {
    submitting.value = false
  }
}

onMounted(fetchArtist)
</script>

<style scoped>
.artist-edit {
  max-width: 48rem;
  margin: 0 auto;
  padding: 2rem;
  padding-bottom: 12rem;
}
.page-header {
  margin-bottom: 2rem;
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
.back-link:hover {
  color: #000;
}
.page-title {
  font-size: 2rem;
  font-weight: 900;
  letter-spacing: -0.04em;
  margin: 0;
}
.loading {
  color: #6b7280;
  padding: 2rem;
}
.edit-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}
.field {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}
.field-row {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr;
  gap: 1rem;
}
.field-label {
  font-size: 0.625rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: #6b7280;
}
.required {
  color: #dc2626;
}
.field-input,
.field-textarea {
  border: 2px solid #000;
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  font-weight: 500;
  outline: none;
  font-family: inherit;
  resize: vertical;
  transition: box-shadow 0.15s;
}
.field-input:focus,
.field-textarea:focus {
  box-shadow: 5px 5px 0px 0px rgba(0, 0, 0, 1);
}
.aliases-list {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  margin-bottom: 0.5rem;
}
.alias-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.375rem 0.75rem;
  border: 1px solid #e5e7eb;
  font-size: 0.875rem;
}
.alias-name {
  flex: 1;
  font-weight: 600;
}
.alias-main {
  font-size: 0.75rem;
  color: #6b7280;
}
.alias-delete {
  border: 1px solid #dc2626;
  color: #dc2626;
  padding: 0.125rem 0.5rem;
  font-size: 0.625rem;
  font-weight: 900;
  text-transform: uppercase;
  cursor: pointer;
  background: transparent;
}
.alias-delete:hover {
  background: #dc2626;
  color: #fff;
}
.alias-add {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}
.alias-input {
  flex: 1;
}
.alias-main-label {
  font-size: 0.75rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 0.25rem;
  white-space: nowrap;
}
.btn-add-alias {
  border: 2px solid #000;
  padding: 0.375rem 1rem;
  font-size: 0.625rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  cursor: pointer;
  background: #fff;
  transition: all 0.2s;
  white-space: nowrap;
}
.btn-add-alias:hover {
  background: #000;
  color: #fff;
}
.submit-error {
  color: #dc2626;
  font-size: 0.875rem;
  font-weight: 600;
  padding: 0.75rem;
  border: 2px solid #dc2626;
}
.form-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}
.btn-submit {
  border: 2px solid #000;
  padding: 0.625rem 2rem;
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  cursor: pointer;
  background: #000;
  color: #fff;
  transition: all 0.2s;
}
.btn-submit:hover:not(:disabled) {
  background: #fff;
  color: #000;
}
.btn-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.btn-cancel {
  font-size: 0.75rem;
  font-weight: 700;
  color: #6b7280;
  text-decoration: none;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.btn-cancel:hover {
  color: #000;
}
</style>
