<template>
  <div class="a-card" :class="hover ? 'a-card-hover' : ''" v-bind="$attrs">
    <slot />
  </div>
</template>

<script setup lang="ts">
defineOptions({ inheritAttrs: false })
withDefaults(defineProps<{ hover?: boolean }>(), { hover: false })
</script>
