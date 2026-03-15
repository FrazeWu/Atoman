
<script setup lang="ts">
import { reactive, ref, onMounted, computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useApi } from '@/composables/useApi';
import { useAuthStore } from '@/stores/auth';
import { usePlayerStore } from '@/stores/player';
import ArtistSelect from '@/components/ArtistSelect.vue';
import AConfirm from '@/components/ui/AConfirm.vue';
interface TrackItem {
  id: string;
  title: string;
  track_number: number;
  isExisting: boolean;
  file?: File;
  songId?: number;
}

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();
const playerStore = usePlayerStore();
const api = useApi();

const singerName = decodeURIComponent(route.params.artist as string).replace(/_/g, ' ');
const albumName = decodeURIComponent(route.params.album as string).replace(/_/g, ' ');

const formData = reactive({
  artist: '',
  album: '',
  releaseDate: '',
});

const tracks = ref<TrackItem[]>([]);
const fileInput = ref<HTMLInputElement | null>(null);
const coverInput = ref<HTMLInputElement | null>(null);
const coverFile = ref<File | null>(null);
const coverPreview = ref<string>('');
const originalCoverUrl = ref<string>('');
const draggingIndex = ref<number | null>(null);

const originalFormData = reactive({
  artist: '',
  album: '',
  releaseDate: '',
});

const reason = ref('');

const isLoading = ref(true);
const isSaving = ref(false);
const currentTrackIndex = ref(0);
const totalTracks = ref(0);
const showDeleteConfirm = ref(false);
const deleteConfirmTitle = ref('请确认删除');
const deleteConfirmMessage = ref('该操作不可撤销，是否继续？');
let pendingDeleteAction: (() => void) | null = null;

const albumSongs = computed(() => {
  return playerStore.songs.filter(song => song.album === albumName && song.artist === singerName);
});

onMounted(async () => {
  if (!authStore.isAuthenticated) {
    router.push('/login');
    return;
  }

  await playerStore.fetchSongs();
  
  if (albumSongs.value.length === 0) {
    alert('专辑未找到');
    router.push('/');
    return;
  }

  const firstSong = albumSongs.value[0];
  formData.artist = firstSong.artist;
  formData.album = firstSong.album;
  formData.releaseDate = firstSong.release_date;

  originalFormData.artist = firstSong.artist;
  originalFormData.album = firstSong.album;
  originalFormData.releaseDate = firstSong.release_date;
  
  originalCoverUrl.value = firstSong.cover_url || '';
  coverPreview.value = firstSong.cover_url || '';

  tracks.value = albumSongs.value.map((song, index) => ({
    id: `existing-${song.id}`,
    title: song.title,
    track_number: index + 1,
    isExisting: true,
    songId: song.id
  }));

  isLoading.value = false;
});

