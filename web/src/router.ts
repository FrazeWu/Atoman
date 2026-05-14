import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  { path: '/', redirect: '/feed' },
  { path: '/music', component: () => import('@/views/music/HomeView.vue') },
  { path: '/music/artists', component: () => import('@/views/music/HomeView.vue') },
  { path: '/about', component: () => import('@/views/music/AboutView.vue') },
  // Album routes (canonical, UUID-based)
  { path: '/music/albums/new', component: () => import('@/views/music/UploadView.vue'), meta: { requiresAuth: true } },
  { path: '/music/albums/:albumId', component: () => import('@/views/music/AlbumDetailView.vue') },
  { path: '/music/albums/:albumId/edit', component: () => import('@/views/music/EditAlbumView.vue'), meta: { requiresAuth: true } },
  { path: '/music/albums/:albumId/history', component: () => import('@/views/music/AlbumHistoryView.vue') },
  { path: '/music/albums/:albumId/discussion', component: () => import('@/views/music/AlbumDiscussionView.vue') },
  // Song-level wiki routes
  { path: '/music/songs/:songId/history', component: () => import('@/views/music/SongHistoryView.vue') },
  { path: '/music/songs/:songId/discussion', component: () => import('@/views/music/SongDiscussionView.vue') },
  // Artist routes (canonical, ID-based)
  {
    path: '/music/artists/new',
    component: () => import('@/views/music/AddArtistView.vue'),
    meta: { requiresAuth: true },
  },
  { path: '/music/artists/:artistId', component: () => import('@/views/music/ArtistDetailView.vue') },
  {
    path: '/music/artists/:artistId/edit',
    component: () => import('@/views/music/ArtistEditView.vue'),
    meta: { requiresAuth: true },
  },
  { path: '/music/artists/:artistId/history', component: () => import('@/views/music/ArtistHistoryView.vue') },
  { path: '/music/artists/:artistId/discussion', component: () => import('@/views/music/ArtistDiscussionView.vue') },
  // Legacy aliases for backward compatibility
  { path: '/music/contribute', redirect: '/music/albums/new' },
  { path: '/login', component: () => import('@/views/auth/LoginView.vue') },
  { path: '/register', component: () => import('@/views/auth/LoginView.vue') },
  {
    path: '/music/admin/review',
    component: () => import('@/views/music/AdminReviewView.vue'),
    meta: { requiresAuth: true, requiresAdmin: true },
  },
  { path: '/blog', component: () => import('@/views/blog/BlogHomeView.vue') },
  // /blog/explore redirects to /blog (discovery page merged)
  { path: '/blog/explore', redirect: '/blog' },
  { path: '/explore', redirect: '/blog' },
  {
    path: '/blog/bookmarks',
    component: () => import('@/views/blog/BookmarkView.vue'),
    meta: { requiresAuth: true },
  },
  { path: '/blog/manage', component: () => import('@/views/blog/BlogManageView.vue'), meta: { requiresAuth: true } },
  { path: '/channels', component: () => import('@/views/blog/ChannelManageView.vue'), meta: { requiresAuth: true } },
  // New slug-based channel routes (canonical)
  { path: '/channel/:slug', component: () => import('@/views/blog/ChannelView.vue') },
  {
    path: '/channel/:slug/manage',
    component: () => import('@/views/blog/ChannelManageDetailView.vue'),
    meta: { requiresAuth: true },
  },
  // Legacy ID-based routes kept for compatibility
  { path: '/blog/channel/:id', component: () => import('@/views/blog/ChannelView.vue') },
  { path: '/collections', component: () => import('@/views/blog/CollectionManageView.vue'), meta: { requiresAuth: true } },
  { path: '/collection/:id', component: () => import('@/views/blog/CollectionView.vue') },
  { path: '/editor/:id?', component: () => import('@/views/blog/PostEditorView.vue'), meta: { requiresAuth: true } },
  { path: '/post/:id', component: () => import('@/views/blog/PostDetailView.vue') },
  {
    path: '/post/new',
    component: () => import('@/views/blog/PostEditorView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/post/:id/edit',
    component: () => import('@/views/blog/PostEditorView.vue'),
    meta: { requiresAuth: true },
  },
  { path: '/user/:username', component: () => import('@/views/blog/ProfileView.vue') },
  {
    path: '/blog/settings',
    component: () => import('@/views/blog/BlogSettingsView.vue'),
    meta: { requiresAuth: true },
  },
  { path: '/feed', component: () => import('@/views/feed/FeedView.vue'), meta: { requiresAuth: true } },
  { path: '/feed/stats', component: () => import('@/views/feed/FeedStatsView.vue'), meta: { requiresAuth: true } },
  { path: '/feed/item/:id', component: () => import('@/views/feed/FeedItemDetailView.vue'), meta: { requiresAuth: true } },
  { path: '/feed/starred', component: () => import('@/views/feed/FeedStarredView.vue'), meta: { requiresAuth: true } },
  { path: '/feed/reading-list', component: () => import('@/views/feed/FeedReadingListView.vue'), meta: { requiresAuth: true } },
  { path: '/forum', component: () => import('@/views/forum/ForumHomeView.vue') },
  { path: '/forum/search', component: () => import('@/views/forum/ForumSearchView.vue') },
  { path: '/topic/:id', component: () => import('@/views/forum/ForumTopicView.vue') },
  {
    path: '/forum/new',
    component: () => import('@/views/forum/ForumNewTopicView.vue'),
    meta: { requiresAuth: true },
  },
  { path: '/debate', component: () => import('@/views/debate/DebateHomeView.vue') },
  { path: '/debate/:id', component: () => import('@/views/debate/DebateTopicView.vue') },
  { path: '/timeline', component: () => import('@/views/timeline/TimelineHomeView.vue') },
  { path: '/timeline/persons', component: () => import('@/views/timeline/PersonListView.vue') },
  { path: '/timeline/persons/:id', component: () => import('@/views/timeline/PersonMapView.vue') },
  { path: '/:pathMatch(.*)*', component: () => import('@/views/NotFoundView.vue') },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(_to, _from, savedPosition) {
    return savedPosition ?? { top: 0 }
  },
})

router.beforeEach((to) => {
  const authStore = useAuthStore()
  const hasValidSession = authStore.validateSession()

  if (to.meta.requiresAuth && !hasValidSession) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }

  if (to.meta.requiresAdmin && authStore.user?.role !== 'admin') {
    return '/feed'
  }
})

export default router
