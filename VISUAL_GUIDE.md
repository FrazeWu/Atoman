# Visual Guide - Music Wiki Navigation Implementation

## Album Detail Page Layout (AFTER Phase 1)

```
┌─────────────────────────────────────────────────────────────────┐
│                       ALBUM DETAIL PAGE                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ← 返回时间线                                                   │
│                                                                 │
│  ┌──────────────┐  Album Title                                │
│  │              │  Artist Name                                │
│  │              │  12 tracks                                   │
│  │   COVER      │                                             │
│  │              │  ┌──────────────┐  ┌────────────────┐      │
│  │   IMAGE      │  │ ▶ 播放专辑   │  │ 编辑专辑       │      │
│  │              │  └──────────────┘  └────────────────┘      │
│  └──────────────┘                                              │
│                     ┌──────────────┐  ┌──────────────┐  🆕    │
│                     │ 📖 修订历史   │  │ 💬 讨论      │  NEW!  │
│                     └──────────────┘  └──────────────┘        │
│                                                                 │
│                     ┌─────────────────────────────────┐ 🆕     │
│                     │ 🔒 半保护                        │ NEW!   │
│                     └─────────────────────────────────┘        │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│                       TRACK LIST                                │
├─────────────────────────────────────────────────────────────────┤
│  01  Track Title One        [ ▶ 播放 ]                        │
│  02  Track Title Two        [ ▶ 播放 ]                        │
│  03  Track Title Three      [ ▶ 播放 ]                        │
│  ...                                                            │
└─────────────────────────────────────────────────────────────────┘
```

## Navigation Flow Diagram

```
┌─────────────────────────────────────┐
│     Music Timeline / Home Page       │
│   (Album cards with metadata)        │
└────────────────┬────────────────────┘
                 │
        [Click Album Card]
                 │
                 ▼
┌─────────────────────────────────────┐
│    Album Detail Page                │
│  (Album Info + Track List)          │
│                                     │
│  ┌─────────────────────────────┐   │
│  │ 📖 修订历史 Link ← NEW!     │   │
│  │ 💬 讨论 Link ← NEW!         │   │
│  │ 🔒 Protection Badge ← NEW!  │   │
│  └─────────────────────────────┘   │
└────────┬─────────────────┬──────────┘
         │                 │
    [Click Links]    [Click Links]
         │                 │
    [If Protected] [Show Edit Restrictions]
         │
         ├─────────────────┬──────────────┐
         │                 │              │
         ▼                 ▼              ▼
   ┌─────────────┐  ┌──────────────┐  ┌──────────────┐
   │   History   │  │ Discussion   │  │ Admin Queue  │
   │   Revisions │  │   Threads    │  │  (Admins)    │
   └─────────────┘  └──────────────┘  └──────────────┘
```

## User Journey - Different User Types

### Regular User Flow
```
User on Music Timeline
    ↓
    Sees Album Card
    ↓
    Clicks Album
    ↓
    Views Album Detail Page
    ├─ Sees "📖 修订历史" Link
    ├─ Sees "💬 讨论" Link
    └─ Sees "🔒 半保护" Badge (if protected)
    ↓
    Can Click Links to:
    ├─ View Revision History
    │  ├─ See who changed what
    │  ├─ Compare versions
    │  └─ Understand edit timeline
    │
    └─ View Discussions
       ├─ Create new discussion
       ├─ Reply to discussions
       └─ See edit context and rationale
```

### Admin User Flow
```
Admin on Music Timeline
    ↓
    [Same as Regular User +]
    ↓
    Additional Capabilities:
    ├─ Can Manage Protection Level
    │  ├─ Set Semi/Full Protection
    │  ├─ Protect/Unprotect Albums
    │  └─ Set Protection Expiration
    │
    ├─ Can Revert Revisions
    │  ├─ View all revision versions
    │  ├─ Revert to any previous version
    │  └─ Add revert reason
    │
    └─ Can Approve/Reject Edits
       ├─ Go to /music/admin/review
       ├─ See pending revisions
       ├─ Approve or Reject
       └─ Add review notes
```

## Protection Status Display

### Unprotected Album (Default)
```
┌─────────────────────────────┐
│                             │
│      Album Detail           │
│                             │
│  ┌───────────┐              │
│  │ 📖 History│              │  ← No protection badge
│  │ 💬 Discuss│              │     (only shows if protected)
│  └───────────┘              │
│                             │
│                             │
└─────────────────────────────┘
```

### Semi-Protected Album (Admin Approval Required)
```
┌──────────────────────────────────────┐
│                                      │
│       Album Detail                   │
│                                      │
│  ┌────────────┐                      │
│  │ 📖 History │                      │
│  │ 💬 Discuss │                      │
│  └────────────┘                      │
│                                      │
│  ┌──────────────────────────────┐   │
│  │ 🔒 半保护                     │   │  ← Yellow badge
│  │    (pending approval)         │   │     (admin approval needed)
│  └──────────────────────────────┘   │
│                                      │
└──────────────────────────────────────┘
```

