
<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink } from 'vue-router';
import { usePlayerStore } from '@/stores/player.ts';

const player = usePlayerStore();
player.fetchSongs();

interface AlbumGroup {
  album: string;
  year: number;
  release_date: string;
  cover_url: string;
  artist: string;
  songs: typeof player.songs;
}

const albumGroups = computed(() => {
  const songs = [...player.songs];
  const groups = new Map<string, AlbumGroup>();
  
  songs.forEach(song => {
    const key = `${song.album}-${song.year}`;
    if (!groups.has(key)) {
      groups.set(key, {
        album: song.album,
        year: song.year,
        release_date: song.release_date,
        cover_url: song.cover_url,
        artist: song.artist,
        songs: []
      });
    }
    groups.get(key)!.songs.push(song);
  });
  
  return Array.from(groups.values()).sort((a, b) => b.year - a.year);
});

const shouldShowYear = (index: number) => {
  return index === 0 || albumGroups.value[index - 1].year !== albumGroups.value[index].year;
};
</script>

<template>
  <div class="relative pt-12 pb-48">
    <div class="max-w-6xl mx-auto px-8 mb-12">
      <h1 class="text-5xl font-black italic tracking-tighter border-l-8 border-black pl-6">
        ATOMAN<br />TIMELINE
      </h1>
      <p class="mt-4 text-gray-500 max-w-lg">
        An interactive archival project cataloging the evolution of sound, fashion, and culture through the lens of Ye.
      </p>
    </div>

    <div class="relative max-w-6xl mx-auto min-h-[2000px] px-8">
      <div 
        class="absolute left-1/3 top-0 bottom-0 w-1 bg-black z-0 opacity-100" 
        style="height: 100%"
      ></div>

      <div class="space-y-24 relative z-10">
        <div v-for="(albumGroup, index) in albumGroups" :key="`${albumGroup.album}-${albumGroup.year}`" class="relative flex items-center">
          <div v-if="shouldShowYear(index)" class="absolute left-1/3 -translate-x-1/2 -top-12 z-20">
            <span class="bg-black text-white px-4 py-1 text-sm font-bold tracking-widest">
              {{ albumGroup.year }}
            </span>
          </div>

          <div 
            class="absolute left-1/3 -translate-x-1/2 w-6 h-6 rounded-full border-4 border-white z-20 bg-black"
          ></div>

          <div class="ml-[calc(33.333%+2rem)] w-[calc(66.666%-2rem)]">
            <div class="bg-white border-2 border-black p-6 hover:shadow-[10px_10px_0px_0px_rgba(0,0,0,1)] transition-all duration-300">
              <div class="flex gap-6">
                <img :src="albumGroup.cover_url || 'data:image/svg+xml,%3Csvg xmlns=%22http://www.w3.org/2000/svg%22 width=%22200%22 height=%22200%22%3E%3Crect width=%22200%22 height=%22200%22 fill=%22%23000%22/%3E%3C/svg%3E'" 
                     class="w-32 h-32 border-2 border-black object-cover flex-shrink-0" 
                     :alt="albumGroup.album" />
                <div class="flex flex-col justify-center flex-grow">
                  <h3 class="text-2xl font-black tracking-tight leading-tight">{{ albumGroup.album }}</h3>
                  <p class="text-sm font-bold text-gray-600">{{ albumGroup.artist }}</p>
                  <p class="text-xs text-gray-500 mt-1">{{ albumGroup.release_date }}</p>
                  <p class="text-xs text-gray-400">{{ albumGroup.songs.length }} {{ albumGroup.songs.length === 1 ? 'track' : 'tracks' }}</p>
                  
                  <div class="flex gap-3 mt-4">
                    <button 
                      @click="player.playSong(albumGroup.songs[0])"
                      class="border-2 border-black px-4 py-2 font-black text-xs uppercase tracking-widest transition-all hover:bg-black hover:text-white"
                    >
                      ▶ 播放
                    </button>
                     <RouterLink
                       :to="`/artist=${encodeURIComponent(albumGroup.artist.replace(/ /g, '_'))}/album=${encodeURIComponent(albumGroup.album.replace(/ /g, '_'))}`"
                       class="border-2 border-black px-4 py-2 font-black text-xs uppercase tracking-widest transition-all hover:bg-black hover:text-white inline-block"
                     >
                       详情
                     </RouterLink>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
