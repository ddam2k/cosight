# Cosight MVP RBAC matrix

## 역할

- System Administrator: Keycloak client role. 조직·사용자 할당과 시스템 설정을 관리한다.
- Organization Member: DB membership. 조직 프로젝트 생성·초대 자격만 제공한다.
- Project Admin: 프로젝트 설정·구성원·저장소·분석·공유 관리
- Developer: 코드 열람, 분석 실행과 프로젝트 탐색 저장
- Viewer: 코드·그래프 열람과 개인 탐색 저장

System Administrator와 Organization Member는 프로젝트 코드 권한을 자동으로 부여하지 않는다.

## 프로젝트 권한

| Permission | Project Admin | Developer | Viewer |
| --- | ---: | ---: | ---: |
| `project.read` | O | O | O |
| `project.manage` | O | X | X |
| `project.delete` | O | X | X |
| `member.invite` | O | X | X |
| `member.manage` | O | X | X |
| `repository.connect` | O | X | X |
| `repository.sync` | O | O | X |
| `source.read` | O | O | O |
| `source.search` | O | O | O |
| `graph.read` | O | O | O |
| `analysis.run` | O | O | X |
| `analysis.cancel` | O | 요청자만 | X |
| `analysis.result.read` | O | O | O |
| `exploration.create` | O | O | 개인만 |
| `exploration.update` | 소유 또는 전체 관리 | 소유만 | 소유만 |
| `exploration.share` | O | 정책 허용 시 | X |
| `ai.use` | O | O | 프로젝트 정책에 따라 |
| `audit.read` | 프로젝트 범위 | X | X |

## Scope 추가 조건

| 동작 | 추가 조건 |
| --- | --- |
| 개인 프로젝트 생성 | active 로그인 사용자 |
| 조직 프로젝트 생성 | 현재 organization membership |
| 조직 프로젝트 조회 | organization membership + project membership |
| 조직 프로젝트 초대 | 초대자 Project Admin + 대상 organization membership |
| 조직 프로젝트 역할 변경 | 대상 organization membership |
| 조직원 해제 | System Administrator + 유일 관리자 교체 완료 |
| 시스템 관리 API | 유효한 `cosight-system-admin` 역할 |
| AI 요청 | `ai.use` + 코드 반출 정책 + quota |

## Endpoint matrix

| Endpoint group | Required permission/role |
| --- | --- |
| `/auth/*` | callback/login 제외 authenticated |
| `GET /projects` | authenticated; membership으로 결과 제한 |
| `POST /projects` | authenticated; organization scope이면 조직원 |
| `GET /projects/:id` | `project.read` |
| `PATCH /projects/:id` | `project.manage` |
| `DELETE /projects/:id` | `project.delete` |
| `/projects/:id/members*` | `member.manage` |
| `/projects/:id/invitations*` | `member.invite` 또는 `member.manage` |
| `/projects/:id/index` | `analysis.run` |
| `/projects/:id/files*` | `source.read` |
| `/projects/:id/search` | `source.search` |
| `/projects/:id/graph*` | `graph.read` |
| `/projects/:id/impact` | `analysis.run` |
| `/projects/:id/explorations*` | 소유권 + exploration permission |
| `/projects/:id/ai*` | `ai.use`; 설정은 Project Admin |
| `/jobs/:id*` | 현재 프로젝트 권한 재검사 |
| `/system/*` | System Administrator |

## 거부 규칙

- 알 수 없는 프로젝트와 권한 없는 프로젝트는 모두 `404 PROJECT_NOT_FOUND`로 응답한다.
- 인증은 되었지만 허용된 프로젝트 안에서 동작 권한만 부족하면 `403 PROJECT_ACTION_DENIED`를 반환한다.
- 조직 프로젝트에서 organization membership이 사라지면 기존 세션·SSE·download도 다음 검사에서 즉시 거부한다.
- 프런트 `PermissionGate`는 UX 용도이며 서버 검사를 대체하지 않는다.

## 필수 테스트

- [ ] 역할별 모든 endpoint allow/deny table test
- [ ] 다른 프로젝트 UUID를 사용한 horizontal escalation
- [ ] 조직 프로젝트의 비조직원 초대·수락·조회
- [ ] 조직원 해제 후 기존 session·SSE·download 재사용
- [ ] System Administrator의 project membership 없는 코드 조회
- [ ] 마지막 Project Admin 강등·제거와 조직 해제
- [ ] private exploration의 타 사용자 조회
