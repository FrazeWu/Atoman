<template>
  <div class="page-container">
    <h1 class="page-title">添加/补全艺术家</h1>
    <p class="page-desc">如果上传页搜索不到艺术家，请先在这里创建，再返回贡献页面选择。</p>

    <div class="rules-box">
      <p class="rules-title">命名规则</p>
      <p class="rules-item">1. 全部使用小写</p>
      <p class="rules-item">2. 空格使用 _ 代替</p>
      <p class="rules-item">3. 每个艺术家使用唯一 uuidv7 表示</p>
    </div>

    <form class="form-stack" @submit.prevent="handleCreate">
      <div class="field">
        <label class="field-label">艺术家名称</label>
        <input v-model.trim="name" class="form-input" type="text" required maxlength="120" placeholder="例如：kanye_west" />
      </div>

      <div class="field">
        <label class="field-label">简介（可选）</label>
        <textarea
          v-model.trim="bio"
          class="form-textarea"
          rows="5"
          maxlength="2000"
          placeholder="可填写艺术家简介"
        />
      </div>

      <p v-if="error" class="error-text">{{ error }}</p>
      <p v-if="success" class="success-text">{{ success }}</p>

      <div class="actions">
        <ABtn to="/music/contribute" outline>返回贡献页</ABtn>
        <ABtn :loading="creating" loadingText="创建中..." @click="handleCreate">创建艺术家</ABtn>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import ABtn from '@/components/ui/ABtn.vue'
import { useApi } from '@/composables/useApi'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const api = useApi()
const authStore = useAuthStore()

const name = ref(typeof route.query.name === 'string' ? route.query.name : '')
const bio = ref('')
const creating = ref(false)
const error = ref('')
const success = ref('')

const handleCreate = async () => {
  error.value = ''
  success.value = ''

  const normalizedName = name.value.trim().toLowerCase().replace(/\s+/g, '_')

  if (!normalizedName) {
    error.value = '请输入艺术家名称'
    return
  }

  name.value = normalizedName

  creating.value = true
  try {
    const res = await fetch(api.artists || '/api/artists', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`,
      },
      body: JSON.stringify({ name: normalizedName, bio: bio.value || undefined }),
    })

    if (res.ok) {
      success.value = '艺术家创建成功，正在返回贡献页...'
      setTimeout(() => {
        router.push('/music/contribute')
      }, 800)
      return
    }

    const payload = await res.json().catch(() => ({}))
    error.value = payload.error || '创建失败，请稍后重试'
  } catch (e) {
    console.error(e)
    error.value = '创建失败，请检查网络后重试'
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
.page-container {
  max-width: 48rem;
  margin: 0 auto;
  padding: 5rem 2rem 12rem;
}

.page-title {
  margin: 0 0 0.5rem;
  font-size: 2.25rem;
  font-weight: 900;
  letter-spacing: -0.04em;
}

.page-desc {
  margin: 0 0 2rem;
  color: #6b7280;
}

.rules-box {
  border: 2px solid #000;
  padding: 1rem;
  margin-bottom: 1.5rem;
  background: #f9fafb;
}

.rules-title {
  margin: 0 0 0.5rem;
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.rules-item {
  margin: 0.25rem 0;
  font-size: 0.875rem;
  font-weight: 700;
}

.form-stack {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.field-label {
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.form-input,
.form-textarea {
  width: 100%;
  border: 2px solid #000;
  padding: 0.9rem 1rem;
  outline: none;
  background: #fff;
  font-size: 0.95rem;
  box-sizing: border-box;
}

.form-input:focus,
.form-textarea:focus {
  box-shadow: 5px 5px 0 0 rgba(0, 0, 0, 1);
}

.form-textarea {
  resize: vertical;
}

.actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
  flex-wrap: wrap;
}

.error-text {
  margin: 0;
  color: #dc2626;
  font-size: 0.875rem;
  font-weight: 700;
}

.success-text {
  margin: 0;
  color: #2bb24c;
  font-size: 0.875rem;
  font-weight: 700;
}
</style>
