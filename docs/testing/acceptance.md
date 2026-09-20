# MVP acceptance checklist

## 인증·조직·프로젝트

- [ ] 조직 없는 사용자가 개인 프로젝트 생성
- [ ] System Administrator가 조직 생성과 사용자 할당
- [ ] 조직원이 조직 프로젝트 생성
- [ ] 비조직원 초대·수락·직접 접근 거부
- [ ] 조직원 해제 후 session·SSE·download 접근 거부
- [ ] Project Admin/Developer/Viewer allow·deny matrix 통과

## 분석

- [ ] Go package, type, interface, direct call과 Gin route evidence
- [ ] TypeScript import/export/call, Pinia와 HTTP request evidence
- [ ] Vue component/prop/emit/template/router와 원본 source map evidence
- [ ] Vue→TypeScript API→Gin handler path 재현
- [ ] parse/type 오류가 부분 결과와 warning으로 표시
- [ ] active index 교체 중 이전 버전 계속 조회

## 탐색 UI

- [ ] 검색→그래프→코드 evidence 이동
- [ ] 구조·흐름·영향 mode와 점진 확장
- [ ] node 500/edge 1,500 상한에서 상호작용 가능
- [ ] keyboard와 graph text fallback
- [ ] stale snapshot 및 partial result 표시

## LLM

- [ ] 전송 문맥 preview와 정책 거부
- [ ] facts의 evidence ID 검증
- [ ] RPM·TPM·동시성·일월 한도
- [ ] streaming cancel, timeout, 429와 usage 정산

## 보안

- [ ] CSRF, traversal, symlink race, SSRF와 XSS fixture
- [ ] log/trace에 source, token, secret 없음
- [ ] cache key와 RLS project isolation
- [ ] System Administrator의 무권한 code read 거부

## 성능 기록

각 기준 저장소마다 file/node/edge 수, full/incremental index 시간, search p95, graph p95와 browser memory를 기록한다.
