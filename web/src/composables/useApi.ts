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
      meSettings: `${apiUrl}/users/me/settings`, // app settings (notifications, privacy)
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
      read: (id: number) => `${apiUrl}/notifications/${id}/read`,
      readAll: `${apiUrl}/notifications/read-all`,
    }
  };
}
