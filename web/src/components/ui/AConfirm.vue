<template>
  <AModal v-if="show" @close="cancel" size="sm" :title="title">
    <p style="font-size:.9rem;color:#374151;line-height:1.6;white-space:pre-wrap">{{ message }}</p>
    <template #footer>
      <ABtn variant="secondary" @click="cancel">{{ cancelText }}</ABtn>
      <ABtn :variant="danger ? 'danger' : 'primary'" @click="confirm">{{ confirmText }}</ABtn>
    </template>
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

const emit = defineEmits<{ confirm: []; cancel: [] }>()
const confirm = () => emit('confirm')
const cancel = () => emit('cancel')
</script>
