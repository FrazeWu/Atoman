<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import APageHeader from '@/components/ui/APageHeader.vue'
import ABtn from '@/components/ui/ABtn.vue'
import AInput from '@/components/ui/AInput.vue'
import ATextarea from '@/components/ui/ATextarea.vue'
import ASelect from '@/components/ui/ASelect.vue'
import AConfirm from '@/components/ui/AConfirm.vue'
import type { PodcastEpisode, Channel } from '@/types'

const API_URL = import.meta.env.VITE_API_URL || '/api'
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const isEdit = computed(() => !!route.params.id)
const savingDraft = ref(false)
const publishing = ref(false)
const showPublishConfirm = ref(false)
const draftSaved = ref(false)
const errorMsg = ref('')
const titleError = ref('')
const audioError = ref('')
const channels = ref<Channel[]>([])

// Upload state
const audioUploadProgress = ref(0)   // 0-100, -1 = error
const audioUploading = ref(false)
const coverUploading = ref(false)

const form = ref({
  channel_id: '' as string,
  title: '',
  shownotes: '',
  audio_url: '',
  episode_cover_url: '',
  season_number: 1,
  episode_number: 1,
})

const channelOptions = computed(() => [
  { label: '请选择节目频道', value: '' },
  ...channels.value.map(ch => ({ label: ch.name, value: ch.id })),
])

// ── Upload helpers ─────────────────────────────────────────

function uploadWithProgress(
  url: string,
  formData: FormData,
  onProgress: (pct: number) => void,
): Promise<{ url: string }> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest()
    xhr.open('POST', url)
    xhr.setRequestHeader('Authorization', `Bearer ${authStore.token}`)
    xhr.upload.addEventListener('progress', (e) => {
      if (e.lengthComputable) onProgress(Math.round((e.loaded / e.total) * 100))
    })
    xhr.addEventListener('load', () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve(JSON.parse(xhr.responseText))
      } else {
        try { reject(JSON.parse(xhr.responseText)) } catch { reject({ error: '上传失败' }) }
      }
    })
    xhr.addEventListener('error', () => reject({ error: '网络错误，请重试' }))
    xhr.send(formData)
  })
}

async function onAudioFileChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  audioUploading.value = true
  audioUploadProgress.value = 0
  audioError.value = ''
  try {
    const fd = new FormData()
    fd.append('audio', file)
    const result = await uploadWithProgress(
      `${API_URL}/podcast/upload-audio`,
      fd,
      (pct) => { audioUploadProgress.value = pct },
    )
    form.value.audio_url = result.url
  } catch (err: any) {
    audioUploadProgress.value = -1
    audioError.value = err?.error || '音频上传失败'
  } finally {
    audioUploading.value = false
  }
}

