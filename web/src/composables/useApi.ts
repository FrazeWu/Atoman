export function useApiUrl() {
  return import.meta.env.VITE_API_URL || '/api';
}

export function useApi() {
  const apiUrl = useApiUrl();

  return {
    url: apiUrl,
    songs: `${apiUrl}/songs`,
    song: (id: number | string) => `${apiUrl}/songs/${id}`,
    albums: `${apiUrl}/albums`,
    album: (id: number | string) => `${apiUrl}/albums/${id}`,
    artists: `${apiUrl}/artists`,
    corrections: `${apiUrl}/corrections`,

    music: {
      albums: `${apiUrl}/albums`,
      album: (id: number | string) => `${apiUrl}/albums/${id}`,
      albumRevisions: (id: number | string) => `${apiUrl}/albums/${id}/revisions`,
      albumRevision: (id: number | string, version: number | string) => `${apiUrl}/albums/${id}/revisions/${version}`,
      albumRevisionDiff: (id: number | string) => `${apiUrl}/albums/${id}/revisions/diff`,
      albumRevert: (id: number | string, version: number | string) => `${apiUrl}/albums/${id}/revert/${version}`,
      albumDiscussions: (id: number | string) => `${apiUrl}/albums/${id}/discussions`,
      albumEntryStatus: (id: number | string) => `${apiUrl}/albums/${id}/status`,
      albumProtection: (id: number | string) => `${apiUrl}/albums/${id}/protection`,
      artists: `${apiUrl}/artists`,
      artist: (id: number | string) => `${apiUrl}/artists/${id}`,
      artistRevisions: (id: number | string) => `${apiUrl}/artists/${id}/revisions`,
      artistAliases: (id: number | string) => `${apiUrl}/artists/${id}/aliases`,
      artistEntryStatus: (id: number | string) => `${apiUrl}/artists/${id}/status`,
      artistDiscussions: (id: number | string) => `${apiUrl}/artists/${id}/discussions`,
      songAnnotations: (id: number | string) => `${apiUrl}/songs/${id}/annotations`,
      adminMusicReview: `${apiUrl}/admin/music/entries`,
      adminMusicConfirm: (id: number | string, type: 'album' | 'artist') =>
        type === 'album' ? `${apiUrl}/albums/${id}/status` : `${apiUrl}/artists/${id}/status`,
    },
    
    blog: {
      channels: `${apiUrl}/blog/channels`,
      channel: (id: number | string) => `${apiUrl}/blog/channels/${id}`,
      channelEnsureDefault: `${apiUrl}/blog/channels/ensure-default`,
      channelCollections: (id: number | string) => `${apiUrl}/blog/channels/${id}/collections`,
      channelBySlug: (slug: string) => `${apiUrl}/blog/channels/slug/${slug}`,
      channelCollectionsBySlug: (slug: string) => `${apiUrl}/blog/channels/slug/${slug}/collections`,
      collections: `${apiUrl}/blog/collections`,
      collection: (id: number | string) => `${apiUrl}/blog/collections/${id}`,
      
      posts: `${apiUrl}/blog/posts`,
      post: (id: number | string) => `${apiUrl}/blog/posts/${id}`,
      postPublish: (id: number | string) => `${apiUrl}/blog/posts/${id}/publish`,
      postUnpublish: (id: number | string) => `${apiUrl}/blog/posts/${id}/unpublish`,
      postPin: (id: number | string) => `${apiUrl}/blog/posts/${id}/pin`,
      postUnpin: (id: number | string) => `${apiUrl}/blog/posts/${id}/unpin`,
      draft: `${apiUrl}/blog/drafts`,
      drafts: `${apiUrl}/blog/posts/drafts`,
      postCollections: (id: number | string) => `${apiUrl}/blog/posts/${id}/collections`,
      postCollection: (id: number | string, collectionId: number | string) => `${apiUrl}/blog/posts/${id}/collections/${collectionId}`,
      uploadImage: `${apiUrl}/blog/upload-image`,
      
      comments: `${apiUrl}/blog/comments`,
      postComments: (id: number | string) => `${apiUrl}/blog/posts/${id}/comments`,
      
      likes: `${apiUrl}/blog/likes`,
      postLikesCount: (id: number | string) => `${apiUrl}/blog/posts/${id}/likes/count`,
      
      explore: `${apiUrl}/blog/explore`,
      bookmarkFolders: `${apiUrl}/blog/bookmark-folders`,
      bookmarkFolder: (id: number | string) => `${apiUrl}/blog/bookmark-folders/${id}`,
      bookmarks: `${apiUrl}/blog/bookmarks`,
    },
    
    auth: {
      register: `${apiUrl}/auth/register`,
      login: `${apiUrl}/auth/login`,
      sendVerification: `${apiUrl}/auth/send-verification`,
      verifyEmail: `${apiUrl}/auth/verify-email`,
    },
    
    users: {
      me: `${apiUrl}/users/me`,
      settings: `${apiUrl}/users/me`,          // profile update (display_name, bio, etc)
      meSettings: `${apiUrl}/users/me/settings`,
      profile: (username: string) => `${apiUrl}/users/by-username/${username}`,
      follow: (userUuid: string) => `${apiUrl}/users/${userUuid}/follow`,
      followers: (userUuid: string) => `${apiUrl}/users/${userUuid}/followers`,
      following: (userUuid: string) => `${apiUrl}/users/${userUuid}/following`,
    },
    
    feed: {
      subscriptions: `${apiUrl}/feed/subscriptions`,
      subscription: (id: number) => `${apiUrl}/feed/subscriptions/${id}`,
      timeline: `${apiUrl}/feed/timeline`,
      rss: (username: string) => `${apiUrl}/feed/rss/${username}`,
    },

    notifications: {
      list: `${apiUrl}/notifications`,
      unreadCount: `${apiUrl}/notifications/unread-count`,
      markRead: (id: string) => `${apiUrl}/notifications/${id}/read`,
      markAllRead: `${apiUrl}/notifications/read-all`,
    },

    dm: {
      conversations: `${apiUrl}/dm/conversations`,
      conversation: (username: string) => `${apiUrl}/dm/conversations/${username}`,
      markRead: (username: string) => `${apiUrl}/dm/conversations/${username}/read`,
      unreadCount: `${apiUrl}/dm/unread-count`,
      upload: `${apiUrl}/dm/upload`,
    },

  };
}
