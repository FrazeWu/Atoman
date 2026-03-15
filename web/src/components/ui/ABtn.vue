<template>
  <component
    :is="to ? 'RouterLink' : tag"
    :to="to"
    class="a-btn"
    :class="[
      outline ? 'a-btn-outline' : 'a-btn',
      size === 'sm' ? 'a-btn-sm' : size === 'lg' ? 'a-btn-lg' : '',
      block ? 'a-btn-block' : '',
      outline && size === 'sm' ? 'a-btn-outline-sm' : '',
    ]"
    :disabled="disabled || loading"
    v-bind="$attrs"
  >
    <slot>{{ loading ? loadingText : label }}</slot>
  </component>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'

defineOptions({ inheritAttrs: false })

withDefaults(defineProps<{
  tag?: string
  to?: string
  label?: string
  outline?: boolean
  size?: 'sm' | 'md' | 'lg'
  block?: boolean
  disabled?: boolean
  loading?: boolean
  loadingText?: string
}>(), {
  tag: 'button',
  size: 'md',
  loadingText: '处理中...',
})
</script>
