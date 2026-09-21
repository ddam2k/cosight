<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import type { CurrentUser } from '../types'

const props = defineProps<{ user: CurrentUser }>()
const route = useRoute()
const title = computed(() => String(route.meta.title || 'Cosight'))
const description = computed(() => String(route.meta.description || ''))
const name = computed(() => String(route.name || ''))

const projects = [
  { name: 'cosight-platform', scope: 'WAS MING', role: '관리자', lang: 'Go · TypeScript · Vue', status: '정상', updated: '8분 전' },
  { name: 'billing-service', scope: 'WAS MING', role: '일반 개발자', lang: 'Go · Gin', status: '정상', updated: '32분 전' },
  { name: 'vue-fixture-kit', scope: '개인', role: '관리자', lang: 'TypeScript · Vue', status: '초안', updated: '어제' },
]
</script>

<template>
  <div class="page-heading"><div><p>{{ route.meta.section }}</p><h1>{{ title }}</h1><span>{{ description }}</span></div><RouterLink v-if="name === 'projects'" class="primary-button" to="/projects/new">+ 프로젝트 만들기</RouterLink></div>

  <template v-if="name === 'projects'">
    <div class="notice"><span>✉</span><div><b>대기 중인 프로젝트 초대가 1건 있습니다.</b><small>초대 내용을 확인하고 프로젝트에 참여하세요.</small></div><RouterLink to="/invitations">초대 확인</RouterLink></div>
    <div class="toolbar"><label><span>⌕</span><input placeholder="프로젝트 검색"></label><button>모든 역할</button><button>최근 업데이트순</button></div>
    <div class="project-grid"><article v-for="project in projects" :key="project.name" class="project-card"><div><span class="scope-badge">{{ project.scope }}</span><span class="state-badge">{{ project.status }}</span></div><h2>{{ project.name }}</h2><p>{{ project.lang }}</p><dl><div><dt>내 역할</dt><dd>{{ project.role }}</dd></div><div><dt>최근 분석</dt><dd>{{ project.updated }}</dd></div></dl><RouterLink to="/projects/current">프로젝트 열기 <span>→</span></RouterLink></article></div>
  </template>

  <template v-else-if="name === 'invitations'">
    <div class="data-card"><div class="list-row"><span class="list-icon">C</span><div><b>commerce-web</b><small>Commerce 조직 · 일반 개발자로 초대됨</small></div><span class="muted">최하은 · 2시간 전</span><button>거절</button><button class="accept">수락</button></div></div>
  </template>

  <template v-else-if="name === 'account'">
    <div class="account-grid"><div class="data-card profile-summary"><span class="large-avatar">{{ (user.displayName || user.preferredUsername).slice(0, 1) }}</span><h2>{{ user.displayName || user.preferredUsername }}</h2><p>{{ user.email }}</p><span class="verified">Keycloak 인증 사용자</span></div><div class="data-card account-details"><h2>계정 정보</h2><dl><div><dt>사용자 ID</dt><dd>{{ user.id }}</dd></div><div><dt>사용자 이름</dt><dd>{{ user.preferredUsername }}</dd></div><div><dt>Realm 역할</dt><dd>{{ user.realmRoles.join(', ') || '없음' }}</dd></div><div><dt>Client 역할</dt><dd>{{ user.clientRoles.join(', ') || '없음' }}</dd></div></dl><p>계정 정보와 비밀번호 변경은 Keycloak에서 수행합니다.</p></div></div>
  </template>

  <template v-else-if="name === 'project-overview'">
    <div class="metric-grid"><div><span>코드 심볼</span><b>18,429</b><small>426 files</small></div><div><span>분석 정확도</span><b>97.8%</b><small>confirmed 관계</small></div><div><span>주의 항목</span><b>24</b><small>동적 호출 17 · 파싱 7</small></div><div><span>마지막 분석</span><b>8분 전</b><small>commit a84c92f</small></div></div><div class="data-card empty-panel"><span>⌘</span><h2>최근 분석이 정상적으로 완료되었습니다.</h2><p>코드 탐색에서 호출 흐름과 영향 범위를 확인할 수 있습니다.</p><RouterLink class="primary-button" to="/projects/current/explore">코드 탐색 열기</RouterLink></div>
  </template>

  <template v-else>
    <div class="data-card empty-panel"><span>{{ route.meta.section === '시스템 관리' ? '◉' : '◇' }}</span><h2>{{ title }} 화면 준비 완료</h2><p>메뉴와 라우트가 연결되었습니다. 다음 단계에서 해당 기능의 API와 상세 UI를 구현합니다.</p><div class="route-label">{{ route.path }}</div></div>
  </template>
</template>
