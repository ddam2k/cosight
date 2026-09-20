# Operations runbook

## 기본 확인

1. `/health/live`, `/health/ready` 확인
2. request ID 또는 job ID로 trace 조회
3. DB job state와 lease 확인
4. Provider·Redis·Helper 상태를 분리 확인

## 인덱싱 정체

- heartbeat와 lease expiry 확인
- Worker가 없으면 재기동; lease 만료 후 자동 reclaim 확인
- 같은 프로젝트 active job unique 충돌 확인
- snapshot 손실이면 job 실패 처리 후 full reindex

## Redis 장애

Rate limit은 fail-closed 또는 운영 정책에 따른 제한 모드로 전환한다. job 내구성은 PostgreSQL에 유지된다. SSE는 polling 안내를 반환한다.

## LLM 장애

연결 health와 Circuit 상태를 확인한다. 자격증명은 재노출하지 않고 새 secret version으로 교체한다. streaming 시작 후 retry하지 않는다.

## Backup과 restore

- PostgreSQL full backup + WAL
- Secret은 Secret Manager 정책으로 backup
- repository source는 원본 Git이 기준
- restore 후 RLS, active index, membership과 audit smoke test 수행

## Migration rollback

데이터 손실 가능 migration은 down을 실행하지 않는다. 이전 binary compatibility를 유지하고 restore 또는 forward fix를 사용한다.
