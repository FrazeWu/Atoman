import { ref, onBeforeUnmount } from 'vue'

type AutoSaveState = 'idle' | 'saving' | 'saved'

interface StoredDraft<T> {
  payload: T
  saved_at: number
}

interface AutoSaveOptions<T> {
  getDraftKey: () => string
  getPayload: () => T
  shouldPersist?: (payload: T) => boolean
}

export function useAutoSave<T>(options: AutoSaveOptions<T>) {
  const autoSaveState = ref<AutoSaveState>('idle')
  const lastSavedAt = ref<number | null>(null)
  let timer: ReturnType<typeof setTimeout> | null = null

  const getDraftKey = () => options.getDraftKey()
  const shouldPersist = options.shouldPersist ?? (() => true)

  function triggerAutoSave() {
    if (timer) clearTimeout(timer)
    autoSaveState.value = 'saving'
    timer = setTimeout(() => {
      const payload = options.getPayload()
      if (shouldPersist(payload)) {
        const savedAt = Date.now()
        localStorage.setItem(
          getDraftKey(),
          JSON.stringify({ payload, saved_at: savedAt }),
        )
        lastSavedAt.value = savedAt
      } else {
        localStorage.removeItem(getDraftKey())
        lastSavedAt.value = null
      }
      autoSaveState.value = 'saved'
    }, 500)
  }

  function loadDraft(): StoredDraft<T> | null {
    const saved = localStorage.getItem(getDraftKey())
    if (!saved) return null
    try {
      const parsed = JSON.parse(saved) as StoredDraft<T>
      if (!parsed || typeof parsed !== 'object' || !('payload' in parsed)) {
        return null
      }
      return parsed
    } catch {
      return null
    }
  }

  function clearDraft() {
    if (timer) {
      clearTimeout(timer)
      timer = null
    }
    localStorage.removeItem(getDraftKey())
    lastSavedAt.value = null
    autoSaveState.value = 'idle'
  }

  onBeforeUnmount(() => {
    if (timer) clearTimeout(timer)
  })

  return {
    autoSaveState,
    lastSavedAt,
    triggerAutoSave,
    loadDraft,
    clearDraft,
  }
}