async function onCoverFileChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  coverUploading.value = true
  try {
    const fd = new FormData()
    fd.append('cover', file)
    const res = await fetch(`${API_URL}/podcast/upload-cover`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authStore.token}` },
      body: fd,
    })
    if (!res.ok) throw await res.json()
    const result = await res.json()
    form.value.episode_cover_url = result.url
  } catch (err: any) {
    errorMsg.value = err?.error || '封面上传失败'
  } finally {
    coverUploading.value = false
  }
}

// ── Form logic ─────────────────────────────────────────────

function validate(): boolean {
  titleError.value = form.value.title.trim() ? '' : '请填写单集标题'
  audioError.value = form.value.audio_url.trim() ? '' : '请先上传音频文件'
  return !titleError.value && !audioError.value
}

function buildPayload(status: 'draft' | 'published') {
  return {
    channel_id: form.value.channel_id,
    title: form.value.title.trim(),
    shownotes: form.value.shownotes,
    audio_url: form.value.audio_url.trim(),
    episode_cover_url: form.value.episode_cover_url,
    season_number: form.value.season_number,
    episode_number: form.value.episode_number,
    status,
  }
}

async function apiSave(payload: ReturnType<typeof buildPayload>): Promise<PodcastEpisode> {
  const headers = { 'Content-Type': 'application/json', Authorization: `Bearer ${authStore.token}` }
  if (isEdit.value) {
    const res = await fetch(`${API_URL}/podcast/episodes/${route.params.id}`, {
      method: 'PUT', headers, body: JSON.stringify(payload),
    })
    if (!res.ok) throw await res.json()
    return res.json()
  } else {
    const res = await fetch(`${API_URL}/podcast/episodes`, {
      method: 'POST', headers, body: JSON.stringify(payload),
    })
    if (!res.ok) throw await res.json()
    return res.json()
  }
}

async function loadChannels() {
  if (!authStore.user) return
  const res = await fetch(
    `${API_URL}/blog/channels?user_id=${authStore.user.id}`,
    { headers: { Authorization: `Bearer ${authStore.token}` } },
  )
  if (res.ok) {
    const data = await res.json()
    channels.value = data.data ?? data
    if (!form.value.channel_id && channels.value.length > 0) {
      form.value.channel_id = channels.value[0].id
    }
  }
}

async function loadEpisode() {
  const id = route.params.id as string
  const res = await fetch(`${API_URL}/podcast/episodes/${id}`, {
    headers: { Authorization: `Bearer ${authStore.token}` },
  })
  if (!res.ok) return
  const ep: PodcastEpisode = await res.json()
  form.value = {
    channel_id: ep.channel_id,
    title: ep.post?.title || '',
    shownotes: ep.post?.content || '',
    audio_url: ep.audio_url,
    episode_cover_url: ep.episode_cover_url,
    season_number: ep.season_number,
    episode_number: ep.episode_number,
  }
}

onMounted(async () => {
  await loadChannels()
  if (isEdit.value) await loadEpisode()
})

async function saveDraft() {
  if (!validate()) return
  savingDraft.value = true
  errorMsg.value = ''
  draftSaved.value = false
  try {
    const ep = await apiSave(buildPayload('draft'))
    if (!isEdit.value) router.replace(`/podcast/${ep.id}/edit`)
    draftSaved.value = true
    setTimeout(() => { draftSaved.value = false }, 3000)
  } catch (e: any) {
    errorMsg.value = e?.error || '保存失败，请重试'
  } finally {
    savingDraft.value = false
  }
}

function requestPublish() {
  if (!validate()) return
  showPublishConfirm.value = true
}

async function doPublish() {
  showPublishConfirm.value = false
  publishing.value = true
  errorMsg.value = ''
  try {
    const ep = await apiSave(buildPayload('published'))
    router.push(`/podcast/${isEdit.value ? route.params.id : ep.id}`)
  } catch (e: any) {
    errorMsg.value = e?.error || '发布失败，请重试'
  } finally {
    publishing.value = false
  }
}
</script>

<template>
  <div class="pe-wrap">
    <APageHeader :title="isEdit ? '编辑单集' : '发布新单集'" accent />

    <div class="pe-layout">
      <!-- 左栏 -->
      <div class="pe-main">
        <!-- 基本信息 -->
        <section class="pe-section">
          <h2 class="pe-section-title">基本信息</h2>
          <AInput
            v-model="form.title"
            label="单集标题 *"
            placeholder="为这集播客起一个吸引听众的标题"
            :error="titleError"
            @input="titleError = ''"
          />
          <ATextarea
            v-model="form.shownotes"
            label="节目说明（Shownotes）"
            placeholder="节目内容简介、时间轴、嘉宾介绍、相关链接…"
            :rows="7"
          />
        </section>

        <!-- 单集编号 -->
        <section class="pe-section">
          <h2 class="pe-section-title">单集编号</h2>
          <div class="pe-row">
            <AInput
              v-model.number="form.season_number"
              label="季"
              type="number"
              :min="1"
            />
            <AInput
              v-model.number="form.episode_number"
              label="集"
              type="number"
              :min="1"
            />
          </div>
        </section>

        <!-- 音频文件 -->
        <section class="pe-section">
          <h2 class="pe-section-title">音频文件</h2>

          <!-- 已上传 -->
          <div v-if="form.audio_url && !audioUploading" class="pe-uploaded">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" style="color:var(--a-color-success,#10b981);flex-shrink:0">
              <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
            </svg>
            <span class="pe-uploaded-name">音频已上传</span>
            <button type="button" class="pe-reupload" @click="form.audio_url = ''">重新上传</button>
          </div>

          <!-- 未上传 / 上传中 -->
          <template v-else>
            <label class="pe-drop-zone" :class="{ 'pe-drop-zone--uploading': audioUploading }">
              <input
                type="file"
                accept="audio/mpeg,audio/ogg,audio/wav,audio/aac,audio/x-m4a,.mp3,.ogg,.wav,.aac,.m4a,.flac"
                class="pe-file-hidden"
                :disabled="audioUploading"
                @change="onAudioFileChange"
              />
              <svg v-if="!audioUploading" width="32" height="32" viewBox="0 0 24 24" fill="currentColor" style="opacity:0.35">
                <path d="M12 3v10.55A4 4 0 1014 17V7h4V3h-6z"/>
              </svg>
              <span v-if="!audioUploading" class="pe-drop-hint">点击选择音频文件</span>
              <span v-if="!audioUploading" class="pe-drop-sub">支持 MP3、AAC、M4A、OGG、WAV、FLAC，最大 500 MB</span>
              <span v-if="audioUploading" class="pe-uploading-label">上传中 {{ audioUploadProgress }}%…</span>
            </label>

            <!-- 进度条 -->
            <div v-if="audioUploading" class="pe-progress-track">
              <div class="pe-progress-bar" :style="{ width: audioUploadProgress + '%' }" />
            </div>

            <p v-if="audioError" class="pe-field-error">{{ audioError }}</p>
          </template>

          <!-- 音频预览 -->
          <div v-if="form.audio_url && !audioUploading" class="pe-audio-preview">
            <audio :src="form.audio_url" controls preload="none" style="width:100%" />
          </div>
        </section>
      </div>

      <!-- 右栏：发布面板 -->
      <aside class="pe-panel">
        <!-- 节目频道 -->
        <ASelect
          label="节目频道 *"
          :model-value="form.channel_id"
          :options="channelOptions"
          @update:model-value="form.channel_id = $event as string"
        />

        <!-- 单集封面 -->
        <div class="pe-cover-section">
          <label class="pe-field-label">单集封面（可选）</label>

          <!-- 有封面：预览 + 重新上传 -->
          <div v-if="form.episode_cover_url" class="pe-cover-preview">
            <img :src="form.episode_cover_url" alt="单集封面" class="pe-cover-img" />
            <label class="pe-cover-reupload">
              <input type="file" accept="image/*" class="pe-file-hidden" :disabled="coverUploading" @change="onCoverFileChange" />
              {{ coverUploading ? '上传中…' : '更换封面' }}
            </label>
          </div>

          <!-- 无封面：上传区 -->
          <label v-else class="pe-cover-empty" :class="{ 'pe-cover-empty--uploading': coverUploading }">
            <input type="file" accept="image/*" class="pe-file-hidden" :disabled="coverUploading" @change="onCoverFileChange" />
            <svg v-if="!coverUploading" width="24" height="24" viewBox="0 0 24 24" fill="currentColor" style="opacity:0.3">
              <path d="M19 3H5a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2V5a2 2 0 00-2-2zm-1 14l-5-6.5-4 5-3-3.5-4 4.5h16z"/>
            </svg>
            <span v-if="!coverUploading" class="pe-cover-hint">点击上传封面</span>
            <span v-else class="pe-cover-hint">上传中…</span>
          </label>
          <p class="pe-cover-sub">不填则显示频道封面</p>
        </div>

        <!-- 全局错误 / 成功 -->
        <p v-if="errorMsg" class="pe-error">{{ errorMsg }}</p>
        <p v-if="draftSaved" class="pe-saved">草稿已保存</p>

        <!-- 操作按钮 -->
        <div class="pe-panel-actions">
          <ABtn
            variant="secondary"
            block
            :loading="savingDraft"
            loading-text="保存中…"
            :disabled="publishing || audioUploading"
            @click="saveDraft"
          >
            保存草稿
          </ABtn>
          <ABtn
            variant="primary"
            block
            size="lg"
            :loading="publishing"
            loading-text="发布中…"
            :disabled="savingDraft || audioUploading"
            @click="requestPublish"
          >
            立即发布
          </ABtn>
        </div>
      </aside>
    </div>

    <!-- 发布确认弹窗 -->
    <AConfirm
      :show="showPublishConfirm"
      title="确认发布单集"
      :message="`《${form.title || '未命名单集'}》将立即对听众公开，发布后可继续编辑。`"
      confirm-text="立即发布"
      cancel-text="再想想"
      @confirm="doPublish"
      @cancel="showPublishConfirm = false"
    />
  </div>
</template>

<style scoped>
.pe-wrap {
  max-width: 60rem;
  margin: 0 auto;
  padding: 1.5rem 1.5rem 6rem;
}

.pe-layout {
  display: grid;
  grid-template-columns: 1fr 17rem;
  gap: 1.5rem;
  align-items: start;
  margin-top: 1.5rem;
}

.pe-main { display: flex; flex-direction: column; gap: 1.5rem; }

.pe-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1.25rem 1.5rem;
  background: var(--a-color-surface);
  border: 1px solid var(--a-color-border, #e5e7eb);
  border-radius: 0.75rem;
}

.pe-section-title {
  font-size: 0.8125rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--a-color-muted, #6b7280);
  margin: 0 0 0.25rem 0;
}

.pe-row { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }

.pe-field-label {
  display: block;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--a-color-fg);
  margin-bottom: 0.375rem;
}

/* ── Audio upload area ── */
.pe-uploaded {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.6rem 0.75rem;
  background: var(--a-color-surface);
  border: 1px solid var(--a-color-border, #e5e7eb);
  border-radius: 0.5rem;
  font-size: 0.8rem;
}
.pe-uploaded-name { flex: 1; color: var(--a-color-fg); font-weight: 500; }
.pe-reupload {
  font-size: 0.75rem;
  color: var(--a-color-muted);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  text-decoration: underline;
}
.pe-reupload:hover { color: var(--a-color-fg); }

.pe-drop-zone {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  padding: 2rem 1rem;
  border: 2px dashed var(--a-color-border, #e5e7eb);
  border-radius: 0.625rem;
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
  text-align: center;
}
.pe-drop-zone:hover:not(.pe-drop-zone--uploading) {
  border-color: var(--a-color-accent, #6366f1);
  background: var(--a-color-surface);
}
.pe-drop-zone--uploading { cursor: default; opacity: 0.7; }
.pe-drop-hint { font-size: 0.875rem; font-weight: 500; color: var(--a-color-fg); }
.pe-drop-sub { font-size: 0.75rem; color: var(--a-color-muted); }
.pe-uploading-label { font-size: 0.875rem; font-weight: 600; color: var(--a-color-accent, #6366f1); }

.pe-progress-track {
  height: 4px;
  background: var(--a-color-border, #e5e7eb);
  border-radius: 9999px;
  overflow: hidden;
}
.pe-progress-bar {
  height: 100%;
  background: var(--a-color-accent, #6366f1);
  border-radius: 9999px;
  transition: width 0.2s ease;
}

.pe-field-error { font-size: 0.8rem; color: var(--a-color-danger, #ef4444); margin: 0; }

.pe-audio-preview { margin-top: 0.25rem; }

/* ── Publish panel ── */
.pe-panel {
  position: sticky;
  top: 5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1.25rem;
  background: var(--a-color-surface);
  border: 1px solid var(--a-color-border, #e5e7eb);
  border-radius: 0.75rem;
}

/* Cover */
.pe-cover-section { display: flex; flex-direction: column; gap: 0.375rem; }

.pe-cover-preview {
  position: relative;
  width: 100%;
  aspect-ratio: 1/1;
  border-radius: 0.5rem;
  overflow: hidden;
  background: var(--a-color-border, #f3f4f6);
}
.pe-cover-img { width: 100%; height: 100%; object-fit: cover; display: block; }
.pe-cover-reupload {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0,0,0,0.45);
  color: #fff;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.15s;
}
.pe-cover-preview:hover .pe-cover-reupload { opacity: 1; }

.pe-cover-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.3rem;
  width: 100%;
  aspect-ratio: 1/1;
  border: 2px dashed var(--a-color-border, #e5e7eb);
  border-radius: 0.5rem;
  cursor: pointer;
  transition: border-color 0.15s;
}
.pe-cover-empty:hover:not(.pe-cover-empty--uploading) {
  border-color: var(--a-color-accent, #6366f1);
}
.pe-cover-empty--uploading { cursor: default; opacity: 0.6; }
.pe-cover-hint { font-size: 0.75rem; color: var(--a-color-muted); }
.pe-cover-sub { font-size: 0.7rem; color: var(--a-color-muted); margin: 0; }

.pe-file-hidden {
  position: absolute;
  width: 0;
  height: 0;
  opacity: 0;
  pointer-events: none;
}

.pe-panel-actions { display: flex; flex-direction: column; gap: 0.625rem; }

.pe-error { font-size: 0.8rem; color: var(--a-color-danger, #ef4444); margin: 0; }
.pe-saved { font-size: 0.8rem; color: var(--a-color-success, #10b981); margin: 0; }

@media (max-width: 768px) {
  .pe-layout { grid-template-columns: 1fr; }
  .pe-panel { position: static; order: -1; }
}
</style>
