<template>
  <div class="a-field">
    <label v-if="label" class="a-field-label">{{ label }}</label>
    <input
      class="a-input"
      v-bind="$attrs"
      :value="modelValue"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
  </div>
</template>

<script setup lang="ts">
defineOptions({ inheritAttrs: false })
defineProps<{ modelValue?: string; label?: string }>()
defineEmits<{ 'update:modelValue': [v: string] }>()
</script>
