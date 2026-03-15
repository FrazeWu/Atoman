import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";
import { useAuthStore } from "@/stores/auth";

const routes: RouteRecordRaw[] = [
  { path: "/", component: () => import("@/views/feed/FeedView.vue") },
  { path: "/music", component: () => import("@/views/HomeView.vue") },
  { path: "/about", component: () => import("@/views/AboutView.vue") },
  { path: "/artist=:artist/album=:album", component: () => import("@/views/AlbumDetailView.vue") },
  { path: "/artist=:artist/album=:album/edit", component: () => import("@/views/EditAlbumView.vue") },
  { path: "/upload", component: () => import("@/views/UploadView.vue") },
  { path: "/login", component: () => import("@/views/LoginView.vue") },
  { path: "/register", component: () => import("@/views/LoginView.vue") },
  {
    path: "/admin/review",
    component: () => import("@/views/AdminReviewView.vue"),
    beforeEnter: (to, from, next) => {
      const authStore = useAuthStore();
      if (authStore.user?.role === "admin") {
        next();
      } else {
        next("/");
      }
    },
  },
  // Blog routes
  { path: "/blog", component: () => import("@/views/blog/BlogHomeView.vue") },
  { path: "/blog/explore", component: () => import("@/views/blog/ExploreView.vue") },
  { path: "/blog/posts/:id", component: () => import("@/views/blog/PostDetailView.vue") },
  { 
    path: "/blog/posts/new", 
    component: () => import("@/views/blog/PostEditorView.vue"),
    beforeEnter: (to, from, next) => {
      const authStore = useAuthStore();
      if (authStore.isAuthenticated) next();
      else next("/login");
    }
  },
  { 
    path: "/blog/posts/:id/edit", 
    component: () => import("@/views/blog/PostEditorView.vue"),
    beforeEnter: (to, from, next) => {
      const authStore = useAuthStore();
      if (authStore.isAuthenticated) next();
      else next("/login");
    }
  },
  { path: "/blog/@:username", component: () => import("@/views/blog/ProfileView.vue") },
  { 
    path: "/blog/bookmarks", 
    component: () => import("@/views/blog/BookmarkView.vue"),
    beforeEnter: (to, from, next) => {
      const authStore = useAuthStore();
      if (authStore.isAuthenticated) next();
      else next("/login");
    }
  },
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
