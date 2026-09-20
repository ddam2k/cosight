# Analyzer Helper protocol

> 상태: Draft 0.1

## 실행 모델

Go Worker가 장기 실행 Node Helper process를 시작하고 NDJSON을 stdin/stdout으로 교환한다. stdout은 protocol 전용이며 로그는 stderr로만 출력한다.

- encoding: UTF-8 NDJSON, 한 줄에 JSON object 하나
- protocol version: `1.0`
- 최대 request: 32 MiB
- 최대 response chunk: 8 MiB
- 기본 request timeout: 파일 분석 30초, project type-check 5분
- Worker 종료 시 `shutdown`, 강제 종료 전 grace period 10초
- Helper가 crash하면 진행 request를 retryable failure로 처리하고 최대 3회 재시작

## Envelope

```json
{
  "protocolVersion": "1.0",
  "requestId": "uuid",
  "type": "analyze",
  "payload": {}
}
```

응답:

```json
{
  "protocolVersion": "1.0",
  "requestId": "uuid",
  "type": "result",
  "payload": {},
  "error": null
}
```

`requestId`가 다른 응답은 결합하지 않는다. 알 수 없는 major version은 `PROTOCOL_VERSION_UNSUPPORTED`로 거부한다.

## Message types

| Type | 방향 | 설명 |
| --- | --- | --- |
| `hello` | Helper → Worker | 버전, analyzer와 capability |
| `initialize` | Worker → Helper | workspace와 설정 초기화 |
| `analyze` | Worker → Helper | analysis unit 분석 |
| `cancel` | Worker → Helper | request cooperative cancel |
| `ping` / `pong` | 양방향 | health와 deadlock 탐지 |
| `shutdown` | Worker → Helper | 정상 종료 |
| `progress` | Helper → Worker | 단계와 진행률 |
| `result` | Helper → Worker | 정규화 전 분석 결과 |
| `error` | Helper → Worker | 구조화 오류 |

## Initialize

```json
{
  "protocolVersion": "1.0",
  "requestId": "uuid",
  "type": "initialize",
  "payload": {
    "workspaceId": "project-uuid:index-version",
    "snapshotRoot": "/validated/snapshot",
    "tsconfigPaths": ["tsconfig.json"],
    "packageManager": "pnpm",
    "excludePatterns": ["dist/**"],
    "httpClients": [
      {"module": "@/api/client", "methods": {"get": "GET", "post": "POST"}}
    ],
    "graphSchemaVersion": "1.0"
  }
}
```

Helper는 `snapshotRoot` 밖의 경로를 읽지 않는다. dependency resolution도 허용된 workspace와 설치된 dependency root 안으로 제한한다.

## Analyze

```json
{
  "protocolVersion": "1.0",
  "requestId": "uuid",
  "type": "analyze",
  "payload": {
    "language": "vue",
    "unitId": "src/pages/ProjectDetail.vue",
    "files": [
      {"path": "src/pages/ProjectDetail.vue", "contentHash": "sha256:..."}
    ],
    "reason": "changed",
    "deadlineUnixMs": 1789800000000
  }
}
```

파일 content 자체는 snapshot에서 읽는다. request에 임의 절대 경로나 source 본문을 전달하지 않는다.

## Result

```json
{
  "protocolVersion": "1.0",
  "requestId": "uuid",
  "type": "result",
  "payload": {
    "analyzer": {"name": "vue-typescript", "version": "0.1.0"},
    "schemaVersion": "1.0",
    "unitId": "src/pages/ProjectDetail.vue",
    "nodes": [],
    "edges": [],
    "evidence": [],
    "warnings": [],
    "statistics": {"files": 1, "durationMs": 84}
  },
  "error": null
}
```

Node·edge·evidence는 [graph-schema.md](./graph-schema.md)를 따른다. Worker는 저장 전에 JSON schema, stable key 중복, edge target과 evidence range를 검증한다.

## Error

```json
{
  "protocolVersion": "1.0",
  "requestId": "uuid",
  "type": "error",
  "payload": null,
  "error": {
    "code": "TYPECHECK_FAILED",
    "message": "TypeScript project could not be checked",
    "retryable": false,
    "details": {"tsconfig": "tsconfig.json", "diagnosticCount": 12}
  }
}
```

| Code | Retry | 의미 |
| --- | ---: | --- |
| `INVALID_REQUEST` | X | schema 또는 경로 오류 |
| `PROTOCOL_VERSION_UNSUPPORTED` | X | major version 불일치 |
| `SNAPSHOT_NOT_FOUND` | X | snapshot 손실 |
| `PARSE_FAILED` | X | 파일 parse 실패; 다른 unit 계속 |
| `TYPECHECK_FAILED` | X | 부분 AST 결과와 warning 가능 |
| `TIMEOUT` | O | deadline 초과 |
| `CANCELLED` | X | 사용자 또는 상위 job 취소 |
| `HELPER_OVERLOADED` | O | 동시 요청 상한 |
| `INTERNAL` | 1회 | 예상하지 못한 Helper 오류 |

## Cancellation과 backpressure

- Worker는 request별 `cancel`을 보내고 5초 안에 종료되지 않으면 process를 재시작한다.
- 한 Helper process의 동시 분석 unit은 기본 2개다.
- Worker는 response를 읽지 못하는 상태에서 새 request를 보내지 않는다.
- progress event는 request당 초당 2회로 제한한다.

## 호환성

- protocol major가 같으면 알 수 없는 optional field를 무시한다.
- graph schema major가 바뀌면 full reindex가 필요하다.
- analyzer version 변경 시 관련 analysis unit을 무효화한다.
- initialize 설정 hash를 index version에 기록한다.

## 보안

- Helper는 외부 network를 사용할 수 없는 실행 환경을 기본으로 한다.
- shell command 실행 기능을 제공하지 않는다.
- stderr 로그에 source 본문, 환경변수와 token을 기록하지 않는다.
- symlink를 해석한 실제 경로가 snapshot root 밖이면 거부한다.
