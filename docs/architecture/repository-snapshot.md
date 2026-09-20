# Repository snapshot and source consistency

## Git 저장소

1. 로컬 연결은 등록 경로의 realpath와 allowlist를 검증한다. Git URL 연결은 HTTPS/SSH scheme, host allowlist와 Secret Manager의 read-only credential reference를 검증한 뒤 Worker 전용 volume에 clone한다.
2. 작업 시작 시 `HEAD` commit과 submodule commit을 기록한다.
3. tracked file inventory와 content는 commit object에서 읽는다.
4. analyzer에는 immutable snapshot root만 전달한다.
5. file content API도 index version의 commit에서 읽는다.

dirty·untracked 파일은 MVP에서 분석하지 않고 warning으로 표시한다. branch가 이동해도 활성 index의 commit 내용은 계속 조회할 수 있어야 하므로 운영 clone은 필요한 object를 보존한다.

## 비 Git 경로

파일을 content hash 기반 snapshot directory에 hardlink 또는 copy한다. snapshot manifest에는 relative path, hash, size, mode를 기록하고 symlink는 따라가지 않는다.

## 보존과 정리

- active와 직전 successful snapshot 보존
- building snapshot은 job terminal 후 24시간 안에 정리
- failed snapshot metadata는 30일 보존
- project delete는 신규 접근을 즉시 차단하고 삭제 job에서 snapshot과 관리 clone을 영구 제거한다. 복구 유예는 제공하지 않는다.

## Stale 처리

API 응답은 `indexVersion`, `snapshotId`, `contentHash`를 포함한다. snapshot을 읽을 수 없으면 live file로 대체하지 않고 `410 SOURCE_SNAPSHOT_GONE`을 반환한다.

## 보안

- 절대 경로와 `..`는 manifest에서 거부한다.
- 각 read 전에 root-relative open과 symlink escape를 검사한다.
- Git command는 인자 배열과 read-only allowlist를 사용한다.
- archive 추출과 hook 실행은 금지한다.
