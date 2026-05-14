<template>
  <div class="artist-detail">
    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="artist">
      <!-- Header -->
      <div class="artist-header">
        <div class="artist-header-inner">
          <div class="artist-img-wrap" v-if="artist.image_url">
            <img :src="artist.image_url" :alt="artist.name" class="artist-img" />
          </div>
          <div class="artist-meta">
            <div class="artist-status-row">
              <span
                v-if="artist.entry_status === 'confirmed'"
                class="status-badge status-confirmed"
              >已确认</span>
              <span
                v-else-if="artist.entry_status === 'disputed'"
                class="status-badge status-disputed"
              >争议</span>
            </div>
            <h1 class="artist-name">{{ artist.name }}</h1>
            <div class="artist-info-row">
              <span v-if="artist.nationality" class="info-item">{{ artist.nationality }}</span>
              <span v-if="artist.birth_year" class="info-item">
                {{ artist.birth_year }}{{ artist.death_year ? ' – ' + artist.death_year : ' –' }}
              </span>
            </div>
            <p v-if="artist.bio" class="artist-bio">{{ artist.bio }}</p>
            <p v-if="artist.members" class="artist-members">
              <span class="label">成员</span> {{ artist.members }}
            </p>
            <!-- Aliases -->
            <div v-if="artist.aliases && artist.aliases.length" class="aliases-row">
              <span class="label">别名</span>
              <span v-for="alias in artist.aliases" :key="alias.id" class="alias-tag">
                {{ alias.alias }}
              </span>
            </div>
          </div>
        </div>

        <!-- Nav actions -->
        <div class="artist-nav">
          <RouterLink :to="`/music/artists/${artistId}/edit`" class="nav-btn">编辑</RouterLink>
          <RouterLink :to="`/music/artists/${artistId}/history`" class="nav-btn">历史</RouterLink>
        </div>

        <!-- Admin status actions -->
        <div v-if="authStore.user?.role === 'admin'" class="admin-actions">
          <button
            v-if="artist.entry_status !== 'confirmed'"
            @click="changeStatus('confirmed')"
            class="action-btn action-confirm"
          >确认条目</button>
          <button
            v-if="artist.entry_status === 'confirmed'"
            @click="changeStatus('disputed')"
            class="action-btn action-dispute"
          >设为争议</button>
          <button
            v-if="artist.entry_status === 'disputed'"
            @click="changeStatus('open')"
            class="action-btn action-open"
          >重新开放</button>
        </div>
      </div>

      <!-- Albums -->
      <div class="albums-section">
        <h2 class="section-title">DISCOGRAPHY</h2>
        <div v-if="artist.albums && artist.albums.length" class="albums-grid">
          <div v-for="album in artist.albums" :key="album.id" class="album-card">
            <RouterLink :to="`/music/albums/${album.id}`">
              <img
                :src="album.cover_url || ''"
                :alt="album.title"
                class="album-cover"
              />
            </RouterLink>
            <div class="album-info">
              <div class="album-badges">
                <span v-if="album.album_type" class="badge">{{ album.album_type.toUpperCase() }}</span>
                <span v-if="album.entry_status === 'disputed'" class="badge badge-disputed">争议</span>
              </div>
              <RouterLink :to="`/music/albums/${album.id}`" class="album-title">
                {{ album.title }}
              </RouterLink>
              <p class="album-year">{{ album.year }}</p>
            </div>
          </div>
        </div>
        <div v-else class="empty">暂无专辑</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import type { Artist, MusicEntryStatus } from '@/types'

const route = useRoute()
const authStore = useAuthStore()
const API_URL = import.meta.env.VITE_API_URL || '/api'

const artistId = route.params.artistId as string
const artist = ref<Artist & { albums?: any[] } | null>(null)
const loading = ref(true)
const error = ref('')

const fetchArtist = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await fetch(`${API_URL}/artists/${artistId}`, {
      headers: authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {},
    })
    if (!res.ok) throw new Error('Artist not found')
    const data = await res.json()
    artist.value = data.data
  } catch (e: any) {
    error.value = e.message || 'Failed to load artist'
  } finally {
    loading.value = false
  }
}

