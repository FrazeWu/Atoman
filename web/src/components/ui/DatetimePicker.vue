<template>
  <div class="dtp-wrap" ref="wrapEl">
    <!-- Trigger input -->
    <div class="dtp-input-wrap" @click="open = !open">
      <input
        class="a-input dtp-input"
        readonly
        :value="displayValue"
        :placeholder="placeholder"
      />
      <span class="dtp-icon">{{ open ? '▲' : '▼' }}</span>
    </div>

    <!-- Dropdown panel -->
    <div v-if="open" class="dtp-panel">
      <!-- Month navigation -->
      <div class="dtp-nav">
        <button class="dtp-nav-btn" @click.stop="prevMonth">‹</button>
        <div class="dtp-nav-center">
          <!-- Year: directly editable input -->
          <input
            class="dtp-year-input"
            type="number"
            :value="viewYear"
            @input="onYearInput"
            @blur="onYearBlur"
            @keydown.enter.stop="onYearBlur"
            @click.stop
          />
          <span class="dtp-nav-sep">年</span>
          <span class="dtp-nav-month">{{ viewMonth + 1 }}月</span>
        </div>
        <button class="dtp-nav-btn" @click.stop="nextMonth">›</button>
      </div>

      <!-- Weekday headers -->
      <div class="dtp-grid">
        <div v-for="d in weekDays" :key="d" class="dtp-weekday">{{ d }}</div>

        <!-- Empty cells before first day -->
        <div v-for="n in firstDayOfMonth" :key="'empty-' + n" />

        <!-- Day cells -->
        <button
          v-for="day in daysInMonth"
          :key="day"
          class="dtp-day"
          :class="{
            'dtp-day-today': isToday(day),
            'dtp-day-selected': isSelected(day),
          }"
          @click.stop="selectDay(day)"
        >{{ day }}</button>
      </div>

      <!-- Time row (only when showTime) -->
      <div v-if="showTime" class="dtp-time-row">
        <span class="dtp-time-label">时间</span>
        <ASelect v-model="selHour" :options="hourSelectOptions" class="dtp-select-wrap" />
        <span class="dtp-colon">:</span>
        <ASelect v-model="selMinute" :options="minuteSelectOptions" class="dtp-select-wrap" />
      </div>

      <!-- Action row -->
      <div class="dtp-actions">
        <button class="dtp-clear" @click.stop="clearValue">清除</button>
        <button class="dtp-confirm" @click.stop="confirm">确定</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import ASelect from '@/components/ui/ASelect.vue'

const props = withDefaults(defineProps<{
  modelValue?: string
  placeholder?: string
  /** whether to show time picker (default true) */
  showTime?: boolean
}>(), {
  placeholder: '选择日期时间',
  showTime: true,
})

const emit = defineEmits<{ 'update:modelValue': [v: string] }>()

// ── State ──────────────────────────────────────────────
const open = ref(false)
const wrapEl = ref<HTMLElement | null>(null)

const now = new Date()
const viewYear = ref(now.getFullYear())
const viewMonth = ref(now.getMonth()) // 0-indexed

const selYear = ref<number | null>(null)
const selMonth = ref<number | null>(null)
const selDay = ref<number | null>(null)
const selHour = ref(0)
const selMinute = ref(0)

// Temporary string while user is typing in year input
const yearInputDraft = ref<string>('')

const weekDays = ['日', '一', '二', '三', '四', '五', '六']
const minuteOptions = [0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55]
const hourSelectOptions = Array.from({ length: 24 }, (_, h) => ({
  label: String(h).padStart(2, '0'),
  value: h,
}))
const minuteSelectOptions = minuteOptions.map((m) => ({
  label: String(m).padStart(2, '0'),
  value: m,
}))

// ── Computed ───────────────────────────────────────────
const daysInMonth = computed(() => {
  // Use abs to handle BC years in JS Date
  const y = viewYear.value
  return new Date(y, viewMonth.value + 1, 0).getDate()
})

