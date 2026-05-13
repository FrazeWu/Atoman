<template>
  <div class="min-h-screen bg-white p-8">
    <!-- Loading -->
    <div v-if="loading" class="p-8 text-center">
      <p class="font-bold">加载中...</p>
    </div>

    <template v-else-if="debate">
      <!-- Header -->
      <div class="border-b-2 border-black p-6 mb-6">
        <div class="flex items-start justify-between mb-4">
          <div>
            <h1 class="text-4xl font-black tracking-tighter mb-2">{{ debate.title }}</h1>
            <p class="text-lg font-medium text-gray-700">{{ debate.description }}</p>
          </div>
          <span
            class="text-xs font-black uppercase tracking-widest px-3 py-2 border-2 border-black"
            :class="{
              'bg-green-100': debate.status === 'open',
              'bg-gray-200': debate.status === 'concluded',
              'bg-gray-100': debate.status === 'archived',
            }"
          >
            {{ statusLabels[debate.status] }}
          </span>
        </div>

        <!-- Tags -->
        <div class="flex flex-wrap gap-2 mb-4">
          <span
            v-for="tag in debate.tags"
            :key="tag"
            class="text-xs font-bold px-3 py-1 bg-gray-100 border border-black"
          >
            #{{ tag }}
          </span>
        </div>

        <!-- Stats & Actions -->
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-6 text-sm font-bold text-gray-600">
            <span>{{ debate.argument_count }} 论点</span>
            <span>{{ debate.vote_count }} 投票</span>
            <span>{{ debate.view_count }}浏览</span>
            <span>{{ formatDate(debate.created_at) }}</span>
          </div>

          <div class="flex gap-2 flex-wrap justify-end">
            <!-- Vote to conclude (any auth user, when open) -->
            <ABtn
              v-if="debate.status === 'open' && authStore.isAuthenticated && !canConclude"
              variant="secondary"
              outline
              @click="handleVoteToConclude"
              :disabled="concludeVoting"
            >
              投票结题 ({{ debate.conclude_vote_count ?? 0 }}/{{ debate.conclude_threshold ?? 10 }})
            </ABtn>

            <!-- Conclude (owner or admin) -->
            <ABtn
              v-if="canConclude"
              variant="secondary"
              @click="showConcludeModal = true"
            >
              结题
            </ABtn>

            <!-- Reopen (admin only) -->
            <ABtn
              v-if="canReopen"
              outline
              @click="handleReopen"
              :disabled="reopening"
            >
              重开辩论
            </ABtn>

            <ABtn
              v-if="canEdit"
              variant="secondary"
              @click="openEditDebateModal"
            >
              编辑
            </ABtn>
          </div>
        </div>
      </div>

      <!-- Conclusion Banner (when concluded) -->
      <div
        v-if="debate.status === 'concluded' && debate.conclusion_type"
        class="mx-6 mb-6 p-4 border-2 flex items-start gap-4"
        :style="conclusionBannerStyles[debate.conclusion_type]"
      >
        <div>
          <div class="text-xs font-black uppercase tracking-widest mb-1">结论</div>
          <div class="text-2xl font-black">{{ conclusionLabels[debate.conclusion_type] }}</div>
          <p v-if="debate.conclusion_summary" class="text-sm mt-2 font-medium">
            {{ debate.conclusion_summary }}
          </p>
        </div>
      </div>

      <!-- Content -->
      <div v-if="debate.content" class="border-b-2 border-black p-6">
        <h2 class="text-sm font-black uppercase tracking-widest mb-4">背景内容</h2>
        <div class="prose max-w-none" v-html="renderMarkdown(debate.content)"></div>
      </div>

      <!-- Arguments Section -->
      <div class="p-6">
        <div class="flex items-center justify-between mb-6">
          <h2 class="text-2xl font-black tracking-tight">论点列表</h2>
          <ABtn
            v-if="debate.status === 'open' && authStore.isAuthenticated"
            variant="primary"
            @click="showAddArgumentModal = true"
          >
            添加论点
          </ABtn>
        </div>

        <!-- Argument Tree -->
        <div v-if="argumentsList.length > 0" class="space-y-4">
          <ArgumentNode
            v-for="arg in argumentsList"
            :key="arg.id"
            :argument="arg"
            :debate="debate"
            :quoted-argument="getQuotedArgument(arg)"
            :user-votes="debateStore.userVotes"
            @vote="handleVote"
            @reply="handleReply"
            @edit="openEditArgumentModal"
            @delete="handleDeleteArgument"
            @reference="openReferenceModal"
          />
        </div>

        <AEmpty v-else message="暂无论点，快来添加第一个论点吧！" />
      </div>
    </template>

    <!-- Edit Debate Modal -->
    <AModal v-if="showEditModal" @close="showEditModal = false">
      <div class="p-6">
        <h3 class="a-title-sm mb-6">编辑辩论</h3>
        <form @submit.prevent="handleUpdate" class="space-y-4">
          <div class="a-field">
            <label class="a-field-label">标题</label>
            <input
              v-model="editForm.title"
              required
              class="a-input"
            />
          </div>
          <div class="a-field">
            <label class="a-field-label">描述</label>
            <input
              v-model="editForm.description"
              required
              class="a-input"
            />
          </div>
          <div class="a-field">
            <label class="a-field-label">内容</label>
            <textarea
              v-model="editForm.content"
              rows="4"
              class="a-textarea"
            ></textarea>
          </div>
          <div class="a-field">
            <label class="a-field-label">标签</label>
            <input
              v-model="tagsInput"
              class="a-input"
            />
          </div>
          <div class="flex justify-end gap-4 mt-6">
            <ABtn outline type="button" @click="showEditModal = false">取消</ABtn>
            <ABtn type="submit">保存</ABtn>
          </div>
        </form>
      </div>
    </AModal>

    <!-- Add Argument Modal -->
    <AModal v-if="showAddArgumentModal" @close="handleCloseArgumentModal">
      <div class="p-6">
        <h3 class="a-title-sm mb-6">
          {{ selectedParentId ? '引用论点' : '添加论点' }}
        </h3>

        <!-- Show selected quote if quoting -->
        <div v-if="selectedParentId" class="a-card-sm mb-4">
          <div class="a-label mb-2">引用:</div>
          <p class="text-sm">{{ selectedParentContent }}</p>
        </div>

        <form @submit.prevent="handleCreateArgument" class="space-y-4">
          <div class="a-field">
            <label class="a-field-label">论点类型</label>
            <ASelect
              v-model="newArgument.argument_type"
              :options="argumentTypeOptions"
            />
          </div>
          <div class="a-field">
            <label class="a-field-label">论点内容</label>
            <AEditor
              v-model="newArgument.content"
              mode="plain"
              placeholder="阐述你的观点…"
            />
          </div>
          <div class="flex justify-end gap-4 mt-6">
            <ABtn outline type="button" @click="handleCloseArgumentModal">取消</ABtn>
            <ABtn type="submit">添加</ABtn>
          </div>
        </form>
      </div>
    </AModal>

    <!-- Edit Argument Modal -->
    <AModal v-if="showEditArgumentModal" @close="showEditArgumentModal = false">
      <div class="p-6">
        <h3 class="a-title-sm mb-6">编辑论点</h3>
        <form @submit.prevent="handleEditArgumentSubmit" class="space-y-4">
          <div class="a-field">
            <label class="a-field-label">论点类型</label>
            <ASelect v-model="editArgumentForm.argument_type" :options="argumentTypeOptions" />
          </div>
          <div class="a-field">
            <label class="a-field-label">论点内容</label>
            <AEditor
              v-model="editArgumentForm.content"
              mode="plain"
              placeholder="阐述你的观点…"
            />
          </div>
          <div class="flex justify-end gap-4 mt-6">
            <ABtn outline type="button" @click="showEditArgumentModal = false">取消</ABtn>
            <ABtn type="submit" :disabled="editArgumentSaving">
              {{ editArgumentSaving ? '保存中...' : '保存' }}
            </ABtn>
          </div>
        </form>
      </div>
    </AModal>

    <!-- Conclude Modal -->
    <AModal v-if="showConcludeModal" @close="showConcludeModal = false">
      <div class="p-6">
        <h3 class="a-title-sm mb-6">结题</h3>
        <form @submit.prevent="handleConcludeSubmit" class="space-y-4">
          <div class="a-field">
            <label class="a-field-label">结论</label>
            <div class="flex gap-4 mt-2">
              <label
                v-for="opt in conclusionOptions"
                :key="opt.value"
                class="flex items-center gap-2 cursor-pointer"
              >
                <input
                  type="radio"
                  v-model="concludeForm.conclusion_type"
                  :value="opt.value"
                  class="w-4 h-4"
                />
                <span
                  class="text-sm font-bold px-3 py-1 border-2"
                  :style="opt.style"
                >
                  {{ opt.label }}
                </span>
              </label>
            </div>
          </div>
          <div class="a-field">
            <label class="a-field-label">结论说明（可选）</label>
            <textarea
              v-model="concludeForm.conclusion_summary"
              rows="3"
              class="a-textarea"
              placeholder="详细说明结论..."
            ></textarea>
          </div>
          <div class="flex justify-end gap-4 mt-6">
            <ABtn outline type="button" @click="showConcludeModal = false">取消</ABtn>
            <ABtn type="submit" :disabled="concluding || !concludeForm.conclusion_type">
              {{ concluding ? '结题中...' : '确认结题' }}
            </ABtn>
          </div>
        </form>
      </div>
    </AModal>

    <!-- Reference Debate Modal -->
    <AModal v-if="showReferenceModal" @close="showReferenceModal = false">
      <div class="p-6">
        <h3 class="a-title-sm mb-6">引用辩题</h3>
        <div class="mb-4">
          <div class="a-field">
            <label class="a-label">搜索辩题</label>
            <div class="flex gap-2 mt-1">
              <input
                v-model="referenceSearchQuery"
                class="a-input flex-1"
                placeholder="输入辩题标题..."
                @keyup.enter="searchForDebates"
              />
              <ABtn outline @click="searchForDebates" :disabled="referenceSearching">
                {{ referenceSearching ? '搜索中...' : '搜索' }}
              </ABtn>
            </div>
          </div>
        </div>

        <!-- Search Results -->
        <div v-if="referenceSearchResults.length > 0" class="space-y-2 max-h-64 overflow-y-auto">
          <div
            v-for="d in referenceSearchResults"
            :key="d.id"
            class="border-2 border-black p-3 cursor-pointer hover:bg-gray-50 transition-colors"
            :class="{ 'bg-black text-white': referenceSelectedDebateId === d.id }"
            @click="referenceSelectedDebateId = d.id"
          >
            <div class="flex items-center gap-2">
              <span
                v-if="d.conclusion_type"
                class="text-xs font-black px-2 py-0.5 border"
                :class="referenceSelectedDebateId === d.id ? 'border-white' : ''"
                :style="referenceSelectedDebateId !== d.id ? conclusionBadgeStyles[d.conclusion_type] : {}"
              >
                {{ conclusionLabels[d.conclusion_type] }}
              </span>
              <span class="text-xs font-bold" :class="referenceSelectedDebateId === d.id ? 'text-gray-300' : 'text-gray-500'">
                {{ statusLabels[d.status] }}
              </span>
            </div>
            <p class="text-sm font-medium mt-1">{{ d.title }}</p>
          </div>
        </div>
        <div v-else-if="referenceSearchQuery && !referenceSearching" class="text-sm text-gray-500 text-center py-4">
          未找到相关辩题
        </div>

        <div class="flex justify-end gap-4 mt-6">
          <ABtn outline type="button" @click="showReferenceModal = false">取消</ABtn>
          <ABtn
            @click="handleAddDebateReference"
            :disabled="!referenceSelectedDebateId || referenceAdding"
          >
            {{ referenceAdding ? '添加中...' : '添加引用' }}
          </ABtn>
        </div>
      </div>
    </AModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useDebateStore } from '@/stores/debate'
