# ADR-0004: PostgreSQL 기반 통합 검색

- 상태: Accepted
- 결정일: 2026-09-20

## 결정

MVP 검색은 PostgreSQL `tsvector`와 `pg_trgm`을 사용한다.

- exact qualified name, prefix, trigram, full-text 순으로 가중치를 적용한다.
- 모든 document는 `project_id`와 `index_version`을 가진다.
- active index version과 권한 scope를 SQL 조건에서 강제한다.
- source body 전체는 검색 document에 저장하지 않고 symbol, path, route와 오류 문자열만 색인한다.
- 검색 결과는 stable cursor로 pagination한다.

## 이유

별도 검색 cluster 없이 MVP 규모의 검색과 프로젝트 격리를 구현할 수 있다.

## 결과

대규모 저장소 기준을 초과하면 OpenSearch 등 별도 엔진은 후속 ADR로 검토한다.
