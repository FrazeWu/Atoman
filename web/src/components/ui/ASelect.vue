<template>
  <div class="a-field">
    <label v-if="label" class="a-field-label">{{ label }}</label>
    <div ref="rootRef" class="a-select-root">
      <button
        type="button"
        class="a-select-trigger"
        :class="[
          open ? 'a-select-trigger--open' : '',
          error ? 'a-select-trigger--error' : '',
        ]"
        :disabled="disabled"
        @click="toggleOpen"
      >
        <span :class="selectedOption ? '' : 'a-select-value--placeholder'">
          {{ selectedOption ? selectedOption.label : placeholder }}
        </span>
        <span class="a-select-chevron">▾</span>
      </button>

      <div v-if="open" class="a-select-panel">
        <button
          v-for="option in options"
          :key="String(option.value)"
          type="button"
          class="a-select-option"
          :disabled="option.disabled"
          @click="selectOption(option)"
        >
          <span class="a-select-marker">{{ option.value === modelValue ? '•' : '' }}</span>
          <span>{{ option.label }}</span>
        </button>
        <div v-if="options.length === 0" class="a-select-empty">暂无选项</div>
      </div>
    </div>
    <div v-if="error" class="a-field-error">{{ error }}</div>
    <div v-else-if="hint" class="a-field-hint">{{ hint }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

interface SelectOption {
  label: string
  value: string | number
  disabled?: boolean
}

const props = withDefaults(
  defineProps<{
    modelValue?: string | number | null
    options?: SelectOption[]
    placeholder?: string
    label?: string
    hint?: string
    error?: string
    disabled?: boolean
  }>(),
  {
    modelValue: null,
    options: () => [],
    placeholder: '请选择',
    disabled: false,
  }
)

const emit = defineEmits<{ 'update:modelValue': [value: string | number] }>()

const open = ref(false)
const rootRef = ref<HTMLElement | null>(null)

const selectedOption = computed(() => props.options.find(option => option.value === props.modelValue))

const close = () => {
  open.value = false
}

const toggleOpen = () => {
  if (props.disabled) return
  open.value = !open.value
}

const selectOption = (option: SelectOption) => {
  if (option.disabled) return
  emit('update:modelValue', option.value)
  close()
}

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as Node
  if (rootRef.value && !rootRef.value.contains(target)) {
    close()
  }
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', handleClickOutside))
</script>
