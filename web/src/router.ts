import { createRouter, createWebHistory } from 'vue-router'
import PageContent from './pages/PageContent.vue'

const page = (name: string, path: string, title: string, description: string, section: string) => ({
  name,
  path,
  component: PageContent,
  meta: { title, description, section },
})

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    page('projects', '/', '내 프로젝트', '개인 프로젝트와 소속 조직의 프로젝트를 확인합니다.', '프로젝트'),
    page('project-new', '/projects/new', '프로젝트 만들기', '저장소를 연결하고 첫 분석을 준비합니다.', '프로젝트'),
    page('invitations', '/invitations', '내 초대', '대기 중인 프로젝트 초대를 확인하고 참여합니다.', '프로젝트'),
    page('project-overview', '/projects/current', '프로젝트 개요', 'cosight-platform의 분석 상태와 최근 변경입니다.', '현재 프로젝트'),
    page('project-explore', '/projects/current/explore', '코드 탐색', '코드 구조와 실행 흐름을 그래프로 탐색합니다.', '현재 프로젝트'),
    page('project-explorations', '/projects/current/explorations', '저장된 탐색', '저장하거나 공유한 코드 탐색 세션입니다.', '현재 프로젝트'),
    page('project-members', '/projects/current/members', '구성원', '프로젝트 구성원과 역할을 관리합니다.', '현재 프로젝트'),
    page('project-settings', '/projects/current/settings/general', '프로젝트 설정', '일반, 저장소, 인덱싱과 AI 설정을 관리합니다.', '현재 프로젝트'),
    page('system-dashboard', '/system', '시스템 대시보드', '서비스와 분석 Worker의 운영 상태를 확인합니다.', '시스템 관리'),
    page('system-organizations', '/system/organizations', '조직 관리', '조직을 만들고 사용자를 할당합니다.', '시스템 관리'),
    page('system-llm-connections', '/system/llm/connections', 'LLM 연결', 'DeepSeek 등 OpenAI-compatible 연결을 관리합니다.', '시스템 관리'),
    page('system-llm-models', '/system/llm/models', '모델 레지스트리', '실제 모델과 기능을 등록합니다.', '시스템 관리'),
    page('system-llm-profiles', '/system/llm/profiles', '논리 모델 프로필', '프로젝트가 사용할 논리 모델을 구성합니다.', '시스템 관리'),
    page('system-projects', '/system/projects', '프로젝트 할당과 한도', '프로젝트별 모델과 사용량 한도를 설정합니다.', '시스템 관리'),
    page('system-usage', '/system/llm/usage', '사용량과 상태', '토큰, 오류율과 응답 지연을 확인합니다.', '시스템 관리'),
    page('system-audit', '/system/audit', '감사 로그', '주요 보안 및 관리 작업을 추적합니다.', '시스템 관리'),
    page('account', '/account', '내 계정', 'Keycloak에서 동기화된 내 정보를 확인합니다.', '사용자'),
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})
