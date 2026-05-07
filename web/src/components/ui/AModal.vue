<template>
  <Teleport to="body">
    <div class="a-modal-backdrop" @click.self="$emit('close')">
      <div class="a-modal" :class="`a-modal-${size}`">
        <slot />
        <slot name="footer" />
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ size?: 'sm' | 'md' | 'lg' }>(), { size: 'md' })
defineEmits<{ close: [] }>()
</script>
