<template>
  <div class="a-field">
    <label v-if="label" class="a-field-label">{{ label }}</label>
    <input
      class="a-input"
      :class="error ? 'a-input--error' : ''"
      v-bind="$attrs"
      :value="modelValue"
      :disabled="disabled"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
    <div v-if="error" class="a-field-error">{{ error }}</div>
    <div v-else-if="hint" class="a-field-hint">{{ hint }}</div>
  </div>
</template>

<script setup lang="ts">
defineOptions({ inheritAttrs: false })

withDefaults(
  defineProps<{
    modelValue?: string | number
    label?: string
    hint?: string
    error?: string
    disabled?: boolean
  }>(),
  {
    modelValue: '',
    disabled: false,
  }
)

defineEmits<{ 'update:modelValue': [v: string] }>()
</script>
