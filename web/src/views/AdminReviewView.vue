<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useAuthStore } from '@/stores/auth';
import { useApi } from '@/composables/useApi';
import type { Song, User } from '@/types';

// 统一的审核项接口
interface ReviewItem {
  id: string; // 唯一标识
  type: 'song_batch' | 'song_correction' | 'album_correction' | 'album_upload';
  created_at: string;
  user: User;
  
  // 歌曲批次上传
  batch_id?: string;
  songs?: Song[];
  song_ids?: number[];
  
  // 纠错
  correction_id?: number;
  field_name?: string;
  current_value?: string;
  corrected_value?: string;
  reason?: string;
  target_title?: string; // 歌曲或专辑名称
  
  // 元数据
  artist?: string;
  album?: string;
  album_id?: number;
  cover_url?: string;
  cover_source?: 'local' | 's3';
  count?: number;
}

const authStore = useAuthStore();
const api = useApi();
const reviewItems = ref<ReviewItem[]>([]);
const loading = ref(true);
const processingItems = ref<Set<string>>(new Set());

// 获取所有待审核内容并统一格式
const fetchAllPendingItems = async () => {
  try {
    const headers = { 'Authorization': `Bearer ${authStore.token}` };
    
    // 并行获取所有待审核数据
    const [songsRes, songCorrectionsRes, albumCorrectionsRes, albumsRes] = await Promise.all([
      fetch(`${api.url}/admin/pending`, { headers }),
      fetch(`${api.url}/admin/pending-song-corrections`, { headers }),
      fetch(`${api.url}/admin/pending-album-corrections`, { headers }),
      fetch(`${api.url}/admin/pending-albums`, { headers })
    ]);

    const items: ReviewItem[] = [];

    // 处理歌曲批次上传 - 按 batch_id 分组
    if (songsRes.ok) {
      const songs = await songsRes.json();
      if (Array.isArray(songs) && songs.length > 0) {
        // 按 batch_id 分组
        const batchGroups = new Map<string, any[]>();
        
        songs.forEach((song: any) => {
          const batchId = song.batch_id || `single_${song.id}`;
          if (!batchGroups.has(batchId)) {
            batchGroups.set(batchId, []);
          }
          batchGroups.get(batchId)!.push(song);
        });
        
        // 为每个批次创建一个 ReviewItem
        batchGroups.forEach((batchSongs, batchId) => {
          // 使用第一首歌的信息作为批次信息
          const firstSong = batchSongs[0];
          
          items.push({
            id: `batch_${batchId}`,
            type: 'song_batch',
            batch_id: batchId,
            created_at: firstSong.created_at,
            user: firstSong.user || { username: 'Unknown' },
            songs: batchSongs,
            song_ids: batchSongs.map(s => s.id),
            artist: firstSong.artist,
            album: firstSong.album,
            count: batchSongs.length,
            cover_url: firstSong.cover_url,
            cover_source: firstSong.cover_source,
            target_title: firstSong.album // 使用专辑名作为标题
          });
        });
      }
    }

    // 处理歌曲纠错
    if (songCorrectionsRes.ok) {
      const corrections = await songCorrectionsRes.json();
      corrections.forEach((corr: any) => {
        items.push({
          id: `song_corr_${corr.id}`,
          type: 'song_correction',
          created_at: corr.created_at,
          user: corr.user,
          correction_id: corr.id,
          field_name: corr.field_name,
          current_value: corr.current_value,
          corrected_value: corr.corrected_value,
          reason: corr.reason,
          target_title: corr.song?.title,
          artist: corr.song?.artist,
          album: corr.song?.album
        });
      });
    }

    // 处理专辑纠错（包括封面和元数据）
    if (albumCorrectionsRes.ok) {
      const albumCorr = await albumCorrectionsRes.json();
      albumCorr.forEach((corr: any) => {
        const artistName = corr.album?.artists?.length > 0 
          ? corr.album.artists[0].name 
          : 'Unknown Artist';
        
        items.push({
          id: `album_corr_${corr.id}`,
          type: 'album_correction',
          created_at: corr.created_at,
          user: corr.user,
          correction_id: corr.id,
          field_name: corr.corrected_cover_url ? 'cover_url' : 'title',
          current_value: corr.original_title || corr.original_cover_url,
          corrected_value: corr.corrected_title || corr.corrected_cover_url,
          reason: corr.reason,
          target_title: corr.album?.title,
          artist: artistName,
          cover_url: corr.corrected_cover_url || corr.album?.cover_url
        });
      });
    }

    // 处理专辑上传
    if (albumsRes.ok) {
      const albums = await albumsRes.json();
      albums.forEach((album: any) => {
        const artistName = album.artists?.length > 0 
          ? album.artists[0].name 
          : 'Unknown Artist';
        
        items.push({
          id: `album_${album.id}`,
          type: 'album_upload',
          created_at: album.created_at,
          user: album.user || { username: 'Unknown' },
          album_id: album.id,
          target_title: album.title,
          artist: artistName,
          cover_url: album.cover_url,
          cover_source: album.cover_source
        });
      });
    }

    // 按时间倒序排列（最新的在前）
    items.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
    
    reviewItems.value = items;
  } catch (e) {
    console.error('Failed to fetch pending items:', e);
  } finally {
    loading.value = false;
  }
};

