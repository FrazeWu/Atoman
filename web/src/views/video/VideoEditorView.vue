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
import type { Video, Channel } from '@/types'

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
const urlError = ref('')
const channels = ref<Channel[]>([])

// Upload state
const videoUploadProgress = ref(0)   // 0-100, -1 = error
const videoUploading = ref(false)
const coverUploading = ref(false)

const form = ref({
  channel_id: '' as string,
  title: '',
  description: '',
  storage_type: 'external' as 'local' | 'external',
  video_url: '',
  thumbnail_url: '',
  visibility: 'public' as 'public' | 'followers' | 'private',
  tags: '',
})

const channelOptions = computed(() => [
  { label: '不关联频道', value: '' },
  ...channels.value.map(ch => ({ label: ch.name, value: ch.id })),
])

const storageOptions = [
  { label: '外部链接（YouTube / Bilibili / 其他）', value: 'external' },
  { label: '本地上传（S3/MinIO）', value: 'local' },
]

const visibilityOptions = [
  { label: '公开', value: 'public' },
  { label: '仅关注者', value: 'followers' },
  { label: '私密', value: 'private' },
]

// ── Upload helpers ────────────────────────────────────────

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

async function onVideoFileChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  videoUploading.value = true
  videoUploadProgress.value = 0
  urlError.value = ''
  try {
    const fd = new FormData()
    fd.append('video', file)
    const result = await uploadWithProgress(
      `${API_URL}/videos/upload-video`,
      fd,
      (pct) => { videoUploadProgress.value = pct },
    )
    form.value.video_url = result.url
  } catch (err: any) {
    videoUploadProgress.value = -1
    urlError.value = err?.error || '视频上传失败'
  } finally {
    videoUploading.value = false
  }
}

