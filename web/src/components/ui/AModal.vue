<template>
  <Teleport to="body">
    <div v-if="visible" class="a-modal-backdrop" @click.self="handleClose">
      <div class="a-modal" :class="`a-modal-${size}`" role="dialog" aria-modal="true">
        <div class="a-modal-header">
          <span class="a-modal-title">{{ title }}</span>
          <button v-if="closable" class="a-modal-close" type="button" @click="handleClose">✕</button>
        </div>
        <div class="a-modal-body">
          <slot />
        </div>
        <div v-if="hasFooter" class="a-modal-footer">
          <slot name="footer" />
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, useSlots } from 'vue'

const props = withDefaults(
  defineProps<{
    size?: 'sm' | 'md' | 'lg'
    title?: string
    modelValue?: boolean
    closable?: boolean
  }>(),
  {
    size: 'md',
    title: '',
    modelValue: undefined,
    closable: true,
  }
)

const emit = defineEmits<{ close: []; 'update:modelValue': [value: boolean] }>()
const slots = useSlots()

const visible = computed(() => (props.modelValue === undefined ? true : props.modelValue))

const hasFooter = computed(() => Boolean(slots.footer))

const handleClose = () => {
  emit('update:modelValue', false)
  emit('close')
}
</script>