### Full-Protected Album (Admin Only)
```
┌──────────────────────────────────────┐
│                                      │
│       Album Detail                   │
│                                      │
│  ┌────────────┐                      │
│  │ 📖 History │                      │
│  │ 💬 Discuss │                      │
│  └────────────┘                      │
│                                      │
│  ┌──────────────────────────────┐   │
│  │ 🔒 完全保护                   │   │  ← Red badge
│  │    仅管理员可编辑             │   │     (admin only)
│  └──────────────────────────────┘   │
│                                      │
└──────────────────────────────────────┘
```

## Responsive Design

### Desktop (>768px)
```
┌────────────────────────────────────────┐
│ ← 返回时间线                           │
│ ┌──────────┐  Title                   │
│ │  Cover   │  Artist                  │
│ │  Image   │  12 tracks               │
│ └──────────┘  ┌────┐ ┌────┐           │
│              │Play│ │Edit│           │
│              └────┘ └────┘           │
│              ┌──────┐ ┌──────┐ (Inline)
│              │Hist. │ │Disc. │       │
│              └──────┘ └──────┘       │
│              ┌─────────────────┐     │
│              │🔒 Protection    │     │
│              └─────────────────┘     │
└────────────────────────────────────────┘
```

### Mobile (<768px)
```
┌──────────────────────┐
│ ← 返回时间线         │
│                      │
│ ┌──────────────────┐│
│ │     Cover        ││
│ │     Image        ││
│ └──────────────────┘│
│                      │
│ Title               │
│ Artist              │
│ 12 tracks           │
│                      │
│ ┌──────────────────┐│
│ │ ▶ 播放专辑       ││
│ └──────────────────┘│
│ ┌──────────────────┐│
│ │ 编辑专辑         ││
│ └──────────────────┘│
│                      │
│ ┌──────────────────┐│
│ │ 📖 修订历史      ││
│ └──────────────────┘│
│ ┌──────────────────┐│
│ │ 💬 讨论          ││
│ └──────────────────┘│
│                      │
│ ┌──────────────────┐│
│ │ 🔒 半保护        ││
│ └──────────────────┘│
│                      │
│ 歌曲列表             │
└──────────────────────┘
```

## Interaction Examples

### Clicking "📖 修订历史" Link
```
Current URL: /music/albums/abc123
User clicks: [📖 修订历史]
    ↓
Navigates to: /music/albums/abc123/history
    ↓
Loads: AlbumHistoryView.vue
    ↓
Shows:
    ├─ Revision List
    ├─ Diff Comparison Tool
    ├─ Revert Button (admin only)
    └─ Back to Album Link
```

### Clicking "💬 讨论" Link
```
Current URL: /music/albums/abc123
User clicks: [💬 讨论]
    ↓
Navigates to: /music/albums/abc123/discussion
    ↓
Loads: AlbumDiscussionView.vue
    ↓
Shows:
    ├─ Create Discussion Form (if logged in)
    ├─ Discussion Threads
    ├─ Reply Forms
    ├─ User Avatars
    ├─ Timestamps
    └─ Back to Album Link
```

## CSS Classes Reference

```
.wiki-links               ← Container for wiki navigation buttons
├─ .wiki-link            ← Individual navigation link
│  ├─ background: white   ← Default background
│  ├─ border: 2px solid   ← Bold border
│  ├─ padding: 0.5rem 1rem
│  └─ hover: black background, white text

.wiki-meta               ← Container for protection badge
├─ background: #f9fafb   ← Light gray background
├─ padding: 1rem
├─ border: 2px solid

.protection-badge        ← Protection status display
├─ font-weight: 700
├─ text-transform: uppercase
├─ .protection-full      ← Full protection styling
│  ├─ background: #dc2626 ← Red
│  └─ color: white
└─ .protection-semi      ← Semi protection styling
   ├─ background: #facc15 ← Yellow
   └─ color: black

.status-badge            ← Edit restriction indicator
└─ .status-draft         ← Edit restriction styling
   ├─ background: #6b7280 ← Gray
   └─ color: white
```

## Before & After Comparison

### BEFORE Phase 1 ❌
```
Album Detail Page
    ├─ Album Cover & Info
    ├─ Play/Edit Buttons
    ├─ Track List
    └─ [No Wiki Links]
    └─ [No Protection Display]

User Experience:
- Wiki features exist but are HIDDEN
- No way to discover them from UI
- Users don't know revisions/discussions exist
- Protection status invisible
```

### AFTER Phase 1 ✅
```
Album Detail Page
    ├─ Album Cover & Info
    ├─ Play/Edit Buttons
    ├─ 📖 Revision History Link ← NEW
    ├─ 💬 Discussion Link ← NEW
    ├─ 🔒 Protection Badge ← NEW
    ├─ Track List
    └─ [Wiki Features Now Discoverable!]

User Experience:
- Wiki features are VISIBLE and accessible
- Clear navigation to revision history
- Can participate in discussions
- Understands edit restrictions
- Significantly improved UX
```

---

This visual guide helps understand the implementation and user experience improvements from Phase 1.

