# AGENTS.md - Atoman Coding Guidelines

This file provides guidelines for AI coding agents working on the Atoman codebase.

## Project Overview

Atoman is an open platform for freedom of speech, featuring blogs, forums, and content
subscriptions. It supports private/self-hosted deployment.

**Modules:**
- **feed** - RSS-based subscription feed
- **blog** - Personal blogging (short & long form, Markdown + LaTeX)
- **video** - IPFS-based video sharing (planned)
- **music** - IPFS-based music sharing (planned)

**Development Notes:**
- **临时逻辑修改**: 暂时不需要登录也可以修改文章内容 (Temporarily allow article editing without login)

**License:** GPL

## Tech Stack

### Frontend (web/)
- Vue 3.5 + Vite 7 + TypeScript 5.9
- State Management: Pinia 3
- UI Library: Naive UI
- Styling: Tailwind CSS v4
- Routing: Vue Router 4

### Backend (server/) - Planned
- Go + Gin + GORM
- PostgreSQL + Redis

### Future
- P2P/Tor distributed routing
- IPFS distributed storage

## Repository Structure

```
Atoman/
├── web/                    # Vue 3 frontend application
│   ├── src/
│   │   ├── assets/         # CSS, images, static assets
│   │   ├── components/     # Reusable Vue components
│   │   ├── views/          # Page/route components
│   │   ├── router/         # Vue Router configuration
│   │   ├── stores/         # Pinia stores
│   │   ├── App.vue         # Root component
│   │   └── main.ts         # Application entry
│   └── package.json
├── server/                 # Go backend (placeholder)
└── doc/                    # Documentation
```

## Development Commands

All frontend commands run from the `web/` directory:

```bash
cd web

# Install dependencies
npm install

# Development server (hot-reload)
npm run dev

# Production build (includes type-check)
npm run build

# Type checking only
npm run type-check

# Lint with auto-fix
npm run lint

# Format with Prettier
npm run format
```

**Node.js version:** `^20.19.0 || >=22.12.0`

### Testing

Testing framework is not yet configured. When added:
- Expected framework: Vitest
- Test file location: `src/**/__tests__/`
- Test file naming: `*.test.ts` or `*.spec.ts`

## Code Style Guidelines

### Page Style
feedly style

### Formatting (Prettier)

```json
{
  "semi": false,
  "singleQuote": true,
  "printWidth": 100
}
```

- **No semicolons** at end of statements
- **Single quotes** for strings
- **100 character** line width
- Run `npm run format` to apply

### TypeScript

- Prefer **type inference** over explicit annotations
- Use path alias `@/` for imports from `src/`
- Keep types minimal; add explicit types only when needed for clarity

```typescript
// Good - inference
const count = ref(0)
const items = ref<Item[]>([])

// Path alias
import { useCounterStore } from '@/stores/counter'
```

### Vue Components

**SFC Order:** `<template>` -> `<script setup lang="ts">` -> `<style scoped>`

```vue
<template>
  <!-- Template content -->
</template>

<script setup lang="ts">
// Composition API with TypeScript
</script>

<style scoped>
/* Component-scoped styles */
</style>
```

- Always use `<script setup lang="ts">`
- Always use `<style scoped>` unless global styles needed
- Use Composition API exclusively

### Import Ordering

1. Node built-ins (`node:url`, `node:path`)
2. Vue core (`vue`, `vue-router`, `pinia`)
3. Third-party libraries (`naive-ui`, etc.)
4. Local aliases (`@/components/...`, `@/stores/...`)
5. Relative imports (`./`, `../`)

```typescript
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { NButton } from 'naive-ui'
import Header from '@/components/Header.vue'
import { useCounterStore } from '@/stores/counter'
```

### Naming Conventions

