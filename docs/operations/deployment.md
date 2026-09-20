# MVP deployment

배포 대상은 Kubernetes다. 운영 서비스 기준 URL과 OIDC Callback은 `https://cosight.wasming.com`, Keycloak Realm은 `cosight`, 권장 client ID는 `cosight-web`이다. Ingress/BFF는 root Callback 요청을 내부 `/api/v1/auth/callback` 처리기로 전달한다.

## 실행 단위

- `cosight api`: stateless HTTP/SSE, OIDC, authorization, graph query, LLM gateway
- `cosight worker`: durable jobs, repository snapshot, analyzers
- `node analyzer helper`: Worker Pod sidecar, localhost IPC 외 network disabled
- PostgreSQL 16, Redis, Keycloak, Vault/Secret Manager

API와 Worker는 별도 service account와 DB role을 사용한다. repository volume은 Worker에만 read-only로 mount한다.

## Network

- Browser → API: TLS
- API → Keycloak/LLM/Vault: explicit egress allowlist
- Worker → repository: filesystem only
- Helper: network deny
- PostgreSQL/Redis: private network, TLS where supported

## Configuration

non-secret은 environment, secret은 Secret Manager reference로 주입한다. startup 시 required config를 검증하고 secret value를 log하지 않는다.

- API와 Worker 기준 platform: Linux/amd64, `CGO_ENABLED=0`
- Node Helper: Node.js 22 LTS, pnpm 10 workspace
- Secret Manager: HashiCorp Vault, Kubernetes workload identity, 90일 이내 credential rotation

## Release

1. backup 확인
2. expand migration
3. API/Worker rolling deploy
4. readiness와 smoke test
5. backfill job
6. contract migration은 다음 release

Git URL 저장소는 Worker 전용 persistent volume에 clone하고 API Pod에는 mount하지 않는다. 로컬 저장소 연결은 관리자가 허용한 root를 Worker Pod에 read-only mount한다. 저장소별 최대 byte·file 수는 시스템 설정으로 관리하며 초기 기본값은 2 GiB와 분석 대상 파일 100,000개다.

## 미확정

- [ ] replica 수와 autoscaling
- [ ] backup product와 RPO/RTO