const changeStatus = async (status: MusicEntryStatus) => {
  try {
    await fetch(`${API_URL}/artists/${artistId}/status`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`,
      },
      body: JSON.stringify({ status }),
    })
    if (artist.value) artist.value.entry_status = status
  } catch (e) {
    console.error('Failed to change status:', e)
  }
}

onMounted(fetchArtist)
</script>

<style scoped>
.artist-detail {
  max-width: 72rem;
  margin: 0 auto;
  padding: 2rem;
  padding-bottom: 12rem;
}
.loading,
.error {
  padding: 4rem;
  text-align: center;
  color: #6b7280;
}
.artist-header {
  border: 2px solid #000;
  padding: 2rem;
  margin-bottom: 3rem;
  box-shadow: 10px 10px 0px 0px rgba(0, 0, 0, 1);
}
.artist-header-inner {
  display: flex;
  gap: 2rem;
  align-items: flex-start;
}
.artist-img {
  width: 160px;
  height: 160px;
  object-fit: cover;
  border: 2px solid #000;
  filter: grayscale(100%);
  flex-shrink: 0;
}
.artist-meta {
  flex: 1;
}
.artist-status-row {
  margin-bottom: 0.5rem;
}
.status-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  font-size: 0.625rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  border: 2px solid;
}
.status-confirmed {
  border-color: #166534;
  color: #166534;
}
.status-disputed {
  border-color: #991b1b;
  color: #991b1b;
}
.artist-name {
  font-size: 2.5rem;
  font-weight: 900;
  letter-spacing: -0.04em;
  line-height: 1;
  margin: 0 0 0.5rem;
}
.artist-info-row {
  display: flex;
  gap: 1rem;
  margin-bottom: 0.75rem;
}
.info-item {
  font-size: 0.875rem;
  font-weight: 700;
  color: #6b7280;
}
.artist-bio {
  font-size: 0.875rem;
  line-height: 1.6;
  color: #374151;
  max-width: 48rem;
  margin: 0 0 0.75rem;
}
.artist-members {
  font-size: 0.875rem;
  margin: 0 0 0.5rem;
}
.aliases-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-top: 0.5rem;
}
.label {
  font-size: 0.625rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: #6b7280;
}
.alias-tag {
  border: 1px solid #d1d5db;
  padding: 0.125rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 600;
}
.artist-nav {
  display: flex;
  gap: 0.75rem;
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid #e5e7eb;
}
.nav-btn {
  border: 2px solid #000;
  padding: 0.375rem 1rem;
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  text-decoration: none;
  color: #000;
  background: #fff;
  transition: all 0.2s;
}
.nav-btn:hover {
  background: #000;
  color: #fff;
}
.admin-actions {
  display: flex;
  gap: 0.75rem;
  margin-top: 0.75rem;
}
.action-btn {
  border: 2px solid;
  padding: 0.375rem 1rem;
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  cursor: pointer;
  transition: all 0.2s;
  background: #fff;
}
.action-confirm {
  border-color: #166534;
  color: #166534;
}
.action-confirm:hover {
  background: #166534;
  color: #fff;
}
.action-dispute {
  border-color: #991b1b;
  color: #991b1b;
}
.action-dispute:hover {
  background: #991b1b;
  color: #fff;
}
.action-open {
  border-color: #000;
  color: #000;
}
.action-open:hover {
  background: #000;
  color: #fff;
}
.section-title {
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  border-left: 4px solid #000;
  padding-left: 0.75rem;
  margin: 0 0 1.5rem;
}
.albums-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 1.5rem;
}
.album-card {
  border: 2px solid #000;
  transition: box-shadow 0.2s;
}
.album-card:hover {
  box-shadow: 6px 6px 0px 0px rgba(0, 0, 0, 1);
}
.album-cover {
  width: 100%;
  aspect-ratio: 1;
  object-fit: cover;
  display: block;
  filter: grayscale(100%);
}
.album-info {
  padding: 0.75rem;
}
.album-badges {
  display: flex;
  gap: 0.25rem;
  margin-bottom: 0.25rem;
}
.badge {
  font-size: 0.5rem;
  font-weight: 900;
  letter-spacing: 0.1em;
  border: 1px solid #000;
  padding: 0.125rem 0.375rem;
}
.badge-disputed {
  border-color: #991b1b;
  color: #991b1b;
}
.album-title {
  display: block;
  font-size: 0.875rem;
  font-weight: 900;
  text-decoration: none;
  color: #000;
  line-height: 1.3;
  margin-bottom: 0.25rem;
}
.album-title:hover {
  text-decoration: underline;
}
.album-year {
  font-size: 0.75rem;
  color: #6b7280;
  margin: 0;
}
.empty {
  color: #9ca3af;
  padding: 2rem 0;
}
</style>