import { useAuthStore } from '@/stores/auth'
import type { Debate, Argument, ArgumentType } from '@/types'
import ABtn from '@/components/ui/ABtn.vue'
import AModal from '@/components/ui/AModal.vue'
import AEmpty from '@/components/ui/AEmpty.vue'
import ArgumentNode from '@/components/debate/ArgumentNode.vue'
import { useMarkdownRenderer } from '@/composables/useMarkdownRenderer'
import AEditor from '@/components/shared/AEditor.vue'
import ASelect from '@/components/ui/ASelect.vue'

const router = useRouter()
const route = useRoute()
const debateStore = useDebateStore()
const authStore = useAuthStore()
const { renderMarkdown } = useMarkdownRenderer()

const debate = computed(() => debateStore.currentDebate)
const argumentsList = computed(() => debateStore.argumentList)
const loading = computed(() => debateStore.loading)

const selectedParentId = ref<string | null>(null)

const canEdit = computed(() => {
  if (!authStore.isAuthenticated) return false
  if (authStore.user?.role === 'admin') return true
  return String(debate.value?.user_id) === String(authStore.user?.id)
})

const canConclude = computed(() => {
  if (!authStore.isAuthenticated) return false
  if (debate.value?.status !== 'open') return false
  if (authStore.user?.role === 'admin') return true
  return String(debate.value?.user_id) === String(authStore.user?.id)
})

