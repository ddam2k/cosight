# Analyzer golden fixtures

## Fixture layout

```text
testdata/analysis/<case>/
 ├─ repo/                 # input source
 ├─ config.json           # build tags, tsconfig, HTTP wrappers
 └─ expected/
     ├─ nodes.json
     ├─ edges.json
     ├─ evidence.json
     └─ warnings.json
```

Expected JSON은 DB UUID와 duration을 제외하고 stable key 기준으로 정렬한다.

## 필수 cases

| Case | 검증 |
| --- | --- |
| `go-basic` | package, struct, interface, method, direct call |
| `go-interface` | 단일·복수 구현과 probable dispatch |
| `go-build-tags` | 활성·비활성 file과 설정 기록 |
| `gin-nested` | Group path, middleware 순서, handler |
| `ts-project-references` | tsconfig reference, paths, re-export |
| `ts-calls` | overload, alias, callback, unresolved dynamic property |
| `ts-pinia-http` | store state/action, fetch, Axios, custom wrapper |
| `vue-script-setup` | props, emits, composable, template reference |
| `vue-router` | nested/lazy route와 component |
| `vue-dynamic` | dynamic component warning과 source map |
| `cross-http` | Vue→TS request→Gin route 연결 |
| `incremental` | edit, rename, delete, config/analyzer version 변경 |

## 초기 합격 기준

- confirmed node·edge precision 98% 이상
- Gin route method/path/handler 100%, middleware 순서 95% 이상
- Vue source evidence range 성공률 99% 이상
- HTTP request→Gin route precision 95% 이상, recall 85% 이상
- 잘못된 confirmed edge 0건
- 모든 persisted edge에 유효 evidence 존재

기준은 M0 prototype 측정 후 변경 사유를 ADR로 기록한다.
