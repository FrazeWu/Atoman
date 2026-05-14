
<template>
  <div class="music-form-page">
    <div class="music-form-header">
      <h1 class="music-form-title">编辑专辑</h1>
      <p class="music-form-desc">修改专辑信息、封面，以及添加、删除、排序歌曲。</p>
    </div>

    <div v-if="isLoading" class="form-loading">加载中...</div>

    <form v-else class="music-form">
      <!-- Artist + Album row -->
      <div class="music-grid">
        <div class="music-field">
          <label class="music-label">艺术家</label>
          <ArtistSelect v-model="formData.artist" :disabled="isSaving" class="music-control" />
        </div>
        <div class="music-field">
          <label class="music-label">专辑名称</label>
          <input
            type="text"
            required
            class="music-input"
            v-model="formData.album"
          />
        </div>
      </div>

      <!-- Release date -->
      <div class="music-field">
        <label class="music-label">发行日期</label>
        <input
          type="date"
          required
          class="music-input music-date-input"
          v-model="formData.releaseDate"
        />
      </div>

      <!-- Cover upload -->
      <div class="upload-card">
        <label class="music-label">专辑封面</label>
        <input
          type="file"
          ref="coverInput"
          class="hidden"
          accept="image/*"
          @change="handleCoverChange"
        />
        <div
          v-if="!coverPreview"
          @click="triggerCoverInput"
          class="dropzone"
        >
          <p class="dropzone-title">点击上传新封面图片</p>
          <p class="dropzone-desc">不上传将保持原封面或默认为纯黑色</p>
        </div>
        <div v-else class="cover-preview-wrap">
          <img :src="coverPreview" class="cover-preview" alt="封面预览" />
          <button type="button" @click="requestRemoveCover" class="cover-action danger-action">删除</button>
          <button type="button" @click="triggerCoverInput" class="cover-action">更换</button>
        </div>
      </div>

      <!-- Add new audio files -->
      <div class="upload-card">
        <label class="music-label">添加新歌曲 (支持多选)</label>
        <input
          type="file"
          ref="fileInput"
          class="hidden"
          accept="audio/*"
          multiple
          @change="handleFileChange"
        />
        <div @click="triggerFileInput" class="dropzone">
          <p class="dropzone-title">点击选择音频文件</p>
          <p class="dropzone-desc">支持批量添加</p>
        </div>
      </div>

      <!-- Track list -->
      <div class="track-section">
        <div class="track-header">
          <label class="music-label">歌曲列表 (拖拽排序)</label>
          <button
            v-if="tracks.length > 0"
            type="button"
            @click="requestRemoveAllTracks"
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
                type="text"
                v-model="track.title"
                class="track-title-input"
                placeholder="歌曲名称"
              />
              <p class="track-meta">
                {{ track.isExisting ? '现有歌曲' : track.file?.name }}
              </p>
            </div>
            <span v-if="track.isExisting" class="track-pill">已存在</span>
            <span v-else class="track-pill track-pill-new">新增</span>
            <button
              type="button"
              @click="requestRemoveTrack(index)"
              class="track-remove"
            >
              移除
            </button>
          </div>
        </div>
      </div>

      <!-- Progress -->
      <div v-if="isSaving" class="progress-panel">
        <div class="progress-row">
          <span>正在保存: {{ currentTrackIndex }} / {{ totalTracks }}</span>
          <span>{{ Math.round((currentTrackIndex / totalTracks) * 100) }}%</span>
        </div>
        <div class="progress-bar">
          <div
            class="progress-bar-fill"
            :style="{ width: `${(currentTrackIndex / totalTracks) * 100}%` }"
          ></div>
        </div>
        <p class="progress-text">正在处理: {{ tracks[currentTrackIndex - 1]?.title }}...</p>
      </div>

      <!-- Action buttons -->
      <div class="form-actions stacked-actions">
        <button
          type="button"
          @click="handleSubmit"
          class="primary-action"
          :disabled="tracks.length === 0 || isSaving"
          :class="{ 'is-disabled': tracks.length === 0 || isSaving }"
        >
          {{ isSaving ? '正在保存...' : `保存更改 (${tracks.length} 首)` }}
        </button>
        <button
          type="button"
          @click="cancel"
          class="secondary-action"
          :disabled="isSaving"
          :class="{ 'is-disabled': isSaving }"
        >
          取消
        </button>
      </div>

      <p v-if="saveMessage" class="save-message" :class="saveError ? 'save-message-error' : 'save-message-ok'">
        {{ saveMessage }}
      </p>
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

