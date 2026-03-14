<script setup lang="ts">
import { ref, watch, computed, onMounted } from 'vue';
import { useApi } from '@/composables/useApi';

interface Artist {
  id: number;
  name: string;
  bio?: string;
}

const props = defineProps<{
  modelValue: string;
  placeholder?: string;
  disabled?: boolean;
}>();

const emit = defineEmits<{
  'update:modelValue': [value: string];
}>();

const api = useApi();
const artists = ref<Artist[]>([]);
const searchQuery = ref('');
const showDropdown = ref(false);
const showAddNew = ref(false);
const newArtistName = ref('');
const loading = ref(false);

const filteredArtists = computed(() => {
  if (!searchQuery.value) return artists.value;
  const query = searchQuery.value.toLowerCase();
  return artists.value.filter(a =>
    a.name.toLowerCase().includes(query)
  );
});

const selectedArtist = computed(() => {
  return artists.value.find(a => a.name === props.modelValue);
});

const inputRef = ref<HTMLInputElement | null>(null);

const fetchArtists = async () => {
  loading.value = true;
  try {
    const response = await fetch(`${api.url}/artists`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    });
    if (response.ok) {
      artists.value = await response.json();
    }
  } catch (error) {
    console.error('Failed to fetch artists:', error);
  } finally {
    loading.value = false;
  }
};

const handleSelect = (artistName: string) => {
  emit('update:modelValue', artistName);
  showDropdown.value = false;
  searchQuery.value = '';
};

const handleAddNew = async () => {
  if (!newArtistName.value.trim()) return;

  try {
    const response = await fetch(`${api.url}/artists`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify({
        name: newArtistName.value.trim(),
        bio: ''
      })
    });

    if (response.ok) {
      const newArtist = await response.json();
      artists.value.push(newArtist);
      emit('update:modelValue', newArtist.name);
      showAddNew.value = false;
      newArtistName.value = '';
      showDropdown.value = false;
    } else {
      const error = await response.json();
      alert(error.error || 'Failed to add artist');
    }
  } catch (error) {
    console.error('Failed to add artist:', error);
    alert('Failed to add artist');
  }
};

const handleFocus = () => {
  if (!props.disabled && artists.value.length === 0) {
    fetchArtists();
  }
  showDropdown.value = true;
};

const handleClickOutside = (event: MouseEvent) => {
  if (inputRef.value && !inputRef.value.contains(event.target as Node)) {
    showDropdown.value = false;
    showAddNew.value = false;
  }
};

onMounted(() => {
  fetchArtists();
  document.addEventListener('click', handleClickOutside);
});

const handleInput = (event: Event) => {
  const target = event.target as HTMLInputElement;
  searchQuery.value = target.value;
  emit('update:modelValue', target.value);
  showDropdown.value = true;
  showAddNew.value = false;
};
</script>

<template>
  <div class="relative" ref="inputRef">
    <input
      type="text"
      :value="modelValue"
      @input="handleInput"
      @focus="handleFocus"
      :placeholder="placeholder || '选择艺术家'"
      :disabled="disabled"
      class="w-full border-2 border-black p-3 outline-none focus:bg-gray-50 transition-colors"
    />

    <div
      v-if="showDropdown && !disabled"
      class="absolute z-10 w-full bg-white border-2 border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] max-h-60 overflow-y-auto"
    >
      <!-- Search input -->
      <div class="p-3 border-b border-black">
        <input
          type="text"
          v-model="searchQuery"
          placeholder="搜索艺术家..."
          class="w-full border-2 border-black p-2 outline-none focus:bg-gray-50"
        />
      </div>

      <!-- Artist list -->
      <div v-if="!showAddNew">
        <div
          v-for="artist in filteredArtists"
          :key="artist.id"
          @click="handleSelect(artist.name)"
          class="px-4 py-2 hover:bg-black hover:text-white cursor-pointer transition-colors border-b border-black last:border-b-0"
        >
          {{ artist.name }}
        </div>

        <!-- Add new artist option -->
        <div
          @click.stop="showAddNew = true"
          class="px-4 py-2 hover:bg-black hover:text-white cursor-pointer transition-colors font-bold"
        >
          + 添加新艺术家
        </div>
      </div>

      <!-- Add new artist form -->
      <div v-else class="p-3 border-t border-black">
        <input
          type="text"
          v-model="newArtistName"
          placeholder="新艺术家名称"
          class="w-full border-2 border-black p-2 mb-2 outline-none focus:bg-gray-50"
        />
        <div class="flex gap-2">
          <button
            @click="handleAddNew"
            :disabled="!newArtistName.trim()"
            class="flex-1 bg-black text-white px-4 py-2 hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            添加
          </button>
          <button
            @click="showAddNew = false"
            class="flex-1 border-2 border-black px-4 py-2 hover:bg-gray-50 transition-colors"
          >
            取消
          </button>
        </div>
      </div>

      <!-- Loading state -->
      <div v-if="loading" class="p-4 text-center text-gray-500">
        加载中...
      </div>

      <!-- No results -->
      <div
        v-if="!loading && !showAddNew && filteredArtists.length === 0"
        class="p-4 text-center text-gray-500"
      >
        未找到艺术家
      </div>
    </div>
  </div>
</template>
