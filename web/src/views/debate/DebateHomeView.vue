<template>
  <div class="a-page-xl" style="padding-bottom:6rem">
    <APageHeader title="辩论" accent sub="结构化讨论，解决问题">
      <template #action>
        <ABtn v-if="authStore.isAuthenticated" @click="showCreateModal = true">发起辩论</ABtn>
      </template>
    </APageHeader>

    <!-- Filters -->
    <div style="display:flex;gap:1rem;margin-bottom:1.5rem;flex-wrap:wrap">
      <select
        v-model="filterStatus"
        @change="loadDebates"
        class="a-select"
        style="max-width:200px"
      >
        <option value="">全部状态</option>
        <option value="open">进行中</option>
        <option value="concluded">已结题</option>
        <option value="archived">已归档</option>
      </select>

      <input
        v-model="filterTag"
        @keyup.enter="loadDebates"
        placeholder="标签筛选"
        class="a-input"
        style="max-width:300px"
      />

      <ABtn outline @click="loadDebates">筛选</ABtn>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="p-8 text-center">
      <p class="font-bold">加载中...</p>
    </div>

    <!-- Empty State -->
    <AEmpty v-else-if="debates.length === 0" text="暂无辩论" />

    <!-- Debate List -->
    <div v-else class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
      <div
        v-for="debate in debates"
        :key="debate.id"
        class="a-card a-card-hover cursor-pointer"
        @click="goToDebate(debate.id)"
      >
        <div class="flex items-start justify-between mb-4">
          <h3 class="text-xl font-black tracking-tight flex-1">{{ debate.title }}</h3>
          <div class="flex items-center gap-2 flex-shrink-0">
            <span
              v-if="debate.conclusion_type"
              class="text-xs font-black px-2 py-1 border-2"
              :style="conclusionBadgeStyles[debate.conclusion_type]"
            >
              {{ conclusionLabels[debate.conclusion_type] }}
            </span>
            <span
              class="a-badge"
              :class="{
                'a-badge-fill': debate.status === 'open',
              }"
              :style="{
                backgroundColor: debate.status === 'open' ? '#16a34a' : undefined,
                borderColor: debate.status === 'open' ? '#16a34a' : undefined,
                color: debate.status === 'open' ? '#fff' : undefined,
              }"
            >
              {{ statusLabels[debate.status] }}
            </span>
          </div>
        </div>

        <p class="text-sm font-medium text-gray-700 mb-4 line-clamp-2">
          {{ debate.description }}
        </p>

        <!-- Tags -->
        <div class="flex flex-wrap gap-2 mb-4">
          <span
            v-for="tag in debate.tags"
            :key="tag"
            class="text-xs font-bold px-2 py-1 bg-gray-100 border border-black"
          >
            #{{ tag }}
          </span>
        </div>

        <!-- Stats -->
        <div class="flex items-center gap-4 text-xs font-bold text-gray-600">
          <span>{{ debate.argument_count }} 论点</span>
          <span>{{ debate.vote_count }} 投票</span>
          <span>{{ debate.view_count }}浏览</span>
        </div>

        <!-- Creator & Date -->
        <div class="mt-4 pt-4 border-t border-gray-200 text-xs text-gray-500">
          <span>{{ debate.user.username }}</span>
          <span class="mx-2">·</span>
          <span>{{ formatDate(debate.created_at) }}</span>
        </div>
      </div>
    </div>

    <!-- Load More -->
    <div v-if="debates.length > 0 && debates.length < debatesTotal" class="mt-6 text-center">
      <ABtn outline @click="loadMore">加载更多</ABtn>
    </div>

    <!-- Create Modal -->
    <AModal v-if="showCreateModal" @close="showCreateModal = false">
      <div class="p-6">
        <h3 class="a-title-sm mb-6">发起辩论</h3>

        <form @submit.prevent="handleCreate" class="space-y-4">
          <div class="a-field">
            <label class="a-field-label">标题</label>
            <input
              v-model="createForm.title"
              required
              class="a-input"
              placeholder="辩论主题"
            />
          </div>

          <div class="a-field">
            <label class="a-field-label">描述</label>
            <input
              v-model="createForm.description"
              required
              class="a-input"
              placeholder="简短描述"
            />
          </div>

          <div class="a-field">
            <label class="a-field-label">背景内容</label>
            <textarea
              v-model="createForm.content"
              rows="4"
              class="a-textarea"
              placeholder="详细说明..."
            ></textarea>
          </div>

          <div class="a-field">
            <label class="a-field-label">标签（逗号分隔）</label>
            <input
              v-model="tagsInput"
              class="a-input"
              placeholder="例如：科技，社会，哲学"
            />
          </div>

          <div class="flex justify-end gap-4 mt-6">
            <ABtn outline type="button" @click="showCreateModal = false">取消</ABtn>
            <ABtn type="submit" :disabled="creating">
              {{ creating ? '创建中...' : '创建' }}
            </ABtn>
          </div>
        </form>
      </div>
    </AModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useDebateStore } from '@/stores/debate'
import { useAuthStore } from '@/stores/auth'
import type { Debate } from '@/types'
import ABtn from '@/components/ui/ABtn.vue'
import AModal from '@/components/ui/AModal.vue'
import AEmpty from '@/components/ui/AEmpty.vue'
import APageHeader from '@/components/ui/APageHeader.vue'

const router = useRouter()
const debateStore = useDebateStore()
const authStore = useAuthStore()

const debates = computed(() => debateStore.debates)
const debatesTotal = computed(() => debateStore.debatesTotal)
const loading = computed(() => debateStore.loading)

const filterStatus = ref('')
const filterTag = ref('')
const currentPage = ref(1)
const limit = 12

const showCreateModal = ref(false)
const creating = ref(false)
const createForm = ref({
  title: '',
  description: '',
  content: '',
})
const tagsInput = ref('')

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

const conclusionBadgeStyles: Record<string, any> = {
  yes: { color: '#16a34a', borderColor: '#16a34a' },
  no: { color: '#dc2626', borderColor: '#dc2626' },
  inconclusive: { color: '#6b7280', borderColor: '#6b7280' },
}

const loadDebates = async () => {
  currentPage.value = 1
  await debateStore.fetchDebates({
    status: filterStatus.value as 'open' | 'concluded' | 'archived' | undefined,
    tag: filterTag.value.trim() || undefined,
    page: currentPage.value,
    limit,
  })
}

const loadMore = async () => {
  currentPage.value++
  const currentCount = debates.value.length
  await debateStore.fetchDebates({
    status: filterStatus.value as 'open' | 'concluded' | 'archived' | undefined,
    tag: filterTag.value.trim() || undefined,
    page: currentPage.value,
    limit,
  })
}

const handleCreate = async () => {
  if (!authStore.isAuthenticated) {
    router.push('/login')
    return
  }

  creating.value = true
  const tags = tagsInput.value
    .split(',')
    .map(t => t.trim())
    .filter(Boolean)

  const debate = await debateStore.createDebate({
    ...createForm.value,
    tags,
  })

  if (debate) {
    showCreateModal.value = false
    createForm.value = { title: '', description: '', content: '' }
    tagsInput.value = ''
    loadDebates()
  }
  creating.value = false
}

const goToDebate = (id: string) => {
  router.push(`/debate/${id}`)
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('zh-CN')
}

onMounted(() => {
  loadDebates()
})
</script>
