# Wails 3 마이그레이션 + 디렉토리 재구조화 + 리팩터 플랜

작성일: 2026-05-17
대상 브랜치: `wails3-migration` (워크트리 격리)
참조 버전: Wails v3.0.0-alpha.92 (master 브랜치 기준)

> **Status: 완료 (`v0.3.0`, 2026-05-17)**
> 8개 Phase 모두 main에 머지·푸시되었습니다 (`v0.3.0` 태그 → 커밋 `b666aa6`).
> 이 문서는 historical record로 보존됩니다. 현재 빌드/실행 가이드는 `README.md`와 `AGENTS.md`를 참조하세요.

## 0. 문제 정의

현재 코드(v2.11.0)는 안정적이지만 다음 부담이 누적되어 있다:
1. Wails v2 API에 강결합된 단일 `App` 구조체(16개 노출 메서드)
2. Windows 컨텍스트 메뉴를 위한 자체 레지스트리 조작 코드(`windows/registry_windows.go` + `cmd/install.go` + `admin_windows.go`)
3. 평면 디렉토리 구조 — `internal/` 미사용, 도메인 경계 불명확
4. 감사에서 발견된 critical/important 이슈 (excludePatterns 키 불일치, semaphore 초기화/acquire 검증 누락 등)

목표:
- v3 alpha 채택 (`Services` 모델 + 이벤트 객체 + `FileAssociations` 네이티브)
- Go 관례 디렉토리 (`internal/`, `cmd/<app>/`)
- 감사 critical 이슈 동반 해결
- v3 alpha의 instability를 흡수할 수 있는 모듈 경계

## 1. 결정과 트레이드오프

| 결정 | 이유 |
|------|------|
| v3 alpha 채택 | 사용자가 알파 상태 인식 하에 진행 결정 (2026-05-17 대화). 공식 마이그레이션 가이드는 beta 도달 시 작성 예정 — 예제 소스로 매핑 |
| `wails3 generate bindings` 출력은 커밋 | v2의 `wailsjs/`처럼 자동생성이지만 빌드 의존이라 추적 필요 |
| `windows/registry_windows.go` 및 install/uninstall CLI 제거 | v3 `FileAssociations` + NSIS 자동 생성으로 대체. UAC 댄스 불필요 |
| `internal/` 도입, `cmd/convert4share/main.go` 패턴 | Go 관례, 외부 import 방지 |
| Service 단위 분해 | v3 model 강제 + 도메인 경계 명확화 |
| viper, slog, lumberjack 유지 | v3와 호환. slog는 v3가 표준으로 채택 |
| 감사 Critical 이슈는 **Phase 0 선행 처리** | 마이그레이션과 결합 시 회귀 추적 불가 — v2 위에서 먼저 고정 |

## 2. 새 디렉토리 구조

