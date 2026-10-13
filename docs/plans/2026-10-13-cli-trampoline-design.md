# Convert4Share CLI 모드 및 C 최소 트램폴린 설계 문서

- **날짜**: 2026-10-13
- **상태**: 승인됨 (Approved)
- **작성자**: Antigravity & Minje Jeon

---

## 1. 개요 및 배경 (Context & Problem Statement)

Convert4Share는 Wails v3 기반의 데스크톱 애플리케이션으로, MOV 및 HEIC 파일을 MP4 및 JPG로 변환하는 도구입니다.

### 요구사항
1. **단일 메인 바이너리 용량 유지**: 전체 GUI와 변환 엔진을 포함한 Go/Wails 바이너리를 두 벌(CLI용, GUI용)로 빌드하여 설치 용량이 두 배가 되는 것을 방지하고, 메인 바이너리 1개의 용량만 유지해야 합니다.
2. **전체 기능의 CLI 지원**: GUI에서 제공하는 모든 변환 기능(화질, 하드웨어 가속, Live Photo 자동 감지, 충돌 해결, 크기 제한, 동시성 등)을 터미널 명령행 인터페이스로 온전히 사용할 수 있어야 합니다.
3. **C 최소 트램폴린(Trampoline) 연동**:
   - Windows 환경에서 GUI 서브시스템(`-H windowsgui`)으로 빌드된 바이너리를 콘솔에서 직접 실행하면, 셸(`cmd.exe`, PowerShell)이 프로세스 종료를 기다리지 않고 즉시 다음 프롬프트를 출력하거나 출력이 비동기적으로 섞이는 문제가 발생합니다.
   - 반대로 콘솔 서브시스템으로 빌드하면 GUI 실행 시 검은 콘솔 창이 번쩍이는(flashing) 문제가 있습니다.
   - 이를 해결하기 위해 약 15~25KB 크기의 순수 C 기반 최소 트램폴린(`c4s`)을 콘솔 서브시스템으로 제공하여, 셸과의 완벽한 입출력 상속, 종료 대기, 종료 코드 중계를 수행합니다.

---

## 2. 전체 아키텍처 (Overall Architecture)

```
[사용자 터미널 (cmd / powershell / bash)]
       │
       ▼
  c4s (C 최소 트램폴린, ~20KB, Console Subsystem)
       │
       │  동일 디렉터리의 convert4share(.exe) 탐색
       │  `convert4share [args...] --cli` 형태로 인자 전달 및 실행
       ▼
  convert4share (메인 Go/Wails 바이너리)
       │
       ├─ [인자에 --cli 포함?] ──▶ Yes ──▶ [internal/cli 엔진 실행]
       │                                       │ Wails/Webview 로딩 없이
       │                                       │ 콘솔 stdout/stderr 출력 후 종료
       ▼ No
  [Wails v3 GUI 런타임 초기화 및 창 표시]
```

---

## 3. 세부 설계 (Detailed Design)

### 3.1. C 최소 트램폴린 (`cmd/c4s/main.c`)
- **언어/규격**: 순수 C99, 외부 라이브러리 의존성 없음 (100줄 이내).
- **실행 바이너리명**: `c4s` (Linux) / `c4s.exe` (Windows).
- **핵심 동작**:
  1. **실행 파일 위치 탐색**:
     - Windows: `GetModuleFileNameW()` 호출 후 디렉터리 경로 추출 ➔ 동일 위치의 `convert4share.exe` 절대 경로 조합.
     - Linux: `/proc/self/exe` (`readlink()`) 호출 후 디렉터리 경로 추출 ➔ 동일 위치의 `convert4share` 절대 경로 조합.
  2. **명령행 인자 재구성**:
     - 첫 번째 인자로 메인 실행 파일 경로 지정.
     - 숨은 플래그 `--cli` 삽입.
     - 사용자가 전달한 원본 `argv[1..argc-1]`을 그대로 뒤에 연결.
  3. **프로세스 실행 및 입출력 상속**:
     - Windows: `CreateProcessW()` 호출. `c4s.exe`가 콘솔 서브시스템이므로 부모의 `stdin`, `stdout`, `stderr` 핸들이 자식에 상속됨. `WaitForSingleObject()`로 종료를 대기하고, `GetExitCodeProcess()`로 자식의 종료 코드를 받아 `exit(code)` 수행.
     - Linux: `execv()`를 호출하여 현재 프로세스 이미지를 `convert4share`로 직접 치환 (PID/파일 디스크립터 그대로 유지, 오버헤드 0).
- **크기**: 컴파일 시 스트립(`-s`) 및 최적화(`-O2`) 적용으로 약 15~25KB.

### 3.2. 메인 바이너리 진입점 분기 (`cmd/convert4share/main.go`)
- `main()` 함수의 최우선 순위에서 `--cli` 인자 포함 여부를 검사:
  ```go
  if isCLIMode(os.Args) {
      os.Exit(cli.Run(stripCliFlag(os.Args[1:])))
  }
  ```
- `--cli`가 주어지면 Wails `application.New()`, Webview2, 프론트엔드 임베드 에셋 초기화 루틴을 완전히 건너뛰고 `internal/cli` 패키지를 직접 실행.
- GUI 실행 파일이므로 탐색기 더블클릭이나 OS 파일 연결(File Association) 실행 시에는 콘솔 창이 전혀 뜨지 않음.