// 批准歌曲批次
const approveSongBatch = async (item: ReviewItem) => {
  if (!item.song_ids || item.song_ids.length === 0) return;
  
  // 防止重复点击
  if (processingItems.value.has(item.id)) return;
  
  processingItems.value.add(item.id);
  
  try {
    let failedCount = 0;
    
    // 循环批准每首歌
    for (const songId of item.song_ids) {
      const response = await fetch(`${api.url}/admin/approve/${songId}`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${authStore.token}` }
      });
      
      if (!response.ok) {
        failedCount++;
      }
    }
    
    if (failedCount > 0) {
      alert(`批准失败：${failedCount} 首歌曲处理失败`);
    } else {
      // 成功后从列表移除
      reviewItems.value = reviewItems.value.filter(i => i.id !== item.id);
    }
  } catch (e) {
    console.error('批准歌曲批次失败:', e);
    alert('操作失败，请重试');
  } finally {
    processingItems.value.delete(item.id);
  }
};

// 驳回歌曲批次
const rejectSongBatch = async (item: ReviewItem) => {
  if (!item.song_ids || item.song_ids.length === 0) return;
  
  // 防止重复点击
  if (processingItems.value.has(item.id)) return;
  
  processingItems.value.add(item.id);
  
  try {
    let failedCount = 0;
    
    // 循环驳回每首歌
    for (const songId of item.song_ids) {
      const response = await fetch(`${api.url}/admin/reject/${songId}`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${authStore.token}` }
      });
      
      if (!response.ok) {
        failedCount++;
      }
    }
    
    if (failedCount > 0) {
      alert(`驳回失败：${failedCount} 首歌曲处理失败`);
    } else {
      // 成功后从列表移除
      reviewItems.value = reviewItems.value.filter(i => i.id !== item.id);
    }
  } catch (e) {
    console.error('驳回歌曲批次失败:', e);
    alert('操作失败，请重试');
  } finally {
    processingItems.value.delete(item.id);
  }
};

// 批准纠错
const approveCorrection = async (item: ReviewItem) => {
  const endpoint = item.type === 'song_correction' 
    ? `${api.url}/admin/approve-song-correction/${item.correction_id}`
    : `${api.url}/admin/approve-album-correction/${item.correction_id}`;
  
  try {
    const response = await fetch(endpoint, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${authStore.token}` }
    });
    
    if (response.ok) {
      reviewItems.value = reviewItems.value.filter(i => i.id !== item.id);
    } else {
      alert('操作失败');
    }
  } catch (e) {
    alert('操作失败');
  }
};

// 驳回纠错
const rejectCorrection = async (item: ReviewItem) => {
  const endpoint = item.type === 'song_correction'
    ? `${api.url}/admin/reject-song-correction/${item.correction_id}`
    : `${api.url}/admin/reject-album-correction/${item.correction_id}`;
  
  try {
    const response = await fetch(endpoint, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${authStore.token}` }
    });
    
    if (response.ok) {
      reviewItems.value = reviewItems.value.filter(i => i.id !== item.id);
    } else {
      alert('操作失败');
    }
  } catch (e) {
    alert('操作失败');
  }
};

