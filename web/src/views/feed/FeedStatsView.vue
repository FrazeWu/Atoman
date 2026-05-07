<template>
  <div class="a-page-xl" style="padding-bottom:8rem">
    <APageHeader title="阅读统计" accent sub="按时间和订阅源查看你的 RSS 阅读趋势" style="margin-bottom:2rem">
      <template #action>
        <div style="display:flex;gap:.75rem;flex-wrap:wrap;justify-content:flex-end">
          <ABtn outline to="/feed">返回订阅</ABtn>
        </div>
      </template>
    </APageHeader>

    <div v-if="!authStore.isAuthenticated" style="min-height:40vh;display:flex;flex-direction:column;align-items:center;justify-content:center;text-align:center">
      <p class="a-title-xl a-muted" style="margin-bottom:1rem">阅读统计</p>
      <p class="a-muted" style="max-width:28rem;margin-bottom:1.5rem">登录后即可查看你的 RSS 阅读趋势和来源分布。</p>
      <ABtn to="/login">登录</ABtn>
    </div>

    <template v-else>
      <div style="display:flex;justify-content:space-between;align-items:flex-start;gap:1rem;flex-wrap:wrap;margin-bottom:1.5rem">
        <div style="display:flex;gap:.5rem;flex-wrap:wrap">
          <button
            v-for="option in periodOptions"
            :key="option.value"
            @click="selectPeriod(option.value)"
            style="font-weight:900;font-size:.72rem;text-transform:uppercase;letter-spacing:.08em;padding:.55rem 1rem;border:2px solid #000;cursor:pointer;transition:all .2s;background:#fff"
            :style="period === option.value ? 'background:#000;color:#fff' : ''"
          >{{ option.label }}</button>
        </div>
        <button
          @click="fetchStats"
          style="font-weight:900;font-size:.72rem;text-transform:uppercase;letter-spacing:.08em;padding:.55rem 1rem;border:2px solid #000;cursor:pointer;transition:all .2s;background:#fff"
          :style="loading ? 'opacity:.5;cursor:not-allowed' : ''"
          :disabled="loading"
        >{{ loading ? '刷新中...' : '刷新' }}</button>
      </div>

      <div v-if="stats" style="display:grid;grid-template-columns:repeat(auto-fit,minmax(12rem,1fr));gap:1rem;margin-bottom:1.5rem">
        <div class="a-card" style="padding:1.25rem">
          <div class="a-label a-muted" style="margin-bottom:.5rem">阅读总量</div>
          <div style="font-size:2rem;font-weight:900;letter-spacing:-0.04em">{{ stats.total_read }}</div>
        </div>
        <div class="a-card" style="padding:1.25rem">
          <div class="a-label a-muted" style="margin-bottom:.5rem">活跃订阅源</div>
          <div style="font-size:2rem;font-weight:900;letter-spacing:-0.04em">{{ stats.source_breakdown.length }}</div>
        </div>
        <div class="a-card" style="padding:1.25rem">
          <div class="a-label a-muted" style="margin-bottom:.5rem">最常阅读</div>
          <div style="font-size:1.2rem;font-weight:900;letter-spacing:-0.03em">{{ topSourceLabel }}</div>
          <div style="font-size:.8rem;color:#6b7280;margin-top:.35rem">{{ topSourceReads }}</div>
        </div>
      </div>

      <div v-if="loading" style="display:grid;grid-template-columns:1.4fr 1fr;gap:1.25rem">
        <div class="a-skeleton" style="height:24rem"></div>
        <div class="a-skeleton" style="height:24rem"></div>
      </div>

      <AEmpty v-else-if="!stats || !stats.total_read" text="还没有阅读记录，先去刷一会儿 RSS 再来看统计" />

      <div v-else style="display:grid;grid-template-columns:minmax(0,1.5fr) minmax(20rem,1fr);gap:1.25rem;align-items:start">
        <section class="a-card" style="padding:1.25rem">
          <div style="display:flex;align-items:center;justify-content:space-between;gap:1rem;margin-bottom:1rem;flex-wrap:wrap">
            <div>
              <div class="a-label a-muted" style="margin-bottom:.35rem">趋势</div>
              <h2 style="margin:0;font-size:1.35rem;font-weight:900;letter-spacing:-0.04em">{{ periodTitle }}</h2>
            </div>
            <div style="font-size:.78rem;color:#6b7280">单位：已读文章数</div>
          </div>
          <div style="position:relative;height:22rem">
            <canvas ref="trendCanvas"></canvas>
          </div>
        </section>

        <section class="a-card" style="padding:1.25rem">
          <div style="display:flex;align-items:center;justify-content:space-between;gap:1rem;margin-bottom:1rem;flex-wrap:wrap">
            <div>
              <div class="a-label a-muted" style="margin-bottom:.35rem">来源分布</div>
              <h2 style="margin:0;font-size:1.35rem;font-weight:900;letter-spacing:-0.04em">订阅源 Top 10</h2>
            </div>
          </div>
          <div style="position:relative;height:22rem">
            <canvas ref="sourceCanvas"></canvas>
          </div>
        </section>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import Chart from 'chart.js/auto'
