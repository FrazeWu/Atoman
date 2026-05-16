/**
 * useImageUpload.ts
 * Image upload composable for the MarkdownEditor.
 *
 * Backend: POST /api/blog/upload-image
 * - Accepts: multipart/form-data with field "image"
 * - Validates: JPEG/PNG/GIF/WebP, max 5MB
 * - Returns: { url: string }
 * - Auth: Bearer token required
 *
 * Supports three input methods:
 * 1. Toolbar button click → file picker
 * 2. Paste image from clipboard
 * 3. Drag & drop onto editor
 *
 * Insert flow:
 * 1. Insert placeholder ![上传中...](uploading-{uuid}) at cursor
 * 2. Upload with XHR (for progress events)
 * 3. On success: replace placeholder with ![filename](url)
 * 4. On error: replace placeholder with error message
 */

import { ref } from 'vue'
import { useApi } from '@/composables/useApi'
import { useAuthStore } from '@/stores/auth'

const ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
const MAX_SIZE = 5 * 1024 * 1024 // 5 MB

export function useImageUpload(
  insertAtCursor: (text: string) => void,
  replaceText: (search: string, replacement: string) => void,
) {
  const api = useApi()
  const authStore = useAuthStore()
  const isUploading = ref(false)
  const uploadProgress = ref(0)
  const uploadError = ref<string | null>(null)

  function generateId(): string {
    return Math.random().toString(36).slice(2, 10)
  }

  function validateFile(file: File): string | null {
    if (!ALLOWED_TYPES.includes(file.type)) {
      return '只支持 JPEG、PNG、GIF、WebP 格式的图片'
    }
    if (file.size > MAX_SIZE) {
      return '图片不能超过 5MB'
    }
    return null
  }

  async function uploadFile(file: File): Promise<void> {
    const error = validateFile(file)
    if (error) {
      uploadError.value = error
      return
    }

    const id = generateId()
    const placeholder = `![上传中...](uploading-${id})`
    insertAtCursor(placeholder)

    isUploading.value = true
    uploadProgress.value = 0
    uploadError.value = null

    try {
      const url = await uploadWithProgress(file)
      const altText = file.name.replace(/\.[^.]+$/, '') || '图片'
      replaceText(placeholder, `![${altText}](${url})`)
    } catch (err) {
      const message = err instanceof Error ? err.message : '上传失败'
      uploadError.value = message
      replaceText(placeholder, `![上传失败]()`)
    } finally {
      isUploading.value = false
      uploadProgress.value = 0
    }
  }

  function uploadWithProgress(file: File): Promise<string> {
    return new Promise((resolve, reject) => {
      const formData = new FormData()
      formData.append('image', file)

      const xhr = new XMLHttpRequest()
      xhr.open('POST', api.blog.uploadImage)

      // Set auth token
      const token = authStore.token
      if (token) {
        xhr.setRequestHeader('Authorization', `Bearer ${token}`)
      }

      xhr.upload.addEventListener('progress', (e) => {
        if (e.lengthComputable) {
          uploadProgress.value = Math.round((e.loaded / e.total) * 100)
        }
      })

      xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            const data = JSON.parse(xhr.responseText)
            if (data.url) {
              resolve(data.url)
            } else {
              reject(new Error('服务器返回了无效的响应'))
            }
          } catch {
            reject(new Error('无法解析服务器响应'))
          }
        } else {
          try {
            const data = JSON.parse(xhr.responseText)
            reject(new Error(data.error || `上传失败 (${xhr.status})`))
          } catch {
            reject(new Error(`上传失败 (${xhr.status})`))
          }
        }
      })

      xhr.addEventListener('error', () => {
        reject(new Error('网络错误，上传失败'))
      })

      xhr.addEventListener('abort', () => {
        reject(new Error('上传已取消'))
      })

      xhr.send(formData)
    })
  }

  /** Handle paste event — check for image in clipboard */
  function handlePaste(e: ClipboardEvent): boolean {
    const items = e.clipboardData?.items
    if (!items) return false

    for (const item of Array.from(items)) {
      if (item.type.startsWith('image/')) {
        e.preventDefault()
        const file = item.getAsFile()
        if (file) {
          uploadFile(file)
          return true
        }
      }
    }
    return false
  }

  /** Handle drop event — check for image files */
  function handleDrop(e: DragEvent): boolean {
    const files = e.dataTransfer?.files
    if (!files || files.length === 0) return false

    const imageFiles = Array.from(files).filter((f) => f.type.startsWith('image/'))
    if (imageFiles.length === 0) return false

    e.preventDefault()
    // Upload the first image only
    uploadFile(imageFiles[0])
    return true
  }

  /** Open file picker and upload selected file */
  function openFilePicker(inputEl: HTMLInputElement) {
    inputEl.click()
  }

  /** Handle file input change event */
  function handleFileInput(e: Event) {
    const input = e.target as HTMLInputElement
    const file = input.files?.[0]
    if (file) {
      uploadFile(file)
      // Reset input so same file can be selected again
      input.value = ''
    }
  }

  return {
    isUploading,
    uploadProgress,
    uploadError,
    uploadFile,
    handlePaste,
    handleDrop,
    openFilePicker,
    handleFileInput,
  }
}
