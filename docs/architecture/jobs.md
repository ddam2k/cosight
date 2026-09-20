# Background job architecture

## 상태 전이

```text
queued → running → completed
   └────→ cancelled
running → cancel_requested → cancelled
running → queued            (retryable failure and attempts remain)
running → failed            (terminal failure)
```

## Claim과 lease

Worker는 transaction 안에서 다음 형태로 한 작업을 claim한다.

```sql
SELECT id FROM analysis_jobs
WHERE status = 'queued'
ORDER BY created_at
FOR UPDATE SKIP LOCKED
LIMIT 1;
```

claim 시 `running`, `lease_owner`, `lease_expires_at=now()+30s`, `attempt+1`을 저장한다. Worker는 10초마다 heartbeat를 갱신한다. Reaper는 lease가 60초 이상 만료된 작업을 재시도하거나 실패 처리한다.

## 동시성과 중복 방지

- 프로젝트별 활성 index job은 partial unique index로 하나만 허용한다.
- mutation API는 `Idempotency-Key`와 actor·route·request hash를 24시간 저장한다.
- 같은 key와 다른 body는 `409 IDEMPOTENCY_KEY_REUSED`다.
- 완료 job 재호출은 기존 job ID와 결과를 반환한다.

## 취소와 retry

- API는 `cancel_requested_at`을 기록한다.
- scanner, analyzer unit, merge 단계 경계에서 취소를 확인한다.
- 활성 index 교체 transaction이 시작된 뒤에는 취소하지 않고 완료한다.
- network/process crash와 timeout만 retryable이다. schema·parse·permission 오류는 재시도하지 않는다.
- 기본 최대 attempt는 3, exponential backoff는 5초·30초다.

## 진행 이벤트

Worker는 PostgreSQL 상태를 갱신한 뒤 Redis pub/sub에 best-effort event를 보낸다. SSE 재연결 시 DB snapshot을 먼저 보내고 이후 event를 구독한다.

```text
snapshot, stage, progress, warning, completed, failed, cancelled
```

## 복구

- API 재시작: DB 상태로 정상 동작
- Worker 재시작: lease 만료 후 reclaim
- Redis 장애: SSE 실시간성만 저하, polling 가능
- DB 장애: 작업 중단, snapshot 활성 버전 유지
- Helper crash: 현재 unit 1회 재시도 후 warning 또는 job 실패