<script setup lang="ts">
import { reactive, ref, onMounted, computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useApi } from '@/composables/useApi';
import { useAuthStore } from '@/stores/auth';
import { usePlayerStore } from '@/stores/player';
import AConfirm from '@/components/ui/AConfirm.vue';
import ArtistSelect from '@/components/ArtistSelect.vue';
import type { Artist } from '@/types';

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

// Support both UUID-based routing (new canonical) and name-based routing (legacy)
const albumIdParam = decodeURIComponent(route.params.albumId as string);
const isUuidRoute = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(albumIdParam);

// Resolved UUID — set during onMounted
const resolvedAlbumUuid = ref<string | null>(isUuidRoute ? albumIdParam : null);

// Legacy name-based fallback via player store
const albumSongsByName = computed(() => {
  if (isUuidRoute) return [];
  return playerStore.songs.filter(
    song => song.album.toLowerCase() === albumIdParam.toLowerCase()
  );
});

const albumUuid = computed(() => resolvedAlbumUuid.value);

const albumSongs = computed(() => {
  const uuid = albumUuid.value;
  if (!uuid) return [];
  return playerStore.songs.filter(song => String(song.album_id) === uuid);
});

const formData = reactive({
  artist: [] as Artist[],
  album: '',
  releaseDate: '',
});

const originalFormData = reactive({
  artist: [] as Artist[],
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

const isLoading = ref(true);
const isSaving = ref(false);
const currentTrackIndex = ref(0);
const totalTracks = ref(0);
const saveMessage = ref('');
const saveError = ref(false);
const showDeleteConfirm = ref(false);
const deleteConfirmTitle = ref('请确认删除');
const deleteConfirmMessage = ref('该操作不可撤销，是否继续？');
let pendingDeleteAction: (() => void) | null = null;

onMounted(async () => {
  if (!authStore.isAuthenticated) {
    router.push('/login');
    return;
  }

  if (isUuidRoute) {
    // UUID-based: load album directly from API
    try {
      const res = await fetch(`${api.url}/albums/${albumIdParam}`, {
        headers: { Authorization: `Bearer ${authStore.token}` },
      });
      if (!res.ok) throw new Error('Album not found');
      const albumData = await res.json();
      const album = albumData;

      // Populate form from API response
      formData.artist = (album.artists || []).map((a: any) => ({ id: a.id, name: a.name }));
      formData.album = album.title || '';
      formData.releaseDate = album.release_date
        ? album.release_date.substring(0, 10)
        : '';

      originalFormData.artist = [...formData.artist];
      originalFormData.album = formData.album;
      originalFormData.releaseDate = formData.releaseDate;

      originalCoverUrl.value = album.cover_url || '';
      coverPreview.value = album.cover_url || '';

      // Load songs for the album
      await playerStore.fetchSongs();
      tracks.value = albumSongs.value.map((song, index) => ({
        id: `existing-${song.id}`,
        title: song.title,
        track_number: song.track_number || index + 1,
        isExisting: true,
        songId: song.id,
      }));
    } catch {
      alert('专辑未找到');
      router.push('/music');
      return;
    }
  } else {
    // Legacy name-based: load from player store
    await playerStore.fetchSongs();

    // Resolve UUID from name
    const matchingSong = albumSongsByName.value[0];
    if (!matchingSong) {
      alert('专辑未找到');
      router.push('/music');
      return;
    }
    resolvedAlbumUuid.value = String(matchingSong.album_id);

    const firstSong = albumSongs.value[0];
    if (!firstSong) {
      alert('专辑未找到');
      router.push('/music');
      return;
    }

    formData.artist = firstSong.artist ? [{ id: (firstSong as any).artist_id || 0, name: firstSong.artist }] : [];
    formData.album = firstSong.album;
    formData.releaseDate = firstSong.release_date;

    originalFormData.artist = [...formData.artist];
    originalFormData.album = firstSong.album;
    originalFormData.releaseDate = firstSong.release_date;

    originalCoverUrl.value = firstSong.cover_url || '';
    coverPreview.value = firstSong.cover_url || '';

    tracks.value = albumSongs.value.map((song, index) => ({
      id: `existing-${song.id}`,
      title: song.title,
      track_number: index + 1,
      isExisting: true,
      songId: song.id,
    }));
  }

  isLoading.value = false;
});

const parseAndSortTracks = (files: FileList) => {
  const newTracks: TrackItem[] = [];
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    const title = file.name.replace(/\.[^/.]+$/, '');
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
    if (numA !== 9999 || numB !== 9999) return numA - numB;
    return a.title.localeCompare(b.title);
  });

  tracks.value = [...tracks.value, ...newTracks];
};

