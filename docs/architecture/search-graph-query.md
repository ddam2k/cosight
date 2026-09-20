# Search and graph query design

## 검색

PostgreSQL `tsvector`와 `pg_trgm`을 사용한다. query를 정규화하고 다음 점수를 합산한다.

| Match | Weight |
| --- | ---: |
| qualified name exact | 100 |
| name exact | 90 |
| name prefix | 70 |
| route/path exact | 65 |
| trigram similarity | 최대 50 |
| full-text rank | 최대 40 |

정렬은 score desc, kind priority, qualified name, node ID 순서다. cursor에는 마지막 sort tuple과 index version을 서명해 넣는다.

## 권한과 버전

query의 첫 조건은 `project_id`와 active `index_version`이다. 조직 프로젝트는 organization membership과 project membership을 모두 검사한다. 권한 필터 후 rank를 계산해 결과 개수도 노출하지 않는다.

## 그래프 traversal

- recursive CTE 또는 application BFS를 사용한다.
- 방문 key는 `(node_id, direction, depth)`다.
- 기본 depth 2, 최대 5; 기본 node 200, 최대 500; edge 최대 1,500
- 같은 source·target·kind edge는 evidence count로 병합할 수 있다.
- cycle edge는 보존하지만 이미 방문한 node를 재확장하지 않는다.
- query timeout은 2초이며 partial result와 warning을 반환할 수 있다.

## Aggregate node

상한을 넘으면 우선순위가 낮은 node를 package/module, directory, kind, impact distance 순으로 묶는다. aggregate에는 hidden node·edge count와 다음 expand token을 넣는다.

## 영향 분석

incoming `calls`, `references`, `implements`, `renders`, `tested_by`를 사용한다. MVP는 직접 영향과 거리 2까지만 제공하고 confidence별로 구분한다.

## Cache

key는 user permission version, project, index version, mode, normalized query hash를 포함한다. membership 변경과 active index 교체 시 관련 namespace를 무효화한다.
