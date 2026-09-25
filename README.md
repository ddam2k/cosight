# Cosight

## 로그인 개발 환경

Cosight는 Keycloak이 제공하는 로그인 UI를 사용한다. 프런트엔드는 백엔드의 공개 설정 API에서 Keycloak URL, Realm, Client ID와 Scope를 읽고 Authorization Code Flow + PKCE로 로그인한다. 이후 모든 보호 API에 Keycloak Access Token을 Bearer Token으로 전송한다.

```bash
cp .env.example .env
# .env에 실제 Keycloak 정보를 입력한 뒤 실행한다.
./run.sh
```

`run.sh`는 필요한 경우 프런트엔드 의존성을 설치하고 Go API와 Vue 개발 서버를 함께 실행한다. 다른 환경 파일을 사용하려면 `COSIGHT_ENV_FILE=/path/to/env ./run.sh`로 지정한다.

Keycloak Client에는 개발용 Valid Redirect URI `http://localhost:5173/*`와 Web Origin `http://localhost:5173`을 등록해야 한다. 운영 환경에서는 `https://cosight.wasming.com/*`만 허용한다.

### 환경변수

| 이름 | 설명 |
| --- | --- |
| `COSIGHT_KEYCLOAK_URL` | Keycloak 서버 URL |
| `COSIGHT_KEYCLOAK_REALM` | Realm (`cosight`) |
| `COSIGHT_KEYCLOAK_CLIENT_ID` | Public SPA Client ID |
| `COSIGHT_KEYCLOAK_SCOPE` | 프런트 요청 scope |
| `COSIGHT_KEYCLOAK_ISSUER` | 선택적 issuer override |
| `COSIGHT_ALLOWED_ORIGINS` | 쉼표로 구분한 CORS Origin |
| `COSIGHT_POSTGRES_URL` | PostgreSQL 전체 연결 URL. 설정 시 개별 연결 값보다 우선 |
| `COSIGHT_POSTGRES_HOST` | PostgreSQL 호스트 (기본 `127.0.0.1`) |
| `COSIGHT_POSTGRES_PORT` | PostgreSQL 포트 (기본 `5432`) |
| `COSIGHT_POSTGRES_USER` | PostgreSQL 사용자 |
| `COSIGHT_POSTGRES_PASSWORD` | PostgreSQL 비밀번호 |
| `COSIGHT_POSTGRES_DATABASE` | PostgreSQL 데이터베이스 이름 |
| `COSIGHT_POSTGRES_SSLMODE` | PostgreSQL SSL 모드 (기본 `require`) |
| `COSIGHT_POSTGRES_MAX_OPEN_CONNS` | 최대 열린 연결 수 (기본 `20`) |
| `COSIGHT_POSTGRES_MAX_IDLE_CONNS` | 최대 유휴 연결 수 (기본 `5`) |
| `COSIGHT_POSTGRES_CONN_MAX_LIFETIME` | 연결 최대 수명 (기본 `30m`) |

`/api/v1/public/auth-config`는 브라우저에 공개해도 되는 값만 반환한다. Client Secret, 관리자 계정과 내부 Keycloak 자격증명은 응답에 포함하지 않는다.

## 데이터베이스 준비

API는 시작할 때 PostgreSQL 연결을 확인하고 [내장 스키마](/internal/database/schema.sql)를 자동 적용한다. 여러 API 인스턴스가 동시에 시작해도 PostgreSQL advisory lock으로 초기화를 직렬화하며 기존 데이터는 유지하고 없는 enum, 테이블과 인덱스만 생성한다. 스키마 생성 권한이나 연결에 문제가 있으면 API는 요청을 받기 전에 종료한다.

애플리케이션 쿼리는 `sqlx`와 pgx 드라이버를 사용하며 프로젝트 생성과 생성자 `project_admin` 등록을 하나의 트랜잭션으로 처리한다.
