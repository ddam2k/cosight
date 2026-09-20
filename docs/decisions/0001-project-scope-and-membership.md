# ADR-0001: 개인·조직 프로젝트와 멤버십

- 상태: Accepted
- 결정일: 2026-09-20

## 결정

프로젝트는 `personal`과 `organization` scope 중 하나를 가진다.

- 모든 active 사용자는 개인 프로젝트를 만들 수 있다.
- 조직원은 소속 조직의 조직 프로젝트도 만들 수 있다.
- 프로젝트 생성자는 원자적으로 `project_admin`이 된다.
- 개인 프로젝트는 active 사용자를 초대할 수 있다.
- 조직 프로젝트는 현재 조직원만 초대할 수 있다.
- 조직 프로젝트 접근에는 organization membership과 project membership이 모두 필요하다.
- System Administrator는 조직과 조직원만 관리하며 project membership 없이는 코드에 접근하지 못한다.
- 조직원 해제 시 관련 초대와 project membership을 회수한다. 유일한 Project Admin이면 대체 관리자를 요구한다.

## 이유

조직 없는 사용자도 제품을 사용할 수 있게 하면서 조직 코드가 조직 밖으로 공유되는 것을 차단하기 위함이다. 조직 소속만으로 모든 프로젝트를 자동 공개하지 않아 최소 권한을 유지한다.

## 결과

`projects.organization_id`는 nullable이며 `scope`와 일관성 check가 필요하다. 모든 조직 프로젝트 API는 두 membership을 검사해야 한다.
