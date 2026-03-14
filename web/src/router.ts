import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";
import HomeView from "@/views/HomeView.vue";
import UploadView from "@/views/UploadView.vue";
import LoginView from "@/views/LoginView.vue";
import AdminReviewView from "@/views/AdminReviewView.vue";
import AlbumDetailView from "@/views/AlbumDetailView.vue";
import EditAlbumView from "@/views/EditAlbumView.vue";
import AboutView from "@/views/AboutView.vue";
import { useAuthStore } from "@/stores/auth";

const routes: RouteRecordRaw[] = [
  { path: "/", component: () => import("@/views/orbit/OrbitView.vue") },
  { path: "/music", component: HomeView },
  { path: "/about", component: AboutView },
  { path: "/artist=:artist/album=:album", component: AlbumDetailView },
  { path: "/artist=:artist/album=:album/edit", component: EditAlbumView },
  { path: "/upload", component: UploadView },
  { path: "/login", component: LoginView },
  { path: "/register", component: LoginView },
  {
    path: "/admin/review",
    component: AdminReviewView,
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
  // Orbit routes
  // Redirect /orbit to / for consistency since it's now the home page
  { path: "/orbit", redirect: "/" },
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
