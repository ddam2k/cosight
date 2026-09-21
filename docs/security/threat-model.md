# Cosight MVP threat model

## 자산과 경계

자산은 source code, graph metadata, Keycloak token, invitation, LLM credential, prompt context와 audit log다. 경계는 Browser↔Keycloak, Browser↔API, API↔Keycloak/JWKS, API↔DB/Redis/Vault/LLM, Worker↔repository/Helper다.

## 위협과 통제

| 위협 | 통제 | 검증 |
| --- | --- | --- |
| Token 탈취 | 메모리 전용 보관, CSP, 로그 redaction, 짧은 Access Token 수명 | storage·log token scan |
| Token 위조·오용 | issuer, signature, expiry, audience/azp 검증 | 다른 realm·client token 거부 |
| horizontal escalation | project ID마다 membership 재검사, RLS | 다른 project UUID test |
| 조직 밖 공유 | 조직+프로젝트 이중 membership | 비조직원 초대·접근 test |
| 관리자 권한 오용 | Keycloak role, code access 분리, audit | admin code read 거부 |
| repository traversal | realpath allowlist, symlink 검사, immutable snapshot | traversal·race test |
| Helper RCE/network | no shell, network 차단, snapshot read-only | sandbox integration |
| LLM SSRF | base URL allowlist, DNS/IP 재검증, private range policy | rebinding test |
| prompt secret leakage | secret pattern scan, excluded files, preview | seeded secret test |
| token leakage | hash invitation/session token, log redaction | log scan |
| XSS/source rendering | Monaco text model, no HTML execution, CSP | payload fixture |
| cache data leak | user permission/project/index key | membership change test |

## HTTP 보안

- CSP: `default-src 'self'`; script/style/connect endpoint를 명시한다.
- `X-Content-Type-Options: nosniff`, frame-ancestors deny, strict referrer policy를 적용한다.
- mutation은 CSRF header와 Origin을 모두 확인한다.
- invitation token은 URL query나 access log에 넣지 않고 POST body로 전달한다.

## 미해결 확인

- [ ] 운영 reverse proxy의 trusted header 목록
- [ ] 외부 LLM 허용 domain과 private network 정책
- [ ] Keycloak MFA와 session lifetime 정책
- [ ] 조직 초대 이메일 사용 여부
