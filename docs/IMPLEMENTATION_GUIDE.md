# Cluster Management 기능 구현 가이드

## 개요

이 문서는 클러스터 관리 기능 (Register, Deregister, List)의 구현 상세사항과 확장 방법을 설명합니다.

---

## 1. 프로젝트 구조

```
internal/
├── domain/                          # 도메인 계층
│   ├── common/
│   │   └── errors.go               # 도메인 에러 정의
│   └── cluster/
│       ├── entity.go               # Cluster 엔티티
│       └── repository.go           # Repository 인터페이스
│
├── application/                     # 애플리케이션 계층
│   ├── dto/
│   │   └── cluster.go              # DTO 정의
│   └── services/
│       ├── cluster_service.go      # 비즈니스 로직
│       └── cluster_service_test.go # 단위 테스트
│
├── infrastructure/                  # 인프라 계층
│   ├── proxmox/
│   │   └── client.go              # Proxmox API 클라이언트
│   └── persistence/
│       ├── memory_repository.go    # 메모리 저장소
│       └── memory_repository_test.go
│
├── api/
│   └── http/
│       ├── routes.go               # 라우팅 설정
│       └── handler/
│           ├── cluster_handler.go  # HTTP 핸들러
│           └── response.go         # 응답 처리
│
└── config/
    └── app.go                      # 의존성 주입
```

---

## 2. 각 계층별 구현

### 2.1 Domain Layer (internal/domain/)

#### 2.1.1 에러 정의 (`domain/common/errors.go`)

도메인 레이어에서 발생할 수 있는 모든 에러를 정의합니다.

```go
var (
    ErrClusterNotFound       = errors.New("cluster not found")
    ErrClusterAlreadyExists  = errors.New("cluster already exists")
    ErrInvalidClusterID      = errors.New("invalid cluster id")
    ErrInvalidCredentials    = errors.New("invalid credentials")
    ErrAuthenticationFailed  = errors.New("authentication failed")
)
```

**특징:**
- 재사용 가능한 에러 변수
- `errors.Is()` 로 에러 타입 확인 가능
- 계층 간 에러 체이닝 용이

#### 2.1.2 엔티티 정의 (`domain/cluster/entity.go`)

Cluster 엔티티는 비즈니스 규칙을 캡슐화합니다.

```go
type Cluster struct {
    ID              string          // UUID
    Name            string          // 클러스터 이름
    APIEndpoint     string          // API URL
    Username        string          // 사용자명
    Password        string          // 비밀번호
    Status          ClusterStatus   // 상태
    ProxmoxVersion  string          // 버전
    NodeCount       int             // 노드 수
    CreatedAt       time.Time       // 생성 시간
    UpdatedAt       time.Time       // 수정 시간
}
```

**메서드:**
- `NewCluster()`: 팩토리 메서드
- `UpdateStatus()`: 상태 변경
- `UpdateNodeCount()`: 노드 수 업데이트
- `Validate()`: 유효성 검증

#### 2.1.3 Repository 인터페이스 (`domain/cluster/repository.go`)

저장소의 인터페이스를 정의하여 구현 세부사항으로부터 분리합니다.

```go
type Repository interface {
    Save(ctx context.Context, cluster *Cluster) error
    FindByID(ctx context.Context, id string) (*Cluster, error)
    FindByName(ctx context.Context, name string) (*Cluster, error)
    List(ctx context.Context) ([]*Cluster, error)
    Delete(ctx context.Context, id string) error
    Exists(ctx context.Context, id string) (bool, error)
}
```

**설계 원칙:**
- 작은 인터페이스 (Go idiom)
- 구현체는 인터페이스를 import 하지 않음 (implicit satisfaction)
- 모든 메서드는 context 파라미터 포함

---

### 2.2 Application Layer (internal/application/)

#### 2.2.1 DTO 정의 (`application/dto/cluster.go`)

계층 간 데이터 전달을 위한 DTO를 정의합니다.

```go
// 요청 DTO
type RegisterClusterRequest struct {
    Name        string `json:"name" binding:"required,max=255"`
    APIEndpoint string `json:"api_endpoint" binding:"required,url"`
    Username    string `json:"username" binding:"required,max=255"`
    Password    string `json:"password" binding:"required,min=1"`
}

// 응답 DTO
type ClusterResponse struct {
    ID             string    `json:"id"`
    Name           string    `json:"name"`
    APIEndpoint    string    `json:"api_endpoint"`
    Status         string    `json:"status"`
    ProxmoxVersion string    `json:"proxmox_version"`
    NodeCount      int       `json:"node_count"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}
