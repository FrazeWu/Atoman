
<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useApi } from '@/composables/useApi';
import { useAuthStore } from '@/stores/auth';
import ArtistSelect from '@/components/ArtistSelect.vue';
import AConfirm from '@/components/ui/AConfirm.vue';
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
const showDeleteConfirm = ref(false);
const deleteConfirmTitle = ref('请确认删除');
const deleteConfirmMessage = ref('该操作不可撤销，是否继续？');
let pendingDeleteAction: (() => void) | null = null;

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

const removeTrack = (index: number) => {
  tracks.value.splice(index, 1);
};

const removeAllTracks = () => {
  tracks.value = [];
};

const requestDeleteAction = (title: string, message: string, action: () => void) => {
  deleteConfirmTitle.value = title;
  deleteConfirmMessage.value = message;
  pendingDeleteAction = action;
  showDeleteConfirm.value = true;
};

const cancelDeleteAction = () => {
  showDeleteConfirm.value = false;
  pendingDeleteAction = null;
};

const confirmDeleteAction = () => {
  const action = pendingDeleteAction;
  cancelDeleteAction();
  if (action) action();
};

const requestRemoveCover = () => {
  requestDeleteAction('删除封面', '确定删除当前封面吗？', removeCover);
};

const requestRemoveTrack = (index: number) => {
  requestDeleteAction('移除歌曲', '确定移除这首歌曲吗？', () => removeTrack(index));
};

