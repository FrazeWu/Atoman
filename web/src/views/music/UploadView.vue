
<script setup lang="ts">
import { computed, reactive, ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useApi } from '@/composables/useApi';
import { useAuthStore } from '@/stores/auth';
import ArtistSelect from '@/components/ArtistSelect.vue';
import type { Artist } from '@/types';
interface TrackFile {
  id: string;
  file: File;
  title: string;
}

const router = useRouter();
const authStore = useAuthStore();
const api = useApi();

const formData = reactive({
  artist: [] as Artist[],
  album: '',
  releaseDate: new Date().toISOString().split('T')[0],
  source: '',
  albumType: 'album' as 'single' | 'ep' | 'album',
  editSummary: '',
});

const tracks = ref<TrackFile[]>([]);
const fileInput = ref<HTMLInputElement | null>(null);
const coverInput = ref<HTMLInputElement | null>(null);
const coverFile = ref<File | null>(null);
const coverPreview = ref<string>('');
const draggingIndex = ref<number | null>(null);

const isUploading = ref(false);
const currentTrackIndex = ref(0);
const totalTracks = ref(0);
const isAdmin = computed(() => authStore.user?.role === 'admin');

onMounted(() => {
  if (!authStore.isAuthenticated) {
    router.push('/login');
  }
});

const parseAndSortTracks = (files: FileList) => {
  const newTracks: TrackFile[] = [];
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    const title = file.name.replace(/\.[^/.]+$/, "");
    newTracks.push({
      id: Math.random().toString(36).substr(2, 9),
      file,
      title
    });
  }

  newTracks.sort((a, b) => {
    const numA = parseInt(a.title.match(/^(\d+)/)?.[0] || '9999');
    const numB = parseInt(b.title.match(/^(\d+)/)?.[0] || '9999');

    if (numA !== 9999 || numB !== 9999) {
      return numA - numB;
    }
    return a.title.localeCompare(b.title);
  });
  
  tracks.value = [...tracks.value, ...newTracks];
  // Auto-suggest album type
  const count = tracks.value.length;
  if (count === 1) formData.albumType = 'single';
  else if (count >= 2 && count <= 6) formData.albumType = 'ep';
  else formData.albumType = 'album';
};

const triggerFileInput = () => {
  fileInput.value?.click();
};

const triggerCoverInput = () => {
  coverInput.value?.click();
};

const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement;
  if (target.files && target.files.length > 0) {
    parseAndSortTracks(target.files);
    target.value = '';
  }
};

const handleCoverChange = (event: Event) => {
  const target = event.target as HTMLInputElement;
  if (target.files && target.files.length > 0) {
    coverFile.value = target.files[0];
    const reader = new FileReader();
    reader.onload = (e) => {
      coverPreview.value = e.target?.result as string;
    };
    reader.readAsDataURL(target.files[0]);
  }
};

const removeCover = () => {
  coverFile.value = null;
  coverPreview.value = '';
  if (coverInput.value) {
    coverInput.value.value = '';
  }
};

const removeArtist = (indexToRemove: number) => {
  formData.artist.splice(indexToRemove, 1);
};

const removeTrack = (index: number) => {
  tracks.value.splice(index, 1);
};

const removeAllTracks = () => {
  if (confirm('确定要删除所有歌曲吗？')) {
    tracks.value = [];
  }
};

const onDragStart = (index: number) => {
  draggingIndex.value = index;
};

const onDragOver = (event: DragEvent) => {
  event.preventDefault();
};

const onDrop = (dropIndex: number) => {
  const dragIndex = draggingIndex.value;
  if (dragIndex !== null && dragIndex !== dropIndex) {
    const itemToMove = tracks.value[dragIndex];
    tracks.value.splice(dragIndex, 1);
    tracks.value.splice(dropIndex, 0, itemToMove);
  }
  draggingIndex.value = null;
};

