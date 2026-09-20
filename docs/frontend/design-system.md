# Frontend design system

## Tokens

CSS custom property를 단일 source로 사용한다.

```text
spacing: 4, 8, 12, 16, 24, 32
radius: 4, 8, 12
font: 12, 14, 16, 20, 24
breakpoints: 1280, 1440, 1920
graph node height: 36; min width: 120; max width: 280
```

상태는 색상과 함께 icon·text·line style로 표현한다. WCAG AA contrast를 기준으로 한다.

## Layout

- 최소 지원 viewport: 1280×720
- 기본 sidebar 240px, collapsed 56px
- Inspector 360px, 사용자가 280~640px 사이 resize
- 1280 미만은 분석 workspace 사용 제한 안내를 표시한다.

## Components

- Naive UI primitive를 먼저 사용한다.
- async 목록은 `AsyncContent`, `CursorPagination`, `RequestErrorPanel` 조합을 사용한다.
- destructive action은 대상 이름 재입력과 영향 설명을 요구한다.
- role, confidence, language와 job status variant를 중앙 mapping으로 관리한다.
- form error는 field와 summary에 함께 표시한다.

## Accessibility

- 모든 icon button에 accessible name 제공
- focus ring 제거 금지
- graph text fallback과 keyboard node traversal 제공
- ECharts 옆에 같은 데이터 표 제공
- reduced motion에서 layout animation 비활성화

## 상태별 Story

각 page는 loading, empty, partial, error, forbidden, stale와 success 상태를 Storybook 또는 동등한 preview로 검증한다.
