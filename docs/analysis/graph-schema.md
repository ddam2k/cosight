# Code graph schema

> schema version: `1.0`

## 공통 규칙

- line과 column은 1부터 시작하고 end는 exclusive다.
- local ID는 한 analyzer result 안에서만 유효하다.
- Worker가 DB UUID를 할당한다.
- edge에는 evidence가 최소 1개 필요하다.
- `metadata`는 node/edge kind별 허용 field만 저장한다.

## Node

```json
{
  "localId": "n1",
  "stableKey": "go:example/internal/project:method:*Service.Get",
  "kind": "method",
  "name": "Get",
  "qualifiedName": "example/internal/project.(*Service).Get",
  "language": "go",
  "filePath": "internal/project/service.go",
  "range": {"startLine": 42, "startColumn": 1, "endLine": 58, "endColumn": 2},
  "parentLocalId": "n0",
  "metadata": {"signature": "func (s *Service) Get(...)"}
}
```

Required: `localId`, `stableKey`, `kind`, `name`, `language`. `filePath`와 `range`는 external·aggregate node가 아니면 필수다.

### Node kinds

```text
repository directory file package module
struct interface class type_alias enum
function method constructor field property variable constant
parameter type_parameter
vue_component vue_template vue_composable vue_route pinia_store
vue_prop vue_emit vue_slot
http_route http_request middleware
test external_system unresolved aggregate
```

### Metadata

| Kind | 허용 필드 |
| --- | --- |
| function/method | `signature`, `async`, `exported`, `receiver`, `packagePath`, `modulePath` |
| struct/interface/class | `typeParameters`, `exported`, `packagePath`, `modulePath` |
| vue_component | `componentName`, `scriptSetup`, `props`, `emits`, `slots`, `routePaths` |
| vue_route | `path`, `name`, `lazy` |
| pinia_store | `storeId`, `stateNames`, `getterNames`, `actionNames` |
| http_route/http_request | `method`, `pathPattern`, `rawPath`, `dynamicSegments` |
| test | `framework`, `testName`, `suiteName` |
| unresolved | `reason`, `rawExpression` |
| aggregate | `groupBy`, `hiddenNodeCount`, `hiddenEdgeCount` |

## Edge

```json
{
  "localId": "e1",
  "sourceLocalId": "n1",
  "target": {"localId": "n2"},
  "kind": "calls",
  "confidence": "confirmed",
  "metadata": {"dispatch": "static"}
}
```

다른 analysis unit target은 stable key로 참조한다.

```json
{"target": {"stableKey": "ts:app:src/api.ts:function:getProject"}}
```

### Edge kinds

```text
contains imports defines references calls implements
extends embeds exports overrides
registers_route applies_middleware
renders emits_event handles_event uses_composable
reads_state writes_state
http_calls maps_to_route
tested_by changed_with
```

### Confidence

| 값 | 의미 |
| --- | --- |
| `confirmed` | AST와 type/symbol 정보로 단일 대상을 확인 |
| `probable` | 제한된 후보 또는 정적 URL pattern으로 높은 가능성 |
| `inferred` | 문자열·명명 규칙 등 heuristic |

LLM 추론은 graph edge로 저장하지 않는다.

## Evidence

```json
{
  "localId": "v1",
  "subject": {"edgeLocalId": "e1"},
  "filePath": "internal/project/service.go",
  "range": {"startLine": 51, "startColumn": 9, "endLine": 51, "endColumn": 31},
  "analyzer": "go",
  "analyzerVersion": "0.1.0",
  "reasonCode": "STATIC_CALL",
  "reason": "go/types resolved the selected method to repository.FindByID"
}
```

`subject`는 `nodeLocalId` 또는 `edgeLocalId` 중 하나만 가진다. Vue virtual file range는 저장 전에 원본 `.vue` range로 역매핑한다.

## Warning

```json
{
  "code": "DYNAMIC_HTTP_URL",
  "severity": "warning",
  "filePath": "src/api/projects.ts",
  "range": {"startLine": 12, "startColumn": 3, "endLine": 12, "endColumn": 42},
  "message": "HTTP URL could not be normalized",
  "metadata": {"expressionKind": "CallExpression"}
}
```

Severity: `info`, `warning`, `error`. Error warning은 해당 파일 결과가 불완전하다는 뜻이며 전체 index build 실패와 동일하지 않다.

## Stable key

```text
go:<module>/<package>:<kind>:<receiver>.<symbol>
ts:<tsconfig-id>:<module-path>:<kind>:<lexical-qualified-name>
vue:<relative-sfc-path>:<block>:<symbol-kind>:<name>
```

이름 없는 callback은 부모 key와 AST ordinal을 사용한다. file rename은 Git rename과 content hash로 이전 symbol을 연결하지만 UUID를 재사용하지 않는다.

## Validation

- [ ] stable key가 index version 안에서 unique
- [ ] node range가 file line/column 범위 안에 있음
- [ ] edge source와 target이 존재하거나 external/unresolved로 해석됨
- [ ] 모든 edge에 evidence 존재
- [ ] Vue evidence가 virtual file을 가리키지 않음
- [ ] metadata에 허용되지 않은 key가 없음
- [ ] project와 index version을 넘는 참조가 없음
