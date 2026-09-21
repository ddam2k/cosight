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

`/api/v1/public/auth-config`는 브라우저에 공개해도 되는 값만 반환한다. Client Secret, 관리자 계정과 내부 Keycloak 자격증명은 응답에 포함하지 않는다.