const handleSubmit = async () => {
  if (tracks.value.length === 0) {
    alert('请至少选择一个音频文件');
    return;
  }
  if (!formData.editSummary.trim()) {
    alert('请填写编辑摘要');
    return;
  }
  
  isUploading.value = true;
  totalTracks.value = tracks.value.length;
  currentTrackIndex.value = 0;

  const batchId = crypto.randomUUID();

  let successCount = 0;
  let failCount = 0;

  for (let i = 0; i < tracks.value.length; i++) {
    const track = tracks.value[i];
    currentTrackIndex.value = i + 1;

    const data = new FormData();
     data.append('title', track.title);
     data.append('artist', formData.artist.map(a => a.name).join(', '));
     data.append('album', formData.album);
     data.append('release_date', formData.releaseDate);
     data.append('source', formData.source);
     data.append('track_number', (i + 1).toString());
     data.append('audio', track.file);
     data.append('batch_id', batchId);
     data.append('album_type', formData.albumType);
    
    if (coverFile.value && i === 0) {
      data.append('cover', coverFile.value);
    }

    try {
      const response = await fetch(`${api.url}/songs`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${authStore.token}`
        },
        body: data
      });

      if (response.ok) {
        successCount++;
      } else {
        failCount++;
        console.error(`Failed to upload ${track.title}`, await response.json());
      }
    } catch (e) {
      failCount++;
      console.error(`Error uploading ${track.title}`, e);
    }
  }
  
   isUploading.value = false;
   currentTrackIndex.value = 0;
   totalTracks.value = 0;

   if (successCount > 0) {
     const message = `上传完成！成功: ${successCount} 首，失败: ${failCount} 首。已直接显示（状态：开放）。`;
     alert(message);

      if (failCount === 0) {
        formData.album = '';
        formData.source = '';
        tracks.value = [];
        coverFile.value = null;
        coverPreview.value = '';
      }
   } else {
     alert('上传失败，请重试。');
   }
};

const handleCreateAlbumOnly = async () => {
  if (!isAdmin.value) {
    alert('只有管理员可以直接创建空专辑');
    return;
  }
  if (!formData.album.trim() || formData.artist.length === 0) {
    alert('请填写专辑名称并选择艺术家');
    return;
  }

  isUploading.value = true;

  const data = new FormData();
  data.append('title', formData.album.trim());
  data.append('artist', formData.artist.map(a => a.name).join(', '));
  data.append('release_date', formData.releaseDate);

  if (coverFile.value) {
    data.append('cover', coverFile.value);
  }

  try {
    const response = await fetch(`${api.url}/albums`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${authStore.token}`
      },
      body: data
    });
    const result = await response.json().catch(() => ({}));

    if (!response.ok) {
      throw new Error(result.error || '创建专辑失败');
    }

    alert('专辑已创建');
    router.push(`/music/albums/${result.id}`);
  } catch (e) {
    alert(e instanceof Error ? e.message : '创建专辑失败');
  } finally {
    isUploading.value = false;
  }
};
</script>