| Item | Convention | Example |
|------|------------|---------|
| Vue components | PascalCase | `Header.vue`, `SideBar.vue` |
| View components | PascalCase + View suffix | `HomeView.vue`, `BlogView.vue` |
| TypeScript files | camelCase | `counter.ts`, `apiClient.ts` |
| Pinia stores | `use` + Name + `Store` | `useCounterStore` |
| Route names | lowercase | `'home'`, `'blog'`, `'feed'` |
| CSS classes | Tailwind utilities or kebab-case | `text-slate-900`, `my-class` |

### Pinia Stores

Use **Setup Store** syntax (Composition API style):

```typescript
import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

export const useExampleStore = defineStore('example', () => {
  // State
  const count = ref(0)

  // Getters
  const doubleCount = computed(() => count.value * 2)

  // Actions
  function increment() {
    count.value++
  }

  return { count, doubleCount, increment }
})
```

### Vue Router

- Home route: **direct import** (faster initial load)
- Other routes: **lazy load** with dynamic imports

```typescript
import HomeView from '../views/HomeView.vue'

const routes = [
  { path: '/', name: 'home', component: HomeView },
  { path: '/about', name: 'about', component: () => import('../views/AboutView.vue') },
]
```

### Styling with Tailwind CSS

- Use **Tailwind v4** utility classes directly in templates
- Common utilities: `flex`, `grid`, `p-4`, `text-slate-900`, `rounded-lg`
- Color palette: slate (e.g., `text-slate-600`, `bg-slate-50`)
- Responsive: `sm:`, `md:`, `lg:` prefixes
- Use CSS custom properties in `assets/base.css` for theming

```vue
<template>
  <div class="flex items-center gap-4 p-4 bg-white rounded-lg">
    <span class="text-lg font-semibold text-slate-900">Title</span>
  </div>
</template>
```

### Error Handling

- Use `try/catch` for async operations in store actions
- Handle errors gracefully with user-friendly messages
- Log errors for debugging

## Common Tasks

### Adding a New Component

1. Create `web/src/components/MyComponent.vue`
2. Use PascalCase filename
3. Follow SFC structure with `<script setup lang="ts">`

### Adding a New View/Route

1. Create `web/src/views/MyView.vue` (with `View` suffix)
2. Add route in `web/src/router/index.ts`
3. Use lazy loading: `component: () => import('../views/MyView.vue')`

### Adding a New Store

1. Create `web/src/stores/myStore.ts`
2. Use Setup Store syntax
3. Export as `useMyStore`

---

## Sub-module: Agents Configuration for Kanye Archive

### Overview
This document defines AI agents and tools to assist in building a Vue 3 website for collecting all Kanye West songs, featuring authentication, uploads, corrections, and a minimalist UI, with a clear definition of user roles and permissions.

### Database Schema (Updated 2026-01-18)

#### Current Architecture
The system uses an optimized schema with many-to-many relationships and typed correction tables.

**Core Tables:**
- `Users` - User accounts with role-based access
- `Artists` - Artist information
- `Albums` - Album metadata with status tracking
- `Songs` - Song metadata with status tracking

**Junction Tables (Many-to-Many):**
- `album_artists` - Albums can have multiple artists (collaboration support)
- `song_artists` - Songs can have multiple artists (featured artists support)

**Correction Tables (Typed):**
- `song_corrections` - Structured song corrections (field-specific)
- `album_corrections` - Structured album corrections (title, cover, release date, artists)

**Key Features:**
- File source tracking (`audio_source`, `cover_source`: 'local' or 's3')
- Audit trail (ApprovedBy, RejectedBy, ApprovedAt, RejectedAt)
- Status management ('pending', 'approved', 'rejected')

### User Roles and Permissions

To ensure content quality and proper management, the system implements a user role and permission mechanism.

*   **Anonymous Users**:
    *   **Allowed Actions**: Browse all approved songs and albums, play music.
    *   **Restrictions**: Cannot upload, modify content, submit corrections, or access any authenticated features.