```

**특징:**
- JSON 태그로 직렬화 제어
- `binding` 태그로 유효성 검증 규칙 정의

#### 2.2.2 비즈니스 로직 (`application/services/cluster_service.go`)

클러스터 관리의 비즈니스 로직을 구현합니다.

```go
type ClusterService struct {
    clusterRepo   cluster.Repository
    proxmoxClient ProxmoxClient  // 인터페이스 기반 의존성
    logger        Logger
}
```

**핵심 메서드:**

1. **RegisterCluster()**
   ```
   1. 요청 유효성 검증
   2. 중복 이름 확인
   3. Proxmox API 인증
   4. 버전/노드 정보 조회 (선택)
   5. Cluster 엔티티 생성
   6. Repository에 저장
   ```

2. **DeregisterCluster()**
   ```
   1. 클러스터 ID 유효성 검증
   2. 존재 여부 확인
   3. Repository에서 삭제
   ```

3. **ListClusters()**
   ```
   1. Repository에서 모든 클러스터 조회
   2. DTO로 변환
   3. 응답 반환
   ```

**의존성 주입:**

```go
func NewClusterService(
    repo cluster.Repository,
    client ProxmoxClient,
    logger Logger,
) *ClusterService {
    // 닐 체크 및 기본값 설정
    if logger == nil {
        logger = NewSimpleLogger(log.Default())
    }

    return &ClusterService{
        clusterRepo:   repo,
        proxmoxClient: client,
        logger:        logger,
    }
}
```

**인터페이스 기반 설계:**

```go
// ProxmoxClient 인터페이스
type ProxmoxClient interface {
    Authenticate(ctx context.Context, username, password string) (ticket, csrf string, err error)
    GetVersion(ctx context.Context, ticket string) (version string, err error)
    GetNodeCount(ctx context.Context, ticket string) (count int, err error)
}
```

이를 통해:
- 테스트 시 Mock 클라이언트 주입 가능
- Proxmox API 구현 변경 시 인터페이스 유지

---

### 2.3 Infrastructure Layer (internal/infrastructure/)

#### 2.3.1 In-Memory Repository (`infrastructure/persistence/memory_repository.go`)

메모리 기반 저장소 구현입니다.

```go
type MemoryRepository struct {
    mu       sync.RWMutex
    clusters map[string]*cluster.Cluster
}
```

**특징:**
- 스레드 안전성 (sync.RWMutex)
- 맵 기반 O(1) 검색 성능
- MVP 개발용 빠른 피드백

**주요 메서드:**
```go
func (r *MemoryRepository) Save(ctx context.Context, c *cluster.Cluster) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if err := c.Validate(); err != nil {
        return fmt.Errorf("invalid cluster: %w", err)
    }

    r.clusters[c.ID] = c
    return nil
}
```

#### 2.3.2 Proxmox API 클라이언트 (`infrastructure/proxmox/client.go`)

Proxmox REST API와의 통신을 담당합니다.

```go
type Client struct {
    baseURL    string
    httpClient *http.Client
    timeout    time.Duration
}
```

**인증 흐름:**

```
POST /api2/json/access/ticket
├─ Body: username=xxx&password=xxx
└─ Response:
   {
     "data": {
       "ticket": "PVE:...",
       "csrf": "..."
     }
   }
```

**Proxmox 인증 메서드:**

```go
func (c *Client) Authenticate(ctx context.Context, username, password string) (ticket string, csrf string, err error) {
    authURL := fmt.Sprintf("%s/api2/json/access/ticket", c.baseURL)

    // 1. 요청 생성
    data := url.Values{}
    data.Set("username", username)
    data.Set("password", password)

    req, err := http.NewRequestWithContext(ctx, "POST", authURL, bytes.NewBufferString(data.Encode()))

    // 2. 요청 실행
    resp, err := c.httpClient.Do(req)

    // 3. 응답 파싱
    var authResp AuthenticateResponse
    json.Unmarshal(body, &authResp)

    return authResp.Data.Ticket, authResp.Data.CSRF, nil
}
```

---

### 2.4 API Layer (internal/api/)

#### 2.4.1 응답 처리 (`api/http/handler/response.go`)

HTTP 응답 작성을 통합합니다.

```go
type ResponseWriter struct {
    logger *log.Logger
}

func (rw *ResponseWriter) WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    return json.NewEncoder(w).Encode(data)
}