const canReopen = computed(() => {
  if (!authStore.isAuthenticated) return false
  if (debate.value?.status === 'open') return false
  return authStore.user?.role === 'admin'
})

const statusLabels: Record<string, string> = {
  open: '进行中',
  concluded: '已结题',
  archived: '已归档',
}

const conclusionLabels: Record<string, string> = {
  yes: '是',
  no: '否',
  inconclusive: '无定论',
}

const conclusionOptions = [
  { value: 'yes', label: '是', style: { color: 'var(--a-color-success)', borderColor: 'var(--a-color-success)' } },
  { value: 'no', label: '否', style: { color: 'var(--a-color-danger)', borderColor: 'var(--a-color-danger)' } },
  { value: 'inconclusive', label: '无定论', style: { color: 'var(--a-color-muted)', borderColor: 'var(--a-color-muted)' } },
]

const conclusionBannerStyles: Record<string, any> = {
  yes: { backgroundColor: 'var(--a-color-success-bg)', borderColor: 'var(--a-color-success)' },
  no: { backgroundColor: '#fef2f2', borderColor: 'var(--a-color-danger)' },
  inconclusive: { backgroundColor: 'var(--a-color-surface)', borderColor: 'var(--a-color-muted)' },
}

const conclusionBadgeStyles: Record<string, any> = {
  yes: { color: 'var(--a-color-success)', borderColor: 'var(--a-color-success)' },
  no: { color: 'var(--a-color-danger)', borderColor: 'var(--a-color-danger)' },
  inconclusive: { color: 'var(--a-color-muted)', borderColor: 'var(--a-color-muted)' },
}

