
<template>
  <div class="max-w-3xl mx-auto px-8 py-20">
    <h1 class="text-4xl font-black tracking-tighter mb-2">编辑专辑</h1>
    <p class="text-gray-500 mb-12">修改专辑信息、封面，以及添加、删除、排序歌曲。</p>

    <div v-if="isLoading" class="text-center py-20 text-gray-400 font-bold text-xl">加载中...</div>

    <form v-else class="space-y-8">
      <!-- Artist + Album row -->
      <div class="grid grid-cols-2 gap-8">
        <div class="flex flex-col gap-3">
          <label class="text-sm font-black uppercase tracking-widest">艺术家</label>
          <ArtistSelect v-model="formData.artist" :disabled="isSaving" />
        </div>
        <div class="flex flex-col gap-3">
          <label class="text-sm font-black uppercase tracking-widest">专辑名称</label>
          <input
            type="text"
            required
            class="w-full bg-white border-2 border-black p-4 outline-none focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] transition-all"
            v-model="formData.album"
          />
        </div>
      </div>

      <!-- Release date -->
      <div class="flex flex-col gap-3">
        <label class="text-sm font-black uppercase tracking-widest">发行日期</label>
        <input
          type="date"
          required
          class="w-full max-w-md bg-white border-2 border-black p-4 outline-none focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] transition-all"
          v-model="formData.releaseDate"
        />
      </div>

      <!-- Cover upload -->
      <div class="flex flex-col gap-3">
        <label class="text-sm font-black uppercase tracking-widest">专辑封面</label>
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
          class="border-2 border-dashed border-black p-12 text-center cursor-pointer hover:bg-gray-100 transition-colors"
        >
          <p class="font-bold">点击上传新封面图片</p>
          <p class="text-xs text-gray-400 mt-2">不上传将保持原封面或默认为纯黑色</p>
        </div>
        <div v-else class="relative border-2 border-black inline-block">
          <img :src="coverPreview" class="w-48 h-48 object-cover grayscale block" alt="封面预览" />
          <button type="button" @click="requestRemoveCover" class="absolute top-2 right-2 bg-black text-white px-3 py-1 text-xs font-bold hover:bg-red-600 transition-colors">删除</button>
          <button type="button" @click="triggerCoverInput" class="absolute bottom-2 right-2 bg-black text-white px-3 py-1 text-xs font-bold hover:bg-gray-700 transition-colors">更换</button>
        </div>
      </div>

      <!-- Add new audio files -->
      <div class="flex flex-col gap-3">
        <label class="text-sm font-black uppercase tracking-widest">添加新歌曲 (支持多选)</label>
        <input
          type="file"
          ref="fileInput"
          class="hidden"
          accept="audio/*"
          multiple
          @change="handleFileChange"
        />
        <div @click="triggerFileInput" class="border-2 border-dashed border-black p-12 text-center cursor-pointer hover:bg-gray-100 transition-colors">
          <p class="font-bold">点击选择音频文件</p>
          <p class="text-xs text-gray-400 mt-2">支持批量添加</p>
        </div>
      </div>

      <!-- Track list -->
      <div class="flex flex-col gap-3">
        <div class="flex justify-between items-center">
          <label class="text-sm font-black uppercase tracking-widest">歌曲列表 (拖拽排序)</label>
          <button
            v-if="tracks.length > 0"
            type="button"
            @click="requestRemoveAllTracks"
            class="px-3 py-1 text-xs font-black border-2 border-black hover:bg-black hover:text-white transition-colors"
          >
            删除所有
          </button>
        </div>
        <div class="border-2 border-black p-4 bg-gray-50 flex flex-col gap-2">
          <div
            v-for="(track, index) in tracks"
            :key="track.id"
            draggable="true"
            @dragstart="onDragStart(index)"
            @dragover="onDragOver"
            @drop="onDrop(index)"
            class="bg-white border-2 border-black p-4 flex items-center gap-4 cursor-move hover:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] transition-shadow"
            :class="{ 'opacity-50': draggingIndex === index }"
          >
            <span class="font-mono text-gray-400 w-8 text-center flex-shrink-0">{{ index + 1 }}</span>
            <div class="flex-1 min-w-0">
              <input
                type="text"
                v-model="track.title"
                class="w-full font-bold outline-none border-b border-transparent focus:border-black transition-colors bg-transparent text-sm"
                placeholder="歌曲名称"
              />
              <p class="text-xs text-gray-400 truncate mt-1">
                {{ track.isExisting ? '现有歌曲' : track.file?.name }}
              </p>
            </div>
            <span v-if="track.isExisting" class="text-xs font-black px-2 py-0.5 border-2 border-black flex-shrink-0">已存在</span>
            <span v-else class="text-xs font-black px-2 py-0.5 border-2 border-black bg-black text-white flex-shrink-0">新增</span>
            <button
              type="button"
              @click="requestRemoveTrack(index)"
              class="text-red-500 font-bold hover:underline text-sm flex-shrink-0"
            >
              移除
            </button>
          </div>
        </div>
      </div>


      <!-- Progress -->
      <div v-if="isSaving" class="space-y-2 pt-4">
        <div class="flex justify-between items-center text-sm font-bold">
          <span>正在保存: {{ currentTrackIndex }} / {{ totalTracks }}</span>
          <span>{{ Math.round((currentTrackIndex / totalTracks) * 100) }}%</span>
        </div>
        <div class="w-full bg-gray-200 h-2">
          <div
            class="bg-black h-2 transition-all duration-300"
            :style="{ width: `${(currentTrackIndex / totalTracks) * 100}%` }"
          ></div>
        </div>
        <p class="text-sm text-gray-500">正在处理: {{ tracks[currentTrackIndex - 1]?.title }}...</p>
      </div>

      <!-- Action buttons -->
      <div class="flex gap-4">
        <button
          type="button"
          @click="handleSubmit"
          class="flex-1 bg-black text-white py-6 font-black uppercase tracking-widest border-2 border-black hover:bg-white hover:text-black transition-all"
          :disabled="tracks.length === 0 || isSaving"
          :class="{ 'opacity-50 cursor-not-allowed': tracks.length === 0 || isSaving }"
        >
          {{ isSaving ? '正在保存...' : `保存更改 (${tracks.length} 首)` }}
        </button>
        <button
          type="button"
          @click="cancel"
          class="flex-1 bg-white text-black py-6 font-black uppercase tracking-widest border-2 border-black hover:bg-black hover:text-white transition-all"
          :disabled="isSaving"
          :class="{ 'opacity-50 cursor-not-allowed': isSaving }"
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