func (rw *ResponseWriter) HandleError(w http.ResponseWriter, err error) {
    // 에러 타입에 따른 상태 코드 매핑
    if errors.Is(err, common.ErrClusterNotFound) {
        statusCode = http.StatusNotFound
    } else if errors.Is(err, common.ErrClusterAlreadyExists) {
        statusCode = http.StatusConflict
    }

    rw.WriteError(w, statusCode, message)
}
```

#### 2.4.2 HTTP 핸들러 (`api/http/handler/cluster_handler.go`)

HTTP 요청을 처리합니다.

```go
type ClusterHandler struct {
    clusterService *services.ClusterService
    responseWriter *ResponseWriter
    logger         *log.Logger
}

func (h *ClusterHandler) RegisterCluster(w http.ResponseWriter, r *http.Request) {
    // 1. JSON 파싱
    var req dto.RegisterClusterRequest
    json.NewDecoder(r.Body).Decode(&req)

    // 2. 서비스 호출
    response, err := h.clusterService.RegisterCluster(r.Context(), &req)
    if err != nil {
        h.responseWriter.HandleError(w, err)
        return
    }

    // 3. 응답 작성
    h.responseWriter.WriteJSON(w, http.StatusCreated, response)
}
```

#### 2.4.3 라우팅 설정 (`api/http/routes.go`)

HTTP 라우트를 설정합니다.

```go
type Router struct {
    mux            *http.ServeMux
    clusterHandler *handler.ClusterHandler
    logger         *log.Logger
}

func (r *Router) setupRoutes() {
    r.mux.HandleFunc("POST /api/v1/clusters", r.clusterHandler.RegisterCluster)
    r.mux.HandleFunc("GET /api/v1/clusters", r.clusterHandler.ListClusters)
    r.mux.HandleFunc("GET /api/v1/clusters/{id}", r.clusterHandler.GetCluster)
    r.mux.HandleFunc("DELETE /api/v1/clusters/{id}", r.clusterHandler.DeregisterCluster)
    r.mux.HandleFunc("GET /health", healthCheckHandler)
}
```

---

### 2.5 의존성 주입 (`config/app.go`)

모든 컴포넌트를 초기화하고 연결합니다.

```go
func InitializeApp(config *AppConfig) (*http.Router, error) {
    // 1. Repository 초기화
    clusterRepo := persistence.NewMemoryRepository()

    // 2. 외부 API 클라이언트 초기화
    proxmoxClient := proxmox.NewClient("https://pve.local", config.ProxmoxTimeout)

    // 3. Service 초기화
    clusterService := services.NewClusterService(clusterRepo, proxmoxClient, nil)

    // 4. HTTP Router 초기화
    router := http.NewRouter(clusterService, config.Logger)

    return router, nil
}
```

---

## 3. 실행 방법

### 3.1 API 서버 시작

```bash
go run ./cmd/proxmoxer-api
```

출력:
```
==============================================
Proxmoxer API Server Starting
==============================================
Version: 0.1.0-mvp
Server Port: 8080
Initializing application components...
✓ Cluster repository initialized (in-memory)
✓ Proxmox client initialized
✓ Cluster service initialized
✓ HTTP router initialized
Application initialization completed successfully!
==============================================
Starting server on :8080
==============================================
```

### 3.2 API 호출

```bash
# 클러스터 등록
curl -X POST http://localhost:8080/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "prod-cluster",
    "api_endpoint": "https://pve.example.com:8006",
    "username": "root@pam",
    "password": "password"
  }'

# 모든 클러스터 조회
curl -X GET http://localhost:8080/api/v1/clusters

# 특정 클러스터 조회
curl -X GET http://localhost:8080/api/v1/clusters/{id}

# 클러스터 삭제
curl -X DELETE http://localhost:8080/api/v1/clusters/{id}
```

---

## 4. 테스트 실행

### 4.1 모든 단위 테스트

```bash
go test -v -cover ./internal/...
```

출력:
```
=== RUN   TestRegisterCluster_Success
--- PASS: TestRegisterCluster_Success (0.00s)

=== RUN   TestMemoryRepository_Save
--- PASS: TestMemoryRepository_Save (0.00s)

PASS
coverage: 79.1% of statements
```

### 4.2 특정 테스트만 실행

```bash
go test -v -run TestRegisterCluster ./internal/application/services

go test -v -run TestMemoryRepository ./internal/infrastructure/persistence
```

### 4.3 커버리지 보고서

```bash
go test -cover ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## 5. 향후 확장

### 5.1 데이터베이스 마이그레이션

1. **PostgreSQL Repository 구현**
   ```go
   // internal/infrastructure/persistence/postgres_repository.go
   type PostgresRepository struct {
       db *sql.DB
   }
   ```

