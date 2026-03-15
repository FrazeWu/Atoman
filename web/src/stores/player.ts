import { defineStore } from 'pinia';
import { ref, watch } from 'vue';
import type { Song, RepeatMode } from '@/types';

const API_URL = import.meta.env.VITE_API_URL || '/api';

export const usePlayerStore = defineStore('player', () => {
  const songs = ref<Song[]>([]);
  const currentSong = ref<Song | null>(null);
  const isPlaying = ref(false);
  const isShuffled = ref(false);
  const repeatMode = ref<RepeatMode>('none');
  const volume = ref(1);
  const currentTime = ref(0);
  const duration = ref(0);

  const audio = new Audio();

  audio.addEventListener('timeupdate', () => {
    currentTime.value = audio.currentTime;
  });
  audio.addEventListener('durationchange', () => {
    duration.value = audio.duration;
  });
  audio.addEventListener('ended', () => {
    playNext();
  });

  const savePlaybackState = () => {
    if (currentSong.value) {
      const state = {
        songId: currentSong.value.id,
        currentTime: audio.currentTime,
        volume: volume.value,
        isPlaying: isPlaying.value,
        isShuffled: isShuffled.value,
        repeatMode: repeatMode.value
      };
      localStorage.setItem('playbackState', JSON.stringify(state));
    }
  };

  const loadPlaybackState = async () => {
    const savedState = localStorage.getItem('playbackState');
    if (savedState) {
      try {
        const state = JSON.parse(savedState);
        volume.value = state.volume || 1;
        isShuffled.value = state.isShuffled || false;
        repeatMode.value = state.repeatMode || 'none';
        
        if (state.songId) {
          const song = songs.value.find(s => s.id === state.songId);
          if (song) {
            audio.src = song.audio_url;
            audio.volume = volume.value;
            audio.currentTime = state.currentTime || 0;
            currentSong.value = song;
            
            if (state.isPlaying) {
              await audio.play();
              isPlaying.value = true;
            }
          }
        }
      } catch (error) {
        console.error('Failed to restore playback state:', error);
      }
    }
  };

  watch([currentSong, currentTime, volume, isPlaying, isShuffled, repeatMode], () => {
    savePlaybackState();
  }, { deep: true });

  const fetchSongs = async () => {
    try {
      const response = await fetch(`${API_URL}/songs`);
      if (response.ok) {
        songs.value = await response.json();
        await loadPlaybackState();
      }
    } catch (error) {
      console.error('Failed to fetch songs:', error);
    }
  };

  const playSong = (song: Song) => {
    if (currentSong.value?.id === song.id) {
      togglePlay();
    } else {
      audio.src = song.audio_url;
      audio.volume = volume.value;
      audio.play();
      currentSong.value = song;
      isPlaying.value = true;
    }
  };

  const togglePlay = () => {
    if (currentSong.value) {
      if (isPlaying.value) {
        audio.pause();
      } else {
        audio.play();
      }
      isPlaying.value = !isPlaying.value;
    }
  };

  const playNext = () => {
    if (!currentSong.value || songs.value.length === 0) return;
    const currentIndex = songs.value.findIndex(s => s.id === currentSong.value?.id);

    let nextIndex;
    if (isShuffled.value) {
      nextIndex = Math.floor(Math.random() * songs.value.length);
    } else {
      if (repeatMode.value === 'one') {
        // Single loop: replay current song
        audio.currentTime = 0;
        audio.play();
        isPlaying.value = true;
        return;
      } else if (repeatMode.value === 'all' || currentIndex < songs.value.length - 1) {
         nextIndex = (currentIndex + 1) % songs.value.length;
      } else {
         // End of list and not repeat all
         isPlaying.value = false;
         return;
      }
    }
    playSong(songs.value[nextIndex]);
  };

  const playPrevious = () => {
    if (!currentSong.value || songs.value.length === 0) return;
    const currentIndex = songs.value.findIndex(s => s.id === currentSong.value?.id);
    let prevIndex = (currentIndex - 1 + songs.value.length) % songs.value.length;
    playSong(songs.value[prevIndex]);
  };

  const toggleShuffle = () => {
    isShuffled.value = !isShuffled.value;
  };

  const toggleRepeat = () => {
    const modes: RepeatMode[] = ['none', 'all', 'one'];
    const nextMode = modes[(modes.indexOf(repeatMode.value) + 1) % modes.length];
    repeatMode.value = nextMode;
  };

  const setVolume = (v: number) => {
    audio.volume = v;
    volume.value = v;
  };

  const seek = (time: number) => {
    audio.currentTime = time;
    currentTime.value = time;
  };

  return {
    songs,
    currentSong,
    isPlaying,
    isShuffled,
    repeatMode,
    volume,
    currentTime,
    duration,
    fetchSongs,
    playSong,
    togglePlay,
    playNext,
    playPrevious,
    toggleShuffle,
    toggleRepeat,
    setVolume,
    seek
  };
});
