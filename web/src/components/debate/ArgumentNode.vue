<template>
<div class="a-card a-card-hover">
  <!-- Argument Header -->
  <div class="flex items-start gap-3 mb-3">
    <!-- Type Badge -->
    <span
      class="a-badge flex-shrink-0"
      :class="typeBadgeClasses[argument.argument_type]"
      :style="typeBadgeStyles[argument.argument_type]"
    >
      {{ typeLabels[argument.argument_type] }}
    </span>

    <!-- Vote Controls -->
    <div class="flex items-center gap-2 flex-shrink-0">
      <button
        v-if="canVote"
        @click="$emit('vote', argument.id, 1)"
        class="w-8 h-8 flex items-center justify-center border-2 border-black font-bold transition-colors"
        :class="{
          'bg-green-600 text-white': userVote === 1,
          'hover:bg-green-100': userVote !== 1,
        }"
      >
        ▲
      </button>
      <button
        v-if="canVote"
        @click="$emit('vote', argument.id, -1)"
        class="w-8 h-8 flex items-center justify-center border-2 border-black font-bold transition-colors"
        :class="{
          'bg-red-600 text-white': userVote === -1,
          'hover:bg-red-100': userVote !== -1,
        }"
      >
        ▼
      </button>
      <span class="font-black text-lg min-w-[3ch] text-center">
        {{ argument.vote_count }}
      </span>
    </div>

    <!-- Actions -->
    <div class="flex items-center gap-2 ml-auto">
      <ABtn v-if="canEdit" size="sm" outline @click="$emit('edit', argument)">
        编辑
      </ABtn>
      <ABtn v-if="canReply" size="sm" outline @click="$emit('reply', argument)">
        引用
      </ABtn>
      <ABtn v-if="canReply" size="sm" outline @click="$emit('reference', argument.id, '')">
        引用辩题
      </ABtn>
      <!-- Admin fold/unfold -->
      <ABtn v-if="authStore.user?.role === 'admin' && !localIsFolded" size="sm" outline @click="foldArgument">
        折叠
      </ABtn>
      <ABtn v-if="authStore.user?.role === 'admin' && localIsFolded" size="sm" outline @click="unfoldArgument">
        展开
      </ABtn>
      <ABtn v-if="canDelete" size="sm" variant="danger" @click="$emit('delete', argument)">
        删除
      </ABtn>
    </div>
  </div>

  <!-- Folded overlay -->
  <div v-if="localIsFolded && authStore.user?.role !== 'admin'" style="padding:.5rem .75rem;color:var(--a-color-muted);font-size:.75rem;font-style:italic">
    [此论点已被管理员折叠<span v-if="argument.fold_note">：{{ argument.fold_note }}</span>]
  </div>

  <!-- Content (dimmed for admin when folded) -->
  <div v-if="!localIsFolded || authStore.user?.role === 'admin'" :style="localIsFolded ? 'opacity:.4' : ''">
    <!-- Quote -->
    <div v-if="quotedArgument" class="mb-3 border-l-4 border-black bg-gray-50 p-3">
      <div class="text-xs font-black uppercase tracking-widest mb-1">
        引用 {{ quotedAuthorName }}
      </div>
      <p class="text-sm text-gray-600 leading-6">
        {{ quotedArgument.content.substring(0, 140) }}
      </p>
    </div>

    <!-- Content -->
    <div class="mb-3">
      <p class="font-medium">{{ argument.content }}</p>
    </div>

    <!-- Evidence source card -->
    <div
      v-if="argument.argument_type === 'evidence' && argument.source_url"
      style="margin-top:.5rem;padding:.5rem .75rem;border:1px solid var(--a-border);border-radius:.375rem;font-size:.75rem;margin-bottom:.75rem"
    >
      <a :href="argument.source_url" target="_blank" rel="noopener noreferrer" style="font-weight:700;display:block;margin-bottom:.2rem">
        {{ argument.source_title || argument.source_url }}
      </a>
      <p v-if="argument.source_excerpt" style="color:var(--a-color-muted);margin:0;font-style:italic">
        "{{ argument.source_excerpt }}"
      </p>
    </div>
  </div>

  <!-- References -->
  <div v-if="argument.references && argument.references.length > 0" class="mb-3">
    <div class="a-label mb-2">引用:</div>
    <div class="space-y-2">
      <div
        v-for="ref in argument.references"
        :key="ref.id"
        class="bg-gray-50 border-2 border-black p-3"
      >
        <span class="font-bold text-xs">[{{ typeLabels[ref.argument_type] }}]</span>
        <p class="text-sm mt-1">{{ ref.content.substring(0, 80) }}...</p>
      </div>
    </div>
  </div>

  <!-- Referenced Debates -->
  <div v-if="argument.referenced_debates && argument.referenced_debates.length > 0" class="mb-3">
    <div class="a-label mb-2">引用辩题:</div>
    <div class="space-y-2">
      <div
        v-for="d in argument.referenced_debates"
        :key="d.id"
        class="bg-blue-50 border-2 border-black p-3"
      >
        <div class="flex items-center gap-2">
          <span
            v-if="d.conclusion_type"
            class="text-xs font-black px-2 py-0.5 border border-current"
            :style="conclusionBadgeStyles[d.conclusion_type]"
          >
            {{ conclusionLabels[d.conclusion_type] }}
          </span>
          <span class="text-xs font-bold text-gray-500">辩题</span>
        </div>
        <p class="text-sm font-medium mt-1">{{ d.title }}</p>
      </div>
    </div>
  </div>

  <!-- Conclusion -->
  <div v-if="argument.conclusion" class="mb-3 p-3 bg-green-50 border-2 border-green-600">
    <div class="a-label text-green-700 mb-1">结论:</div>
    <p class="text-sm font-medium">{{ argument.conclusion }}</p>
  </div>

  <!-- Creator & Date -->
  <div class="flex items-center justify-between text-xs text-gray-500 mb-3">
    <span>{{ argument.user?.username || '未知用户' }}</span>
    <span>{{ formatDate(argument.created_at) }}</span>
  </div>
