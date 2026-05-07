import { onMounted, onUnmounted } from 'vue'

export interface KeyboardShortcut {
  key: string
  description: string
  handler: () => void
  ctrl?: boolean
  meta?: boolean
  shift?: boolean
}

export function useKeyboardShortcuts(shortcuts: KeyboardShortcut[]) {
  const handleKeyDown = (event: KeyboardEvent) => {
    // Ignore if user is typing in input/textarea
    const target = event.target as HTMLElement
    if (
      target.tagName === 'INPUT' ||
      target.tagName === 'TEXTAREA' ||
      target.contentEditable === 'true'
    ) {
      return
    }

    for (const shortcut of shortcuts) {
      const keyMatches = event.key.toLowerCase() === shortcut.key.toLowerCase()
      const ctrlMatches = shortcut.ctrl ? event.ctrlKey : !event.ctrlKey
      const metaMatches = shortcut.meta ? event.metaKey : !event.metaKey
      const shiftMatches = shortcut.shift ? event.shiftKey : !event.shiftKey

      if (keyMatches && ctrlMatches && metaMatches && shiftMatches) {
        event.preventDefault()
        shortcut.handler()
        break
      }
    }
  }

  onMounted(() => {
    window.addEventListener('keydown', handleKeyDown)
  })

  onUnmounted(() => {
    window.removeEventListener('keydown', handleKeyDown)
  })

  return {
    shortcuts,
  }
}
