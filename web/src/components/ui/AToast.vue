<template>
  <transition name="fade">
    <div
      v-if="visible"
      class="a-toast"
      :style="{ top: `${top}px` }"
      @mouseenter="clearTimer"
      @mouseleave="startTimer"
    >
      {{ message }}
    </div>
  </transition>
</template>

<script setup lang="ts">
import { ref, onUnmounted, watch } from 'vue'

const props = defineProps<{
  modelValue: boolean
  message: string
  duration?: number
  top?: number
}>()

const emit = defineEmits(['update:modelValue'])

const visible = ref(props.modelValue)
const timer = ref<number | null>(null)
const top = props.top ?? 32

watch(
  () => props.modelValue,
  (val) => {
    visible.value = val
    if (val) startTimer()
    else clearTimer()
  },
  { immediate: true }
)

function startTimer() {
  clearTimer()
  if (props.duration !== 0) {
    timer.value = window.setTimeout(() => {
      emit('update:modelValue', false)
    }, props.duration ?? 1800)
  }
}
function clearTimer() {
  if (timer.value) {
    clearTimeout(timer.value)
    timer.value = null
  }
}

onUnmounted(clearTimer)
</script>

<style scoped>
.a-toast {
  position: fixed;
  left: 50%;
  transform: translateX(-50%);
  min-width: 120px;
  max-width: 80vw;
  background: #000;
  color: #fff;
  font-weight: 900;
  font-size: 1rem;
  padding: 0.75rem 1.5rem;
  border-radius: 999px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.12);
  z-index: 9999;
  text-align: center;
  pointer-events: none;
  opacity: 0.96;
}
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.25s;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style>
