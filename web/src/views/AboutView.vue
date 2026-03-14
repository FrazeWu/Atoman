<script setup lang="ts">
// About page for Atoman
</script>

<template>
  <div class="max-w-4xl mx-auto px-8 py-20">
    <RouterLink to="/" class="inline-flex items-center gap-2 mb-8 font-bold hover:underline">
      ← 返回时间线
    </RouterLink>

    <div class="space-y-8">
      <h1 class="text-5xl font-black tracking-tighter border-l-8 border-black pl-6">
        关于这个收藏
      </h1>

      <div class="prose prose-lg text-gray-700 space-y-8">
        <p>
          这是一个关于 Atoman 的互动档案项目。我们致力于提供一个全面可靠的分享平台。
        </p>
        
        <h2 class="text-3xl font-black mb-4">目标</h2>
        <ul class="list-disc pl-6 space-y-2">
          <li>打造最完整的内容时间线</li>
          <li>提供纯粹无广告的体验</li>
          <li>通过社区共同维护数据准确性</li>
        </ul>

        <h2 class="text-3xl font-black mb-4 gap-y-4">团队</h2>
        <p>
          致力于提供一个全面而可靠的开源资源。
        </p>
      </div>

      <div class="border-t-2 border-black pt-8">
        <h2 class="text-2xl font-black mb-4">贡献指南</h2>
        <ul class="space-y-2 text-gray-700">
          <li>• 注册账户后可以上传新的音乐内容</li>
          <li>• 所有上传内容将进入审核队列</li>
          <li>• 如果发现信息错误，可以提交修正请求</li>
          <li>• 保持内容的真实性和准确性</li>
        </ul>
      </div>
    </div>
  </div>
</template>