```
convert4share/
├── cmd/
│   └── convert4share/
│       └── main.go                    # application.New + windows + service 등록
├── internal/
│   ├── services/                      # v3 Services (각각 grouping of methods)
│   │   ├── jobs/                      # AddFiles, ConvertFiles, Cancel, Pause/Resume
│   │   │   ├── service.go
│   │   │   └── service_test.go
│   │   ├── settings/                  # GetSettings, SaveSettings
│   │   │   └── service.go
│   │   └── tools/                     # SelectFiles, GetThumbnail, CopyFileToClipboard, DetectBinaries, InstallTool
│   │       └── service.go
│   ├── converter/                     # 도메인: FFmpeg/ImageMagick (현 converter/ 이전)
│   │   ├── ffmpeg.go
│   │   ├── magick.go
│   │   ├── converter.go
│   │   ├── cmd_windows.go
│   │   ├── cmd_unix.go
│   │   ├── converter_test.go
│   │   └── regex_test.go
│   ├── config/                        # viper 래퍼, 기본값, Settings struct
│   │   ├── settings.go
│   │   └── defaults.go
│   ├── livephoto/                     # Live Photo 감지 (.heic ↔ .mov 페어링)
│   │   └── detector.go
│   ├── concurrency/                   # semaphore 패턴 추출 (BUG-001 교훈 내장)
│   │   └── semaphore.go               # acquire 성공 검증 + ctx 취소 통합
│   ├── logging/                       # slog + lumberjack, windowsgui-safe
│   │   └── logger.go
│   └── platform/
│       └── windows/                   # Windows 전용 (Wails 비의존 부분)
│           ├── clipboard.go           # CF_HDROP / Set-Clipboard
│           ├── clipboard_dummy.go
│           ├── winget.go              # winget 탐지
│           └── winget_dummy.go
├── frontend/
│   ├── src/
│   │   ├── bindings/                  # NEW: wails3 generate bindings 출력 (구 wailsjs/ 대체)
│   │   ├── components/                # 기존 유지
│   │   ├── hooks/                     # 기존 유지
│   │   └── ...
│   ├── package.json                   # @wailsio/runtime 추가
│   └── vite.config.ts                 # WAILS_VITE_PORT 환경변수 사용
├── build/                             # NEW: v3 빌드 구성
│   ├── config.yml                     # info + fileAssociations (.mov, .heic)
│   ├── Taskfile.windows.yml
│   ├── Taskfile.common.yml
│   ├── info.json
│   ├── wails.exe.manifest
│   ├── icon.ico
│   ├── appicon.png
│   └── nsis/
│       └── project.nsi
├── docs/
│   └── plans/
│       └── 2026-05-17-wails3-migration.md  # 이 파일
├── Taskfile.yml                       # NEW: 루트 dispatcher
├── go.mod                             # wails/v2 → wails/v3
├── AGENTS.md                          # 기존 유지
├── README.md
└── LICENSE
```

제거 대상:
- `app.go`, `app_jobs.go`, `app_settings.go`, `app_test.go`, `app_tools.go` (루트) — 서비스로 분해 후 삭제
- `main.go` (루트) — `cmd/convert4share/main.go`로 이동
- `cmd/install.go`, `cmd/uninstall.go`, `cmd/install_dummy.go`, `cmd/uninstall_dummy.go`, `cmd/root.go` — v3 FileAssociations로 대체
- `converter/` (루트) — `internal/converter/`로 이동
- `windows/registry_*.go`, `windows/admin_*.go` — v3 FileAssociations로 대체
- `wails.json` — `build/config.yml` + `Taskfile.yml`로 대체
- `frontend/src/wailsjs/` — `frontend/src/bindings/` + npm `@wailsio/runtime`로 대체
- `build_dev.go`, `build_prod.go` — v3에서는 빌드 태그 필요시 새 위치로 이동 또는 제거
- `verification/` (만약 v2 마이그레이션 전용 스크립트라면 검토)

## 3. v2 → v3 API 대응표 (확인된 항목)

