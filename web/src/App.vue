<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { apiFetch } from './api'
import { initializeAuth, logout } from './auth'
import AppShell from './components/AppShell.vue'
import { normalizeCurrentUser, type CurrentUser, type CurrentUserResponse } from './types'

const state = ref<'loading' | 'ready' | 'error'>('loading')
const message = ref('Keycloak 연결 정보를 확인하고 있습니다…')
const user = ref<CurrentUser>()

function reload(): void {
  window.location.reload()
}

onMounted(async () => {
  try {
    message.value = 'Keycloak 로그인 상태를 확인하고 있습니다…'
    await initializeAuth()
    const response = await apiFetch<CurrentUserResponse>('/api/v1/me')
    user.value = normalizeCurrentUser(response)
    state.value = 'ready'
  } catch (error) {
    message.value = error instanceof Error ? error.message : '로그인 초기화에 실패했습니다.'
    state.value = 'error'
  }
})
</script>

<template>
  <main v-if="state !== 'ready'" class="auth-shell">
    <section class="brand-panel">
      <div class="brand"><span class="mark">C</span><span>Cosight</span></div>
      <div class="brand-copy"><p>CODE INTELLIGENCE</p><h1>코드의 연결을<br>한눈에 이해하세요.</h1><span>Go, TypeScript, Vue 프로젝트의 구조와 실행 흐름을 탐색합니다.</span></div>
      <div class="security-note">OIDC · Authorization Code Flow · PKCE S256</div>
    </section>
    <section class="content-panel">
      <div v-if="state === 'loading'" class="status-card" aria-live="polite"><div class="spinner"></div><h2>로그인 준비 중</h2><p>{{ message }}</p><small>로그인 화면은 Keycloak에서 제공합니다.</small></div>
      <div v-else-if="state === 'error'" class="status-card error" role="alert"><span class="status-icon">!</span><h2>로그인 연결 실패</h2><p>{{ message }}</p><button @click="reload">다시 시도</button></div>
    </section>
  </main>
  <AppShell v-else-if="user" :user="user" @logout="logout" />
</template>
