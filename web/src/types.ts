export type MusicEntryStatus = 'open' | 'confirmed' | 'disputed'

export interface ArtistAlias {
  id: string
  artist_id: string
  alias: string
  is_main_name: boolean
  created_at: string
}

export interface LyricAnnotation {
  id: string
  song_id: string
  line_number: number
  content: string
  user?: User
  created_at: string
}

export interface Artist {
  id: number;
  name: string;
  bio?: string;
  image_url?: string;
  nationality?: string;
  birth_year?: number;
  death_year?: number;
  members?: string;
  entry_status?: MusicEntryStatus;
  aliases?: ArtistAlias[];
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
  status: 'open' | 'closed' | 'pending' | 'approved' | 'rejected';
  album_type?: string;
  entry_status?: MusicEntryStatus;
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
  status: 'open' | 'closed' | 'pending' | 'approved' | 'rejected';
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

export interface ArtistCorrection {
  id: string
  artist_id: string
  artist?: Artist
  user_id?: string
  user?: User
  description: string
  reason?: string
  status: 'pending' | 'approved' | 'rejected'
  approved_by?: string
  approved_at?: string
  created_at: string
  updated_at: string
}

export type RepeatMode = 'none' | 'one' | 'all'

export interface User {
  id?: number
  uuid?: string
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
  id: string
  user_id: string
  user?: User
  name: string
  slug: string
  description?: string
  cover_url?: string
  is_default?: boolean
  created_at: string
  updated_at: string
}

export interface Collection {
  id: string
  channel_id: string
  channel?: Channel
  name: string
  is_default?: boolean
  description?: string
  cover_url?: string
  created_at: string
  updated_at: string
}

export interface Post {
  id: string
  user_id: string
  user?: User
  channel_id?: string
  channel?: Channel
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

export interface BlogDraft {
  id: string
  user_id: string
  context_key: string
  source_post_id?: string
  title: string
  content: string
  summary?: string
  cover_url?: string
  allow_comments: boolean
  channel_id?: string
  collection_ids: string[]
  created_at: string
  updated_at: string
}

export interface Comment {
  id: string
  post_id: string
  user_id: string
  user?: User
  content: string
  status: 'visible' | 'hidden'
  created_at: string
  updated_at: string
}

export interface Like {
  id: string
  user_id: string
  target_type: 'post' | 'comment'
  target_id: string
  created_at: string
}

// ===== Feed / Types =====

export interface FeedSource {
  id: string
  source_type: 'internal_user' | 'internal_channel' | 'internal_collection' | 'external_rss'
  source_id?: string
  rss_url?: string
  hash: string
  title?: string
  last_fetched_at?: string
  created_at: string
}

export interface SubscriptionGroup {
  id: string
  user_id: string
  name: string
  created_at: string
  updated_at: string
}

export interface Subscription {
  id: string
  user_id: string
  feed_source_id: string
  feed_source?: FeedSource
  title?: string
  subscription_group_id?: string
  subscription_group?: SubscriptionGroup
  health_status?: 'healthy' | 'warning' | 'error'
  error_message?: string
  last_checked?: string
  created_at: string
}

export interface FeedItem {
  id: string
  feed_source_id: string
  feed_source?: FeedSource
  guid: string
  title: string
  link: string
  summary: string
  author: string
  published_at: string
  fetched_at: string
  enclosure_url?: string
  enclosure_type?: string
  duration?: string
  image_url?: string
  is_duplicate?: boolean
  duplicate_count?: number
  duplicate_of_id?: string
  duplicate_sources?: string[]
  is_starred?: boolean
}

// Unified timeline item returned by GET /api/feed/timeline
export interface TimelineItem {
  type: 'post' | 'feed_item' | 'orbit_item'
  post?: Post
  feed_item?: FeedItem
  orbit_item?: OrbitItem
  published_at: string
  is_read: boolean
}

// ===== Bookmark Types =====

export interface BookmarkFolder {
  id: string
  user_id: string
  name: string
  created_at: string
  updated_at: string
}

export interface Bookmark {
  id: string
  user_id: string
  post_id: string
  post?: Post
  bookmark_folder_id?: string
  bookmark_folder?: BookmarkFolder
  created_at: string
}

// ===== Orbit Types =====

export interface OrbitItem {
  id: string
  feed_source_id: string
  feed_source?: FeedSource
  guid: string
  title: string
  link: string
  summary: string
  author: string
  published_at: string
  fetched_at: string
  enclosure_url?: string
  enclosure_type?: string
  duration?: string
  image_url?: string
  is_starred?: boolean
}

// ===== Forum Types =====

export interface ForumCategory {
  id: string
  name: string
  description?: string
  color: string
  topic_count?: number
  created_at: string
}

export interface ForumTopic {
  id: string
  user_id: string
  user?: User
  category_id: string
  category?: ForumCategory
  title: string
  content: string          // raw Markdown
  tags: string[]
  pinned: boolean
  featured: boolean
  closed: boolean
  reply_count: number
  like_count: number
  view_count: number
  last_reply_at?: string
  is_liked: boolean
  is_bookmarked: boolean
  created_at: string
  updated_at: string
}

export interface ForumReply {
  id: string
  topic_id: string
  user_id: string
  user?: User
  parent_reply_id?: string // quoted reply id
  content: string          // raw Markdown
  path: string
  floor_number: number
  like_count: number
  is_liked: boolean
  created_at: string
  updated_at: string
}

export interface ForumDraft {
  id?: string
  context_key: string
  title?: string
  content: string
  tags?: string
  updated_at?: string
}

// ===== Profile Types =====

export interface UserProfile {
  id: number
  uuid: string
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

// ===== Debate Types =====

export type DebateStatus = 'open' | 'concluded' | 'archived'
export type ArgumentType = 'support' | 'oppose' | 'neutral' | 'evidence' | 'question' | 'counter'

export interface Debate {
  id: string
  user_id: string
  user?: User
  title: string
  description: string
  content: string
  status: DebateStatus
  tags: string[]
  view_count: number
  argument_count: number
  vote_count: number
  conclusion_type?: 'yes' | 'no' | 'inconclusive' | ''
  conclusion_summary?: string
  conclude_vote_count?: number
  conclude_threshold?: number
  created_at: string
  updated_at: string
  concluded_at?: string
}

export interface Argument {
  id: string
  debate_id: string
  debate?: Debate
  parent_id?: string // quoted argument id
  parent?: Argument
  user_id: string
  user?: User
  content: string
  argument_type: ArgumentType
  vote_count: number
  references?: Argument[]
  referenced_debates?: Debate[]
  is_concluded: boolean
  conclusion?: string
  source_url?: string
  source_title?: string
  source_excerpt?: string
  is_folded?: boolean
  fold_note?: string
  created_at: string
  updated_at: string
}

export interface DebateVote {
  id: string
  argument_id: string
  argument?: Argument
  user_id: string
  user?: User
  vote_type: number // +1 or -1
  created_at: string
  updated_at: string
}

export interface VoteHistory {
  id: string
  argument_id: string
  user_id: string
  old_vote_type: number
  new_vote_type: number
  created_at: string
}

export interface TimelineEvent {
  id: string
  user_id: string
  user?: User
  title: string
  description: string
  content: string
  event_date: string
  end_date?: string
  location: string
  latitude?: number
  longitude?: number
  source: string
  category: string
  tags: string[]
  is_public: boolean
  created_at: string
  updated_at: string
}

export interface TimelinePerson {
  id: string
  user_id: string
  user?: User
  name: string
  bio: string
  birth_date?: string
  death_date?: string
  tags: string[]
  is_public: boolean
  locations?: PersonLocation[]
  created_at: string
  updated_at: string
}

export interface PersonLocation {
  id: string
  person_id: string
  date: string
  end_date?: string
  place_name: string
  latitude: number
  longitude: number
  source: string
  note: string
  created_at: string
  updated_at: string
}

export interface PodcastEpisode {
  id: string
  post_id: string
  post?: Post
  channel_id: string
  channel?: Channel
  audio_url: string
  duration_sec: number
  episode_cover_url: string
  season_number: number
  episode_number: number
  created_at: string
  updated_at: string
}