const requestRemoveAllTracks = () => {
  requestDeleteAction('删除所有歌曲', '确定删除列表中的所有歌曲吗？', removeAllTracks);
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
     const isAdmin = authStore.user?.role === 'admin';
     const message = isAdmin
       ? `上传完成！成功: ${successCount} 首，失败: ${failCount} 首。内容已立即生效。`
       : `上传完成！成功: ${successCount} 首，失败: ${failCount} 首。等待管理员审核。`;
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
</script>

<template>
  <div class="page-container">
    <h1 class="page-title">贡献新档案</h1>
    <p class="page-desc">帮助我们完善 Ye 的音乐史料库。支持批量上传和拖拽排序。</p>

    <form class="form-stack">
      <!-- Artist + Album row -->
      <div class="form-row-2">
        <div class="field">
          <label class="field-label">艺术家</label>
          <ArtistSelect
            v-model="formData.artist"
            placeholder="选择艺术家"
            :disabled="isUploading"
          />
        </div>
        <div class="field">
          <label class="field-label">专辑名称</label>
          <input
            type="text"
            required
            class="form-input"
            v-model="formData.album"
          />
        </div>
      </div>

      <!-- Release date -->
      <div class="field">
        <label class="field-label">发行日期</label>
        <input
          type="date"
          required
          class="form-input"
          v-model="formData.releaseDate"
        />
      </div>

      <!-- Source -->
      <div class="field">
        <label class="field-label">信息来源</label>
        <input
          type="text"
          placeholder="例如：官方网站、维基百科、音乐平台等"
          class="form-input"
          v-model="formData.source"
        />
      </div>

      <!-- Cover upload -->
      <div class="field">
        <label class="field-label">专辑封面（可选，不上传默认为纯黑）</label>
        <input
          type="file"
          ref="coverInput"
          style="display:none"
          accept="image/*"
          @change="handleCoverChange"
        />
        <div
          v-if="!coverPreview"
          @click="triggerCoverInput"
          class="upload-zone"
        >
          <p class="upload-zone-title">点击上传封面图片</p>
          <p class="upload-zone-hint">不上传将默认为纯黑色</p>
        </div>
        <div v-else class="cover-preview-wrapper">
          <img :src="coverPreview" class="cover-preview-img" alt="封面预览" />
          <button
            type="button"
            @click="requestRemoveCover"
            class="cover-remove-btn"
          >
            删除
          </button>
        </div>
      </div>

      <!-- Audio files upload -->
      <div class="field">
        <label class="field-label">音频文件 (支持多选)</label>
        <input
          type="file"
          ref="fileInput"
          style="display:none"
          accept="audio/*"
          multiple
          @change="handleFileChange"
        />
        <div @click="triggerFileInput" class="upload-zone">
          <p class="upload-zone-title">点击选择多个音频文件</p>
          <p class="upload-zone-hint">支持批量上传</p>
        </div>
      </div>

      <!-- Track list -->
      <div class="field">
        <div class="tracklist-header">
          <label class="field-label" style="margin-bottom:0">歌曲列表 (拖拽排序)</label>
          <button
            v-if="tracks.length > 0"
            type="button"
            @click="requestRemoveAllTracks"
            class="delete-all-btn"
          >
            删除所有
          </button>
        </div>
        <div class="tracklist">
          <div
            v-for="(track, index) in tracks"
            :key="track.id"
            draggable="true"
            @dragstart="onDragStart(index)"
            @dragover="onDragOver"
            @drop="onDrop(index)"
            class="track-item"
            :style="draggingIndex === index ? 'opacity:0.5' : ''"
          >
            <span class="track-num">{{ index + 1 }}</span>
            <div class="track-info">
              <input
                type="text"
                v-model="track.title"
                class="track-title-input"
                placeholder="歌曲名称"
              />
              <p class="track-filename">{{ track.file.name }}</p>
            </div>
            <button
              type="button"
              @click="requestRemoveTrack(index)"
              class="track-remove-btn"
            >
              移除
            </button>
          </div>
        </div>
      </div>

      <!-- Progress -->
      <div v-if="isUploading" class="progress-section">
        <div class="progress-header">
          <span>正在上传: {{ currentTrackIndex }} / {{ totalTracks }}</span>
          <span>{{ Math.round((currentTrackIndex / totalTracks) * 100) }}%</span>
        </div>
        <div class="progress-bar-bg">
          <div
            class="progress-bar-fill"
            :style="{ width: `${(currentTrackIndex / totalTracks) * 100}%` }"
          ></div>
        </div>
        <p class="progress-desc">
          正在处理: {{ tracks[currentTrackIndex - 1]?.title }}...
        </p>
      </div>

      <!-- Submit -->
      <button
        type="button"
        @click="handleSubmit"
        class="submit-btn"
        :disabled="tracks.length === 0 || isUploading"
        :style="tracks.length === 0 || isUploading ? 'opacity:0.5;cursor:not-allowed' : ''"
      >
        {{ isUploading ? '正在提交...' : `提交至审核队列 (${tracks.length} 首)` }}
      </button>
    </form>

    <AConfirm
      :show="showDeleteConfirm"
      :title="deleteConfirmTitle"
      :message="deleteConfirmMessage"
      confirm-text="删除"
      cancel-text="取消"
      danger
      @confirm="confirmDeleteAction"
      @cancel="cancelDeleteAction"
    />
  </div>
</template>

<style scoped>
.page-container {
  max-width: 768px;
  margin: 0 auto;
  padding: 5rem 2rem;
}

.page-title {
  font-size: 2.5rem;
  font-weight: 900;
  letter-spacing: -0.05em;
  margin: 0 0 0.5rem 0;
}

.page-desc {
  color: #6b7280;
  margin-bottom: 3rem;
}

.form-stack {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.form-row-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.field-label {
  display: block;
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.form-input {
  width: 100%;
  background: #fff;
  border: 2px solid #000;
  padding: 1rem;
  outline: none;
  transition: all 0.2s;
  font-size: 1rem;
  box-sizing: border-box;
}
.form-input:focus {
  box-shadow: 5px 5px 0px 0px rgba(0,0,0,1);
}

.upload-zone {
  border: 2px dashed #000;
  padding: 3rem;
  text-align: center;
  cursor: pointer;
  transition: background 0.2s;
}
.upload-zone:hover { background: #f3f4f6; }

.upload-zone-title {
  font-weight: 700;
  margin: 0 0 0.5rem 0;
}

.upload-zone-hint {
  font-size: 0.75rem;
  color: #9ca3af;
  margin: 0;
}

.cover-preview-wrapper {
  position: relative;
  border: 2px solid #000;
  display: inline-block;
}

.cover-preview-img {
  width: 12rem;
  height: 12rem;
  object-fit: cover;
  display: block;
}

.cover-remove-btn {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  background: #000;
  color: #fff;
  padding: 0.25rem 0.75rem;
  font-size: 0.75rem;
  font-weight: 700;
  border: none;
  cursor: pointer;
  transition: background 0.2s;
}
.cover-remove-btn:hover { background: #dc2626; }

.tracklist-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.delete-all-btn {
  padding: 0.25rem 0.75rem;
  font-size: 0.75rem;
  border: 1px solid #ef4444;
  color: #ef4444;
  background: transparent;
  cursor: pointer;
  transition: all 0.2s;
}
.delete-all-btn:hover {
  background: #ef4444;
  color: #fff;
}

.tracklist {
  border: 2px solid #000;
  padding: 1rem;
  background: #f9fafb;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.track-item {
  background: #fff;
  border: 1px solid #d1d5db;
  padding: 1rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  cursor: move;
  transition: box-shadow 0.2s;
}
.track-item:hover { box-shadow: 0 2px 8px rgba(0,0,0,0.1); }

.track-num {
  font-family: monospace;
  color: #9ca3af;
  width: 2rem;
  text-align: center;
  flex-shrink: 0;
}

.track-info {
  flex: 1;
  min-width: 0;
}

.track-title-input {
  width: 100%;
  font-weight: 700;
  outline: none;
  border: none;
  border-bottom: 1px solid transparent;
  transition: border-color 0.2s;
  background: transparent;
  font-size: 0.875rem;
}
.track-title-input:focus {
  border-bottom-color: #000;
}

.track-filename {
  font-size: 0.75rem;
  color: #9ca3af;
  margin: 0.25rem 0 0 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.track-remove-btn {
  color: #ef4444;
  font-weight: 700;
  font-size: 0.875rem;
  background: transparent;
  border: none;
  cursor: pointer;
  flex-shrink: 0;
}
.track-remove-btn:hover { text-decoration: underline; }

.progress-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding-top: 1rem;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.875rem;
  font-weight: 700;
}

.progress-bar-bg {
  width: 100%;
  background: #e5e7eb;
  height: 0.5rem;
}

.progress-bar-fill {
  background: #000;
  height: 0.5rem;
  transition: width 0.3s;
}

.progress-desc {
  font-size: 0.875rem;
  color: #6b7280;
  margin: 0;
}

.submit-btn {
  width: 100%;
  background: #000;
  color: #fff;
  padding: 1.5rem;
  font-weight: 900;
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  border: 2px solid #000;
  cursor: pointer;
  transition: all 0.2s;
}
.submit-btn:hover:not(:disabled) {
  background: #fff;
  color: #000;
}
</style>