// 批准专辑上传
const approveAlbum = async (item: ReviewItem) => {
  try {
    const response = await fetch(`${api.url}/admin/approve-album/${item.album_id}`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${authStore.token}` }
    });
    
    if (response.ok) {
      reviewItems.value = reviewItems.value.filter(i => i.id !== item.id);
    } else {
      alert('操作失败');
    }
  } catch (e) {
    alert('操作失败');
  }
};

// 驳回专辑上传
const rejectAlbum = async (item: ReviewItem) => {
  try {
    const response = await fetch(`${api.url}/admin/reject-album/${item.album_id}`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${authStore.token}` }
    });
    
    if (response.ok) {
      reviewItems.value = reviewItems.value.filter(i => i.id !== item.id);
    } else {
      alert('操作失败');
    }
  } catch (e) {
    alert('操作失败');
  }
};

// 试听歌曲
const playAudio = (url: string) => {
  const audio = new Audio(url);
  audio.play();
};

// 类型标签文本
const getTypeLabel = (type: string) => {
  const labels: Record<string, string> = {
    'song_batch': '歌曲上传',
    'song_correction': '歌曲纠错',
    'album_correction': '专辑修正',
    'album_upload': '专辑上传'
  };
  return labels[type] || type;
};

// 字段名称中文映射
const getFieldLabel = (field: string) => {
  const labels: Record<string, string> = {
    'title': '标题',
    'artist': '艺术家',
    'album': '专辑',
    'year': '年份',
    'release_date': '发行日期',
    'lyrics': '歌词',
    'cover_url': '专辑封面',
    'track_number': '曲目编号'
  };
  return labels[field] || field;
};

// 格式化日期
const formatDate = (dateString: string | undefined) => {
  if (!dateString) return '未知';
  const date = new Date(dateString);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
};

// 检查是否本地存储
const isLocalStorage = (item: any) => {
  // Check if any song in the batch has local storage
  if (item.songs && item.songs.length > 0) {
    return item.songs.some((song: any) => 
      song.audio_source === 'local' || 
      song.cover_source === 'local' ||
      (song.audio_url && song.audio_url.startsWith('/uploads/'))
    );
  }
  // Check for album uploads
  if (item.type === 'album_upload' && item.cover_url) {
    return item.cover_url.startsWith('/uploads/');
  }
  return false;
};

// 格式化曲目编号
const formatTrackNumber = (num: number | null | undefined) => {
  if (!num) return '—';
  return String(num).padStart(2, '0');
};