async function onCoverFileChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  coverUploading.value = true
  try {
    const fd = new FormData()
    fd.append('cover', file)
    const res = await fetch(`${API_URL}/videos/upload-cover`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authStore.token}` },
      body: fd,
    })
    if (!res.ok) throw await res.json()
    const result = await res.json()
    form.value.thumbnail_url = result.url
  } catch (err: any) {
    errorMsg.value = err?.error || '封面上传失败'
  } finally {
    coverUploading.value = false
  }
}

// ── Form logic ────────────────────────────────────────────

function validate(): boolean {
  titleError.value = form.value.title.trim() ? '' : '请填写视频标题'
  urlError.value = form.value.video_url.trim() ? '' : (
    form.value.storage_type === 'local' ? '请先上传视频文件' : '请填写视频链接'
  )
  return !titleError.value && !urlError.value
}

function buildPayload(status: 'draft' | 'published') {
  return {
    channel_id: form.value.channel_id || null,
    title: form.value.title.trim(),
    description: form.value.description,
    storage_type: form.value.storage_type,
    video_url: form.value.video_url.trim(),
    thumbnail_url: form.value.thumbnail_url,
    visibility: form.value.visibility,
    status,
    tags: form.value.tags.split(',').map(t => t.trim()).filter(Boolean),
  }
}

async function apiSave(payload: ReturnType<typeof buildPayload>): Promise<Video> {
  const headers = { 'Content-Type': 'application/json', Authorization: `Bearer ${authStore.token}` }
  if (isEdit.value) {
    const res = await fetch(`${API_URL}/videos/${route.params.id}`, {
      method: 'PUT', headers, body: JSON.stringify(payload),
    })
    if (!res.ok) throw await res.json()
    return res.json()
  } else {
    const res = await fetch(`${API_URL}/videos`, {
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
    { headers: { Authorization: `Bearer ${authStore.token}` } }
  )
  if (res.ok) {
    const data = await res.json()
    channels.value = data.data ?? data
    if (!form.value.channel_id && channels.value.length > 0) {
      form.value.channel_id = channels.value[0].id
    }
  }
}

async function loadVideo() {
  const id = route.params.id as string
  const res = await fetch(`${API_URL}/videos/${id}`, {
    headers: { Authorization: `Bearer ${authStore.token}` },
  })
  if (!res.ok) return
  const v: Video = await res.json()
  form.value = {
    channel_id: v.channel_id ?? '',
    title: v.title,
    description: v.description,
    storage_type: v.storage_type,
    video_url: v.video_url,
    thumbnail_url: v.thumbnail_url,
    visibility: v.visibility,
    tags: v.tags?.map(t => t.name).join(', ') ?? '',
  }
}

onMounted(async () => {
  await loadChannels()
  if (isEdit.value) await loadVideo()
})

async function saveDraft() {
  if (!validate()) return
  savingDraft.value = true
  errorMsg.value = ''
  draftSaved.value = false
  try {
    const v = await apiSave(buildPayload('draft'))
    if (!isEdit.value) router.replace(`/video/${v.id}/edit`)
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
    const v = await apiSave(buildPayload('published'))
    router.push(`/video/${isEdit.value ? route.params.id : v.id}`)
  } catch (e: any) {
    errorMsg.value = e?.error || '发布失败，请重试'
  } finally {
    publishing.value = false
  }
}
</script>

<template>
  <div class="ve-wrap">
    <APageHeader :title="isEdit ? '编辑视频' : '上传视频'" accent />

    <div class="ve-layout">
      <!-- 左栏 -->
      <div class="ve-main">
        <!-- 基本信息 -->
        <section class="ve-section">
          <h2 class="ve-section-title">基本信息</h2>
          <AInput
            v-model="form.title"
            label="标题 *"
            placeholder="为视频起一个吸引人的标题"
            :error="titleError"
            @input="titleError = ''"
          />
          <ATextarea
            v-model="form.description"
            label="简介"
            placeholder="介绍视频内容、背景或注意事项…"
            :rows="5"
          />
          <AInput
            v-model="form.tags"
            label="标签"
            placeholder="music, tutorial, vlog（逗号分隔）"
            hint="多个标签用英文逗号隔开"
          />
        </section>

        <!-- 视频来源 -->
        <section class="ve-section">
          <h2 class="ve-section-title">视频来源</h2>

          <ASelect
            label="来源类型"
            :model-value="form.storage_type"
            :options="storageOptions"
            @update:model-value="form.storage_type = $event as 'local' | 'external'; form.video_url = ''; urlError = ''"
          />

          <!-- 外部链接 -->
          <AInput
            v-if="form.storage_type === 'external'"
            v-model="form.video_url"
            label="视频链接 *"
            placeholder="https://youtube.com/watch?v=..."
            hint="支持 YouTube / Bilibili 自动嵌入"
            :error="urlError"
            @input="urlError = ''"
          />

          <!-- 本地上传 -->
          <div v-else class="ve-upload-area">
            <label class="ve-field-label">视频文件 *</label>

            <!-- 已上传 -->
            <div v-if="form.video_url && !videoUploading" class="ve-uploaded">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" style="color:var(--a-color-success,#10b981);flex-shrink:0">
                <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
              </svg>
              <span class="ve-uploaded-name">视频已上传</span>
              <button type="button" class="ve-reupload" @click="form.video_url = ''">重新上传</button>
            </div>

            <!-- 未上传 / 上传中 -->
            <template v-else>
              <label class="ve-drop-zone" :class="{ 've-drop-zone--uploading': videoUploading }">
                <input
                  type="file"
                  accept="video/mp4,video/webm,video/quicktime,.mp4,.webm,.mov"
                  class="ve-file-hidden"
                  :disabled="videoUploading"
                  @change="onVideoFileChange"
                />
                <svg v-if="!videoUploading" width="32" height="32" viewBox="0 0 24 24" fill="currentColor" style="opacity:0.35">
                  <path d="M9 16h6v-6h4l-7-7-7 7h4zm-4 2h14v2H5z"/>
                </svg>
                <span v-if="!videoUploading" class="ve-drop-hint">点击选择或拖拽视频文件</span>
                <span v-if="!videoUploading" class="ve-drop-sub">支持 MP4、WebM、MOV，最大 2 GB</span>
                <span v-if="videoUploading" class="ve-uploading-label">上传中 {{ videoUploadProgress }}%…</span>
              </label>

              <!-- 进度条 -->
              <div v-if="videoUploading" class="ve-progress-track">
                <div class="ve-progress-bar" :style="{ width: videoUploadProgress + '%' }" />
              </div>

              <p v-if="urlError" class="ve-field-error">{{ urlError }}</p>
            </template>
          </div>
        </section>

        <!-- 频道 -->
        <section class="ve-section">
          <h2 class="ve-section-title">关联频道</h2>
          <ASelect
            label="频道"
            :model-value="form.channel_id"
            :options="channelOptions"
            @update:model-value="form.channel_id = $event as string"
          />
        </section>
      </div>

      <!-- 右栏：发布面板 -->
      <aside class="ve-panel">
        <!-- 封面上传 -->
        <div class="ve-cover-section">
          <label class="ve-field-label">封面图</label>

          <!-- 有封面：预览 + 重新上传 -->
          <div v-if="form.thumbnail_url" class="ve-cover-preview">
            <img :src="form.thumbnail_url" alt="封面" class="ve-cover-img" />
            <label class="ve-cover-reupload">
              <input type="file" accept="image/*" class="ve-file-hidden" :disabled="coverUploading" @change="onCoverFileChange" />
              {{ coverUploading ? '上传中…' : '更换封面' }}
            </label>
          </div>

          <!-- 无封面：上传区 -->
          <label v-else class="ve-cover-empty" :class="{ 've-cover-empty--uploading': coverUploading }">
            <input type="file" accept="image/*" class="ve-file-hidden" :disabled="coverUploading" @change="onCoverFileChange" />
            <svg v-if="!coverUploading" width="24" height="24" viewBox="0 0 24 24" fill="currentColor" style="opacity:0.3">
              <path d="M19 3H5a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2V5a2 2 0 00-2-2zm-1 14l-5-6.5-4 5-3-3.5-4 4.5h16z"/>
            </svg>
            <span v-if="!coverUploading" class="ve-cover-hint">点击上传封面</span>
            <span v-else class="ve-cover-hint">上传中…</span>
          </label>
        </div>

        <!-- 可见范围 -->
        <ASelect
          label="可见范围"
          :model-value="form.visibility"
          :options="visibilityOptions"
          @update:model-value="form.visibility = $event as 'public' | 'followers' | 'private'"
        />

        <!-- 全局错误 / 成功 -->
        <p v-if="errorMsg" class="ve-error">{{ errorMsg }}</p>
        <p v-if="draftSaved" class="ve-saved">草稿已保存</p>

        <!-- 操作按钮 -->
        <div class="ve-panel-actions">
          <ABtn
            variant="secondary"
            block
            :loading="savingDraft"
            loading-text="保存中…"
            :disabled="publishing || videoUploading"
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
            :disabled="savingDraft || videoUploading"
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
      title="确认发布视频"
      :message="`《${form.title || '未命名视频'}》将对${
        form.visibility === 'public' ? '所有人' :
        form.visibility === 'followers' ? '关注者' : '仅自己'
      }可见，发布后观众可立即观看。`"
      confirm-text="立即发布"
      cancel-text="再想想"
      @confirm="doPublish"
      @cancel="showPublishConfirm = false"
    />
  </div>
</template>

<style scoped>
.ve-wrap {
  max-width: 64rem;
  margin: 0 auto;
  padding: 1.5rem 1.5rem 6rem;
}

.ve-layout {
  display: grid;
  grid-template-columns: 1fr 18rem;
  gap: 1.5rem;
  align-items: start;
  margin-top: 1.5rem;
}

.ve-main { display: flex; flex-direction: column; gap: 1.5rem; }

.ve-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1.25rem 1.5rem;
  background: var(--a-color-surface);
  border: 1px solid var(--a-color-border, #e5e7eb);
  border-radius: 0.75rem;
}

.ve-section-title {
  font-size: 0.8125rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--a-color-muted, #6b7280);
  margin: 0 0 0.25rem 0;
}

/* Field label (matches AInput label style) */
.ve-field-label {
  display: block;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--a-color-fg);
  margin-bottom: 0.375rem;
}

/* ── Video upload area ── */
.ve-upload-area { display: flex; flex-direction: column; gap: 0.5rem; }

.ve-uploaded {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.6rem 0.75rem;
  background: var(--a-color-surface);
  border: 1px solid var(--a-color-border, #e5e7eb);
  border-radius: 0.5rem;
  font-size: 0.8rem;
}
.ve-uploaded-name { flex: 1; color: var(--a-color-fg); font-weight: 500; }
.ve-reupload {
  font-size: 0.75rem;
  color: var(--a-color-muted);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  text-decoration: underline;
}
.ve-reupload:hover { color: var(--a-color-fg); }

.ve-drop-zone {
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
.ve-drop-zone:hover:not(.ve-drop-zone--uploading) {
  border-color: var(--a-color-accent, #6366f1);
  background: var(--a-color-surface);
}
.ve-drop-zone--uploading { cursor: default; opacity: 0.7; }
.ve-drop-hint { font-size: 0.875rem; font-weight: 500; color: var(--a-color-fg); }
.ve-drop-sub { font-size: 0.75rem; color: var(--a-color-muted); }
.ve-uploading-label { font-size: 0.875rem; font-weight: 600; color: var(--a-color-accent, #6366f1); }

.ve-progress-track {
  height: 4px;
  background: var(--a-color-border, #e5e7eb);
  border-radius: 9999px;
  overflow: hidden;
}
.ve-progress-bar {
  height: 100%;
  background: var(--a-color-accent, #6366f1);
  border-radius: 9999px;
  transition: width 0.2s ease;
}

.ve-field-error { font-size: 0.8rem; color: var(--a-color-danger, #ef4444); margin: 0; }

/* ── Publish panel ── */
.ve-panel {
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
.ve-cover-section { display: flex; flex-direction: column; gap: 0.375rem; }

.ve-cover-preview {
  position: relative;
  width: 100%;
  aspect-ratio: 16/9;
  border-radius: 0.5rem;
  overflow: hidden;
  background: var(--a-color-border, #f3f4f6);
}
.ve-cover-img { width: 100%; height: 100%; object-fit: cover; display: block; }
.ve-cover-reupload {
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
.ve-cover-preview:hover .ve-cover-reupload { opacity: 1; }

.ve-cover-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.3rem;
  width: 100%;
  aspect-ratio: 16/9;
  border: 2px dashed var(--a-color-border, #e5e7eb);
  border-radius: 0.5rem;
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
}
.ve-cover-empty:hover:not(.ve-cover-empty--uploading) {
  border-color: var(--a-color-accent, #6366f1);
}
.ve-cover-empty--uploading { cursor: default; opacity: 0.6; }
.ve-cover-hint { font-size: 0.75rem; color: var(--a-color-muted); }

.ve-file-hidden {
  position: absolute;
  width: 0;
  height: 0;
  opacity: 0;
  pointer-events: none;
}

.ve-panel-actions { display: flex; flex-direction: column; gap: 0.625rem; }

.ve-error { font-size: 0.8rem; color: var(--a-color-danger, #ef4444); margin: 0; }
.ve-saved { font-size: 0.8rem; color: var(--a-color-success, #10b981); margin: 0; }

@media (max-width: 768px) {
  .ve-layout { grid-template-columns: 1fr; }
  .ve-panel { position: static; order: -1; }
}
</style>