*   **Regular Users (`user`)**:
    *   **Authentication**: Register and login to receive JWT.
    *   **Allowed Actions**: Browse all approved songs/albums, play music, manage personal profile.
    *   **Actions Requiring Approval**: Submit new song/album uploads, submit correction requests for existing content. These actions enter a "pending" status and require admin approval before becoming visible/effective.
    *   **Restrictions**: Cannot directly edit or delete public content, no access to admin features.

*   **Admin Users (`admin`)**:
    *   **Authentication**: Login to receive JWT with admin role information.
    *   **Allowed Actions**: All permissions of regular users.
    *   **Immediate Effect Actions**:
        *   **Direct creation, editing, and deletion of any songs/albums**: Admin operations do not require approval and take effect immediately.
        *   **Direct submission of uploads and corrections**: Content submitted by admins via upload or correction interfaces is marked as "approved" and effective immediately, bypassing the "pending" workflow.
    *   **Core Responsibilities**: Review and approve/reject all song/album uploads and correction requests submitted by regular users.
    *   **Management Features**: User management (role changes, account suspension), system configuration.

### Agents

#### Frontend Agent
- **Purpose**: Develop Vue 3 components, implement black/white minimalist UI with topbar, timeline, and player, adapting UI/UX based on user roles and content approval status.
- **Tools**: Vue Router, Pinia for state management, Tailwind CSS for styling, fetch API for HTTP calls.
- **Responsibilities**:
    *   Build `HomeView`, `AlbumDetailView`, `SongDetailView`, `UploadView`, `AdminReviewView`, `LoginView`, `EditAlbumView`, `AboutView`.
    *   Implement `AppTopbar` and `AudioPlayer` components.
    *   **Permissions and Roles**:
        *   Ensure `HomeView`, `AlbumDetailView`, `SongDetailView`, and `AudioPlayer` are accessible to anonymous users, displaying only `approved` content.
        *   Dynamically show/hide UI elements based on user role (anonymous, regular user, admin), such as upload buttons, edit/delete buttons, admin panel entry points.
        *   For regular users: After submitting uploads or corrections, clearly display "pending approval" status in the UI.
        *   For admins: After submitting uploads or making modifications, the UI should reflect that changes are effective immediately, without showing "pending approval" status.
        *   Build `AdminReviewView` to display all pending uploads and correction requests submitted by regular users, providing approve/reject interaction functionality, and showing the submitting user and modified information.
    *   Integrate user authentication status and role information into Pinia store for global access.
    *   **Type Definitions**: Use TypeScript interfaces for `Artist`, `Album`, `Song`, `SongCorrection`, `AlbumCorrection` matching backend schema.

#### Backend Agent
- **Purpose**: Create API endpoints for user auth, song/album management, media upload, corrections, and admin review, with robust role-based access control.
- **Tools**: Go (Gin framework), GORM for database interaction, SQLite (development) / MySQL (production) database, JWT for authentication, AWS SDK for S3-compatible cloud storage.
- **Responsibilities**:
    *   Implement auth routes (register, login, token validation).
    *   Implement CRUD operations for songs and albums.
    *   Integrate cloud storage for media file uploads (AWS S3-compatible).
    *   Implement correction submission and application logic.
    *   **User Management**:
        *   Add `role` field to `User` model (`user`, `admin`), include role information in JWT upon login.
        *   `create_admin` utility tool for creating admin users.
    *   **Permissions and Access Control**:
        *   Set endpoints like `GET /api/albums`, `GET /api/albums/:id`, `GET /api/songs` (non-admin lists) as publicly accessible, but returning only `approved` status content.
        *   Implement middleware to verify permissions for protected API routes based on role information in JWT.
    *   **Content Approval and Status Management**:
        *   Add `status` field to `Song`, `Album`, `SongCorrection`, `AlbumCorrection` models (`pending`, `approved`, `rejected`).
        *   **Regular User Operations**: When submitting uploads or corrections, content is set to `pending` status by default, and immediately updated in the database and S3.
        *   **Admin Operations**: Content submitted by admins via upload, correction, or direct modification interfaces is set to `approved` status and takes effect immediately.
        *   Provide admin-only API endpoints:
            *   Retrieve all pending uploads and correction requests submitted by regular users.
            *   Approve or reject pending uploads and corrections. If approved, update the status to 'approved' and apply changes (for corrections). If rejected, permanently delete the content from the database and associated files from S3.
    *   **Correction Endpoints**:
        *   `POST /api/corrections/song` - Submit song corrections
        *   `POST /api/corrections/album` - Submit album corrections
        *   `GET /api/admin/pending-song-corrections` - List pending song corrections
        *   `POST /api/admin/approve-song-correction/:id` - Approve song correction
        *   `POST /api/admin/reject-song-correction/:id` - Reject song correction
        *   `GET /api/admin/pending-album-corrections` - List pending album corrections
        *   `POST /api/admin/approve-album-correction/:id` - Approve album correction
        *   `POST /api/admin/reject-album-correction/:id` - Reject album correction