const firstDayOfMonth = computed(() => {
  return new Date(viewYear.value, viewMonth.value, 1).getDay()
})

const displayValue = computed(() => {
  if (!props.modelValue) return ''
  const v = props.modelValue
  // Format: YYYY-MM-DDTHH:mm → display as YYYY-MM-DD HH:mm
  return v.replace('T', ' ').slice(0, 16)
})

// ── Sync incoming value ────────────────────────────────
const syncFromValue = () => {
  const v = props.modelValue
  if (!v) {
    selYear.value = null
    selMonth.value = null
    selDay.value = null
    selHour.value = 0
    selMinute.value = 0
    return
  }
  // Parse manually to support years outside JS Date range (e.g. year 50 AD)
  const match = v.match(/^(-?\d{1,4})-(\d{2})-(\d{2})(?:[T ](\d{2}):(\d{2}))?/)
  if (match) {
    selYear.value = parseInt(match[1])
    selMonth.value = parseInt(match[2]) - 1
    selDay.value = parseInt(match[3])
    selHour.value = match[4] ? parseInt(match[4]) : 0
    const rawMin = match[5] ? parseInt(match[5]) : 0
    selMinute.value = minuteOptions.reduce((prev, cur) =>
      Math.abs(cur - rawMin) < Math.abs(prev - rawMin) ? cur : prev
    )
    viewYear.value = selYear.value
    viewMonth.value = selMonth.value
  }
}

watch(() => props.modelValue, syncFromValue, { immediate: true })

// ── Year input handling ────────────────────────────────
const onYearInput = (e: Event) => {
  yearInputDraft.value = (e.target as HTMLInputElement).value
}

const onYearBlur = (e: Event) => {
  const raw = yearInputDraft.value || (e.target as HTMLInputElement).value
  const parsed = parseInt(raw)
  if (!isNaN(parsed)) {
    viewYear.value = parsed
  }
  yearInputDraft.value = ''
  // Clamp month days if needed
  const maxDay = new Date(viewYear.value, viewMonth.value + 1, 0).getDate()
  if (selDay.value && selDay.value > maxDay) selDay.value = maxDay
}

// ── Methods ────────────────────────────────────────────
const prevMonth = () => {
  if (viewMonth.value === 0) {
    viewMonth.value = 11
    viewYear.value--
  } else {
    viewMonth.value--
  }
}

const nextMonth = () => {
  if (viewMonth.value === 11) {
    viewMonth.value = 0
    viewYear.value++
  } else {
    viewMonth.value++
  }
}

const selectDay = (day: number) => {
  selYear.value = viewYear.value
  selMonth.value = viewMonth.value
  selDay.value = day
}

const isToday = (day: number) => {
  return (
    day === now.getDate() &&
    viewMonth.value === now.getMonth() &&
    viewYear.value === now.getFullYear()
  )
}

const isSelected = (day: number) => {
  return (
    selDay.value === day &&
    selMonth.value === viewMonth.value &&
    selYear.value === viewYear.value
  )
}

const padYear = (y: number) => {
  // Support negative (BC) years: -500 → "-0500", 100 → "0100"
  if (y < 0) return '-' + String(Math.abs(y)).padStart(4, '0')
  return String(y).padStart(4, '0')
}

const confirm = () => {
  if (selYear.value === null || selMonth.value === null || selDay.value === null) return
  const y = padYear(selYear.value)
  const mo = String(selMonth.value + 1).padStart(2, '0')
  const d = String(selDay.value).padStart(2, '0')
  const h = String(selHour.value).padStart(2, '0')
  const m = String(selMinute.value).padStart(2, '0')
  const value = props.showTime ? `${y}-${mo}-${d}T${h}:${m}` : `${y}-${mo}-${d}`
  emit('update:modelValue', value)
  open.value = false
}

const clearValue = () => {
  emit('update:modelValue', '')
  selYear.value = null
  selMonth.value = null
  selDay.value = null
  selHour.value = 0
  selMinute.value = 0
  open.value = false
}

