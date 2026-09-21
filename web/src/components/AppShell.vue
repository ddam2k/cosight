<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import type { CurrentUser } from '../types'

const props = defineProps<{ user: CurrentUser }>()
const emit = defineEmits<{ logout: [] }>()
const route = useRoute()
const mobileOpen = ref(false)
const userOpen = ref(false)

interface MenuItem { label: string; to: string; icon: string }
interface MenuGroup { label: string; items: MenuItem[] }

const commonGroups: MenuGroup[] = [
  { label: '프로젝트', items: [
    { label: '내 프로젝트', to: '/', icon: '▦' },
    { label: '프로젝트 만들기', to: '/projects/new', icon: '+' },
    { label: '내 초대', to: '/invitations', icon: '✉' },
  ] },
  { label: '현재 프로젝트', items: [
    { label: '개요', to: '/projects/current', icon: '⌂' },
    { label: '코드 탐색', to: '/projects/current/explore', icon: '⌘' },
    { label: '저장된 탐색', to: '/projects/current/explorations', icon: '◇' },
    { label: '구성원', to: '/projects/current/members', icon: '♧' },
    { label: '설정', to: '/projects/current/settings/general', icon: '⚙' },
  ] },
]

const systemGroup: MenuGroup = { label: '시스템 관리', items: [
  { label: '대시보드', to: '/system', icon: '◉' },
  { label: '조직 관리', to: '/system/organizations', icon: '▤' },
  { label: 'LLM 연결', to: '/system/llm/connections', icon: '⇄' },
  { label: '모델 레지스트리', to: '/system/llm/models', icon: '◆' },
  { label: '논리 모델 프로필', to: '/system/llm/profiles', icon: '◫' },
  { label: '프로젝트 할당과 한도', to: '/system/projects', icon: '◒' },
  { label: '사용량과 상태', to: '/system/llm/usage', icon: '⌁' },
  { label: '감사 로그', to: '/system/audit', icon: '≡' },
] }

const isSystemAdmin = computed(() =>
  [...props.user.realmRoles, ...props.user.clientRoles].includes('cosight-system-admin'),
)
const groups = computed(() => isSystemAdmin.value ? [...commonGroups, systemGroup] : commonGroups)
const initials = computed(() => (props.user.displayName || props.user.preferredUsername || 'U').slice(0, 2).toUpperCase())
const displayName = computed(() => props.user.displayName || props.user.preferredUsername || '사용자')

function closeNavigation(): void { mobileOpen.value = false }
function signOut(): void { userOpen.value = false; emit('logout') }
</script>

<template>
  <div class="app-layout">
    <button v-if="mobileOpen" class="nav-backdrop" aria-label="메뉴 닫기" @click="mobileOpen = false"></button>
    <aside class="sidebar" :class="{ open: mobileOpen }">
      <RouterLink class="app-brand" to="/" @click="closeNavigation"><span class="brand-mark">C</span><span><b>Cosight</b><small>Code Intelligence</small></span></RouterLink>
      <nav aria-label="주 메뉴">
        <section v-for="group in groups" :key="group.label" class="menu-group">
          <h2>{{ group.label }}</h2>
          <RouterLink v-for="item in group.items" :key="item.to" :to="item.to" class="menu-link" @click="closeNavigation"><span class="menu-icon">{{ item.icon }}</span><span>{{ item.label }}</span><em v-if="item.to === '/invitations'">1</em></RouterLink>
        </section>
      </nav>
      <div class="sidebar-project"><span class="status-dot"></span><div><b>cosight-platform</b><small>main · 인덱스 최신</small></div></div>
    </aside>

    <div class="main-area">
      <header class="global-header">
        <button class="mobile-menu" aria-label="주 메뉴 열기" @click="mobileOpen = true">☰</button>
        <div class="breadcrumb"><span>{{ route.meta.section }}</span><b>/</b><strong>{{ route.meta.title }}</strong></div>
        <label class="global-search"><span>⌕</span><input placeholder="파일, 심볼, API 검색" aria-label="전역 검색"><kbd>⌘ K</kbd></label>
        <div class="user-control">
          <button class="user-button" :aria-expanded="userOpen" @click="userOpen = !userOpen"><span class="header-avatar">{{ initials }}</span><span class="user-summary"><b>{{ displayName }}</b><small>{{ isSystemAdmin ? 'System Administrator' : '사용자' }}</small></span><span>⌄</span></button>
          <div v-if="userOpen" class="user-menu">
            <div class="user-card"><b>{{ displayName }}</b><span>{{ user.email || user.preferredUsername }}</span></div>
            <RouterLink to="/account" @click="userOpen = false">내 계정</RouterLink>
            <button @click="signOut">로그아웃</button>
          </div>
        </div>
      </header>
      <main class="page-content"><RouterView :user="user" /></main>
    </div>
  </div>
</template>
