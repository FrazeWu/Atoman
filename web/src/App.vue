
<script setup lang="ts">
import { onMounted, watch } from 'vue'
import AppTopbar from './components/AppTopbar.vue'
import AudioPlayer from './components/AudioPlayer.vue'
import { useAuthStore } from '@/stores/auth'
import { useFeedStore } from '@/stores/feed'

const authStore = useAuthStore()
const feedStore = useFeedStore()

// Start polling when authenticated
onMounted(() => {
  if (authStore.isAuthenticated) {
    feedStore.startPolling()
  }
})

// Also watch for login/logout
watch(() => authStore.isAuthenticated, (authenticated) => {
  if (authenticated) {
    feedStore.startPolling()
  } else {
    feedStore.stopPolling()
  }
})
</script>

<template>
  <div class="flex flex-col min-h-screen">
    <AppTopbar />
    <main class="flex-grow pb-32">
      <router-view></router-view>
    </main>
    <AudioPlayer />
  </div>
</template>