const triggerFileInput = () => fileInput.value?.click();
const triggerCoverInput = () => coverInput.value?.click();

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
    reader.onload = (e) => { coverPreview.value = e.target?.result as string; };
    reader.readAsDataURL(target.files[0]);
  }
};

const removeCover = () => {
  coverFile.value = null;
  coverPreview.value = '';
  if (coverInput.value) coverInput.value.value = '';
};

const removeTrack = (index: number) => tracks.value.splice(index, 1);
const removeAllTracks = () => { tracks.value = []; };

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

const requestRemoveCover = () =>
  requestDeleteAction('删除封面', '确定删除当前封面吗？', removeCover);

const requestRemoveTrack = (index: number) => {
  const track = tracks.value[index];
  if (!track) return;
  const message = track.isExisting
    ? `确定删除歌曲 "${track.title}" 吗？这将从数据库中永久删除该歌曲。`
    : `确定移除新增歌曲 "${track.title}" 吗？`;
  requestDeleteAction('删除歌曲', message, () => removeTrack(index));
};

const requestRemoveAllTracks = () =>
  requestDeleteAction('删除所有歌曲', '确定删除专辑中的所有歌曲吗？', removeAllTracks);

const onDragStart = (index: number) => { draggingIndex.value = index; };
const onDragOver = (event: DragEvent) => event.preventDefault();
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
  saveMessage.value = '';
  saveError.value = false;

  if (tracks.value.length === 0) {
    saveMessage.value = '至少需要保留一首歌曲';
    saveError.value = true;
    return;
  }

  if (!confirm('确定要保存修改吗？保存后将立即生效。')) return;

  const hasMetadataChanges =
    JSON.stringify(formData.artist.map(a => a.id)) !== JSON.stringify(originalFormData.artist.map(a => a.id)) ||
    formData.album !== originalFormData.album ||
    formData.releaseDate !== originalFormData.releaseDate;


  isSaving.value = true;
  totalTracks.value = tracks.value.length;
  currentTrackIndex.value = 0;

  const batchId = crypto.randomUUID();
  let successCount = 0;
  let failCount = 0;

  const firstSong = albumSongs.value[0];
  const albumId = firstSong?.album_id;

  if (!albumId) {
    saveMessage.value = '无法获取专辑 ID，请刷新页面重试';
    saveError.value = true;
    isSaving.value = false;
    return;
  }

  const artistName = formData.artist.map(a => a.name).join(', ');

  // ── 专辑元数据 / 封面更新 ───────────────────────────────────────────────
  if (coverFile.value || hasMetadataChanges) {
    currentTrackIndex.value = 1;
    const albumData = new FormData();
    if (coverFile.value) albumData.append('cover', coverFile.value);
    if (artistName) albumData.append('artist', artistName);
    if (formData.album) albumData.append('title', formData.album);
    if (formData.releaseDate) albumData.append('release_date', formData.releaseDate);

    try {
      const response = await fetch(`${api.url}/albums/${albumId}`, {
        method: 'PUT',
        headers: { 'Authorization': `Bearer ${authStore.token}` },
        body: albumData
      });
      if (response.ok) successCount++;
      else failCount++;
    } catch {
      failCount++;
    }
  }

  // ── 歌曲处理：分别处理新增歌曲和现有歌曲 ────────────────────────────────
  for (let i = 0; i < tracks.value.length; i++) {
    const track = tracks.value[i];
    currentTrackIndex.value = i + 1;

    if (!track.isExisting && track.file) {
      // 新歌曲：POST 上传
      const data = new FormData();
      data.append('title', track.title);
      data.append('artist', artistName);
      data.append('album', formData.album);
      data.append('release_date', formData.releaseDate);
      data.append('track_number', (i + 1).toString());
      data.append('audio', track.file);
      data.append('batch_id', batchId);

      try {
        const response = await fetch(`${api.url}/songs`, {
          method: 'POST',
          headers: { 'Authorization': `Bearer ${authStore.token}` },
          body: data
        });
        if (response.ok) successCount++;
        else failCount++;
      } catch {
        failCount++;
      }
    } else if (track.isExisting && track.songId) {
      // 现有歌曲：仅在标题或顺序有变化时 PUT 更新元数据
      const originalSong = albumSongs.value.find(s => s.id === track.songId);
      const titleChanged = originalSong && originalSong.title !== track.title;
      const orderChanged = originalSong && (i + 1) !== track.track_number;

      if (titleChanged || orderChanged) {
        try {
          const response = await fetch(`${api.url}/songs/${track.songId}`, {
            method: 'PUT',
            headers: {
              'Authorization': `Bearer ${authStore.token}`,
              'Content-Type': 'application/json'
            },
            body: JSON.stringify({ title: track.title, track_number: i + 1 })
          });
          if (response.ok) successCount++;
          else failCount++;
        } catch {
          failCount++;
        }
      } else {
        successCount++; // 无变化，视为成功
      }
    }
  }

  isSaving.value = false;
  currentTrackIndex.value = 0;
  totalTracks.value = 0;

  if (successCount > 0) {
    saveError.value = failCount > 0;
    saveMessage.value = failCount > 0
      ? `提交完成，成功 ${successCount} 项，失败 ${failCount} 项。`
      : `提交完成，成功 ${successCount} 项。`;
    if (failCount === 0) router.push('/music');
  } else {
    saveMessage.value = '提交失败，请重试。';
    saveError.value = true;
  }
};