import ABtn from '@/components/ui/ABtn.vue'
import AEmpty from '@/components/ui/AEmpty.vue'
import APageHeader from '@/components/ui/APageHeader.vue'
import { useAuthStore } from '@/stores/auth'

type FeedStatsPeriod = 'day' | 'week' | 'month'

interface FeedStatsPoint {
  label: string
  count: number
}

interface FeedSourceStat {
  feed_source_id: string
  title: string
  count: number
}

interface FeedStatsData {
  period: FeedStatsPeriod
  total_read: number
  points: FeedStatsPoint[]
  source_breakdown: FeedSourceStat[]
}

const API_URL = import.meta.env.VITE_API_URL || '/api'

const authStore = useAuthStore()

const periodOptions: Array<{ value: FeedStatsPeriod; label: string }> = [
  { value: 'day', label: '日' },
  { value: 'week', label: '周' },
  { value: 'month', label: '月' },
]

const period = ref<FeedStatsPeriod>('day')
const loading = ref(false)
const stats = ref<FeedStatsData | null>(null)
const trendCanvas = ref<HTMLCanvasElement | null>(null)
const sourceCanvas = ref<HTMLCanvasElement | null>(null)

let trendChart: Chart | null = null
let sourceChart: Chart | null = null

const authHeaders = () => ({ Authorization: `Bearer ${authStore.token}` })

const periodTitle = computed(() => {
  if (period.value === 'week') return '最近 8 周阅读趋势'
  if (period.value === 'month') return '最近 6 个月阅读趋势'
  return '最近 7 天阅读趋势'
})

const topSource = computed(() => stats.value?.source_breakdown[0] ?? null)
const topSourceLabel = computed(() => topSource.value?.title || '暂无数据')
const topSourceReads = computed(() => topSource.value ? `${topSource.value.count} 篇已读` : '没有可展示的来源')

const selectPeriod = async (value: FeedStatsPeriod) => {
  if (period.value === value) return
  period.value = value
  await fetchStats()
}

const fetchStats = async () => {
  if (!authStore.isAuthenticated) return
  loading.value = true
  try {
    const params = new URLSearchParams({ period: period.value })
    const res = await fetch(`${API_URL}/feed/stats?${params}`, { headers: authHeaders() })
    if (res.ok) {
      const data = await res.json()
      stats.value = data.data || null
      await nextTick()
      renderCharts()
    }
  } catch (error) {
    console.error('Failed to fetch feed stats', error)
  } finally {
    loading.value = false
  }
}

const renderCharts = () => {
  destroyCharts()
  if (!stats.value || !trendCanvas.value || !sourceCanvas.value) return

  trendChart = new Chart(trendCanvas.value, {
    type: 'bar',
    data: {
      labels: stats.value.points.map((point) => point.label),
      datasets: [
        {
          label: '已读文章',
          data: stats.value.points.map((point) => point.count),
          backgroundColor: '#111111',
          borderColor: '#111111',
          borderWidth: 2,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: false,
      plugins: {
        legend: { display: false },
      },
      scales: {
        x: {
          ticks: { color: '#6b7280', font: { weight: 700 } },
          grid: { display: false },
          border: { color: '#000000' },
        },
        y: {
          beginAtZero: true,
          ticks: { color: '#6b7280', precision: 0, font: { weight: 700 } },
          grid: { color: '#e5e7eb' },
          border: { color: '#000000' },
        },
      },
    },
  })

  sourceChart = new Chart(sourceCanvas.value, {
    type: 'bar',
    data: {
      labels: stats.value.source_breakdown.map((item) => item.title),
      datasets: [
        {
          label: '阅读量',
          data: stats.value.source_breakdown.map((item) => item.count),
          backgroundColor: ['#111111', '#3f3f46', '#71717a', '#a1a1aa', '#d4d4d8', '#18181b', '#52525b', '#a8a29e', '#94a3b8', '#d6d3d1'],
          borderWidth: 0,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: false,
      indexAxis: 'y',
      plugins: {
        legend: { display: false },
      },
      scales: {
        x: {
          beginAtZero: true,
          ticks: { color: '#6b7280', precision: 0, font: { weight: 700 } },
          grid: { color: '#f3f4f6' },
          border: { color: '#000000' },
        },
        y: {
          ticks: { color: '#6b7280', font: { weight: 700 } },
          grid: { display: false },
          border: { color: '#000000' },
        },
      },
    },
  })
}

const destroyCharts = () => {
  trendChart?.destroy()
  sourceChart?.destroy()
  trendChart = null
  sourceChart = null
}

onMounted(async () => {
  if (authStore.isAuthenticated) {
    await fetchStats()
  }
})

onBeforeUnmount(() => {
  destroyCharts()
})
</script>