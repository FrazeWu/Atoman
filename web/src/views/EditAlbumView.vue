
<script setup lang="ts">
import { reactive, ref, onMounted, computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useApi } from '@/composables/useApi';
import { useAuthStore } from '@/stores/auth';
import { usePlayerStore } from '@/stores/player.ts';
import ArtistSelect from '@/components/ArtistSelect.vue';
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
  const track = tracks.value[index];
  if (track.isExisting) {
    if (confirm(`确定要删除歌曲 "${track.title}" 吗？这将从数据库中永久删除该歌曲。`)) {
      tracks.value.splice(index, 1);
    }
  } else {
    tracks.value.splice(index, 1);
  }
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
  <div class="max-w-3xl mx-auto px-8 py-20">
    <h1 class="text-4xl font-black tracking-tighter mb-8">编辑专辑</h1>
    <p class="text-gray-500 mb-12">
      修改专辑信息、封面，以及添加、删除、排序歌曲。
    </p>

    <div v-if="isLoading" class="text-center py-20">
      <p class="text-xl font-bold text-gray-400">加载中...</p>
    </div>

    <form v-else class="space-y-8">
      <div class="grid grid-cols-2 gap-8">
        <div class="space-y-4">
          <label class="block text-sm font-black uppercase tracking-widest">艺术家</label>
          <ArtistSelect
            v-model="formData.artist"
            placeholder="选择艺术家"
            :disabled="isSaving"
          />
        </div>
        <div class="space-y-4">
          <label class="block text-sm font-black uppercase tracking-widest">专辑名称</label>
          <input 
            type="text" 
            required
            class="w-full bg-white border-2 border-black p-4 focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] outline-none transition-all"
            v-model="formData.album"
          />
        </div>
      </div>

       <div class="space-y-4">
         <label class="block text-sm font-black uppercase tracking-widest">发行日期</label>
         <input
           type="date"
           required
           class="w-full bg-white border-2 border-black p-4 focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] outline-none transition-all"
           v-model="formData.releaseDate"
         />
       </div>

      <div class="space-y-4">
        <label class="block text-sm font-black uppercase tracking-widest">专辑封面</label>
        <input 
          type="file" 
          ref="coverInput" 
          class="hidden" 
          accept="image/*"
          @change="handleCoverChange"
        />
        <div v-if="!coverPreview" 
          @click="triggerCoverInput"
          class="border-2 border-dashed border-black p-12 text-center cursor-pointer hover:bg-gray-100 transition-colors"
        >
          <p class="font-bold">点击上传新封面图片</p>
          <p class="text-xs text-gray-400 mt-2">不上传将保持原封面或默认为纯黑色</p>
        </div>
        <div v-else class="relative border-2 border-black inline-block">
          <img :src="coverPreview" class="w-48 h-48 object-cover" alt="封面预览" />
          <button 
            type="button"
            @click="removeCover"
            class="absolute top-2 right-2 bg-black text-white px-3 py-1 text-xs font-bold hover:bg-red-600"
          >
            删除
          </button>
          <button 
            type="button"
            @click="triggerCoverInput"
            class="absolute bottom-2 right-2 bg-black text-white px-3 py-1 text-xs font-bold hover:bg-gray-700"
          >
            更换
          </button>
        </div>
      </div>

      <div class="space-y-4">
        <label class="block text-sm font-black uppercase tracking-widest">添加新歌曲 (支持多选)</label>
        <input 
          type="file" 
          ref="fileInput" 
          class="hidden" 
          accept="audio/*"
          multiple
          @change="handleFileChange"
        />
        <div 
          @click="triggerFileInput"
          class="border-2 border-dashed border-black p-12 text-center cursor-pointer hover:bg-gray-100 transition-colors"
        >
          <p class="font-bold">点击选择音频文件</p>
          <p class="text-xs text-gray-400 mt-2">支持批量添加</p>
        </div>
      </div>

       <!-- Track List -->
       <div class="space-y-4">
         <div class="flex justify-between items-center">
           <label class="block text-sm font-black uppercase tracking-widest">歌曲列表 (拖拽排序)</label>
           <div class="flex gap-2">
             <button
               v-if="tracks.length > 0"
               type="button"
               @click="removeAllTracks"
               class="px-3 py-1 text-xs border border-red-500 text-red-500 hover:bg-red-500 hover:text-white transition-colors"
             >
               删除所有
             </button>
           </div>
         </div>
        <div class="space-y-2 border-2 border-black p-4 bg-gray-50">
          <div 
            v-for="(track, index) in tracks" 
            :key="track.id"
            draggable="true"
            @dragstart="onDragStart(index)"
            @dragover="onDragOver"
            @drop="onDrop(index)"
            class="bg-white border border-gray-300 p-4 flex items-center gap-4 cursor-move hover:shadow-md transition-shadow"
            :class="{ 'opacity-50': draggingIndex === index }"
          >
            <span class="font-mono text-gray-400 w-8 text-center">{{ index + 1 }}</span>
            <div class="flex-1">
              <input 
                type="text" 
                v-model="track.title"
                class="w-full font-bold outline-none border-b border-transparent focus:border-black transition-colors"
                placeholder="歌曲名称"
              />
              <p class="text-xs text-gray-400 truncate">
                {{ track.isExisting ? '现有歌曲' : track.file?.name }}
              </p>
            </div>
            <span v-if="track.isExisting" class="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded font-bold">
              已存在
            </span>
            <span v-else class="text-xs bg-green-100 text-green-800 px-2 py-1 rounded font-bold">
              新增
            </span>
            <button 
              type="button" 
              @click="removeTrack(index)"
              class="text-red-500 font-bold hover:underline text-sm"
            >
              移除
            </button>
          </div>
        </div>
      </div>

       <div class="space-y-4" v-if="authStore.user?.role !== 'admin'">
         <label class="block text-sm font-black uppercase tracking-widest">修正原因 (仅在修改专辑信息时需要)</label>
         <textarea v-model="reason" rows="3"
           class="w-full border-2 border-black p-4 focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] outline-none transition-all resize-none"
           placeholder="请详细说明修正原因"></textarea>
       </div>

       <!-- Progress Display -->
      <div v-if="isSaving" class="space-y-2 pt-4">
        <div class="flex justify-between items-center text-sm font-bold">
          <span>正在保存: {{ currentTrackIndex }} / {{ totalTracks }}</span>
          <span>
            {{ Math.round((currentTrackIndex / totalTracks) * 100) }}%
          </span>
        </div>
        <div class="w-full bg-gray-200 h-2">
          <div 
            class="bg-black h-2 transition-all duration-300" 
            :style="{ width: `${(currentTrackIndex / totalTracks) * 100}%` }"
          ></div>
        </div>
        <p class="text-sm text-gray-500">
          正在处理: {{ tracks[currentTrackIndex - 1]?.title }}...
        </p>
      </div>

      <div class="flex gap-4">
        <button 
          type="button"
          @click="handleSubmit"
          class="flex-1 bg-black text-white py-6 font-black uppercase tracking-widest hover:bg-white hover:text-black border-2 border-black transition-all"
          :disabled="tracks.length === 0 || isSaving"
          :class="{ 'opacity-50 cursor-not-allowed': tracks.length === 0 || isSaving }"
        >
          {{ isSaving ? '正在保存...' : `保存更改 (${tracks.length} 首)` }}
        </button>
        <button 
          type="button"
          @click="cancel"
          class="flex-1 border-2 border-black py-6 font-black uppercase tracking-widest hover:bg-black hover:text-white transition-all"
          :disabled="isSaving"
          :class="{ 'opacity-50 cursor-not-allowed': isSaving }"
        >
          取消
        </button>
      </div>
    </form>
  </div>
</template>