// Get album and artist names from route params
const albumName = decodeURIComponent(route.params.albumId as string);
const artistName = route.params.artistName ? decodeURIComponent(route.params.artistName as string) : null;

// Find album UUID by matching songs
const albumUuid = computed(() => {
  const matchingSong = playerStore.songs.find(song => {
    const albumMatch = song.album.toLowerCase() === albumName.toLowerCase();
    if (artistName) {
      return albumMatch && song.artist.toLowerCase() === artistName.toLowerCase();
    }
    return albumMatch;
  });
  return matchingSong ? String(matchingSong.album_id) : null;
});

const formData = reactive({
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

const originalFormData = reactive({
  artist: [] as Artist[],
  album: '',
  releaseDate: '',
});


const isLoading = ref(true);
const isSaving = ref(false);
const currentTrackIndex = ref(0);
const totalTracks = ref(0);
const showDeleteConfirm = ref(false);
const deleteConfirmTitle = ref('请确认删除');
const deleteConfirmMessage = ref('该操作不可撤销，是否继续？');
let pendingDeleteAction: (() => void) | null = null;

const albumSongs = computed(() => {
  const uuid = albumUuid.value;
  if (!uuid) return [];
  return playerStore.songs.filter(song => String(song.album_id) === uuid);
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

  // Build Artist array from song data
  const artistObj: Artist = { id: (firstSong as any).artist_id || 0, name: firstSong.artist };
  formData.artist = firstSong.artist ? [artistObj] : [];
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
    songId: song.id
  }));

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
  if (tracks.value.length === 0) {
    alert('至少需要保留一首歌曲');
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
    alert('无法获取专辑 ID，请刷新页面重试');
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
    let message = '提交完成！';
    message += `\n成功: ${successCount} 项`;
    if (failCount > 0) message += `\n失败: ${failCount} 项`;
    message += '\n已立即生效。';

    alert(message);
    if (failCount === 0) router.push('/');
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
