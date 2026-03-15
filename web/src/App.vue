<template>
  <n-config-provider :theme-overrides="themeOverrides">
    <div style="display:flex;flex-direction:column;min-height:100vh">
      <AppTopbar />
      <main style="flex:1;padding-bottom:128px">
        <router-view />
      </main>
      <AudioPlayer />
    </div>
  </n-config-provider>
</template>

<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { NConfigProvider } from 'naive-ui'
import AppTopbar from '@/components/AppTopbar.vue'
import AudioPlayer from '@/components/AudioPlayer.vue'
import { useAuthStore } from '@/stores/auth'
import { useFeedStore } from '@/stores/feed'

const themeOverrides = {
  common: {
    primaryColor: '#000000',
    primaryColorHover: '#333333',
    primaryColorPressed: '#000000',
    borderRadius: '0px',
  },
}

const authStore = useAuthStore()
const feedStore = useFeedStore()

onMounted(() => {
  if (authStore.isAuthenticated) feedStore.startPolling()
})

watch(() => authStore.isAuthenticated, (authenticated) => {
  if (authenticated) {
    feedStore.startPolling()
  } else {
    feedStore.stopPolling()
  }
})
</script>