const argumentTypeOptions = [
  { label: '支持', value: 'support' },
  { label: '反对', value: 'oppose' },
  { label: '中立', value: 'neutral' },
  { label: '证据', value: 'evidence' },
  { label: '提问', value: 'question' },
  { label: '反驳', value: 'counter' },
]

// Edit Debate modal
const showEditModal = ref(false)
const editForm = ref({ title: '', description: '', content: '' })
const tagsInput = ref('')

const openEditDebateModal = () => {
  if (!debate.value) return
  editForm.value = {
    title: debate.value.title,
    description: debate.value.description,
    content: debate.value.content,
  }
  tagsInput.value = (debate.value.tags ?? []).join(', ')
  showEditModal.value = true
}

// Add Argument modal
const showAddArgumentModal = ref(false)
const newArgument = ref({
  content: '',
  argument_type: 'support' as ArgumentType,
})

const selectedParentContent = computed(() => {
  if (!selectedParentId.value) return ''
  const parent = argumentsList.value.find(a => a.id === selectedParentId.value)
  return parent ? parent.content.substring(0, 100) + '...' : ''
})

const getQuotedArgument = (argument: Argument) => {
  if (!argument.parent_id) return null
  return argumentsList.value.find(candidate => candidate.id === argument.parent_id) ?? null
}

// Edit Argument modal
const showEditArgumentModal = ref(false)
const editArgumentTarget = ref<Argument | null>(null)
const editArgumentForm = ref({ content: '', argument_type: 'support' as ArgumentType })
const editArgumentSaving = ref(false)

const openEditArgumentModal = (arg: Argument) => {
  editArgumentTarget.value = arg
  editArgumentForm.value = {
    content: arg.content,
    argument_type: arg.argument_type,
  }
  showEditArgumentModal.value = true
}

// Conclude modal
const showConcludeModal = ref(false)
const concluding = ref(false)
const concludeForm = ref<{ conclusion_type: 'yes' | 'no' | 'inconclusive' | ''; conclusion_summary: string }>({
  conclusion_type: '',
  conclusion_summary: '',
})

// Vote to conclude
const concludeVoting = ref(false)

// Reopen
const reopening = ref(false)

// Reference debate modal
const showReferenceModal = ref(false)
const referenceTargetArgumentId = ref<string>('')
const referenceSearchQuery = ref('')
const referenceSearchResults = ref<Debate[]>([])
const referenceSearching = ref(false)
const referenceSelectedDebateId = ref<string>('')
const referenceAdding = ref(false)

const loadDebate = async () => {
  const id = route.params.id as string
  await debateStore.fetchDebate(id)
  await debateStore.fetchArguments(id)
}