const parseAndSortTracks = (files: FileList) => {
  const newTracks: TrackItem[] = [];
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    const title = file.name.replace(/\.[^/.]+$/, "");
    newTracks.push({
      id: Math.random().toString(36).substr(2, 9),
      title,
      track_number: tracks.value.length + i + 1,
      isExisting: false,
      file
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
  const track = tracks.value[index];
  if (!track) return;
  const message = track.isExisting
    ? `确定删除歌曲 "${track.title}" 吗？这将从数据库中永久删除该歌曲。`
    : `确定移除新增歌曲 "${track.title}" 吗？`;
  requestDeleteAction('删除歌曲', message, () => removeTrack(index));
};

const requestRemoveAllTracks = () => {
  requestDeleteAction('删除所有歌曲', '确定删除专辑中的所有歌曲吗？', removeAllTracks);
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
  console.log('handleSubmit called');
  console.log('tracks:', tracks.value);

  if (tracks.value.length === 0) {
    alert('至少需要保留一首歌曲');
    return;
  }

  const isAdmin = authStore.user?.role === 'admin';
  const confirmMessage = isAdmin
    ? '确定要提交修改吗？管理员提交的内容将立即生效。'
    : '提交修改后将进入审核队列，管理员批准后才会生效。确定要提交吗？';

  if (!confirm(confirmMessage)) {
    console.log('User cancelled');
    return;
  }

  const hasMetadataChanges =
    formData.artist !== originalFormData.artist ||
    formData.album !== originalFormData.album ||
    formData.releaseDate !== originalFormData.releaseDate;

  if (!isAdmin && hasMetadataChanges && !reason.value.trim()) {
    alert('请填写修正原因');
    return;
  }

  console.log('Starting submission...');
  isSaving.value = true;
  totalTracks.value = tracks.value.length;
  currentTrackIndex.value = 0;

  const batchId = crypto.randomUUID();
  let successCount = 0;
  let failCount = 0;
  let duplicateCount = 0;

  const firstSong = albumSongs.value[0];
  const albumId = firstSong?.album_id;

  if (!albumId) {
    alert('无法获取专辑 ID，请刷新页面重试');
    isSaving.value = false;
    return;
  }

  if (!isAdmin) {
    if (hasMetadataChanges || coverFile.value) {
      console.log('Regular user submitting album metadata/cover correction');

      const correctionFormData = new FormData();
      correctionFormData.append('album_id', String(albumId));
      correctionFormData.append('corrected_title', formData.album);
      correctionFormData.append('corrected_release_date', formData.releaseDate);
      correctionFormData.append('reason', reason.value);
      if (coverFile.value) {
        correctionFormData.append('cover', coverFile.value);
      }
      try {
        const response = await fetch(`${api.url}/corrections/album`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${authStore.token}`,
          },
          body: correctionFormData
        });

        if (response.ok) {
          console.log('Album correction submitted successfully');
          successCount++;
        } else {
          const error = await response.json();
          console.error('Failed to submit album correction:', error);
          failCount++;
        }
      } catch (e) {
        console.error('Error submitting album correction:', e);
        failCount++;
      }
    }
  } else {
    // Admin 直接更新专辑信息
    const hasMetadataChanges = 
      formData.artist !== originalFormData.artist ||
      formData.album !== originalFormData.album ||
      formData.releaseDate !== originalFormData.releaseDate;

    if (coverFile.value || hasMetadataChanges) {
      currentTrackIndex.value = 1;

      const albumData = new FormData();

      // 只有当有新封面文件时才添加
      if (coverFile.value) {
        albumData.append('cover', coverFile.value);
      }

      // 如果艺术家/专辑名/发行日期有变化，添加到表单
      if (formData.artist) {
        albumData.append('artist', formData.artist);
      }
      if (formData.album) {
        albumData.append('title', formData.album);
      }
      if (formData.releaseDate) {
        albumData.append('release_date', formData.releaseDate);
      }

      try {
        const response = await fetch(`${api.url}/albums/${albumId}`, {
          method: 'PUT',
          headers: {
            'Authorization': `Bearer ${authStore.token}`
          },
          body: albumData
        });

        if (response.ok) {
          successCount++;
          console.log('Album updated successfully');
        } else {
          const error = await response.json();
          console.error('Failed to update album:', error);
          failCount++;
        }
      } catch (e) {
        console.error('Error updating album:', e);
        failCount++;
      }
    }
  }

  // Original song processing logic (apply to all users, but statuses will differ on backend)
  // 对于现有歌曲，如果没有修改则跳过
  for (let i = 0; i < tracks.value.length; i++) {
    const track = tracks.value[i];

    // 如果是现有歌曲，检查是否有修改
    if (track.isExisting && track.songId) {
      const originalSong = albumSongs.value.find(s => s.id === track.songId);
      // 如果歌曲标题没有变化，跳过
      if (originalSong && originalSong.title === track.title) {
        console.log(`Skipping unchanged track: ${track.title}`);
        successCount++; // Count as success since we're keeping the existing song
        continue;
      }
    }

    currentTrackIndex.value = i + 1;

    console.log(`Processing track ${i + 1}/${tracks.value.length}:`, track.title);

    const data = new FormData();
    data.append('title', track.title);
    data.append('artist', formData.artist);
    data.append('album', formData.album);
    data.append('release_date', formData.releaseDate);
    data.append('track_number', (i + 1).toString());
    data.append('batch_id', batchId);

    // 如果是新歌曲，需要上传音频文件
    if (!track.isExisting && track.file) {
      data.append('audio', track.file);
    } else if (track.isExisting) {
      // 对于现有歌曲，使用原有的音频URL（后端需要支持这个字段）
      const originalSong = albumSongs.value.find(s => s.id === track.songId);
      if (originalSong) {
        data.append('audio_url', originalSong.audio_url);
      }
    }

    try {
      console.log(`Submitting track: ${track.title}`);
      const response = await fetch(`${api.url}/songs`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${authStore.token}`
        },
        body: data
      });

      console.log(`Response for ${track.title}:`, response.status, response.ok);

      if (response.ok) {
        const result = await response.json();
        console.log('Result:', result);
        // 检查是否是已存在的歌曲（通过检查返回的歌曲状态）
        if (result.status && result.status !== 'pending') {
          duplicateCount++;
          console.log('Duplicate song detected');
        } else {
          successCount++;
          console.log('New song added');
        }
      } else {
        failCount++;
        const error = await response.json();
        console.error(`Failed to submit ${track.title}`, error);
      }
    } catch (e) {
      failCount++;
      console.error(`Error submitting ${track.title}`, e);
    }
  }

  isSaving.value = false;
  currentTrackIndex.value = 0;
  totalTracks.value = 0;

  if (successCount > 0 || duplicateCount > 0) {
    let message = '提交完成！';
    if (successCount > 0) message += `\n成功: ${successCount} 项`;
    if (duplicateCount > 0) message += `\n已存在(跳过): ${duplicateCount} 首`;
    if (failCount > 0) message += `\n失败: ${failCount} 项`;

    if (!isAdmin && successCount > 0) message += '\n等待管理员批准后生效。';
    if (isAdmin && successCount > 0) message += '\n内容已立即生效。';

    alert(message);
    if (failCount === 0) {
      router.push('/');
    }
  } else {
    alert('提交失败，请重试。');
  }
};

const cancel = () => {
  if (confirm('确定要取消编辑吗？未保存的更改将丢失。')) {
    router.back();
  }
};
</script>

<template>
  <div class="page-container">
    <h1 class="page-title">编辑专辑</h1>
    <p class="page-desc">修改专辑信息、封面，以及添加、删除、排序歌曲。</p>

    <div v-if="isLoading" class="state-loading">
      <p>加载中...</p>
    </div>

    <form v-else class="form-stack">
      <!-- Artist + Album row -->
      <div class="form-row-2">
        <div class="field">
          <label class="field-label">艺术家</label>
          <ArtistSelect
            v-model="formData.artist"
            placeholder="选择艺术家"
            :disabled="isSaving"
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

      <!-- Cover upload -->
      <div class="field">
        <label class="field-label">专辑封面</label>
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
          <p class="upload-zone-title">点击上传新封面图片</p>
          <p class="upload-zone-hint">不上传将保持原封面或默认为纯黑色</p>
        </div>
        <div v-else class="cover-preview-wrapper">
          <img :src="coverPreview" class="cover-preview-img" alt="封面预览" />
          <button type="button" @click="requestRemoveCover" class="cover-btn-delete">删除</button>
          <button type="button" @click="triggerCoverInput" class="cover-btn-change">更换</button>
        </div>
      </div>

      <!-- Add new audio files -->
      <div class="field">
        <label class="field-label">添加新歌曲 (支持多选)</label>
        <input
          type="file"
          ref="fileInput"
          style="display:none"
          accept="audio/*"
          multiple
          @change="handleFileChange"
        />
        <div @click="triggerFileInput" class="upload-zone">
          <p class="upload-zone-title">点击选择音频文件</p>
          <p class="upload-zone-hint">支持批量添加</p>
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
              <p class="track-filename">
                {{ track.isExisting ? '现有歌曲' : track.file?.name }}
              </p>
            </div>
            <span v-if="track.isExisting" class="track-badge badge-existing">已存在</span>
            <span v-else class="track-badge badge-new">新增</span>
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

      <!-- Reason (non-admin only) -->
      <div class="field" v-if="authStore.user?.role !== 'admin'">
        <label class="field-label">修正原因 (仅在修改专辑信息时需要)</label>
        <textarea
          v-model="reason"
          rows="3"
          class="form-textarea"
          placeholder="请详细说明修正原因"
        ></textarea>
      </div>

      <!-- Progress -->
      <div v-if="isSaving" class="progress-section">
        <div class="progress-header">
          <span>正在保存: {{ currentTrackIndex }} / {{ totalTracks }}</span>
          <span>{{ Math.round((currentTrackIndex / totalTracks) * 100) }}%</span>
        </div>
        <div class="progress-bar-bg">
          <div
            class="progress-bar-fill"
            :style="{ width: `${(currentTrackIndex / totalTracks) * 100}%` }"
          ></div>
        </div>
        <p class="progress-desc">正在处理: {{ tracks[currentTrackIndex - 1]?.title }}...</p>
      </div>

      <!-- Action buttons -->
      <div class="form-actions">
        <button
          type="button"
          @click="handleSubmit"
          class="btn-save"
          :disabled="tracks.length === 0 || isSaving"
          :style="tracks.length === 0 || isSaving ? 'opacity:0.5;cursor:not-allowed' : ''"
        >
          {{ isSaving ? '正在保存...' : `保存更改 (${tracks.length} 首)` }}
        </button>
        <button
          type="button"
          @click="cancel"
          class="btn-cancel"
          :disabled="isSaving"
          :style="isSaving ? 'opacity:0.5;cursor:not-allowed' : ''"
        >
          取消
        </button>
      </div>
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

.state-loading {
  text-align: center;
  padding: 5rem 0;
  font-size: 1.25rem;
  font-weight: 700;
  color: #9ca3af;
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

.form-textarea {
  width: 100%;
  border: 2px solid #000;
  padding: 1rem;
  outline: none;
  transition: all 0.2s;
  resize: none;
  font-size: 1rem;
  box-sizing: border-box;
}
.form-textarea:focus {
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

.cover-btn-delete {
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
.cover-btn-delete:hover { background: #dc2626; }

.cover-btn-change {
  position: absolute;
  bottom: 0.5rem;
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
.cover-btn-change:hover { background: #374151; }

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

.track-badge {
  font-size: 0.75rem;
  font-weight: 700;
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  flex-shrink: 0;
}
.badge-existing { background: #dbeafe; color: #1e40af; }
.badge-new { background: #dcfce7; color: #15803d; }

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

.form-actions {
  display: flex;
  gap: 1rem;
}

.btn-save {
  flex: 1;
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
.btn-save:hover:not(:disabled) {
  background: #fff;
  color: #000;
}

.btn-cancel {
  flex: 1;
  background: #fff;
  color: #000;
  padding: 1.5rem;
  font-weight: 900;
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  border: 2px solid #000;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-cancel:hover:not(:disabled) {
  background: #000;
  color: #fff;
}
</style>
