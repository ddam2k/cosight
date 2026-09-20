# ADR-0002: Git commit 기반 저장소 snapshot

- 상태: Accepted
- 결정일: 2026-09-20

## 결정

인덱스는 작업 시작 시 확인한 Git commit과 연결한다. tracked 파일은 분석과 코드 조회 모두 `git show <commit>:<path>`에 해당하는 immutable content를 사용한다.

저장소 입력은 허용 root 아래의 서버 경로와 HTTPS/SSH Git URL을 지원한다. Git URL은 Worker 전용 persistent volume에 clone하며 자격증명은 URL이 아닌 Secret Manager reference로 주입한다.

- dirty working tree는 기본적으로 분석하지 않는다.
- 관리자가 명시적으로 working tree 분석을 요청하는 기능은 MVP에서 제외한다.
- untracked 파일은 MVP 인덱스에서 제외하고 warning에 기록한다.
- submodule은 고정된 submodule commit을 별도 analysis unit으로 처리하되 허용 root 밖이면 제외한다.
- Git 저장소가 아닌 경로는 스캔 시작 시 content-addressed snapshot을 작업 저장소에 만들고 완료 후 활성·직전 버전만 보존한다.

## 이유

라이브 mount의 파일이 바뀌면 그래프 range와 Monaco 코드가 달라질 수 있다. commit 또는 content snapshot을 고정해 evidence 재현성을 보장한다.

## 결과

`code_index_versions.git_commit` 또는 `snapshot_id`가 필요하며 file content API는 활성 index의 snapshot을 읽는다.