| v2 (현재 코드) | v3 (마이그레이션 후) |
|---|---|
| `wails.Run(&options.App{...})` (main.go:138) | `application.New(application.Options{...}).Run()` |
| `Bind: []interface{}{ app }` | `Services: []application.Service{ application.NewService(jobsService), application.NewService(settingsService), application.NewService(toolsService) }` |
| `AssetServer: &assetserver.Options{ Assets: frontend }` | `Assets: application.AssetOptions{ Handler: application.AssetFileServerFS(frontend) }` |
| `SingleInstanceLock: &options.SingleInstanceLock{UniqueId, OnSecondInstanceLaunch}` | `SingleInstance: &application.SingleInstanceOptions{UniqueID, OnSecondInstanceLaunch}` (대소문자 변경) |
| `DragAndDrop: &options.DragAndDrop{EnableFileDrop: true, DisableWebViewDrop: true}` | `WebviewWindowOptions{EnableFileDrop: true}` (단일 옵션으로 통합) |
| `OnStartup: app.startup` (ctx 받음) | `app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(e *ApplicationEvent){...})` |
| `OnDomReady: app.domReady` | `events.Common.WindowDOMReady` (window scoped) |
| `OnShutdown: app.shutdown` | `Options.OnShutdown func()` |
| `OnBeforeClose: app.beforeClose` | `win.RegisterHook(events.Common.WindowClosing, ...)` + `e.Cancel()` |
| `runtime.EventsEmit(ctx, name, data)` | `app.Event.Emit(name, data)` |
| `runtime.EventsOn(ctx, name, cb)` (Go에서 수신) | `app.Event.On(name, func(e *CustomEvent){ /* e.Data */ })` |
| `runtime.WindowShow(ctx)` | `window.Show()` 또는 `window.Focus()` |
| `runtime.WindowUnminimise(ctx)` | `window.Restore()` |
| `runtime.WindowSetAlwaysOnTop(ctx, b)` | `window.SetAlwaysOnTop(b)` 또는 옵션 |
| `runtime.OnFileDrop(...)` (frontend) | `Events.On("files-dropped", e => /* e.data */)` + DOM `data-file-drop-target` |
| `runtime.OpenMultipleFilesDialog(ctx, opts)` | `app.Dialog.OpenFile().AllowMultiple(true).Show()` (fluent) |
| `Logger: customLogger` | `Logger: *slog.Logger, LogLevel: slog.Level` |
| `frontend/src/wailsjs/runtime/runtime` import | `import { Events, Window, Dialog } from "@wailsio/runtime"` |
| `frontend/src/wailsjs/go/main/App` import | `import { JobsService, SettingsService, ToolsService } from "./bindings/..."` |
| `wails.json` | `build/config.yml` + `Taskfile.yml` |
| `wails build` / `wails dev` | `task build` / `task dev` (내부에서 `wails3` CLI 호출) |
| `windows/registry_windows.go` (직접 레지스트리 조작) | `Options.FileAssociations: []string{".mov", ".heic"}` + `build/config.yml`의 `fileAssociations:` |
| `cmd/install.go` (CLI install) | NSIS 인스톨러가 사용자가 인스톨할 때 자동 등록 |
| `app.OnSecondInstanceLaunch`에서 직접 파일 처리 | `OnSecondInstanceLaunch(data SecondInstanceData)`에서 `data.Args` 처리 + `ApplicationOpenedWithFile` 이벤트도 활용 |

## 4. 이벤트 변경

이벤트명은 유지(`conversion-progress`, `queue-paused`, `queue-resumed`, `all-jobs-done`, `file-added`, `files-received`, `frontend-ready`). 프론트엔드 콜백 시그니처가 변경됨:

```typescript
// v2
EventsOn("conversion-progress", (status: JobStatus) => { ... })

// v3
import { Events } from "@wailsio/runtime";
Events.On("conversion-progress", (e: { data: JobStatus }) => {
    const status = e.data;  // 한 단계 wrapping
});
```

`queue-paused`/`queue-resumed`/`all-jobs-done`는 payload 없이 emit하거나 의미 있는 객체로 변경 (감사 권장사항).

## 5. 단계별 작업 (subagent-driven-development 실행 단위)

각 단계는 독립 task로 dispatch, 끝나면 spec-reviewer → code-quality-reviewer.

### Phase 0 — 사전 정리 (v2 위에서 먼저, **선행 PR**)

**Task 0.1: Critical 버그 수정** (`go-implementer`)
- `excludePatterns`/`excludeStringPatterns` 키 통일 (`app_settings.go:21`, `:137`, `:156`, `config.example.yaml`, 프론트엔드 인터페이스 `Settings`)
- `NewApp()`에서 `ffmpegSem`, `magickSem` 초기화 (`app.go`)
- semaphore acquire 성공 검증 후 `defer release` (`app_jobs.go:255-256`, `:276-277`, `app_tools.go:181-184`)
- `regexp.MustCompile` 핫패스 검증 (이미 OK이지만 재확인)
- `logLevel` 키 `config.example.yaml`에 문서화

**Task 0.2: 프론트엔드 이펙트 cleanup 보강** (`react-implementer`)
- `App.tsx:74-91` install polling double-cleanup
- `SettingsIntegration.tsx:31-42` 명시적 cleanup ref
- `useFileQueue.ts:92-130` 의존성 배열 좁히기
- `FileList.tsx:55-60` aria-label

Phase 0는 v2 main에 commit 후 새 워크트리에서 cherry-pick 또는 main에서 시작.

### Phase 1 — 디렉토리 재구조화 (v2 유지, **빌드 가능 보장**)