const handleUpdate = async () => {
  const tags = tagsInput.value
    .split(',')
    .map(t => t.trim())
    .filter(Boolean)

  const updated = await debateStore.updateDebate(debate.value!.id, {
    ...editForm.value,
    tags,
  })

  if (updated) {
    showEditModal.value = false
    await loadDebate()
  }
}

const handleConcludeSubmit = async () => {
  if (!concludeForm.value.conclusion_type) return
  concluding.value = true
  const updated = await debateStore.concludeDebate(debate.value!.id, {
    conclusion_type: concludeForm.value.conclusion_type as 'yes' | 'no' | 'inconclusive',
    conclusion_summary: concludeForm.value.conclusion_summary || undefined,
  })
  concluding.value = false
  if (updated) {
    showConcludeModal.value = false
    concludeForm.value = { conclusion_type: '', conclusion_summary: '' }
    await loadDebate()
  }
}

const handleVoteToConclude = async () => {
  if (!authStore.isAuthenticated) {
    router.push('/login')
    return
  }
  concludeVoting.value = true
  const result = await debateStore.voteToConclude(debate.value!.id)
  concludeVoting.value = false
  if (result) {
    await loadDebate()
  }
}

const handleReopen = async () => {
  if (!confirm('确定要重开这个辩论吗？')) return
  reopening.value = true
  const updated = await debateStore.reopenDebate(debate.value!.id)
  reopening.value = false
  if (updated) {
    await loadDebate()
  }
}

const handleCreateArgument = async () => {
  try {
    const payload = {
      content: newArgument.value.content,
      argument_type: newArgument.value.argument_type,
      parent_id: selectedParentId.value || undefined,
    }

    const result = await debateStore.createArgument(debate.value!.id, payload)
    if (result) {
      handleCloseArgumentModal()
      await debateStore.fetchArguments(debate.value!.id)
    }
  } catch (error) {
    console.error('Failed to create argument:', error)
    alert('创建论点失败：' + (error as Error).message)
  }
}

const handleVote = async (argumentId: string, voteType: number) => {
  if (!authStore.isAuthenticated) {
    router.push('/login')
    return
  }
  await debateStore.voteArgument(argumentId, voteType)
  await debateStore.fetchArguments(debate.value!.id)
}

const handleReply = (parentArg: Argument) => {
  selectedParentId.value = parentArg.id
  newArgument.value = {
    content: '',
    argument_type: 'support',
  }
  showAddArgumentModal.value = true
}

const handleCloseArgumentModal = () => {
  selectedParentId.value = null
  showAddArgumentModal.value = false
  newArgument.value = {
    content: '',
    argument_type: 'support',
  }
}

const handleEditArgumentSubmit = async () => {
  if (!editArgumentTarget.value) return
  editArgumentSaving.value = true
  const updated = await debateStore.updateArgument(editArgumentTarget.value.id, {
    content: editArgumentForm.value.content,
    argument_type: editArgumentForm.value.argument_type,
  })
  editArgumentSaving.value = false
  if (updated) {
    showEditArgumentModal.value = false
    editArgumentTarget.value = null
    await debateStore.fetchArguments(debate.value!.id)
  }
}

const handleDeleteArgument = async (arg: Argument) => {
  if (!confirm('确定要删除这个论点吗？引用它的论点会保留。')) return

  const success = await debateStore.deleteArgument(arg.id)
  if (success) {
    await debateStore.fetchArguments(debate.value!.id)
  }
}

const openReferenceModal = (argumentId: string, _refId: string) => {
  referenceTargetArgumentId.value = argumentId
  referenceSearchQuery.value = ''
  referenceSearchResults.value = []
  referenceSelectedDebateId.value = ''
  showReferenceModal.value = true
}

const searchForDebates = async () => {
  if (!referenceSearchQuery.value.trim()) return
  referenceSearching.value = true
  const results = await debateStore.searchDebates(referenceSearchQuery.value.trim(), 10)
  // Exclude the current debate
  referenceSearchResults.value = results.filter(d => d.id !== debate.value?.id)
  referenceSearching.value = false
}

const handleAddDebateReference = async () => {
  if (!referenceSelectedDebateId.value || !referenceTargetArgumentId.value) return
  referenceAdding.value = true
  const success = await debateStore.addDebateReference(
    referenceTargetArgumentId.value,
    referenceSelectedDebateId.value,
  )
  referenceAdding.value = false
  if (success) {
    showReferenceModal.value = false
    await debateStore.fetchArguments(debate.value!.id)
  }
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  })
}

onMounted(() => {
  loadDebate()
})
</script>
