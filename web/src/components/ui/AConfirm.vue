<template>
  <AModal v-if="show" @close="cancel" size="sm">
    <h3 class="a-subtitle" style="margin-bottom:.75rem">{{ title }}</h3>
    <p style="font-size:.9rem;color:#374151;line-height:1.6;margin-bottom:1.25rem;white-space:pre-wrap">
      {{ message }}
    </p>
    <div style="display:flex;gap:.6rem;justify-content:flex-end">
      <ABtn outline @click="cancel">{{ cancelText }}</ABtn>
      <ABtn :style="danger ? 'background:#dc2626;border-color:#dc2626;color:#fff' : ''" @click="confirm">
        {{ confirmText }}
      </ABtn>
    </div>
  </AModal>
</template>

<script setup lang="ts">
import AModal from '@/components/ui/AModal.vue'
import ABtn from '@/components/ui/ABtn.vue'

withDefaults(defineProps<{
  show: boolean
  title?: string
  message?: string
  confirmText?: string
  cancelText?: string
  danger?: boolean
}>(), {
  title: '请确认操作',
  message: '该操作不可撤销，是否继续？',
  confirmText: '确认',
  cancelText: '取消',
  danger: false,
})

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()

const confirm = () => emit('confirm')
const cancel = () => emit('cancel')
</script>
