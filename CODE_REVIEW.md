# Code Review Report

**Review Date**: 2026-01-11
**Files Reviewed**: 15 Go files
**Overall Assessment**: Approved with suggestions

---

## Critical Issues

### 1. Response Writer 에러 미처리
- **위치**: `internal/api/http/handler/response.go:24-28`
- **문제**: `WriteJSON`, `WriteError` 반환 에러가 호출자에서 무시됨
- **영향**: JSON 인코딩 실패 시 클라이언트가 불완전한 응답을 받음
- **해결**: 에러 로깅 추가
```go
if err := rw.WriteJSON(w, statusCode, errResp); err != nil {
    rw.logger.Printf("[ERROR] Failed to write JSON response: %v", err)
}
```

### 2. 비밀번호 평문 저장
- **위치**: `internal/domain/cluster/entity.go:28-29`
- **문제**: Cluster 엔티티에 Password가 평문으로 저장
- **영향**: 메모리 덤프, 로그, 응답에서 민감 정보 노출 가능
- **해결**: 토큰 기반 인증으로 변경, CredentialStore 분리

### 3. TLS 검증 비활성화
- **위치**: `internal/config/app.go:41`
- **문제**: `InsecureSkipVerify`가 true로 하드코딩
- **영향**: MITM 공격에 취약
- **해결**: 환경변수로 설정 가능하게 변경, 기본값 false

---

## Important Issues

### 4. HTTP 입력 검증 누락
- **위치**: `internal/api/http/handler/cluster_handler.go:39-55`
- **문제**: DTO에 validation 태그 있으나 실제 적용 안됨
- **해결**: `go-playground/validator` 적용

### 5. Cluster ID 추출 방식 취약
- **위치**: `internal/api/http/handler/cluster_handler.go:104, 134`
- **문제**: `strings.TrimPrefix` 사용, 라우팅 변경 시 깨짐
- **해결**: Go 1.22+ `r.PathValue("id")` 사용

### 6. 에러 타입 구분 누락
- **위치**: `internal/application/services/cluster_service.go:92-96`
- **문제**: DB 에러와 NotFound 구분 안됨
- **해결**: `errors.Is()` 사용
```go
if !errors.Is(err, common.ErrClusterNotFound) {
    return nil, fmt.Errorf("failed to verify cluster uniqueness: %w", err)
}
```

### 7. 불완전한 에러 래핑
- **위치**: `internal/infrastructure/proxmox/client.go:109`
- **문제**: 응답 본문이 에러 메시지에 그대로 노출
- **해결**: 민감 정보 제거, 일반화된 에러 반환

### 8. 테스트에 하드코딩된 자격증명
- **위치**: `internal/infrastructure/proxmox/client_integration_test.go:11-16`
- **문제**: 실제 비밀번호가 코드에 노출
- **해결**: 환경변수로 이동
```go
url = os.Getenv("TEST_PROXMOX_URL")
if url == "" {
    t.Skip("TEST_PROXMOX_URL not set - skipping integration test")
}
```

### 9. Request Body 크기 제한 없음
- **위치**: `internal/api/http/handler/cluster_handler.go:49`
- **문제**: 무제한 JSON 디코딩으로 메모리 고갈 공격 가능
- **해결**: `http.MaxBytesReader` 적용
```go
r.Body = http.MaxBytesReader(w, r.Body, 1*1024*1024) // 1MB
```

### 10. Logger 인터페이스 불일치
- **위치**: 여러 파일
- **문제**: 일부는 `*log.Logger`, 일부는 인터페이스 사용
- **해결**: 모든 컴포넌트에서 Logger 인터페이스 사용

### 11. Graceful Shutdown 없음 ✅ 구현완료
- **위치**: `cmd/proxmoxer-api/main.go:44-46`
- **문제**: 진행 중인 요청이 즉시 종료됨
- **해결**: signal 핸들링과 `server.Shutdown()` 구현

#### 구현 후 추가 리뷰 (2026-01-11)

**Important Issues:**
- 서버 종료 로깅 누락: `http.ErrServerClosed`나 정상 종료 시 로그가 없음
- 타임아웃 시 강제 종료: 30초 후 진행 중 요청이 강제 종료됨 (의도된 동작이지만 인지 필요)

**Suggestions:**
- 종료 시간 로깅: shutdown 소요 시간 기록
```go
startTime := time.Now()
// ... shutdown logic ...
appConfig.Logger.Printf("Server shutdown completed in %v\n", time.Since(startTime))
```
- 타임아웃 설정 가능하게: 30초 하드코딩 → `AppConfig.GracefulShutdownTime`으로 이동
- SIGQUIT 지원: 디버깅용 스택 덤프

### 12. Health Check 실제 상태 미확인
- **위치**: `internal/api/http/routes.go:55-59`
- **문제**: 항상 `healthy` 반환
- **해결**: DB, Proxmox 연결 상태 실제 확인

### 13. 인증 시도 로깅 없음
- **위치**: `internal/infrastructure/proxmox/client.go:76-122`
- **문제**: 인증 성공/실패 감사 불가
- **해결**: 인증 시도, 성공, 실패 로깅 추가

---

## Suggestions

### 14. 에러 경로 테스트 커버리지
- 동시 접근, 컨텍스트 취소, nil 값 등 테스트 추가

### 15. Magic Value 상수화
- 타임아웃 값들을 상수로 정의
```go
const (
    DefaultProxmoxTimeout = 30 * time.Second
    ServerReadHeaderTimeout = 30 * time.Second
)
```

### 16. 중복 검증 로직 통합
- DTO, Service, Entity에 분산된 검증을 Entity.Validate()로 통합

### 17. Godoc 문서화
- 주요 함수에 godoc 주석 추가

### 18. 설정값 시작 시 검증
- `AppConfig.Validate()` 구현

---

## 잘된 점

- Clean Architecture 잘 적용됨
- Repository 패턴 올바른 인터페이스 기반 설계
- `sync.RWMutex`와 defer unlock 적절히 사용
- 모든 I/O에 `context.Context` 사용
- Happy path에 대한 좋은 테스트 커버리지

---

## 우선순위

1. **보안 이슈** - 비밀번호 저장 방식, TLS 검증
2. **에러 처리** - Response 에러 로깅, errors.Is() 사용
3. **운영 준비** - Graceful shutdown, Health check, 인증 로깅
