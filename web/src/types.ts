export interface Artist {
  id: number;
  name: string;
  bio?: string;
  image_url?: string;
  created_at?: string;
  updated_at?: string;
}

export interface Album {
  id: number;
  title: string;
  year?: number;
  release_date?: string;
  cover_url?: string;
  cover_source?: 'local' | 's3';
  status: 'pending' | 'approved' | 'rejected';
  uploaded_by?: number;
  artists?: Artist[];
  created_at?: string;
  updated_at?: string;
}

export interface Song {
  id: number;
  title: string;
  artist: string;
  album: string;
  album_id: number;
  year: number;
  release_date: string;
  lyrics: string;
  audio_url: string;
  audio_source?: 'local' | 's3';
  cover_url: string;
  cover_source?: 'local' | 's3';
  track_number?: number;
  status: 'pending' | 'approved' | 'rejected';
  uploaded_by?: number;
  artists?: Artist[];
}

export interface SongCorrection {
  id: number;
  song_id: number;
  song?: Song;
  user_id?: number;
  user?: User;
  status: 'pending' | 'approved' | 'rejected';
  field_name: string;
  current_value: string;
  corrected_value: string;
  reason?: string;
  created_at: string;
  approved_at?: string;
  approved_by?: number;
  rejected_at?: string;
  rejected_by?: number;
}

export interface AlbumCorrection {
  id: number;
  album_id: number;
  album?: Album;
  user_id?: number;
  user?: User;
  status: 'pending' | 'approved' | 'rejected';
  original_title?: string;
  original_cover_url?: string;
  original_release_date?: string;
  original_artist_ids?: string;
  corrected_title?: string;
  corrected_cover_url?: string;
  corrected_cover_source?: 'local' | 's3';
  corrected_release_date?: string;
  corrected_artist_ids?: string;
  reason?: string;
  created_at: string;
  approved_at?: string;
  approved_by?: number;
  rejected_at?: string;
  rejected_by?: number;
}

export type RepeatMode = 'none' | 'one' | 'all';

export interface PlayerState {
  currentSong: Song | null;
  isPlaying: boolean;
  isShuffled: boolean;
  repeatMode: RepeatMode;
  volume: number;
  currentTime: number;
  duration: number;
}

export interface User {
  id?: number
  username: string
  email: string
  role?: 'user' | 'admin'
  display_name?: string
  avatar_url?: string
  bio?: string
  website?: string
  location?: string
  is_active?: boolean
  created_at?: string
  updated_at?: string
}

// ===== Blog Types =====

export interface Channel {
  id: number
  user_id: number
  user?: User
  name: string
  description?: string
  cover_url?: string
  created_at: string
  updated_at: string
}

export interface Collection {
  id: number
  channel_id: number
  channel?: Channel
  name: string
  description?: string
  cover_url?: string
  created_at: string
  updated_at: string
}

export interface Post {
  id: number
  user_id: number
  user?: User
  title: string
  content: string
  summary?: string
  cover_url?: string
  status: 'draft' | 'published'
  allow_comments: boolean
  pinned: boolean
  collections?: Collection[]
  likes_count?: number
  comments_count?: number
  created_at: string
  updated_at: string
}

export interface Comment {
  id: number
  post_id: number
  user_id: number
  user?: User
  content: string
  status: 'visible' | 'hidden'
  created_at: string
  updated_at: string
}

export interface Like {
  id: number
  user_id: number
  target_type: 'post' | 'comment'
  target_id: number
  created_at: string
}

export interface BookmarkFolder {
  id: number
  user_id: number
  name: string
  created_at: string
  updated_at: string
}

export interface Bookmark {
  id: number
  user_id: number
  post_id: number
  post?: Post
  bookmark_folder_id?: number
  created_at: string
}

// ===== Feed / Orbit Types =====

export interface FeedSource {
  id: number
  source_type: 'internal_user' | 'internal_channel' | 'internal_collection' | 'external_rss'
  source_id?: number
  rss_url?: string
  hash: string
  title?: string
  last_fetched_at?: string
  created_at: string
}

export interface Subscription {
  id: number
  user_id: number
  feed_source_id: number
  feed_source?: FeedSource
  title?: string
  created_at: string
}

export interface OrbitItem {
  id: number
  feed_source_id: number
  feed_source?: FeedSource
  guid: string
  title: string
  link: string
  summary: string
  author: string
  published_at: string
  fetched_at: string
}

// Unified timeline item returned by GET /api/feed/timeline
export interface TimelineItem {
  type: 'post' | 'orbit_item'
  post?: Post
  orbit_item?: OrbitItem
  published_at: string
}

// ===== Notification Types =====

export interface Notification {
  id: number
  user_id: number
  type: 'comment' | 'like' | 'bookmark' | 'system'
  content: string
  target_type?: string
  target_id?: number
  read_at?: string
  created_at: string
}

// ===== Profile Types =====

export interface UserProfile {
  id: number
  username: string
  display_name?: string
  avatar_url?: string
  bio?: string
  website?: string
  role: string
  followers_count?: number
  following_count?: number
  posts_count?: number
  created_at: string
}

export interface AuthState {
  token: string | null;
  user: User | null;
  isAuthenticated: boolean;
}
