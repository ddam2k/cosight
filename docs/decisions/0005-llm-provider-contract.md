# ADR-0005: OpenAI-compatible LLM adapter 경계

- 상태: Accepted
- 결정일: 2026-09-20

## 결정

Provider adapter는 Cosight 내부의 단일 request·stream event 계약을 구현한다. 외부 API 형식은 adapter 내부에서만 변환한다.

- 필수 기능은 text input, streaming text output, usage와 request ID다.
- tool calling, image, audio와 batch는 MVP에서 제외한다.
- 내부 event는 `metadata`, `delta`, `evidence`, `usage`, `done`, `error`로 정규화한다.
- Provider가 보고한 token usage를 우선하고 없으면 `estimated`로 기록한다.
- streaming 시작 전의 일시적 네트워크 오류만 한 번 재시도한다.
- Base URL은 시스템 관리자가 등록하되 allowlist와 DNS/IP 재검증으로 SSRF를 방어한다.

## 이유

Provider별 차이를 Gateway 밖으로 노출하지 않고 프로젝트 정책·계량·감사를 일관되게 적용하기 위함이다.
