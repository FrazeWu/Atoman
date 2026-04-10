import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";
import { useAuthStore } from "@/stores/auth";

const routes: RouteRecordRaw[] = [
  { path: "/", component: () => import("@/views/feed/FeedView.vue") },
  { path: "/music", component: () => import("@/views/music/HomeView.vue") },
  { path: "/music/artists/:artistName", component: () => import("@/views/music/HomeView.vue") },
  { path: "/about", component: () => import("@/views/music/AboutView.vue") },
  { path: "/music/albums/:albumId", component: () => import("@/views/music/AlbumDetailView.vue") },
  { path: "/music/artists/:artistName/albums/:albumId", component: () => import("@/views/music/AlbumDetailView.vue") },
  { path: "/music/albums/:albumId/history", component: () => import("@/views/music/AlbumHistoryView.vue") },
  { path: "/music/albums/:albumId/discussion", component: () => import("@/views/music/AlbumDiscussionView.vue") },
  {
    path: "/music/albums/:albumId/edit",
    component: () => import("@/views/music/EditAlbumView.vue"),
    beforeEnter: (to, from, next) => {
      const authStore = useAuthStore()
      if (authStore.isAuthenticated) next()
      else next("/login")
    },
  },
  {
    path: "/music/artists/:artistName/albums/:albumId/edit",
    component: () => import("@/views/music/EditAlbumView.vue"),
    beforeEnter: (to, from, next) => {
      const authStore = useAuthStore()
      if (authStore.isAuthenticated) next()
      else next("/login")
    },
  },
  {
    path: "/music/artists/:artistName/albums/:albumId/edit",
    redirect: to => `/music/albums/${to.params.albumId}/edit`,
  },
  {
    path: "/music/contribute",
    component: () => import("@/views/music/UploadView.vue"),
    beforeEnter: (to, from, next) => {
      const authStore = useAuthStore()
      if (authStore.isAuthenticated) next()
      else next("/login")
    },
  },
  {
    path: "/music/artists/new",
    component: () => import("@/views/music/AddArtistView.vue"),
    beforeEnter: (to, from, next) => {
      const authStore = useAuthStore()
      if (authStore.isAuthenticated) next()
      else next("/login")
    },
  },
  { path: "/login", component: () => import("@/views/auth/LoginView.vue") },
  { path: "/register", component: () => import("@/views/auth/LoginView.vue") },
  {
    path: "/music/admin/review",
    component: () => import("@/views/music/AdminReviewView.vue"),
    beforeEnter: (to, from, next) => {
      const authStore = useAuthStore();
      if (authStore.user?.role === "admin") {
        next();
      } else {
        next("/");
      }
    },
  },
  // Blog routes - simplified to single level
  { path: "/blog", component: () => import("@/views/blog/BlogHomeView.vue") },
  { path: "/blog/manage", component: () => import("@/views/blog/BlogManageView.vue") },
  { path: "/channels", component: () => import("@/views/blog/ChannelManageView.vue"), beforeEnter: (to, from, next) => {
    const authStore = useAuthStore();
    if (authStore.isAuthenticated) next();
    else next("/login");
  } },
  { path: "/channel/:id", component: () => import("@/views/blog/ChannelView.vue") },
  { path: "/collections", component: () => import("@/views/blog/CollectionManageView.vue") },
  { path: "/collection/:id", component: () => import("@/views/blog/CollectionView.vue") },
  { path: "/editor/:id?", component: () => import("@/views/blog/PostEditorView.vue") },
  { path: "/explore", component: () => import("@/views/blog/ExploreView.vue") },
  { path: "/post/:id", component: () => import("@/views/blog/PostDetailView.vue") },
  { 
    path: "/post/new", 
    component: () => import("@/views/blog/PostEditorView.vue"),
    beforeEnter: (to, from, next) => {
      const authStore = useAuthStore();
      if (authStore.isAuthenticated) next();
      else next("/login");
    }
  },
  { 
    path: "/post/:id/edit", 
    component: () => import("@/views/blog/PostEditorView.vue"),
    beforeEnter: (to, from, next) => {
      const authStore = useAuthStore();
      if (authStore.isAuthenticated) next();
      else next("/login");
    }
  },
  { path: "/user/:username", component: () => import("@/views/blog/ProfileView.vue") },
  {
    path: "/blog/settings",
    component: () => import("@/views/blog/BlogSettingsView.vue"),
    beforeEnter: (to, from, next) => {
      const authStore = useAuthStore();
      if (authStore.isAuthenticated) next();
      else next("/login");
    }
  },
  {
    path: "/blog/notifications",
    component: () => import("@/views/blog/NotificationsView.vue"),
    beforeEnter: (to, from, next) => {
      const authStore = useAuthStore();
      if (authStore.isAuthenticated) next();
      else next("/login");
    }
  },
  // Feed routes
  { path: "/feed", component: () => import("@/views/feed/FeedView.vue") },
  { path: "/feed/item/:id", component: () => import("@/views/feed/FeedItemDetailView.vue") },
  // Forum routes - simplified
  { path: "/forum", component: () => import("@/views/forum/ForumHomeView.vue") },
  { path: "/topic/:id", component: () => import("@/views/forum/ForumTopicView.vue") },
  {
    path: "/forum/new",
    component: () => import("@/views/forum/ForumNewTopicView.vue"),
    beforeEnter: (to, from, next) => {
      const authStore = useAuthStore();
      if (authStore.isAuthenticated) next();
      else next("/login");
    },
  },
  // Debate routes
  { path: "/debate", component: () => import("@/views/debate/DebateHomeView.vue") },
  { path: "/debate/:id", component: () => import("@/views/debate/DebateTopicView.vue") },
  { path: "/:pathMatch(.*)*", component: () => import("@/views/NotFoundView.vue") },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition;
    } else {
      return { top: 0 };
    }
  },
});

export default router;