// ── Close on outside click ─────────────────────────────
const handleClickOutside = (e: MouseEvent) => {
  if (wrapEl.value && !wrapEl.value.contains(e.target as Node)) {
    open.value = false
  }
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onUnmounted(() => document.removeEventListener('click', handleClickOutside))
</script>

<style scoped>
.dtp-wrap {
  position: relative;
  display: inline-block;
  width: 100%;
}

.dtp-input-wrap {
  display: flex;
  align-items: center;
  position: relative;
  cursor: pointer;
}

.dtp-input {
  width: 100%;
  cursor: pointer;
  padding-right: 2rem;
  user-select: none;
}

.dtp-icon {
  position: absolute;
  right: 0.6rem;
  font-size: 0.65rem;
  color: #6b7280;
  pointer-events: none;
}

/* Panel */
.dtp-panel {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  z-index: 200;
  background: #fff;
  border: 2px solid #000;
  box-shadow: 6px 6px 0 0 rgba(0,0,0,1);
  width: 290px;
  padding: 0.75rem;
  user-select: none;
}

/* Nav */
.dtp-nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.5rem;
  gap: 0.25rem;
}

.dtp-nav-btn {
  background: none;
  border: 2px solid #000;
  width: 28px;
  height: 28px;
  font-size: 1rem;
  font-weight: 900;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: background 0.1s;
}
.dtp-nav-btn:hover { background: #000; color: #fff; }

.dtp-nav-center {
  display: flex;
  align-items: center;
  gap: 0.2rem;
  flex: 1;
  justify-content: center;
}

/* Year direct-input */
.dtp-year-input {
  border: none;
  border-bottom: 2px solid #000;
  background: transparent;
  font-size: 0.875rem;
  font-weight: 900;
  font-family: inherit;
  width: 5ch;
  text-align: center;
  padding: 0 2px;
  outline: none;
  color: #000;
  /* Hide spin buttons */
  -moz-appearance: textfield;
}
.dtp-year-input::-webkit-inner-spin-button,
.dtp-year-input::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
.dtp-year-input:focus {
  border-bottom-color: #000;
  background: #f9fafb;
}

.dtp-nav-sep,
.dtp-nav-month {
  font-size: 0.875rem;
  font-weight: 900;
  letter-spacing: -0.02em;
  white-space: nowrap;
}

/* Grid */
.dtp-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 2px;
  margin-bottom: 0.5rem;
}

.dtp-weekday {
  font-size: 0.65rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: #9ca3af;
  text-align: center;
  padding: 2px 0;
}

.dtp-day {
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8rem;
  font-weight: 700;
  border: none;
  background: none;
  cursor: pointer;
  border-radius: 0;
  transition: background 0.1s, color 0.1s;
}
.dtp-day:hover { background: #f3f4f6; }
.dtp-day-today { text-decoration: underline; font-weight: 900; }
.dtp-day-selected {
  background: #000;
  color: #fff;
}
.dtp-day-selected:hover { background: #333; }

/* Time row */
.dtp-time-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  border-top: 1px solid #e5e7eb;
  padding-top: 0.5rem;
  margin-bottom: 0.5rem;
}

.dtp-time-label {
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: #6b7280;
  flex-shrink: 0;
}

.dtp-select-wrap {
  width: 84px;
}

.dtp-colon { font-weight: 900; font-size: 1rem; }

/* Actions */
.dtp-actions {
  display: flex;
  justify-content: space-between;
  border-top: 1px solid #e5e7eb;
  padding-top: 0.5rem;
}

.dtp-clear {
  font-size: 0.75rem;
  font-weight: 900;
  color: #6b7280;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  text-transform: uppercase;
  letter-spacing: 0.06em;
}
.dtp-clear:hover { color: #000; text-decoration: underline; }

.dtp-confirm {
  font-size: 0.75rem;
  font-weight: 900;
  color: #fff;
  background: #000;
  border: 2px solid #000;
  cursor: pointer;
  padding: 4px 14px;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  transition: background 0.1s;
}
.dtp-confirm:hover { background: #333; }
</style>
