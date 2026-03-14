export function useApiUrl() {
  return import.meta.env.VITE_API_URL || '/api';
}

export function useApi() {
  const apiUrl = useApiUrl();

  return {
    url: apiUrl,
    songs: `${apiUrl}/songs`,
    corrections: `${apiUrl}/corrections`,
    correctionsBatch: `${apiUrl}/corrections/batch`,
    adminPending: `${apiUrl}/admin/pending`,
    adminApprove: (id: number) => `${apiUrl}/admin/approve/${id}`,
    adminReject: (id: number) => `${apiUrl}/admin/reject/${id}`,
    
    blog: {
      channels: `${apiUrl}/blog/channels`,
      channel: (id: number) => `${apiUrl}/blog/channels/${id}`,
      channelCollections: (id: number) => `${apiUrl}/blog/channels/${id}/collections`,
      collections: `${apiUrl}/blog/collections`,
      collection: (id: number) => `${apiUrl}/blog/collections/${id}`,
      
      posts: `${apiUrl}/blog/posts`,
      post: (id: number) => `${apiUrl}/blog/posts/${id}`,
      postPublish: (id: number) => `${apiUrl}/blog/posts/${id}/publish`,
      postUnpublish: (id: number) => `${apiUrl}/blog/posts/${id}/unpublish`,
      postPin: (id: number) => `${apiUrl}/blog/posts/${id}/pin`,
      postUnpin: (id: number) => `${apiUrl}/blog/posts/${id}/unpin`,
      drafts: `${apiUrl}/blog/posts/drafts`,
      postCollections: (id: number) => `${apiUrl}/blog/posts/${id}/collections`,
      postCollection: (id: number, collectionId: number) => `${apiUrl}/blog/posts/${id}/collections/${collectionId}`,
      
      comments: `${apiUrl}/blog/comments`,
      postComments: (id: number) => `${apiUrl}/blog/posts/${id}/comments`,
      
      likes: `${apiUrl}/blog/likes`,
      postLikesCount: (id: number) => `${apiUrl}/blog/posts/${id}/likes/count`,
      
      bookmarks: `${apiUrl}/blog/bookmarks`,
      bookmark: (id: number) => `${apiUrl}/blog/bookmarks/${id}`,
      bookmarkFolders: `${apiUrl}/blog/bookmark-folders`,
      bookmarkFolder: (id: number) => `${apiUrl}/blog/bookmark-folders/${id}`,
      
      explore: `${apiUrl}/blog/explore`,
    },
    
    users: {
      me: `${apiUrl}/users/me`,
      settings: `${apiUrl}/users/me`,          // profile update (display_name, bio, etc)
      meSettings: `${apiUrl}/users/me/settings`, // app settings (notifications, privacy)
      profile: (username: number | string) => `${apiUrl}/users/by-username/${username}`,
      follow: (id: number | string) => `${apiUrl}/users/${id}/follow`,
      followers: (id: number | string) => `${apiUrl}/users/${id}/followers`,
      following: (id: number | string) => `${apiUrl}/users/${id}/following`,
    },
    
    orbit: {
      subscriptions: `${apiUrl}/feed/subscriptions`,
      subscription: (id: number) => `${apiUrl}/feed/subscriptions/${id}`,
      timeline: `${apiUrl}/feed/timeline`,
      rss: (username: string) => `${apiUrl}/feed/rss/${username}`,
    },
    
    notifications: {
      list: `${apiUrl}/notifications`,
      read: (id: number) => `${apiUrl}/notifications/${id}/read`,
      readAll: `${apiUrl}/notifications/read-all`,
      unreadCount: `${apiUrl}/notifications/unread-count`,
    }
  };
}