**Task 1.1: internal/ 구조 도입 + 코드 이동** (`go-implementer`)
- 새 패키지 생성: `internal/converter`, `internal/config`, `internal/concurrency`, `internal/logging`, `internal/livephoto`, `internal/platform/windows`
- 기존 `converter/` → `internal/converter/`
- `app_settings.go`의 Settings struct와 viper 코드 → `internal/config/settings.go`
- `windows/clipboard*` → `internal/platform/windows/clipboard.go`
- `windows/winget*` → `internal/platform/windows/winget.go`
- `windows/admin*`, `windows/registry*`는 일단 유지 (Phase 6에서 삭제)
- `app.go`의 semaphore 관련 → `internal/concurrency/semaphore.go`
- `main.go`의 logger 설정 → `internal/logging/logger.go`
- import path 갱신
- **검증**: `go build ./...`, `go test ./internal/converter/...`, `wails generate module`, `wails build`

이 phase 끝에 v2 빌드가 여전히 동작해야 한다. Phase 2 진입 전에 PR 분리 가능.

### Phase 2 — v3 골격 (병렬 코드, v2와 공존하지 않음 — clean break)

**Task 2.1: go.mod 교체 + cmd/convert4share/main.go 작성** (`go-implementer`)
- `go.mod`: `wails/v2` → `wails/v3` (alpha.92)
- `cmd/convert4share/main.go` 작성:
  - `application.New(application.Options{Name, Description, Services, Assets, SingleInstance, FileAssociations, Logger, LogLevel, OnShutdown})`
  - `app.Window.NewWithOptions(WebviewWindowOptions{URL, EnableFileDrop, Width, Height, ...})`
  - `app.Run()`
- 루트 `main.go`, `build_dev.go`, `build_prod.go` 삭제
- `cmd/install.go`, `cmd/uninstall.go`, `cmd/root.go` 등 삭제 (v3 FileAssociations로 대체)
- **검증**: `go build ./...` (서비스 비어있어도 컴파일됨)

### Phase 3 — Services 분해 (병렬 dispatch 가능)

세 서비스를 병렬로 dispatch (`dispatching-parallel-agents` 적용):

**Task 3.1: JobsService** (`go-implementer`)
- 위치: `internal/services/jobs/service.go`
- 노출 메서드: `AddFiles`, `ConvertFiles`, `CancelJob`, `PauseQueue`, `ResumeQueue`
- 이벤트 emit: `conversion-progress`, `file-added`, `files-received`, `queue-paused`, `queue-resumed`, `all-jobs-done`
- v3 `*application.App` 주입 (NewService 등록 시) 또는 `application.Get()`로 접근
- semaphore는 `internal/concurrency.Sem` 사용
- TDD: 가능한 부분 (큐 관리, 상태 전이)

**Task 3.2: SettingsService** (`go-implementer`)
- 위치: `internal/services/settings/service.go`
- 노출 메서드: `GetSettings`, `SaveSettings`
- `internal/config` 위임

**Task 3.3: ToolsService** (`go-implementer`)
- 위치: `internal/services/tools/service.go`
- 노출 메서드: `SelectFiles`, `SelectBinaryDialog`, `GetThumbnail`, `CopyFileToClipboard`, `DetectBinaries`, `InstallTool`
- `app.Dialog.OpenFile()` fluent API 사용
- `internal/platform/windows.CopyFileToClipboard` 호출
- `InstallContextMenu`/`UninstallContextMenu`/`GetContextMenuStatus` **삭제** (v3 FileAssociations로 자동 처리)

### Phase 4 — 라이프사이클 이벤트와 윈도우 활성화

**Task 4.1: ApplicationStarted / WindowDOMReady / OnShutdown 매핑** (`go-implementer`)
- 현재 `App.startup(ctx)`의 로직 → `ApplicationStarted` 핸들러
- 현재 `App.domReady(ctx)`의 로직 → `WindowDOMReady` 핸들러
- 현재 `App.shutdown(ctx)`의 로직 → `Options.OnShutdown`
- 현재 `App.beforeClose(ctx)` → `WindowClosing` hook + `e.Cancel()`

**Task 4.2: SingleInstance 콜백에서 윈도우 활성화** (`go-implementer`)
- `OnSecondInstanceLaunch(data SecondInstanceData)`에서 `window.Restore(); window.Focus()` 호출
- `data.Args`를 jobs 서비스로 전달

