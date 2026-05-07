
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
  <div class="max-w-3xl mx-auto px-8 py-20">
    <h1 class="text-4xl font-black tracking-tighter mb-8">贡献新档案</h1>
    <p class="text-gray-500 mb-12">
      帮助我们完善 Ye 的音乐史料库。支持批量上传和拖拽排序。
    </p>

    <form class="space-y-8">
      <div>
        <label class="block text-sm font-black uppercase tracking-widest mb-4">艺术家</label>
        <ArtistSelect
          v-model="formData.artist"
          placeholder="选择艺术家"
          :disabled="isUploading"
          class="w-full max-w-md"
        />
        <!-- Selected Artists Display -->
        <div v-if="formData.artist.length" class="mt-4 flex flex-wrap gap-2">
          <span
            v-for="(artist, idx) in formData.artist"
            :key="idx"
            class="inline-flex items-center bg-gray-100 border-2 border-black px-3 py-1.5 text-sm font-bold"
          >
            {{ artist.name }}
            <button
              type="button"
              @click.prevent="removeArtist(idx)"
              class="ml-2 text-red-600 hover:text-red-700 font-black"
            >
              ×
            </button>
          </span>
        </div>
      </div>
      
      <div>
        <label class="block text-sm font-black uppercase tracking-widest mb-4">专辑名称</label>
        <input 
          type="text" 
          required
          class="w-full max-w-md bg-white border-2 border-black p-4 focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] outline-none transition-all"
          v-model="formData.album"
        />
      </div>

       <div class="space-y-4">
         <label class="block text-sm font-black uppercase tracking-widest">发行日期</label>
         <input
           type="date"
           required
           class="w-full max-w-md bg-white border-2 border-black p-4 focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] outline-none transition-all"
           v-model="formData.releaseDate"
         />
       </div>

       <div class="space-y-4">
         <label class="block text-sm font-black uppercase tracking-widest">信息来源</label>
         <input
           type="text"
           placeholder="例如：官方网站、维基百科、音乐平台等"
           class="w-full max-w-md bg-white border-2 border-black p-4 focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] outline-none transition-all"
           v-model="formData.source"
         />
       </div>

      <div class="space-y-4">
        <label class="block text-sm font-black uppercase tracking-widest">专辑封面（可选，不上传默认为纯黑）</label>
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
          <p class="font-bold">点击上传封面图片</p>
          <p class="text-xs text-gray-400 mt-2">不上传将默认为纯黑色</p>
        </div>
        <div v-else class="relative border-2 border-black inline-block">
          <img :src="coverPreview" class="w-48 h-48 object-cover grayscale" alt="封面预览" />
          <button 
            type="button"
            @click="removeCover"
            class="absolute top-2 right-2 bg-black text-white px-3 py-1 text-xs font-bold hover:bg-red-600"
          >
            删除
          </button>
        </div>
      </div>

      <div class="space-y-4">
        <label class="block text-sm font-black uppercase tracking-widest">音频文件 (支持多选)</label>
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
          <p class="font-bold">点击选择多个音频文件</p>
          <p class="text-xs text-gray-400 mt-2">支持批量上传</p>
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
               class="px-3 py-1 text-xs font-black border-2 border-black hover:bg-black hover:text-white transition-colors"
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
            class="bg-white border-2 border-black p-4 flex items-center gap-4 cursor-move hover:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] transition-shadow"
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
              <p class="text-xs text-gray-400 truncate">{{ track.file.name }}</p>
            </div>
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

      <!-- Progress Display -->
      <div v-if="isUploading" class="space-y-2 pt-4">
        <div class="flex justify-between items-center text-sm font-bold">
          <span>正在上传: {{ currentTrackIndex }} / {{ totalTracks }}</span>
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


      <button 
        type="button"
        @click="handleSubmit"
        class="w-full bg-black text-white py-6 font-black uppercase tracking-widest hover:bg-white hover:text-black border-2 border-black transition-all"
        :disabled="tracks.length === 0 || isUploading"
        :class="{ 'opacity-50 cursor-not-allowed': tracks.length === 0 || isUploading }"
      >
        {{ isUploading ? '正在提交...' : `直接上传 (${tracks.length} 首)` }}
      </button>

      <button
        v-if="isAdmin"
        type="button"
        @click="handleCreateAlbumOnly"
        class="w-full bg-white text-black py-5 font-black uppercase tracking-widest hover:bg-black hover:text-white border-2 border-black transition-all"
        :disabled="tracks.length > 0 || isUploading || formData.artist.length === 0 || !formData.album.trim()"
        :class="{ 'opacity-50 cursor-not-allowed': tracks.length > 0 || isUploading || formData.artist.length === 0 || !formData.album.trim() }"
      >
        仅创建专辑档案
      </button>
    </form>
  </div>
</template>