<template>
  <div class="music-form-page">
    <div class="music-form-header">
      <h1 class="music-form-title">贡献新档案</h1>
      <p class="music-form-desc">帮助我们完善 Ye 的音乐史料库。支持批量上传、封面预览和拖拽排序。</p>
    </div>

    <form class="music-form" @submit.prevent="handleSubmit">
      <div class="music-field">
        <label class="music-label">艺术家</label>
        <ArtistSelect
          v-model="formData.artist"
          placeholder="选择艺术家"
          :disabled="isUploading"
          class="music-control"
        />
        <div v-if="formData.artist.length" class="selected-tags">
          <span
            v-for="(artist, idx) in formData.artist"
            :key="artist.id || idx"
            class="selected-tag"
          >
            {{ artist.name }}
            <button
              type="button"
              @click.prevent="removeArtist(idx)"
              class="selected-tag-remove"
            >
              ×
            </button>
          </span>
        </div>
      </div>

      <div class="music-grid">
        <div class="music-field">
          <label class="music-label">专辑名称</label>
          <input
            v-model="formData.album"
            type="text"
            required
            class="music-input"
          />
        </div>

        <div class="music-field">
          <label class="music-label">发行日期</label>
          <input
            v-model="formData.releaseDate"
            type="date"
            required
            class="music-input"
          />
        </div>
      </div>

      <div class="music-field">
        <label class="music-label">信息来源</label>
        <input
          v-model="formData.source"
          type="text"
          placeholder="例如：官方网站、维基百科、音乐平台等"
          class="music-input"
        />
      </div>

      <div class="upload-card">
        <label class="music-label">专辑类型</label>
        <select v-model="formData.albumType" class="music-input">
          <option value="single">Single（单曲）</option>
          <option value="ep">EP</option>
          <option value="album">Album（专辑）</option>
        </select>
        <p class="field-hint">根据曲目数量自动建议，可手动修改</p>
      </div>

      <div class="upload-card">
        <label class="music-label">编辑摘要 <span class="required">*</span></label>
        <input
          v-model="formData.editSummary"
          type="text"
          placeholder="简要描述本次贡献内容（如：添加专辑《...》）"
          class="music-input"
          required
        />
      </div>

      <div class="upload-card">
        <label class="music-label">专辑封面（可选，不上传默认为纯黑）</label>
        <input
          ref="coverInput"
          type="file"
          class="hidden"
          accept="image/*"
          @change="handleCoverChange"
        />
        <div
          v-if="!coverPreview"
          class="dropzone"
          @click="triggerCoverInput"
        >
          <p class="dropzone-title">点击上传封面图片</p>
          <p class="dropzone-desc">不上传将默认为纯黑色</p>
        </div>
        <div v-else class="cover-preview-wrap">
          <img :src="coverPreview" class="cover-preview" alt="封面预览" />
          <button type="button" @click="removeCover" class="cover-action danger-action">删除</button>
          <button type="button" @click="triggerCoverInput" class="cover-action">更换</button>
        </div>
      </div>

      <div class="upload-card">
        <label class="music-label">音频文件（支持多选）</label>
        <input
          ref="fileInput"
          type="file"
          class="hidden"
          accept="audio/*"
          multiple
          @change="handleFileChange"
        />
        <div class="dropzone" @click="triggerFileInput">
          <p class="dropzone-title">点击选择多个音频文件</p>
          <p class="dropzone-desc">支持批量上传</p>
        </div>
      </div>

      <div class="track-section">
        <div class="track-header">
          <label class="music-label">歌曲列表（拖拽排序）</label>
          <button
            v-if="tracks.length > 0"
            type="button"
            @click="removeAllTracks"
            class="mini-action"
          >
            删除所有
          </button>
        </div>

        <div class="track-list">
          <div
            v-for="(track, index) in tracks"
            :key="track.id"
            draggable="true"
            @dragstart="onDragStart(index)"
            @dragover="onDragOver"
            @drop="onDrop(index)"
            class="track-row"
            :class="{ 'is-dragging': draggingIndex === index }"
          >
            <span class="track-index">{{ index + 1 }}</span>
            <div class="track-body">
              <input
                v-model="track.title"
                type="text"
                class="track-title-input"
                placeholder="歌曲名称"
              />
              <p class="track-meta">{{ track.file.name }}</p>
            </div>
            <button
              type="button"
              @click="removeTrack(index)"
              class="track-remove"
            >
              移除
            </button>
          </div>
        </div>
      </div>

      <div v-if="isUploading" class="progress-panel">
        <div class="progress-row">
          <span>正在上传: {{ currentTrackIndex }} / {{ totalTracks }}</span>
          <span>{{ Math.round((currentTrackIndex / totalTracks) * 100) }}%</span>
        </div>
        <div class="progress-bar">
          <div class="progress-bar-fill" :style="{ width: `${(currentTrackIndex / totalTracks) * 100}%` }"></div>
        </div>
        <p class="progress-text">正在处理: {{ tracks[currentTrackIndex - 1]?.title }}...</p>
      </div>

      <div class="form-actions stacked-actions">
        <button
          type="submit"
          class="primary-action"
          :disabled="tracks.length === 0 || isUploading"
          :class="{ 'is-disabled': tracks.length === 0 || isUploading }"
        >
          {{ isUploading ? '正在提交...' : `直接上传 (${tracks.length} 首)` }}
        </button>

        <button
          v-if="isAdmin"
          type="button"
          @click="handleCreateAlbumOnly"
          class="secondary-action"
          :disabled="tracks.length > 0 || isUploading || formData.artist.length === 0 || !formData.album.trim()"
          :class="{ 'is-disabled': tracks.length > 0 || isUploading || formData.artist.length === 0 || !formData.album.trim() }"
        >
          仅创建专辑档案
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.music-form-page {
  max-width: 56rem;
  margin: 0 auto;
  padding: 5rem 2rem 12rem;
}

.music-form-header {
  margin-bottom: 2.5rem;
}

.music-form-title {
  margin: 0 0 0.75rem;
  font-size: 2.75rem;
  font-weight: 900;
  letter-spacing: -0.05em;
}

.music-form-desc {
  margin: 0;
  max-width: 42rem;
  color: #6b7280;
  line-height: 1.7;
}

.music-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.music-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1.5rem;
}