**Task 4.3: ApplicationOpenedWithFile 이벤트로 파일 연결 처리** (`go-implementer`)
- 컨텍스트 메뉴 클릭으로 앱이 시작되거나 두 번째 인스턴스가 시작될 때 파일 경로 수신

### Phase 5 — 프론트엔드 마이그레이션

**Task 5.1: 패키지 의존성과 runtime import 교체** (`react-implementer`)
- `package.json`에 `@wailsio/runtime: latest` 추가
- `frontend/src/wailsjs/` 폴더 삭제
- `wails3 generate bindings -clean -b -d frontend/src/bindings` 실행 후 결과 커밋
- 모든 import 갱신:
  - `import * as runtime from './wailsjs/runtime/runtime'` → `import { Events, Window, Dialog } from "@wailsio/runtime"`
  - `import { X } from './wailsjs/go/main/App'` → `import { X } from "./bindings/.../jobsservice"` (등)
- `runtime.OnFileDrop(...)` → `Events.On("files-dropped", e => ...)` + HTML에 `data-file-drop-target`
- 이벤트 콜백: payload가 `e.data`에 wrapping된 점 모두 갱신
- `runtime.EventsEmit("frontend-ready")` → `Events.Emit("frontend-ready")`

**Task 5.2: DropZone에 `data-file-drop-target` 적용** (`react-implementer`)
- `DropZone.tsx`에 attribute 추가
- 글로벌 CSS의 `--wails-drop-target` 대신 v3가 자동 토글하는 `.file-drop-target-active` 클래스 사용

### Phase 6 — 빌드 시스템 교체

**Task 6.1: build/config.yml + Taskfile.yml 작성** (`go-implementer`)
- `wails.json` 삭제
- `build/config.yml`:
  ```yaml
  version: '3'
  info:
    companyName: ""
    productName: "Convert4Share"
    productIdentifier: "io.github.minjejeon.convert4share"
    description: "Converts MOV/HEIC to MP4/JPG"
    version: "0.3.0"
  fileAssociations:
    - ext: mov
      name: "QuickTime Video"
      description: "QuickTime Video (handled by Convert4Share)"
      iconName: icon
      role: Editor
    - ext: heic
      name: "HEIC Image"
      description: "High Efficiency Image (handled by Convert4Share)"
      iconName: icon
      role: Editor
  ```
- `Taskfile.yml` 작성 (file-association 예제 참고)
- `build/Taskfile.windows.yml`: `go build -tags production -trimpath -ldflags="-w -s -H windowsgui"`
- `wails3 task common:update:build-assets` 실행 → NSIS 스크립트 생성

**Task 6.2: 기존 `taskfile.yaml` (소문자) 정리** (`go-implementer`)
- 기존 task 흐름을 새 Taskfile.yml로 이전
- `task build`, `task dev`, `task test`, `task release`

### Phase 7 — Windows 통합 코드 정리

**Task 7.1: registry/admin/install/uninstall 코드 제거** (`go-implementer`)
- `windows/registry_windows.go`, `windows/registry_dummy.go` 삭제
- `windows/admin_windows.go`, `windows/admin_dummy.go` 삭제 (v3는 NSIS가 처리)
- `cmd/install*.go`, `cmd/uninstall*.go`, `cmd/root.go` 삭제 (Phase 2에서 이미 했지만 재확인)
- `ToolsService`에서 `InstallContextMenu`, `UninstallContextMenu`, `GetContextMenuStatus` 메서드 제거
- 프론트엔드 `SettingsIntegration.tsx`에서 install/uninstall 버튼 UI 제거 또는 "Open Installer" 안내로 변경

### Phase 8 — 최종 검증

**Task 8.1: 통합 검증** (`go-implementer` + `react-implementer`)
- `task build` 성공
- `task dev` 동작 (Vite + Wails dev 서버)
- NSIS 인스톨러 빌드 후 클린 머신에 설치 → 컨텍스트 메뉴 등록 확인 → .mov 우클릭으로 앱 실행 확인
- 모든 이벤트 동작:
  - 변환 진행률 표시
  - 일시정지/재개
  - 단일 인스턴스 (두 번째 실행 시 윈도우 활성화)
