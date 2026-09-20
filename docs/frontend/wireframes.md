# MVP wireframe checklist

## 전역

```text
┌ Header: project switcher | search | index | invitations | user ┐
├ Sidebar ──────────┬ Main content ───────────────────────────────┤
│ role-based menu   │ page header / actions / async content      │
└───────────────────┴─────────────────────────────────────────────┘
```

## 핵심 화면

- [ ] 로그인: 정상, Keycloak 장애, session 만료
- [ ] 프로젝트 목록: 개인/조직 filter, no project, pending invitation
- [ ] 프로젝트 생성: scope, identity, repository, review, validation failure
- [ ] 조직 관리: 목록, 생성, 사용자 할당, 유일 관리자 교체
- [ ] 프로젝트 개요: no index, indexing, ready, failed
- [ ] 분석 workspace: tree/graph/inspector, partial graph, stale source, no permission
- [ ] 구성원: 초대, 역할 변경, 마지막 관리자 차단
- [ ] 시스템 관리: Provider empty/error, quota warning, audit detail

## 분석 workspace

```text
┌ project / branch / search / save ┐
├ Explorer ┬ Graph canvas ┬ Inspector ┤
│ tree     │ mode/filters │ code      │
│ entries  │ evidence path│ evidence  │
├──────────┴──────────────┴───────────┤
│ job status / warnings / retry       │
└─────────────────────────────────────┘
```

제품 검토 시 각 체크 항목에 desktop screenshot과 interaction note를 연결한다.