### 3.3. CLI 엔진 및 명령어 인터페이스 (`internal/cli`)
- **사용법**: `c4s [flags] <file1> [file2 ...]`
- **지원 플래그**:
  - `-o, --output <dir>`: 변환 결과 파일 저장 디렉터리 (기본값: 원본 파일 위치 또는 제외 패턴 시 config의 기본 디렉터리).
  - `-q, --quality <high|medium|low>`: 비디오 품질 프리셋 (기본값: config 설정 또는 `high`).
  - `--hwaccel <auto|nvenc|amf|vaapi|cpu>`: 하드웨어 가속기 프리셋 (기본값: config 설정 또는 `auto`).
  - `--max-size <px>`: 비디오 최장축 최대 해상도 다운스케일링 (기본값: 0, 무제한).
  - `--max-image-size <px>`: HEIC 이미지 최장축 최대 해상도 (기본값: config 설정 또는 2560).
  - `--collision <rename|overwrite|skip>`: 파일명 충돌 해결 모드 (기본값: `rename`).
  - `--live-photo`: Live Photo MOV 자동 건너뛰기 여부 (기본값: `true`, `--live-photo=false`로 해제).
  - `-j, --concurrency <n>`: 동시 변환 작업 수 (기본값: config 설정 또는 2).
  - `-v, --verbose`: 상세 디버그 로그 활성화 (`stderr`).
  - `-h, --help`: 사용법 안내 출력.
  - `-V, --version`: 버전 정보 출력.
- **설정 파일 상속**:
  - 기본 실행 시 `%APPDATA%\Convert4Share\config.yaml` (Windows) 또는 `~/.config/convert4share/config.yaml` (Linux)의 설정을 기본값으로 로드.
  - CLI 플래그가 주어질 경우 해당 설정값을 덮어씀.
  - `ffmpeg` 및 `magick` 바이너리 경로는 기존 `tools.Service.DetectBinaries`를 재사용하여 자동 감지.

### 3.4. 공통 배치 변환 엔진 분리 (`internal/batch`)
GUI(`jobs.Service`)와 CLI(`internal/cli`)의 코드 중복을 제거하기 위해 공통 배치 변환 로직을 분리/재사용합니다.
- **공통 컴포넌트**:
  - Live Photo 매칭 (`internal/livephoto.PairedHeicStems`)
  - 충돌 파일명 결정 (`resolveDestination`)
  - 동시성 세마포어 할당 (`concurrency.Sem`)
  - FFmpeg 및 ImageMagick 변환 호출 (`internal/converter`)
  - Pass-through 직접 복사 처리
- **진행도 콜백**:
  - GUI: `jobs.Service`가 Wails 이벤트 버스를 통해 `conversion-progress` 발행.
  - CLI: `internal/cli`가 터미널 표준 출력(`stdout`)으로 파일별 진행률 및 속도를 출력.
- **시그널 처리**:
  - `os.Interrupt`, `syscall.SIGTERM` 수신 시 실행 중인 하위 변환 프로세스를 즉시 안전하게 취소(`context.WithCancel`).
  - 취소 시 생성되던 미완성 임시 파일 자동 정리.
- **종료 코드 (Exit Code)**:
  - 모든 변환 성공 시 `0`.
  - 오류 발생 또는 변환 실패 파일 존재 시 `1`.

---

## 4. 빌드 시스템 및 패키징 (Build & Packaging)

### 4.1. Taskfile 설정
- **`build/Taskfile.linux.yml`**:
  ```yaml
  tasks:
    build:c4s:
      cmds:
        - '{{.CC | default "gcc"}} -O2 -s -o {{.BIN_DIR}}/c4s ./cmd/c4s/main.c'
  ```
  `linux:build`의 선행 종속성(`deps`)에 `build:c4s`를 추가.
- **`build/Taskfile.windows.yml`**:
  ```yaml
  tasks:
    build:c4s:
      cmds:
        - '{{.CC | default "gcc"}} -O2 -s -municode -mconsole -o {{.BIN_DIR}}/c4s.exe ./cmd/c4s/main.c'
  ```
  `windows:build`의 선행 종속성에 `build:c4s`를 추가.

### 4.2. 패키징 설정
- **Linux (`build/linux/nfpm/nfpm.yaml`)**:
  - `/usr/local/bin/c4s`에 `bin/c4s`를 추가 패키징하여 `.deb`, `.rpm`, `archlinux` 패키지 설치 시 즉시 터미널에서 `c4s` 사용 가능.
- **Windows (`build/windows/nsis/project.nsi`)**:
  - `$INSTDIR\c4s.exe` 파일을 설치 항목에 추가.

---

## 5. 검증 계획 (Verification Plan)

1. **단위 테스트 및 빌드 검증**:
   - `go test ./...` 실행하여 기존 테스트 통과 확인.
   - `task build` 실행하여 `bin/convert4share` 및 `bin/c4s` 생성 확인.
   - 생성된 `bin/c4s` 크기 측정 (목표: < 50KB).
2. **C 트램폴린 연동 검증**:
   - `bin/c4s --help` 실행 시 정상 사용법 출력 확인.
   - `bin/c4s --version` 실행 시 버전 정보 출력 확인.
3. **CLI 변환 기능 검증**:
   - 샘플 이미지/동영상 변환 테스트 (HEIC -> JPG, MOV -> MP4).
   - Live Photo 건너뛰기 동작 확인.
   - 옵션 플래그(`-o`, `-q`, `--collision` 등) 정상 적용 확인.
   - 잘못된 인자 전달 시 적절한 에러 메시지 및 exit code 1 반환 확인.
4. **GUI 회귀 검증**:
   - `--cli` 플래그 없이 실행 시 기존 Wails GUI가 정상 구동되는지 확인.
