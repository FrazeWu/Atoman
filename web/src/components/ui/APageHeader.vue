<template>
  <div class="a-section-header" :style="{ marginBottom: mb }">
    <div>
      <h1 class="a-title" :class="accent ? 'a-accent-l' : ''">
        <slot>{{ title }}</slot>
      </h1>
      <p v-if="sub" class="a-muted" :style="accent ? 'padding-left:2rem;margin-top:0.5rem' : 'margin-top:0.5rem'">{{ sub }}</p>
    </div>
    <slot name="action" />
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  title?: string
  sub?: string
  accent?: boolean
  mb?: string
}>(), { accent: false, mb: '2.5rem' })
</script>
