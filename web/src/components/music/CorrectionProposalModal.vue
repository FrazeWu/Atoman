<template>
  <AModal
    :model-value="show"
    title="提交修改建议"
    size="md"
    @close="$emit('close')"
    @update:model-value="(v) => { if (!v) $emit('close') }"
  >
    <form @submit.prevent="submit">
      <div class="a-field">
        <label class="a-label">修改说明 <span style="color: var(--a-color-danger, #ef4444)">*</span></label>
        <textarea
          v-model="description"
          class="a-textarea"
          placeholder="描述你希望修改的内容……"
          rows="4"
          required
        />
      </div>
      <div class="a-field" style="margin-top: 0.75rem">
        <label class="a-label">修改理由（可选）</label>
        <textarea
          v-model="reason"
          class="a-textarea"
          placeholder="为什么需要这个修改？"
          rows="2"
        />
      </div>
      <p v-if="errorMsg" style="color: var(--a-color-danger, #ef4444); font-size: 0.75rem; margin-top: 0.5rem">
        {{ errorMsg }}
      </p>
      <div style="display: flex; gap: 0.5rem; margin-top: 1.25rem; justify-content: flex-end">
        <button type="button" class="a-btn a-btn--ghost" @click="$emit('close')">取消</button>
        <button type="submit" class="a-btn a-btn--primary" :disabled="submitting">
          {{ submitting ? '提交中…' : '提交建议' }}
        </button>
      </div>
    </form>
  </AModal>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import AModal from '@/components/ui/AModal.vue'
import { useAuthStore } from '@/stores/auth'

const props = defineProps<{
  show: boolean
  type: 'album' | 'artist'
  targetId: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'submitted'): void
}>()

const API_URL = import.meta.env.VITE_API_URL || '/api'
const authStore = useAuthStore()

const description = ref('')
const reason = ref('')
const submitting = ref(false)
const errorMsg = ref('')

async function submit() {
  if (!description.value.trim()) return
  submitting.value = true
  errorMsg.value = ''

  try {
    const payload: Record<string, string> = {
      description: description.value.trim(),
      reason: reason.value.trim(),
    }
    const endpoint =
      props.type === 'album'
        ? `${API_URL}/corrections/album`
        : `${API_URL}/corrections/artist`

    if (props.type === 'album') {
      payload['album_id'] = props.targetId
    } else {
      payload['artist_id'] = props.targetId
    }

    const res = await fetch(endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {}),
      },
      body: JSON.stringify(payload),
    })

    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error((data as any).error || `请求失败 (${res.status})`)
    }

    description.value = ''
    reason.value = ''
    emit('submitted')
    emit('close')
  } catch (e: unknown) {
    errorMsg.value = e instanceof Error ? e.message : '提交失败，请稍后再试'
  } finally {
    submitting.value = false
  }
}
</script>
