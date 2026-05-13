<template>
  <div class="a-field">
    <label v-if="label" class="a-field-label">{{ label }}</label>
    <textarea
      class="a-textarea"
      :class="error ? 'a-textarea--error' : ''"
      v-bind="$attrs"
      :value="modelValue"
      :rows="rows"
      :disabled="disabled"
      @input="$emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
    />
    <div v-if="error" class="a-field-error">{{ error }}</div>
    <div v-else-if="hint" class="a-field-hint">{{ hint }}</div>
  </div>
</template>

<script setup lang="ts">
defineOptions({ inheritAttrs: false })

withDefaults(
  defineProps<{
    modelValue?: string
    label?: string
    hint?: string
    error?: string
    disabled?: boolean
    rows?: number
  }>(),
  {
    modelValue: '',
    disabled: false,
    rows: 4,
  }
)

defineEmits<{ 'update:modelValue': [v: string] }>()
</script>