#### Data Agent
- **Purpose**: Collect and structure Kanye discography data.
- **Tools**: Web scraping (Puppeteer), APIs (Spotify, Genius), CSV/JSON processing.
- **Responsibilities**: Scrape song metadata, populate initial database.

#### Media Agent
- **Purpose**: Handle media storage, streaming, and optimization.
- **Tools**: AWS S3 SDK (via Backend Agent), CloudFront CDN, FFmpeg for audio processing (may be an external service or integrated tool).
- **Responsibilities**: Upload files to cloud, generate streaming URLs, ensure secure access, optimize audio for playback.

#### Testing Agent
- **Purpose**: Ensure code quality and functionality.
- **Tools**: Vitest for frontend unit tests, Playwright for frontend E2E tests, Go's testing framework for backend unit/integration tests.
- **Responsibilities**: Write tests for auth, uploads, playback, UI interactions, and role-based access control.

### Migration History

#### 2026-01-18: Schema Optimization
**Changes Made:**
1. Album-Artist relationship changed from one-to-many to many-to-many (supports collaboration albums)
2. Correction table split into `SongCorrection` and `AlbumCorrection` with typed fields
3. Added file source tracking (`audio_source`, `cover_source`) for local staging support
4. Added audit trail fields to correction tables

**Files Modified:**
- Backend: `models.go`, `main.go`, `songs.go`, `albums.go`, `admin.go`, `corrections.go`
- Frontend: `types.ts`, `AdminReviewView.vue`

**Admin Credentials:**
- Username: `admin`
- Email: `admin@kanyearchive.com`
- Password: `admin123`

### Usage
Agents work collaboratively: Frontend builds UI, Backend handles logic, Data populates content, Media manages files, Testing validates.

---

## Sub-module: Kanye Archive — UI/UX Design Spec

### Design Style
**极简主义 + 档案馆美学 (Minimalist Archive Aesthetic)**

