# ADR-0003: PostgreSQL lease 기반 작업 queue

- 상태: Accepted
- 결정일: 2026-09-20

## 결정

MVP의 durable 분석 queue는 PostgreSQL `analysis_jobs`를 사용한다. Worker는 `FOR UPDATE SKIP LOCKED`로 작업을 claim하고 lease와 heartbeat를 갱신한다.

- Redis는 진행 이벤트 fan-out과 단기 상태에만 사용한다.
- 동일 프로젝트에는 동시에 하나의 index build만 허용한다.
- lease가 만료된 `running` 작업은 최대 재시도 횟수 안에서 다시 `queued`로 전환한다.
- cancellation은 cooperative 방식이며 활성 index는 교체하지 않는다.
- API와 Worker 재시작 후에도 PostgreSQL 상태로 복구한다.

## 이유

별도 message broker 없이도 MVP의 내구성과 단순한 운영을 확보하고 기존 PostgreSQL 의존성을 활용한다.

## 결과

job schema에 lease owner, lease expiry, heartbeat, attempt와 cancel request가 필요하다.