2. **마이그레이션 파일 생성**
   ```sql
   -- migrations/001_create_clusters.up.sql
   CREATE TABLE clusters (
       id UUID PRIMARY KEY,
       name VARCHAR(255) UNIQUE NOT NULL,
       api_endpoint VARCHAR(500) NOT NULL,
       username VARCHAR(255) NOT NULL,
       password BYTEA NOT NULL,
       status VARCHAR(50) NOT NULL,
       proxmox_version VARCHAR(50),
       node_count INT,
       created_at TIMESTAMP NOT NULL,
       updated_at TIMESTAMP NOT NULL
   );
   ```

3. **Service 초기화 코드 변경**
   ```go
   // config/app.go
   var clusterRepo cluster.Repository
   if config.UsePostgres {
       clusterRepo = persistence.NewPostgresRepository(db)
   } else {
       clusterRepo = persistence.NewMemoryRepository()
   }
   ```

### 5.2 토큰 캐싱

Proxmox 인증 토큰을 캐싱하여 인증 횟수를 줄입니다.

```go
type TokenCache struct {
    mu        sync.RWMutex
    ticket    string
    csrf      string
    expiresAt time.Time
}

func (c *Client) GetOrRefreshToken(ctx context.Context, username, password string) (*Token, error) {
    if c.token.IsValid() {
        return c.token.Get(), nil
    }
    return c.RefreshToken(ctx, username, password)
}
```

### 5.3 비밀번호 암호화

저장된 비밀번호를 암호화합니다.

```go
import "golang.org/x/crypto/bcrypt"

// 비밀번호 해싱
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// 비밀번호 검증
err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
```

### 5.4 감사 로깅

모든 작업을 로깅합니다.

```go
type AuditLog struct {
    UserID      string
    Action      string  // "register", "deregister", "list"
    ResourceID  string  // cluster ID
    Status      string  // "success", "failure"
    Error       string
    CreatedAt   time.Time
}

func (s *ClusterService) logAudit(action, resourceID, status, errMsg string) {
    // 데이터베이스에 저장
}
```

---

## 6. 문제 해결

### 6.1 Proxmox 연결 실패

**증상:** `502 Bad Gateway` 에러

**원인:**
- Proxmox API 엔드포인트가 잘못됨
- Proxmox 서버가 응답하지 않음
- 네트워크 연결 문제

**해결:**
```bash
# Proxmox API 확인
curl -k https://pve.example.com:8006/api2/json/version

# 방화벽 확인
telnet pve.example.com 8006
```

### 6.2 인증 실패

**증상:** `401 Unauthorized` 에러

**원인:**
- 잘못된 사용자명/비밀번호
- 사용자 권한 부족

**해결:**
```bash
# Proxmox 직접 인증 테스트
curl -k -X POST https://pve.example.com:8006/api2/json/access/ticket \
  -d "username=root@pam&password=yourpassword"
```

### 6.3 중복 클러스터 이름

**증상:** `409 Conflict` 에러

**원인:** 동일한 이름의 클러스터가 이미 등록됨

**해결:**
```bash
# 다른 이름으로 등록
curl -X POST http://localhost:8080/api/v1/clusters \
  -d '{
    "name": "prod-cluster-2",
    ...
  }'

# 또는 기존 클러스터 삭제
curl -X DELETE http://localhost:8080/api/v1/clusters/{id}
```

---

## 7. 성능 최적화

### 7.1 Connection Pooling

```go
client := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        MaxConnsPerHost:     10,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 30 * time.Second,
}
```

### 7.2 Rate Limiting

```go
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(10, 1)  // 10 req/sec

func (c *Client) Do(ctx context.Context, ...) {
    if !limiter.Allow() {
        return fmt.Errorf("rate limit exceeded")
    }
    // API 호출
}
```

### 7.3 Context Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := service.RegisterCluster(ctx, req)
```

---

## 8. 보안 체크리스트

- [ ] HTTPS/TLS 암호화 (운영 환경)
- [ ] 요청 유효성 검증
- [ ] SQL Injection 방지 (Parameterized queries)
- [ ] XSS 방지 (JSON encoding)
- [ ] CSRF 토큰 (상태 변경 작업)
- [ ] Rate Limiting
- [ ] 로깅 및 모니터링
- [ ] 비밀번호 암호화 저장
- [ ] API 토큰 암호화

---

## 참고 자료

- [Go 에러 처리](https://go.dev/blog/error-handling-and-go)
- [Go Context 패턴](https://go.dev/blog/context)
- [Proxmox API 문서](https://pve.proxmox.com/pve-docs/api-viewer/)
- [Repository 패턴](https://martinfowler.com/eaaCatalog/repository.html)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
