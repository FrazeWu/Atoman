# Atoman UI Components

A minimal, reusable component library and CSS utility system built for the Atoman frontend.
All components follow the **极简主义档案馆 (Minimalist Archive Aesthetic)**: pure black/white, hard shadows, bold typography, and border-first design.

---

## Table of Contents

- [Design Tokens](#design-tokens)
- [CSS Utility Classes](#css-utility-classes)
  - [Layout](#layout)
  - [Typography](#typography)
  - [Buttons (CSS-only)](#buttons-css-only)
  - [Form Fields](#form-fields)
  - [Cards](#cards)
  - [Modal](#modal)
  - [Feedback](#feedback)
  - [Grid Helpers](#grid-helpers)
  - [Misc Helpers](#misc-helpers)
- [Vue Components](#vue-components)
  - [ABtn](#abtn)
  - [AInput](#ainput)
  - [ATextarea](#atextarea)
  - [AModal](#amodal)
  - [ACard](#acard)
  - [AEmpty](#aempty)
  - [APageHeader](#apageheader)

---

## Design Tokens

| Token | Value |
|---|---|
| Primary black | `#000000` |
| Primary white | `#ffffff` |
| Muted text | `#6b7280` |
| Border | `2px solid #000` |
| Border heavy | `4px solid #000` / `8px solid #000` (accent) |
| Hard shadow | `10px 10px 0px 0px rgba(0,0,0,1)` |
| Modal shadow | `20px 20px 0px 0px rgba(0,0,0,1)` |
| Focus shadow | `5px 5px 0px 0px rgba(0,0,0,1)` |
| Font heading | `font-weight: 900`, `letter-spacing: -0.05em` |
| Font label | `font-weight: 900`, `uppercase`, `letter-spacing: 0.1em` |
| Transition | `all 0.2s` / `0.3s` |
| Grayscale | `filter: grayscale(1)` |

---

## CSS Utility Classes

All classes are defined in `src/style.css` and available globally.

### Layout

```html
<!-- Full-width pages with horizontal centering -->
<div class="a-page">...</div>      <!-- max-w: 72rem, standard pages -->
<div class="a-page-md">...</div>   <!-- max-w: 48rem, medium pages -->
<div class="a-page-sm">...</div>   <!-- max-w: 32rem, forms/settings -->
<div class="a-page-xl">...</div>   <!-- max-w: 80rem, wide dashboards -->
```

All page containers include `padding: 3rem 2rem 12rem` (bottom padding accommodates the fixed audio player).

---

### Typography

```html
<h1 class="a-title">Title</h1>           <!-- 3rem, weight 900 -->
<h1 class="a-title-lg">Large</h1>        <!-- 3.75rem, weight 900 -->
<h1 class="a-title-xl">Extra Large</h1>  <!-- 4.5rem, weight 900 -->
<h2 class="a-title-sm">Small</h2>        <!-- 2rem, weight 900 -->
<h3 class="a-subtitle">Subtitle</h3>     <!-- 1.5rem, weight 900 -->
<span class="a-label">LABEL TEXT</span>  <!-- 0.75rem, uppercase, wide spacing -->
<p class="a-muted">Muted text</p>        <!-- gray-500, weight 500 -->

<!-- Left accent border (8px solid black) -->
<h1 class="a-title a-accent-l">Accented Title</h1>
```

---

### Buttons (CSS-only)

Use directly on `<button>` or `<a>` elements when you don't need the `ABtn` component.

```html
<!-- Primary (black fill) -->
<button class="a-btn">Submit</button>
<button class="a-btn a-btn-sm">Small</button>
<button class="a-btn a-btn-lg">Large</button>
<button class="a-btn a-btn-block">Full Width</button>

<!-- Outline (white fill) -->
<button class="a-btn-outline">Cancel</button>
<button class="a-btn-outline a-btn-outline-sm">Small Outline</button>
<button class="a-btn-outline a-btn-outline-lg">Large Outline</button>
<button class="a-btn-outline a-btn-outline-block">Full Width Outline</button>

<!-- Icon-only button -->
<button class="a-btn-icon">✕</button>

<!-- Link style -->
<a href="/about" class="a-link">← Back</a>
<RouterLink to="/blog" class="a-link">Blog</RouterLink>
```

---

### Form Fields

```html
<!-- Standalone input/textarea/select -->
<input class="a-input" placeholder="Enter value" />
<textarea class="a-textarea" rows="4"></textarea>
<select class="a-select">...</select>

<!-- Field wrapper with label -->
<div class="a-field">
  <label class="a-field-label">Field Label</label>
  <input class="a-input" />
</div>

<div class="a-field">
  <label class="a-field-label">Description</label>
  <textarea class="a-textarea" rows="4"></textarea>
</div>
```

---

### Cards

```html
<!-- Static card -->
<div class="a-card">Content</div>

<!-- Card with hover shadow -->
<div class="a-card a-card-hover">Hoverable content</div>

<!-- Compact card -->
<div class="a-card-sm">Compact content</div>
```

---

### Modal

The `.a-modal-*` classes are used internally by `AModal.vue`. You can use them standalone:

```html
<div class="a-modal-backdrop">
  <div class="a-modal a-modal-md">Modal content</div>
</div>

<!-- Sizes: a-modal-sm (24rem) | a-modal-md (32rem) | a-modal-lg (40rem) -->
```

---

### Feedback

```html
<div class="a-error">Something went wrong.</div>
<div class="a-success">✓ Saved successfully.</div>
<div class="a-empty">No items yet.</div>
```

---

### Grid Helpers

```html
<!-- 2-column grid (1 col on mobile, 2 on md+) -->
<div class="a-grid-2">
  <div>Item</div>
  <div>Item</div>
</div>

<!-- 3-column grid (1 → 2 → 3 cols) -->
<div class="a-grid-3">
  <div>Item</div>
  <div>Item</div>
  <div>Item</div>
</div>
```

---

### Misc Helpers

```html
<!-- Section header: title on left, action on right -->
<div class="a-section-header">
  <h1 class="a-title">Page Title</h1>
  <button class="a-btn">Action</button>
</div>

<!-- Skeleton loading placeholder -->
<div class="a-skeleton" style="height:4rem"></div>

<!-- Divider -->
<hr class="a-divider" />

<!-- Text truncation -->
<span class="a-truncate">Very long text...</span>  <!-- single line -->
<p class="a-clamp-2">Long paragraph...</p>          <!-- 2-line clamp -->

<!-- Grayscale image -->
<img class="a-grayscale" src="cover.jpg" />

<!-- Badges -->
<span class="a-badge">外部</span>            <!-- outlined -->
<span class="a-badge a-badge-fill">博客</span> <!-- filled black -->
```

---

## Vue Components

All components live in `src/components/ui/`. Import individually:

```ts
import ABtn from '@/components/ui/ABtn.vue'
import AInput from '@/components/ui/AInput.vue'
import ATextarea from '@/components/ui/ATextarea.vue'
import AModal from '@/components/ui/AModal.vue'
import ACard from '@/components/ui/ACard.vue'
import AEmpty from '@/components/ui/AEmpty.vue'
import APageHeader from '@/components/ui/APageHeader.vue'
```

---

### ABtn

A versatile button/link component. Renders as `<button>`, `<a>`, or `<RouterLink>` depending on props.

**Props:**

| Prop | Type | Default | Description |
|---|---|---|---|
| `tag` | `string` | `'button'` | HTML tag when no `to` is provided |
| `to` | `string` | — | Vue Router path; renders as `RouterLink` |
| `label` | `string` | — | Button text (use slot for JSX-style) |
| `outline` | `boolean` | `false` | White background, black border variant |
| `size` | `'sm' \| 'md' \| 'lg'` | `'md'` | Button size |
| `block` | `boolean` | `false` | Full-width button |
| `disabled` | `boolean` | `false` | Disabled state |
| `loading` | `boolean` | `false` | Shows loading text, disables click |
| `loadingText` | `string` | `'处理中...'` | Text shown while loading |

All native attributes (`@click`, `type`, `form`, `style`, etc.) are forwarded via `v-bind="$attrs"`.

**Usage:**

```vue
<!-- Basic button -->
<ABtn @click="save">保存</ABtn>

<!-- With label prop -->
<ABtn label="提交" />

<!-- Outline variant -->
<ABtn outline @click="cancel">取消</ABtn>

<!-- Sizes -->
<ABtn size="sm">小按钮</ABtn>
<ABtn size="lg">大按钮</ABtn>

<!-- Full width -->
<ABtn block>全宽按钮</ABtn>

<!-- Router link -->
<ABtn to="/login">去登录</ABtn>
<ABtn to="/blog/posts/new" size="sm" outline>写文章</ABtn>

<!-- External link -->
<ABtn tag="a" href="https://example.com" target="_blank">外部链接</ABtn>

<!-- Form submit -->
<ABtn type="submit" block :loading="saving" loadingText="保存中...">保存更改</ABtn>

<!-- Disabled -->
<ABtn :disabled="!isValid">提交</ABtn>

<!-- Slot content -->
<ABtn outline size="sm">
  <span>+ 新建</span>
</ABtn>
```

---

### AInput

A labeled input field with `v-model` support.

**Props:**

| Prop | Type | Default | Description |
|---|---|---|---|
| `modelValue` | `string` | — | Bound value (`v-model`) |
| `label` | `string` | — | Label text above input |

All native `<input>` attributes (`type`, `placeholder`, `required`, etc.) are forwarded.

**Usage:**

```vue
<AInput v-model="form.email" label="邮箱" type="email" placeholder="you@example.com" />
<AInput v-model="form.password" label="密码" type="password" />
<AInput v-model="form.url" label="网站" type="url" placeholder="https://" />

<!-- Without label (renders bare input) -->
<AInput v-model="query" placeholder="搜索..." />
```

---

### ATextarea

A labeled textarea with `v-model` support.

**Props:**

| Prop | Type | Default | Description |
|---|---|---|---|
| `modelValue` | `string` | — | Bound value (`v-model`) |
| `label` | `string` | — | Label text above textarea |

All native `<textarea>` attributes (`rows`, `placeholder`, etc.) are forwarded.

**Usage:**

```vue
<ATextarea v-model="form.bio" label="个人简介" rows="4" placeholder="介绍一下自己..." />
<ATextarea v-model="post.content" rows="20" />
```

---

### AModal

A teleported, backdrop-click-to-close modal dialog.

**Props:**

| Prop | Type | Default | Description |
|---|---|---|---|
| `size` | `'sm' \| 'md' \| 'lg'` | `'md'` | Modal width (24/32/40rem) |

**Events:**

| Event | Description |
|---|---|
| `close` | Emitted when backdrop is clicked |

Modal is rendered via `<Teleport to="body">` so it always appears above other content.

**Usage:**

```vue
<script setup>
const showModal = ref(false)
</script>

<template>
  <ABtn @click="showModal = true">打开</ABtn>

  <AModal v-if="showModal" @close="showModal = false" size="md">
    <h3 class="a-subtitle" style="margin-bottom:1rem">标题</h3>
    <p>Modal content here.</p>
    <div style="display:flex;gap:.5rem;margin-top:1.5rem">
      <ABtn style="flex:1" @click="confirm">确认</ABtn>
      <ABtn outline @click="showModal = false">取消</ABtn>
    </div>
  </AModal>
</template>
```

**Size examples:**

```vue
<AModal size="sm">...</AModal>  <!-- 24rem — confirmations -->
<AModal size="md">...</AModal>  <!-- 32rem — forms (default) -->
<AModal size="lg">...</AModal>  <!-- 40rem — large forms -->
```

---

### ACard

A simple card wrapper with optional hover shadow.

**Props:**

| Prop | Type | Default | Description |
|---|---|---|---|
| `hover` | `boolean` | `false` | Adds hard shadow on hover |

All native div attributes are forwarded.

**Usage:**

```vue
<!-- Static card -->
<ACard>
  <p>Content</p>
</ACard>

<!-- Hoverable card -->
<ACard hover>
  <h3>Article title</h3>
  <p>Summary...</p>
</ACard>

<!-- With extra styles -->
<ACard hover style="display:flex;gap:1rem">
  <img class="a-grayscale" src="cover.jpg" style="width:5rem;height:5rem" />
  <div>...</div>
</ACard>
```

---

### AEmpty

A dashed empty-state block for zero-data situations.

**Props:**

| Prop | Type | Default | Description |
|---|---|---|---|
| `text` | `string` | — | Message to display |

Supports default slot for custom content.

**Usage:**

```vue
<!-- With text prop -->
<AEmpty text="暂无内容" />
<AEmpty text="还没有文章" />

<!-- With slot -->
<AEmpty>
  <p>暂无收藏</p>
  <ABtn to="/blog" size="sm" style="margin-top:1rem">去发现文章</ABtn>
</AEmpty>
```

---

### APageHeader

A section header with a title on the left and an optional action slot on the right.

**Props:**

| Prop | Type | Default | Description |
|---|---|---|---|
| `title` | `string` | — | Main heading text |
| `sub` | `string` | — | Subtitle below heading |
| `accent` | `boolean` | `false` | Adds left border accent (`a-accent-l`) |
| `mb` | `string` | `'2.5rem'` | Bottom margin (CSS value) |

Supports `#action` slot for right-side content.

**Usage:**

```vue
<!-- Simple header -->
<APageHeader title="我的博客" />

<!-- With subtitle -->
<APageHeader title="探索" sub="发现新内容" />

<!-- With accent border -->
<APageHeader title="订阅" accent sub="聚合你的 RSS 订阅源" />

<!-- With action button -->
<APageHeader title="文章" sub="管理你的文章">
  <template #action>
    <ABtn to="/blog/posts/new">+ 写文章</ABtn>
  </template>
</APageHeader>

<!-- Custom title via slot -->
<APageHeader accent>
  <span>自定义 <em>标题</em></span>
  <template #action>
    <ABtn outline size="sm">筛选</ABtn>
  </template>
</APageHeader>

<!-- Custom bottom margin -->
<APageHeader title="设置" mb="1.5rem" />
```

---

## Common Patterns

### Form Page

```vue
<template>
  <div class="a-page-sm">
    <APageHeader title="创建文章" accent mb="2rem" />

    <form @submit.prevent="submit" style="display:flex;flex-direction:column;gap:1.5rem">
      <AInput v-model="form.title" label="标题" placeholder="文章标题" />
      <ATextarea v-model="form.content" label="内容" rows="12" />

      <div v-if="error" class="a-error">{{ error }}</div>

      <ABtn block type="submit" :loading="saving" loadingText="发布中...">发布</ABtn>
    </form>
  </div>
</template>
```

### Card Grid

```vue
<template>
  <div class="a-page">
    <APageHeader title="文章" sub="所有文章">
      <template #action>
        <ABtn to="/blog/posts/new" size="sm">+ 写文章</ABtn>
      </template>
    </APageHeader>

    <AEmpty v-if="!posts.length" text="暂无文章" />
    <div v-else class="a-grid-3">
      <ACard v-for="post in posts" :key="post.id" hover>
        <h3 class="a-title-sm">{{ post.title }}</h3>
        <p class="a-muted a-clamp-2">{{ post.summary }}</p>
      </ACard>
    </div>
  </div>
</template>
```

### Confirm Modal

```vue
<template>
  <ABtn outline size="sm" @click="showConfirm = true">删除</ABtn>

  <AModal v-if="showConfirm" @close="showConfirm = false" size="sm">
    <h3 class="a-subtitle" style="margin-bottom:.75rem">确认删除</h3>
    <p class="a-muted" style="margin-bottom:1.5rem;font-size:.875rem">此操作不可撤销。</p>
    <div style="display:flex;gap:.5rem">
      <ABtn style="flex:1" @click="deleteItem">删除</ABtn>
      <ABtn outline @click="showConfirm = false">取消</ABtn>
    </div>
  </AModal>
</template>
```

### Loading Skeletons

```vue
<template>
  <div v-if="loading" class="a-grid-2">
    <div v-for="i in 4" :key="i" class="a-skeleton" style="height:8rem" />
  </div>
  <div v-else class="a-grid-2">
    <ACard v-for="item in items" :key="item.id" hover>...</ACard>
  </div>
</template>
```