onMounted(async () => {
  if (!authStore.isAuthenticated || authStore.user?.role !== 'admin') {
    alert('需要管理员权限');
    return;
  }
  await fetchAllPendingItems();
});
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">审核队列</h1>
      <p class="page-desc">
        审核用户提交的内容修正和新增。共 <strong>{{ reviewItems.length }}</strong> 项待审核。
      </p>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="state-loading">
      <p>加载中...</p>
    </div>

    <!-- 空状态 -->
    <div v-else-if="reviewItems.length === 0" class="state-empty">
      <p>暂无待审核内容</p>
    </div>

    <!-- 审核列表 -->
    <div v-else class="review-list">
      <div
        v-for="item in reviewItems"
        :key="item.id"
        class="review-card"
      >
        <!-- 头部：类型标签 + 元信息 -->
        <div class="card-header">
          <div class="card-meta">
            <div class="card-meta-top">
              <span class="type-badge">{{ getTypeLabel(item.type) }}</span>
              <span class="card-date">{{ new Date(item.created_at).toLocaleString('zh-CN') }}</span>
            </div>
            <h2 class="card-title">{{ item.target_title || item.album || '未命名' }}</h2>
            <p class="card-artist">{{ item.artist }}</p>
            <p class="card-submitter">提交者: {{ item.user?.username }}</p>
          </div>

          <!-- 操作按钮 -->
          <div class="action-btns">
            <button
              @click="
                item.type === 'song_batch' ? approveSongBatch(item) :
                item.type === 'album_upload' ? approveAlbum(item) :
                approveCorrection(item)
              "
              :disabled="processingItems.has(item.id)"
              class="btn-approve"
            >
              {{ processingItems.has(item.id) ? '处理中...' : '通过' }}
            </button>
            <button
              @click="
                item.type === 'song_batch' ? rejectSongBatch(item) :
                item.type === 'album_upload' ? rejectAlbum(item) :
                rejectCorrection(item)
              "
              :disabled="processingItems.has(item.id)"
              class="btn-reject"
            >
              {{ processingItems.has(item.id) ? '处理中...' : '驳回' }}
            </button>
          </div>
        </div>

        <!-- 内容区：根据类型显示不同内容 -->

        <!-- 歌曲批次上传 -->
        <div v-if="item.type === 'song_batch'" class="content-section">
          <div class="batch-info-row">
            <div v-if="item.cover_url" class="batch-cover">
              <img :src="item.cover_url" class="batch-cover-img" alt="专辑封面" />
            </div>
            <div v-else class="batch-cover-placeholder">
              <span>无封面</span>
            </div>
            <div class="batch-details">
              <p class="meta-label">专辑信息</p>
              <p class="batch-album">{{ item.album }}</p>
              <p class="batch-artist">{{ item.artist }}</p>
              <p class="batch-count"><strong>{{ item.count }}</strong> 首歌曲</p>
            </div>
          </div>

          <!-- 上传信息 -->
          <div class="info-grid">
            <div class="info-cell">
              <p class="meta-label">上传时间</p>
              <p class="info-value">{{ formatDate(item.created_at) }}</p>
            </div>
            <div class="info-cell">
              <p class="meta-label">上传者</p>
              <p class="info-value">{{ item.user?.username || '匿名用户' }}</p>
            </div>
            <div class="info-cell">
              <p class="meta-label">存储位置</p>
              <span v-if="isLocalStorage(item)" class="storage-badge storage-local">🟡 本地暂存</span>
              <span v-else class="storage-badge storage-cloud">🔵 云端存储</span>
            </div>
          </div>

          <!-- 歌曲列表 -->
          <div class="tracklist-section">
            <p class="meta-label" style="margin-bottom:0.5rem">曲目列表</p>
            <div
              v-for="song in item.songs"
              :key="song.id"
              class="track-row"
            >
              <span class="track-num">{{ formatTrackNumber(song.track_number) }}</span>
              <div class="track-title">{{ song.title }}</div>
              <div class="track-badges">
                <span
                  v-if="song.audio_source === 'local'"
                  class="track-badge badge-local"
                  title="本地暂存"
                >LOCAL</span>
                <span
                  v-else
                  class="track-badge badge-s3"
                  title="云端存储"
                >S3</span>
              </div>
              <button
                @click="playAudio(song.audio_url)"
                class="btn-play"
              >
                ▶ 试听
              </button>
              <a
                :href="song.audio_url"
                target="_blank"
                class="btn-download"
                title="下载"
              >⬇</a>
            </div>
          </div>
        </div>

        <!-- 专辑上传 -->
        <div v-else-if="item.type === 'album_upload'" class="content-section">
          <div class="batch-info-row">
            <div v-if="item.cover_url" class="album-cover">
              <img :src="item.cover_url" class="album-cover-img" alt="专辑封面" />
            </div>
            <div v-else class="album-cover-placeholder">
              <span>无封面</span>
            </div>
            <div class="batch-details">
              <p class="meta-label">专辑信息</p>
              <p class="album-title">{{ item.target_title }}</p>
              <p class="album-artist">{{ item.artist }}</p>
            </div>
          </div>

          <!-- 上传信息 -->
          <div class="info-grid">
            <div class="info-cell">
              <p class="meta-label">上传时间</p>
              <p class="info-value">{{ formatDate(item.created_at) }}</p>
            </div>
            <div class="info-cell">
              <p class="meta-label">上传者</p>
              <p class="info-value">{{ item.user?.username || '匿名用户' }}</p>
            </div>
            <div class="info-cell">
              <p class="meta-label">存储位置</p>
              <span v-if="isLocalStorage(item)" class="storage-badge storage-local">🟡 本地暂存</span>
              <span v-else class="storage-badge storage-cloud">🔵 云端存储</span>
            </div>
          </div>
        </div>

        <!-- 纠错：封面类型 -->
        <div v-else-if="item.field_name === 'cover_url'" class="content-section">
          <div class="correction-grid">
            <div>
              <p class="meta-label">原封面</p>
              <div class="cover-before">
                <img
                  v-if="item.current_value"
                  :src="item.current_value"
                  class="cover-compare-img"
                  alt="原封面"
                />
                <div v-else class="cover-empty"><span>无封面</span></div>
              </div>
            </div>
            <div>
              <p class="meta-label">修改后</p>
              <div class="cover-after">
                <img
                  :src="item.corrected_value"
                  class="cover-compare-img"
                  alt="新封面"
                />
              </div>
            </div>
          </div>
          <div v-if="item.reason" class="reason-box">
            <p class="meta-label">修改原因</p>
            <p class="reason-text">{{ item.reason }}</p>
          </div>
        </div>

        <!-- 纠错：文本类型 -->
        <div v-else class="content-section">
          <div class="correction-grid">
            <div>
              <p class="meta-label">原 {{ getFieldLabel(item.field_name || '') }}</p>
              <div class="text-before">{{ item.current_value || '（无）' }}</div>
            </div>
            <div>
              <p class="meta-label">修改后 {{ getFieldLabel(item.field_name || '') }}</p>
              <div class="text-after">{{ item.corrected_value }}</div>
            </div>
          </div>
          <div v-if="item.reason" class="reason-box">
            <p class="meta-label">修改原因</p>
            <p class="reason-text">{{ item.reason }}</p>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>

