<template>
  <component
    :is="to ? RouterLink : tag"
    :to="to"
    class="a-btn"
    :class="[
      `a-btn--${normalizedVariant}`,
      `a-btn--${size}`,
      block ? 'a-btn--block' : '',
    ]"
    :disabled="isNativeButton ? (disabled || loading) : undefined"
    :aria-disabled="!isNativeButton && (disabled || loading) ? 'true' : undefined"
    v-bind="$attrs"
    @click="handleClick"
  >
    <slot>{{ loading ? loadingText : label }}</slot>
  </component>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

const props = withDefaults(
  defineProps<{
    tag?: string
    to?: string
    label?: string
    variant?: 'primary' | 'secondary' | 'danger' | 'ghost'
    outline?: boolean
    danger?: boolean
    size?: 'sm' | 'md' | 'lg'
    block?: boolean
    disabled?: boolean
    loading?: boolean
    loadingText?: string
  }>(),
  {
    tag: 'button',
    variant: 'primary',
    size: 'md',
    loadingText: '处理中...',
  }
)

const emit = defineEmits<{ click: [event: MouseEvent] }>()

defineOptions({ inheritAttrs: false })

const normalizedVariant = computed(() => {
  if (props.danger) return 'danger'
  if (props.outline) return 'secondary'
  return props.variant
})

const isNativeButton = computed(() => !props.to && props.tag === 'button')

const handleClick = (event: MouseEvent) => {
  if (props.disabled || props.loading) {
    event.preventDefault()
    event.stopPropagation()
    return
  }
  emit('click', event)
}
</script>
