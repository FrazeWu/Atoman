<script setup lang="ts">
import { RouterLink } from 'vue-router'

const techStack = [
  { role: 'Frontend', label: 'Vue 3 + Vite' },
  { role: 'Styling', label: 'Naive UI' },
  { role: 'Backend', label: 'Go + Gin' },
  { role: 'Database', label: 'PostgreSQL' },
  { role: 'Storage', label: 'S3 / MinIO' },
  { role: 'Auth', label: 'JWT' },
  { role: 'Feed', label: 'RSS 2.0' },
  { role: 'License', label: 'GPL v3' },
]
</script>

<template>
  <div class="page-container">
    <RouterLink to="/music" class="back-link">← 返回时间线</RouterLink>

    <div class="content-stack">
      <div>
        <h1 class="page-title">关于 Atoman</h1>
        <p class="page-desc">
          一个开放的言论自由平台，支持博客写作、音乐归档与 RSS 订阅聚合。完全开源、可自托管，遵循 GPL 协议。
        </p>
      </div>

      <!-- Features -->
      <div class="features-grid">
        <RouterLink to="/music" class="feature-card">
          <div class="card-label">MUSIC</div>
          <h2 class="card-title">音乐档案</h2>
          <p class="card-desc">
            以时间轴方式呈现艺术家的完整唱片目录。支持社区上传、信息修正与管理员审核工作流。
          </p>
        </RouterLink>

        <RouterLink to="/blog" class="feature-card">
          <div class="card-label">BLOG</div>
          <h2 class="card-title">个人博客</h2>
          <p class="card-desc">
            支持 Markdown 的所见即所得博客。可发布长短文、创建专栏与合集，支持评论、点赞与书签。
          </p>
        </RouterLink>

        <RouterLink to="/feed" class="feature-card">
          <div class="card-label">FEED</div>
          <h2 class="card-title">订阅聚合</h2>
          <p class="card-desc">
            订阅站内用户频道或任意外部 RSS 源，所有内容聚合为统一时间流。每个用户也生成标准 RSS 源供他人订阅。
          </p>
        </RouterLink>
      </div>

      <!-- Tech stack -->
      <div class="section">
        <h2 class="section-title">技术栈</h2>
        <div class="tech-grid">
          <div v-for="item in techStack" :key="item.label" class="tech-card">
            <div class="tech-role">{{ item.role }}</div>
            <div class="tech-label">{{ item.label }}</div>
          </div>
        </div>
      </div>

      <!-- Contribution guide -->
      <div class="section">
        <h2 class="section-title">参与贡献</h2>
        <ul class="contrib-list">
          <li class="contrib-item">
            <span class="contrib-num">01</span>
            <span class="contrib-text">注册账号后可上传音乐内容或发布博客文章</span>
          </li>
          <li class="contrib-item">
            <span class="contrib-num">02</span>
            <span class="contrib-text">普通用户上传的音乐进入待审核队列，经管理员批准后对外可见</span>
          </li>
          <li class="contrib-item">
            <span class="contrib-num">03</span>
            <span class="contrib-text">发现信息有误？可提交修正申请，同样经审核后生效</span>
          </li>
          <li class="contrib-item">
            <span class="contrib-num">04</span>
            <span class="contrib-text">博客内容由作者自主管理，无需审核即可发布</span>
          </li>
        </ul>
      </div>

      <!-- Links -->
      <div class="links-row">
        <a
          href="https://github.com"
          target="_blank"
          rel="noopener noreferrer"
          class="link-btn"
        >
          GitHub →
        </a>
        <span class="license-badge">GPL License</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page-container {
  max-width: 896px;
  margin: 0 auto;
  padding: 5rem 2rem;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 2rem;
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  border-bottom: 2px solid #000;
  text-decoration: none;
  color: #000;
  transition: opacity 0.2s;
}
.back-link:hover { opacity: 0.6; }

.content-stack {
  display: flex;
  flex-direction: column;
  gap: 3rem;
}

.page-title {
  font-size: 3rem;
  font-weight: 900;
  letter-spacing: -0.05em;
  border-left: 8px solid #000;
  padding-left: 1.5rem;
  margin: 0;
}

.page-desc {
  margin-top: 1rem;
  color: #4b5563;
  font-size: 1.125rem;
  font-weight: 500;
  line-height: 1.75;
  max-width: 42rem;
}

.features-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.5rem;
}
@media (min-width: 768px) {
  .features-grid { grid-template-columns: repeat(3, 1fr); }
}

.feature-card {
  display: block;
  border: 2px solid #000;
  padding: 1.5rem;
  text-decoration: none;
  color: #000;
  transition: all 0.3s;
}
.feature-card:hover {
  box-shadow: 10px 10px 0px 0px rgba(0,0,0,1);
}

.card-label {
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: #9ca3af;
  margin-bottom: 0.75rem;
}

.card-title {
  font-size: 1.25rem;
  font-weight: 900;
  letter-spacing: -0.025em;
  margin: 0 0 0.5rem 0;
}

.card-desc {
  font-size: 0.875rem;
  color: #4b5563;
  font-weight: 500;
  line-height: 1.6;
  margin: 0;
}

.section {
  border-top: 2px solid #000;
  padding-top: 2rem;
}

.section-title {
  font-size: 1.5rem;
  font-weight: 900;
  letter-spacing: -0.025em;
  margin: 0 0 1.5rem 0;
}

.tech-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
}
@media (min-width: 768px) {
  .tech-grid { grid-template-columns: repeat(4, 1fr); }
}

.tech-card {
  border: 1px solid #000;
  padding: 1rem;
}

.tech-role {
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: #9ca3af;
  margin-bottom: 0.25rem;
}

.tech-label {
  font-weight: 700;
  font-size: 0.875rem;
}

.contrib-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.contrib-item {
  display: flex;
  gap: 0.75rem;
}

.contrib-num {
  font-weight: 900;
  color: #000;
  flex-shrink: 0;
}

.contrib-text {
  font-weight: 500;
  color: #374151;
}

.links-row {
  border-top: 2px solid #000;
  padding-top: 2rem;
  display: flex;
  gap: 1rem;
}

.link-btn {
  border: 2px solid #000;
  padding: 0.75rem 1.5rem;
  font-weight: 900;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  text-decoration: none;
  color: #000;
  transition: all 0.2s;
}
.link-btn:hover {
  background: #000;
  color: #fff;
}

.license-badge {
  border: 1px solid #000;
  padding: 0.75rem 1.5rem;
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: #9ca3af;
}
</style>