<style scoped>
.page-container {
  max-width: 1024px;
  margin: 0 auto;
  padding: 5rem 2rem;
}

.page-header {
  margin-bottom: 3rem;
}

.page-title {
  font-size: 2.5rem;
  font-weight: 900;
  letter-spacing: -0.05em;
  margin: 0 0 1rem 0;
}

.page-desc {
  color: #6b7280;
}

.state-loading,
.state-empty {
  text-align: center;
  padding: 5rem 0;
  color: #9ca3af;
  font-weight: 500;
}

.state-empty {
  background: #f9fafb;
  border: 2px dashed #e5e7eb;
}

.review-list {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.review-card {
  background: #fff;
  border: 2px solid #000;
  padding: 2rem;
  transition: box-shadow 0.2s;
}
.review-card:hover {
  box-shadow: 15px 15px 0px 0px rgba(0,0,0,1);
}

/* Card header */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 2px solid #f3f4f6;
}

.card-meta { flex: 1; }

.card-meta-top {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
}

.type-badge {
  background: #000;
  color: #fff;
  padding: 0.25rem 0.75rem;
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.card-date {
  color: #9ca3af;
  font-size: 0.875rem;
}

.card-title {
  font-size: 1.5rem;
  font-weight: 900;
  margin: 0 0 0.25rem 0;
}

.card-artist {
  color: #6b7280;
  font-weight: 700;
  margin: 0 0 0.5rem 0;
}

.card-submitter {
  font-size: 0.875rem;
  color: #9ca3af;
  margin: 0;
}

.action-btns {
  display: flex;
  gap: 0.75rem;
  flex-shrink: 0;
  margin-left: 1rem;
}

.btn-approve {
  background: #000;
  color: #fff;
  padding: 0.75rem 1.5rem;
  font-weight: 700;
  border: none;
  cursor: pointer;
  transition: background 0.2s;
}
.btn-approve:hover:not(:disabled) { background: #16a34a; }
.btn-approve:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-reject {
  background: #fff;
  color: #000;
  border: 2px solid #000;
  padding: 0.75rem 1.5rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-reject:hover:not(:disabled) {
  background: #dc2626;
  color: #fff;
  border-color: #dc2626;
}
.btn-reject:disabled { opacity: 0.5; cursor: not-allowed; }

/* Content sections */
.content-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.batch-info-row {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  margin-bottom: 1rem;
}

.batch-cover {
  width: 8rem;
  height: 8rem;
  border: 2px solid #000;
  flex-shrink: 0;
}
.batch-cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  filter: grayscale(1);
}

.batch-cover-placeholder {
  width: 8rem;
  height: 8rem;
  border: 2px solid #d1d5db;
  background: #f3f4f6;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-size: 0.75rem;
  font-weight: 700;
  color: #9ca3af;
}

.batch-details { flex: 1; }

.meta-label {
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: #6b7280;
  margin: 0 0 0.25rem 0;
}

.batch-album {
  font-size: 1.25rem;
  font-weight: 900;
  margin: 0 0 0.25rem 0;
}

.batch-artist {
  font-weight: 700;
  color: #4b5563;
  margin: 0 0 0.5rem 0;
}

.batch-count {
  font-size: 0.875rem;
  color: #6b7280;
  margin: 0;
}

.album-cover {
  width: 12rem;
  height: 12rem;
  border: 2px solid #000;
  flex-shrink: 0;
}
.album-cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.album-cover-placeholder {
  width: 12rem;
  height: 12rem;
  border: 2px solid #d1d5db;
  background: #f3f4f6;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-weight: 700;
  color: #9ca3af;
}

.album-title {
  font-size: 1.5rem;
  font-weight: 900;
  margin: 0 0 0.5rem 0;
}

.album-artist {
  font-size: 1.125rem;
  font-weight: 700;
  color: #4b5563;
  margin: 0;
}

/* Info grid */
.info-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
  padding: 1rem;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  font-size: 0.875rem;
}

.info-cell { }

.info-value {
  font-weight: 700;
  margin: 0;
}

.storage-badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 900;
}
.storage-local { background: #fef9c3; color: #713f12; }
.storage-cloud { background: #dbeafe; color: #1e40af; }

/* Track list */
.tracklist-section { }

.track-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem;
  background: #f9fafb;
  transition: background 0.2s;
}
.track-row:hover { background: #f3f4f6; }

.track-num {
  font-family: monospace;
  color: #9ca3af;
  width: 2rem;
  text-align: right;
  font-size: 0.875rem;
  flex-shrink: 0;
}

.track-title {
  flex: 1;
  font-weight: 700;
}

.track-badges {
  display: flex;
  gap: 0.5rem;
}

.track-badge {
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 700;
}
.badge-local { background: #fefce8; color: #713f12; border: 1px solid #fde68a; }
.badge-s3 { background: #eff6ff; color: #1d4ed8; border: 1px solid #bfdbfe; }

.btn-play {
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  background: #fff;
  border: 1px solid #000;
  padding: 0.25rem 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
}
.btn-play:hover { background: #000; color: #fff; }

.btn-download {
  color: #9ca3af;
  font-size: 0.875rem;
  text-decoration: none;
  flex-shrink: 0;
  transition: color 0.2s;
}
.btn-download:hover { color: #000; }

/* Correction: cover */
.correction-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
}

.cover-before { border: 2px solid #d1d5db; }
.cover-after { border: 2px solid #000; }

.cover-compare-img {
  width: 100%;
  aspect-ratio: 1/1;
  object-fit: cover;
  display: block;
}

.cover-empty {
  width: 100%;
  aspect-ratio: 1/1;
  background: #e5e7eb;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9ca3af;
}

/* Correction: text */
.text-before {
  padding: 1rem;
  background: #f9fafb;
  border: 2px solid #d1d5db;
  font-family: monospace;
  font-size: 0.875rem;
  min-height: 5rem;
}

.text-after {
  padding: 1rem;
  background: #f0fdf4;
  border: 2px solid #000;
  font-family: monospace;
  font-size: 0.875rem;
  min-height: 5rem;
}

.reason-box {
  padding: 1rem;
  background: #f9fafb;
  border-left: 4px solid #000;
}

.reason-text {
  color: #374151;
  margin: 0;
}
</style>