</div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Argument, Debate } from '@/types'
import { useAuthStore } from '@/stores/auth'
import ABtn from '@/components/ui/ABtn.vue'

const props = defineProps<{
  argument: Argument
  debate: Debate
  quotedArgument?: Argument | null
  userVotes?: Record<string, number>
}>()

const emit = defineEmits<{
  vote: [argumentId: string, voteType: number]
  reply: [argument: Argument]
  edit: [argument: Argument]
  delete: [argument: Argument]
  reference: [argumentId: string, referenceId: string]
}>()

const authStore = useAuthStore()

const apiBase = import.meta.env.VITE_API_URL || '/api'
const localIsFolded = ref(props.argument.is_folded ?? false)

const foldArgument = async () => {
  const note = prompt('折叠理由（可选）：', '') ?? ''
  const res = await fetch(`${apiBase}/debate/arguments/${props.argument.id}/fold`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${authStore.token}` },
    body: JSON.stringify({ fold_note: note }),
  })
  if (res.ok) localIsFolded.value = true
}

const unfoldArgument = async () => {
  const res = await fetch(`${apiBase}/debate/arguments/${props.argument.id}/fold`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${authStore.token}` },
  })
  if (res.ok) localIsFolded.value = false
}

const canVote = computed(() => {
  return authStore.isAuthenticated && props.debate.status === 'open'
})

const canEdit = computed(() => {
  if (!authStore.isAuthenticated) return false
  if (authStore.user?.role === 'admin') return true
  return String(props.argument.user_id) === String(authStore.user?.id)
})

const canDelete = computed(() => {
  if (!authStore.isAuthenticated) return false
  if (authStore.user?.role === 'admin') return true
  return String(props.argument.user_id) === String(authStore.user?.id)
})

const canReply = computed(() => {
  return authStore.isAuthenticated && props.debate.status === 'open'
})

const userVote = computed(() => {
  return props.userVotes?.[props.argument.id] ?? null
})

const quotedAuthorName = computed(() => {
  if (!props.quotedArgument) return ''
  return props.quotedArgument.user?.display_name || props.quotedArgument.user?.username || '未知用户'
})

const typeLabels: Record<string, string> = {
  support: '支持',
  oppose: '反对',
  neutral: '中立',
  evidence: '证据',
  question: '提问',
  counter: '反驳',
}

const typeBadgeClasses: Record<string, string> = {
  support: '',
  oppose: '',
  neutral: '',
  evidence: '',
  question: '',
  counter: '',
}

const typeBadgeStyles: Record<string, any> = {
  support: { backgroundColor: '#16a34a', borderColor: '#16a34a', color: '#fff' },
  oppose: { backgroundColor: '#dc2626', borderColor: '#dc2626', color: '#fff' },
  neutral: { backgroundColor: '#6b7280', borderColor: '#6b7280', color: '#fff' },
  evidence: { backgroundColor: '#2563eb', borderColor: '#2563eb', color: '#fff' },
  question: { backgroundColor: '#ca8a04', borderColor: '#ca8a04', color: '#fff' },
  counter: { backgroundColor: '#9333ea', borderColor: '#9333ea', color: '#fff' },
}

const conclusionLabels: Record<string, string> = {
  yes: '是',
  no: '否',
  inconclusive: '无定论',
}

const conclusionBadgeStyles: Record<string, any> = {
  yes: { color: '#16a34a', borderColor: '#16a34a' },
  no: { color: '#dc2626', borderColor: '#dc2626' },
  inconclusive: { color: '#6b7280', borderColor: '#6b7280' },
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>