- Live Photo 감지 동작
- 썸네일 생성
- Clipboard 복사

**Task 8.2: Final code review** (`code-quality-reviewer` on entire diff)

## 6. 위험과 완화

| 위험 | 완화 |
|---|---|
| Wails v3 alpha API 후속 breaking change | 마이그레이션 후에도 `go.mod`의 v3 버전을 명시적으로 고정. CI에서 v3 latest 빌드도 정기 확인 |
| `wails3 generate bindings` 출력 형식 변동 | 출력을 커밋하지만, regenerate가 깨끗하게 동작하도록 task에 명확히 정의. 차이 발생 시 PR로 별도 처리 |
| FileAssociations + NSIS 자동 처리가 기존 사용자의 우클릭 동작과 충돌 | 사전 uninstall 가이드 README에 추가. v2 인스톨러가 등록한 `HKCU\Software\Classes\SystemFileAssociations\.mov\shell\Convert4Share` 키 정리 |
| HEIC/MOV 컨텍스트 메뉴의 "Send to" 흐름이 NSIS 등록과 다를 수 있음 | Phase 6.1 작업 시 NSIS 생성 결과를 검토하여 `SystemFileAssociations` 키도 자동 등록되는지 확인. 안 되면 NSIS 스크립트 커스터마이즈 |
| Vite 8 + `@wailsio/runtime` 마이너 버전 호환성 | `package-lock.json`에 정확한 버전 고정, dev에서 검증 |
| Frontend Windows drag&drop이 v3에서 회귀 | `EnableFileDrop: true` + DOM 어트리뷰트 흐름을 단계별로 검증. 회귀 시 v2 패턴으로 일시 폴백 가능 |
| Long-running migration → main 충돌 | 워크트리 격리 + 정기 rebase. main에 사소한 변경만 들어가도록 communication |

## 7. 메모리/문서 영향

- `AGENTS.md`: v2 항목들(Wails generate module, wailsjs 폴더 설명, wails.json 트러블슈팅) 갱신 필요 — Phase 8에 포함
- `README.md`: 빌드 명령 변경 (`task build`), prerequisite 갱신 (`go-task`, `wails3`)
- `.claude/agents/wails-bridge.md`: v3 패턴 반영 (Phase 8)
- `memory/agent_team_structure.md`: 변경 없음 (팀 구조 동일)

## 8. 실행 절차

1. **계획 승인**: 이 문서 검토 후 사용자 승인
2. **워크트리 진입**: `EnterWorktree(name: "wails3")` — main 보호
3. **Phase 0 먼저**: main에서 Critical 버그만 수정 후 PR (Wails v2 위에서) → main에 merge
4. **Phase 1부터 워크트리에서**: 각 Phase의 task를 subagent dispatch
5. **각 task 후**: spec-reviewer → 통과 시 code-quality-reviewer → 통과 시 commit
6. **Phase 단위 PR 생성** 권장 (rebase 부담 감소)
7. **모든 Phase 완료 후**: 최종 통합 PR 또는 `wails3-migration` 브랜치 자체를 squash merge

## 9. 미해결 결정 사항

이 플랜을 실행하기 전에 아래 항목 확인 필요:

A. **버전 번호 결정**: 0.2.1 → 0.3.0(minor)? 1.0.0(major)? — v3는 의미상 major break지만 사용자 시인성은 minor 변경 충분
B. **v2 사용자 마이그레이션 안내**: 기존 v2 설치 사용자가 v3로 업그레이드 시 레지스트리 정리가 자동인지 수동인지. 자동이면 마이그레이션 헬퍼 코드 필요 (v3 인스톨러 시작 시 옛 키 정리)
C. **CI/Release 워크플로**: 현재 `taskfile.yaml`의 `release` task 동작 — v3에서 어떻게 표현될지 같이 마이그레이션
D. **`AGENTS.md` 또는 `CLAUDE.md` 갱신 시점**: 마이그레이션 중간(혼란 방지)? 끝(정확성)?

---

이 플랜은 살아있는 문서다. 각 Phase 실행 중 발견되는 v3 API 함정이나 우리 코드의 숨은 의존성은 이 문서에 추가하고, 필요시 phase를 재분할한다.