const cancel = () => {
  if (confirm('确定要取消编辑吗？未保存的更改将丢失。')) {
    router.back();
  }
};
</script>

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
  color: var(--a-color-muted);
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
.music-control { width: 100%; }
.music-label {
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
}
.music-input {
  width: 100%;
  border: var(--a-border);
  background: var(--a-color-bg);
  padding: 1rem;
  font-size: 0.95rem;
  outline: none;
  transition: box-shadow 0.2s;
}
.music-input:focus,
.track-title-input:focus {
  box-shadow: 5px 5px 0 0 rgba(0, 0, 0, 1);
}
.music-date-input { max-width: 24rem; }
.upload-card,
.track-section,
.progress-panel {
  border: var(--a-border);
  background: var(--a-color-bg);
  padding: 1.25rem;
}
.dropzone {
  border: 2px dashed var(--a-color-fg);
  padding: 3rem 1rem;
  text-align: center;
  cursor: pointer;
  transition: background 0.2s;
}
.dropzone:hover { background: var(--a-color-surface); }
.dropzone-title { margin: 0; font-weight: 800; }
.dropzone-desc {
  margin: 0.5rem 0 0;
  color: var(--a-color-muted);
  font-size: 0.875rem;
}
.cover-preview-wrap {
  position: relative;
  display: inline-block;
  border: var(--a-border);
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
  border: var(--a-border);
  background: var(--a-color-bg);
  color: var(--a-color-fg);
  padding: 0.35rem 0.6rem;
  font-size: 0.75rem;
  font-weight: 900;
  cursor: pointer;
}
.cover-action:hover {
  background: var(--a-color-fg);
  color: var(--a-color-bg);
}
.danger-action {
  top: 0.5rem;
  bottom: auto;
}
.danger-action:hover {
  background: var(--a-color-danger);
  border-color: var(--a-color-danger);
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
  border: var(--a-border);
  background: var(--a-color-surface);
  padding: 1rem;
  cursor: move;
}
.track-row:hover { box-shadow: var(--a-shadow-button); }
.track-row.is-dragging { opacity: 0.5; }
.track-index {
  width: 2rem;
  flex-shrink: 0;
  text-align: center;
  color: var(--a-color-muted-soft);
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
  color: var(--a-color-muted);
  font-size: 0.8rem;
}
.track-remove,
.mini-action,
.secondary-action,
.primary-action {
  border: var(--a-border);
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  cursor: pointer;
  transition: all 0.2s;
}
.track-remove,
.mini-action {
  padding: 0.5rem 0.75rem;
  background: var(--a-color-bg);
  color: var(--a-color-fg);
  font-size: 0.75rem;
}
.track-remove:hover,
.mini-action:hover,
.secondary-action:hover {
  background: var(--a-color-fg);
  color: var(--a-color-bg);
}
.progress-bar {
  margin-top: 0.75rem;
  width: 100%;
  height: 0.5rem;
  background: var(--a-color-disabled-border);
}
.progress-bar-fill {
  height: 100%;
  background: var(--a-color-fg);
  transition: width 0.3s;
}
.form-actions {
  display: flex;
  gap: 1rem;
}
.stacked-actions { flex-direction: column; }
.primary-action,
.secondary-action {
  width: 100%;
  padding: 1.1rem 1.5rem;
  font-size: 0.8125rem;
}
.primary-action {
  background: var(--a-color-fg);
  color: var(--a-color-bg);
}
.primary-action:hover {
  background: var(--a-color-bg);
  color: var(--a-color-fg);
}
.secondary-action {
  background: var(--a-color-bg);
  color: var(--a-color-fg);
}
.is-disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.track-pill {
  font-size: 0.75rem;
  font-weight: 900;
  padding: 0.125rem 0.5rem;
  border: var(--a-border);
  flex-shrink: 0;
}
.track-pill-new {
  background: var(--a-color-fg);
  color: var(--a-color-bg);
}
.form-loading {
  text-align: center;
  padding: 5rem 0;
  color: var(--a-color-muted-soft);
  font-weight: 700;
  font-size: 1.125rem;
}
.save-message {
  font-size: 0.875rem;
  font-weight: 700;
}
.save-message-error { color: var(--a-color-danger); }
.save-message-ok { color: var(--a-color-success); }
@media (max-width: 768px) {
  .music-form-page { padding: 3rem 1rem 8rem; }
  .music-grid { grid-template-columns: 1fr; }
  .track-row,
  .track-header,
  .progress-row { align-items: flex-start; flex-direction: column; }
  .cover-preview {
    width: min(100%, 16rem);
    height: auto;
    aspect-ratio: 1;
  }
}
</style>