.music-field {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.music-control {
  width: 100%;
}

.music-label {
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.music-input {
  width: 100%;
  border: 2px solid #000;
  background: #fff;
  padding: 1rem;
  font-size: 0.95rem;
  outline: none;
  transition: box-shadow 0.2s;
}

.music-input:focus,
.track-title-input:focus {
  box-shadow: 5px 5px 0 0 rgba(0, 0, 0, 1);
}

.selected-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.selected-tag {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  border: 2px solid #000;
  background: #f3f4f6;
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  font-weight: 800;
}

.selected-tag-remove {
  border: none;
  background: transparent;
  padding: 0;
  color: #dc2626;
  font-size: 1rem;
  font-weight: 900;
  cursor: pointer;
}

.upload-card,
.track-section,
.progress-panel {
  border: 2px solid #000;
  background: #fff;
  padding: 1.25rem;
}

.dropzone {
  border: 2px dashed #000;
  padding: 3rem 1rem;
  text-align: center;
  cursor: pointer;
  transition: background 0.2s;
}

.dropzone:hover {
  background: #f9fafb;
}

.dropzone-title {
  margin: 0;
  font-weight: 800;
}

.dropzone-desc {
  margin: 0.5rem 0 0;
  color: #6b7280;
  font-size: 0.875rem;
}

.cover-preview-wrap {
  position: relative;
  display: inline-block;
  border: 2px solid #000;
}

.cover-preview {
  display: block;
  width: 12rem;
  height: 12rem;
  object-fit: cover;
  filter: grayscale(100%);
}

.cover-action {
  position: absolute;
  right: 0.5rem;
  bottom: 0.5rem;
  border: 2px solid #000;
  background: #fff;
  color: #000;
  padding: 0.35rem 0.6rem;
  font-size: 0.75rem;
  font-weight: 900;
  cursor: pointer;
}

.cover-action:hover {
  background: #000;
  color: #fff;
}

.danger-action {
  top: 0.5rem;
  bottom: auto;
}

.danger-action:hover {
  background: #dc2626;
  border-color: #dc2626;
}

.track-header,
.progress-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.track-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-top: 1rem;
}

.track-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  border: 2px solid #000;
  background: #f9fafb;
  padding: 1rem;
  cursor: move;
}

.track-row:hover {
  box-shadow: 4px 4px 0 0 rgba(0, 0, 0, 1);
}

.track-row.is-dragging {
  opacity: 0.5;
}

.track-index {
  width: 2rem;
  flex-shrink: 0;
  text-align: center;
  color: #9ca3af;
  font-weight: 900;
}

.track-body {
  flex: 1;
  min-width: 0;
}

.track-title-input {
  width: 100%;
  border: 2px solid transparent;
  background: transparent;
  padding: 0.35rem 0.5rem;
  font-weight: 800;
  outline: none;
}

.track-meta,
.progress-text {
  margin: 0.35rem 0 0;
  color: #6b7280;
  font-size: 0.8rem;
}

.track-remove,
.mini-action,
.secondary-action,
.primary-action {
  border: 2px solid #000;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  cursor: pointer;
  transition: all 0.2s;
}

.track-remove,
.mini-action {
  padding: 0.5rem 0.75rem;
  background: #fff;
  color: #000;
  font-size: 0.75rem;
}

.track-remove:hover,
.mini-action:hover,
.secondary-action:hover {
  background: #000;
  color: #fff;
}

.progress-bar {
  margin-top: 0.75rem;
  width: 100%;
  height: 0.5rem;
  background: #e5e7eb;
}

.progress-bar-fill {
  height: 100%;
  background: #000;
  transition: width 0.3s;
}

.form-actions {
  display: flex;
  gap: 1rem;
}

.stacked-actions {
  flex-direction: column;
}

.primary-action,
.secondary-action {
  width: 100%;
  padding: 1.1rem 1.5rem;
  font-size: 0.8125rem;
}

.primary-action {
  background: #000;
  color: #fff;
}

.primary-action:hover {
  background: #fff;
  color: #000;
}

.secondary-action {
  background: #fff;
  color: #000;
}

.is-disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .music-form-page {
    padding: 3rem 1rem 8rem;
  }

  .music-grid {
    grid-template-columns: 1fr;
  }

  .track-row,
  .track-header,
  .progress-row {
    align-items: flex-start;
    flex-direction: column;
  }

  .cover-preview {
    width: min(100%, 16rem);
    height: auto;
    aspect-ratio: 1;
  }
}

.field-hint {
  font-size: 0.75rem;
  color: #9ca3af;
  margin-top: 0.25rem;
}
.required {
  color: #dc2626;
}
</style>