#### Color System
- 主色调：纯黑 (#000000) + 纯白 (#FFFFFF)
- 辅助色：灰度系统 (gray-100, gray-400, gray-500, gray-600, gray-700)
- 强调色：绿色 (text-green-600) 用于状态显示，红色 (text-red-600) 用于管理入口

#### Typography System
- 主标题：`text-5xl ~ text-6xl`, `font-black`, `tracking-tighter` (极紧字距)
- 次级标题：`text-2xl ~ text-4xl`, `font-black`, `tracking-tight`
- 正文：`text-sm ~ text-xl`, `font-medium` / `font-bold`
- 标签/标识：`text-xs`, `font-black`, `uppercase`, `tracking-widest` (极宽字距)

#### Visual Language
- **边框美学**：所有卡片、按钮、输入框使用 `border-2 border-black` 或 `border-4 border-black`
- **投影效果**：使用硬投影 `shadow-[10px_10px_0px_0px_rgba(0,0,0,1)]` 替代柔和阴影
- **灰度图像**：所有封面图片使用 `grayscale` 滤镜，统一视觉调性
- **过渡动效**：使用 `transition-all` + `duration-300 / duration-500`，强调状态变化
- **悬停反转**：按钮悬停时执行黑白反转 `hover:bg-white hover:text-black`

#### Component Style Specs

**Buttons**
```css
/* Primary button */
bg-black text-white border-2 border-black px-8 py-4 font-black uppercase tracking-widest hover:bg-white hover:text-black transition-all

/* Secondary button */
border-2 border-black px-4 py-2 font-black text-xs uppercase tracking-widest hover:bg-black hover:text-white transition-colors
```

**Inputs**
```css
w-full bg-white border-2 border-black p-4 focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] outline-none transition-all
```

**Cards**
```css
bg-white border-2 border-black p-6 hover:shadow-[10px_10px_0px_0px_rgba(0,0,0,1)] transition-all duration-300
```

### Layout Structure

#### 1. Topbar (Sticky)
- **Position**: `sticky top-0 z-50`
- **Height**: `h-16` (64px)
- **Background**: white + `border-b-2 border-black`
- **Content**:
  - Left: `KANYE ARCHIVE` brand (`text-2xl font-black tracking-tighter`)
  - Right: nav links + user menu (Timeline | Upload | Admin Review | Username dropdown / Login)

#### 2. Main Content Areas

**HomeView (Timeline)**
- Layout: vertical timeline + card display
- Timeline line: `absolute left-1/3`, `w-1 bg-black` through full page
- Year labels: `bg-black text-white px-4 py-1 font-bold tracking-widest`
- Timeline nodes: `w-6 h-6 rounded-full border-4 border-white bg-black`; current playing: `scale-125`
- Song cards:
  - Position: `ml-[calc(33.333%+2rem)] w-[calc(66.666%-2rem)]`
  - Cover: `w-32 h-32 border-2 border-black grayscale`
  - Hover: `hover:shadow-[10px_10px_0px_0px_rgba(0,0,0,1)]`

**SongDetailView**
- Layout: two-column grid `grid grid-cols-1 md:grid-cols-2 gap-16`
- Left: large cover `border-4 border-black grayscale shadow-2xl`
- Right:
  - Archive number label: `YE-0001` format
  - Title + artist info
  - Action buttons: play + correct data
  - Lyrics preview: `italic text-gray-700`
  - Tech specs table: release year, audio format, archive number, status

**UploadView**
- Layout: centered single-column `max-w-2xl mx-auto`
- Title: `text-4xl font-black tracking-tighter`
- File upload area: `border-2 border-dashed border-black p-12`, hover `bg-gray-100`

**LoginView**
- Layout: `min-h-[calc(100vh-64px)] flex items-center justify-center`
- Form container: `max-w-md border-2 border-black p-12 shadow-[20px_20px_0px_0px_rgba(0,0,0,1)]`
- Inputs: `focus:bg-gray-50`

#### 3. AudioPlayer (Fixed Bottom)
- **Position**: `fixed bottom-0 w-full z-50`
- **Background**: white + `border-t-2 border-black`
- **Layout**: three-column (song info | playback controls | volume control)
- **Progress bar**:
  - Container: `h-1 bg-gray-200`
  - Progress: `bg-black` dynamic width
  - Interaction: click to seek
- **Controls**:
  - Play/pause: `px-6 py-2 border-2 border-black bg-black text-white`
  - Secondary: `px-3 py-1 border border-black`
  - Active state: `bg-black text-white`

### Animation System
- Fade-in: `animate-in fade-in slide-in-from-bottom-4 duration-500`
- Loading: `animate-pulse`
- Dropdown: `animate-in fade-in slide-in-from-top-1 duration-200`
- Rotating icon: `transition-transform duration-200 rotate-180`

### Font Weight Usage
- `font-black` (900): titles, labels, buttons, brand name
- `font-bold` (700): secondary headings, emphasized text
- `font-medium` (500): body text, descriptive text

### Spacing System
- Page margin: `px-8` (horizontal) + `py-20` (vertical)
- Card padding: `p-6` (content cards) / `p-12` (form containers)
- Element spacing: `gap-4`, `gap-8`, `gap-16` (increasing by hierarchy)

---

## Sub-module: Kanye Archive — Functional Requirements

### 1. User Authentication
- Email + password login/registration system
- Anonymous users: can browse and play songs
- Logged-in users: can upload, edit, and submit corrections
- Admin role: access to review queue

### 2. Database Architecture
- **Artists table**: id, name, bio, image_url
- **Albums table**: id, title, year, cover_url, artist_id
- **Songs table**: id, title, year, lyrics, audio_url, cover_url, album_id, uploaded_by
- **SongArtists junction**: song_id, artist_id, role (many-to-many)
- **Users table**: id, username, email, password_hash, role
- **Corrections table**: id, song_id, field_name, current_value, corrected_value, reason, user_id, status

**Relationships**:
- Artist (1) ↔ Album (N)
- Album (1) ↔ Song (N)
- Artist (N) ↔ Song (M): supports feat, collaborations
- User (1) → Song (N): uploaded songs

*(Note: Current schema in AGENTS.md above uses the updated many-to-many album-artist model — refer to Database Schema section for the canonical schema.)*

### 3. Song/Album Management
- Auto-create or link artists and albums on upload
- Timeline view sorted by year
- Detail page: song info, playback options, tech specs
- Archive number system: `YE-{id.padStart(4, '0')}` format

### 4. Correction Feature
- Logged-in users can submit song info corrections
- Support re-uploading media files if errors exist
- Corrections enter admin review queue after submission
- Correction status: pending / approved / rejected

### 5. Media Storage
- AWS S3 (or S3-compatible, e.g. MinIO) for audio files and cover images
- **S3 path format**: `music/{ArtistName}/{AlbumName}/{FileName}`
- Streaming support with secure access
- Path safety: replace slashes on upload

### 6. Backend API
- **Stack**: Go (Gin) + GORM
- **Database**: PostgreSQL (production), SQLite (development)
- **Auth**: JWT Token
- **Endpoints**:
  - `POST /api/auth/register` — user registration
  - `POST /api/auth/login` — user login
  - `GET /api/songs` — all songs (with Preload associations)
  - `GET /api/songs/:id` — single song detail
  - `POST /api/songs` — upload new song (requires auth)
  - `/api/corrections` — correction management

### 7. Audio Player
- HTML5 Audio API
- Features: play/pause, prev/next, progress bar scrub, volume control, shuffle, loop modes (off / list loop / single loop)
- State managed in Pinia store
- Fixed at page bottom, does not affect content scroll

### 8. Admin Review System
- Admins view pending uploads and corrections
- Approve / reject actions
- Status tracking and notifications

---

## Sub-module: Kanye Archive — Tech Stack & Priority

### Frontend
- Framework: Vue 3 (Composition API)
- Language: TypeScript
- Build tool: Vite
- State management: Pinia
- Routing: Vue Router
- Styling: Tailwind CSS
- Animation: Tailwind animation utilities

### Backend
- Language: Go
- Framework: Gin
- ORM: GORM
- Auth: JWT
- File upload: AWS SDK for Go (S3)
- Database: PostgreSQL (production) / SQLite (development)

### Infrastructure
- Cloud storage: AWS S3 / S3-compatible (MinIO for dev)
- Deployment: Docker Compose
- Testing: Vitest (frontend), Go testing (backend)

### Priority
- **High**: user auth, database schema, song management, backend API, S3 upload
- **Medium**: UI detail polish, correction feature, admin review system
- **Low**: performance optimization, SEO, internationalization