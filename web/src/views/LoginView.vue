<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRouter, useRoute, RouterLink } from 'vue-router';
import { useAuthStore } from '@/stores/auth.ts';

const email = ref('');
const password = ref('');
const username = ref('');
const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();

const isRegister = computed(() => route.path === '/register');

const handleSubmit = async () => {
  try {
    if (isRegister.value) {
      await authStore.register(username.value, email.value, password.value);
    } else {
      await authStore.loginWithPassword(email.value, password.value);
    }
    
    const redirect = route.query.redirect as string;
    router.push(redirect || '/');
  } catch (error: any) {
    alert(error.message);
  }
};
</script>

<template>
  <div class="min-h-[calc(100vh-64px)] flex items-center justify-center px-8">
    <div class="max-w-md w-full bg-white border-2 border-black p-12 shadow-[20px_20px_0px_0px_rgba(0,0,0,1)]">
      <h1 class="text-4xl font-black tracking-tighter mb-2">{{ isRegister ? '加入我们' : '欢迎回来' }}</h1>
      <p class="text-gray-400 mb-8 font-medium">进入 Atoman 的数字领域</p>

      <form @submit.prevent="handleSubmit" class="space-y-6">
        <div v-if="isRegister" class="space-y-2">
          <label class="text-xs font-black uppercase tracking-widest">用户名</label>
          <input 
            type="text" 
            required
            class="w-full border-2 border-black p-3 outline-none focus:bg-gray-50 transition-colors"
            v-model="username"
          />
        </div>

        <div class="space-y-2">
          <label class="text-xs font-black uppercase tracking-widest">邮箱地址</label>
          <input 
            type="email" 
            required
            class="w-full border-2 border-black p-3 outline-none focus:bg-gray-50 transition-colors"
            v-model="email"
          />
        </div>

        <div class="space-y-2">
          <label class="text-xs font-black uppercase tracking-widest">通行密码</label>
          <input 
            type="password" 
            required
            class="w-full border-2 border-black p-3 outline-none focus:bg-gray-50 transition-colors"
            v-model="password"
          />
        </div>

        <button 
          type="submit"
          class="w-full bg-black text-white py-4 font-black uppercase tracking-widest border-2 border-black hover:bg-white hover:text-black transition-all"
        >
          {{ isRegister ? '注册账号' : '登 录' }}
        </button>
      </form>

      <div class="mt-8 pt-8 border-t border-gray-100 text-center text-sm font-medium">
        <span v-if="isRegister">
          已有账号？ <RouterLink to="/login" class="font-black underline">立即登录</RouterLink>
        </span>
        <span v-else>
          还没有账号？ <RouterLink to="/register" class="font-black underline">立即加入档案室</RouterLink>
        </span>
      </div>
    </div>
  </div>
</template>
