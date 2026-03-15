<template>
  <div class="a-field">
    <label v-if="label" class="a-field-label">{{ label }}</label>
    <textarea
      class="a-textarea"
      v-bind="$attrs"
      :value="modelValue"
      @input="$emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
    />
  </div>
</template>

<script setup lang="ts">
defineOptions({ inheritAttrs: false })
defineProps<{ modelValue?: string; label?: string }>()
defineEmits<{ 'update:modelValue': [v: string] }>()
</script>
