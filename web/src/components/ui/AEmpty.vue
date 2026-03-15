<template>
  <div class="a-empty">
    <slot>{{ text }}</slot>
  </div>
</template>

<script setup lang="ts">
defineProps<{ text?: string }>()
</script>
