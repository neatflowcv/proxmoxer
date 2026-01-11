# Proxmox 클러스터 관리 시스템 아키텍처

## 1. 개요

### 1.1 프로젝트 목표
- Go 기반 Proxmox 클러스터 관리 플랫폼 구축
- 다중 클러스터 관리, 가상 머신/컨테이너 관리, 리소스 모니터링, 자동화

### 1.2 대상 사용자
- 시스템 관리자
- 데브옵스 엔지니어

### 1.3 기술적 제약사항
- Go 1.25+
- Proxmox API (REST 기반)
- 모듈식 설계로 향후 확장 용이

### 1.4 우선순위 품질 속성
1. **확장성**: 다중 클러스터, 높은 동시성 처리
2. **유지보수성**: 깔끔한 코드 구조, 테스트 용이성
3. **보안성**: API 인증/인가, 민감 정보 관리
4. **관찰성**: 로깅, 모니터링, 추적

---

## 2. 시스템 아키텍처

### 2.1 아키텍처 개요

```
┌─────────────────────────────────────────────────────────────┐
│                     사용자 인터페이스 계층                      │
│              (CLI, Web API, 향후 WebUI)                      │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                    HTTP/gRPC API 계층                        │
│  (RESTful API, Request/Response Handlers, Validation)        │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                   비즈니스 로직 계층                           │
│         (Use Cases, Domain Services, Application Logic)      │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                    도메인 계층                                │
│         (Domain Entities, Value Objects, Interfaces)         │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                    인프라 계층                                │
│  (API Clients, Database, Cache, External Services)           │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 계층별 책임

| 계층 | 책임 | 예시 |
|------|------|------|
| **API 계층** | HTTP 핸들링, 요청 유효성 검증, 응답 형식화 | HTTP 라우팅, 미들웨어 |
| **비즈니스 로직** | 도메인 규칙 적용, 사용 사례 구현 | VM 생성 로직, 네트워크 구성 |
| **도메인** | 비즈니스 엔티티, 불변식 | Cluster, Node, VirtualMachine |
| **인프라** | 외부 시스템 통신, 영속성 | Proxmox API, 데이터베이스 |

---

## 3. 디렉토리 구조

```
proxmoxer/
├── cmd/                          # 실행 가능한 프로그램
│   ├── proxmoxer/               # 메인 CLI 도구
│   │   └── main.go
│   └── proxmoxer-api/           # API 서버
│       └── main.go
│
├── internal/                     # 내부 패키지 (외부 사용 불가)
│   ├── domain/                  # 도메인 계층
│   │   ├── cluster/
│   │   │   ├── entity.go        # Cluster 엔티티
│   │   │   ├── repository.go    # 인터페이스
│   │   │   └── service.go       # 도메인 서비스
│   │   ├── node/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── virtualmachine/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   ├── service.go
│   │   │   └── status.go        # VM 상태 정의
│   │   └── common/
│   │       ├── errors.go        # 도메인 에러
│   │       └── types.go         # 공통 타입
│   │
│   ├── application/             # 비즈니스 로직 계층
│   │   ├── dto/
│   │   │   ├── cluster.go       # DTO 정의
│   │   │   ├── node.go
│   │   │   └── virtualmachine.go
│   │   ├── services/
│   │   │   ├── cluster_service.go
│   │   │   ├── node_service.go
│   │   │   ├── vm_service.go
│   │   │   └── auth_service.go
│   │   └── usecases/
│   │       ├── cluster_usecases.go
│   │       ├── node_usecases.go
│   │       ├── vm_usecases.go
│   │       └── interfaces.go
│   │
│   ├── infrastructure/          # 인프라 계층
│   │   ├── proxmox/
│   │   │   ├── client.go        # Proxmox API 클라이언트
│   │   │   ├── auth.go          # 인증 로직
│   │   │   ├── models.go        # API 응답 모델
│   │   │   ├── cluster.go       # 클러스터 관련 API
│   │   │   ├── node.go          # 노드 관련 API
│   │   │   └── virtualmachine.go # VM 관련 API
│   │   ├── persistence/
│   │   │   ├── repository/
│   │   │   │   ├── cluster_repo.go
│   │   │   │   ├── node_repo.go
│   │   │   │   └── virtualmachine_repo.go
│   │   │   └── models/
│   │   │       └── db_models.go
│   │   ├── cache/
│   │   │   ├── redis.go
│   │   │   └── memory.go
│   │   └── logging/
│   │       └── logger.go
│   │
│   ├── api/                     # API 계층 (HTTP/gRPC)
│   │   ├── http/
│   │   │   ├── handler/
│   │   │   │   ├── cluster_handler.go
│   │   │   │   ├── node_handler.go
│   │   │   │   └── vm_handler.go
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go
│   │   │   │   ├── logging.go
│   │   │   │   └── errorhandler.go
│   │   │   ├── routes.go
│   │   │   └── response.go
│   │   └── grpc/               # 향후 확장용
│   │       ├── interceptors/
│   │       └── services/
│   │
│   └── config/                 # 설정 관리
│       ├── config.go
│       ├── env.go
│       └── defaults.go
│
├── pkg/                        # 외부에서 재사용 가능한 패키지
│   ├── proxmox/               # Proxmox API SDK
│   │   ├── client.go
│   │   ├── models.go
│   │   └── errors.go
│   └── retry/                 # 재시도 로직
│       └── backoff.go
│
├── test/                       # 테스트 파일
│   ├── integration/
│   │   ├── cluster_test.go
│   │   └── vm_test.go
│   ├── e2e/
│   │   └── workflow_test.go
│   └── fixtures/
│       └── proxmox_responses.go
│
├── docker/                     # Docker 관련 파일
│   ├── Dockerfile
│   └── docker-compose.yml
│
├── scripts/                    # 개발/배포 스크립트
│   ├── setup.sh
│   └── migrate.sh
│
├── docs/                       # 문서
│   ├── ARCHITECTURE.md         # 아키텍처 문서
│   ├── API.md                  # API 명세
│   └── DEPLOYMENT.md           # 배포 가이드
│
├── go.mod
├── go.sum
├── Makefile
├── .env.example
└── .gitignore
```

---

## 4. 주요 컴포넌트 및 역할

### 4.1 도메인 계층 (Domain Layer)

#### 클러스터 (Cluster)
```go
// internal/domain/cluster/entity.go
type Cluster struct {
    ID           string
    Name         string
    Version      string
    Nodes        []*Node
    Status       ClusterStatus
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type ClusterStatus string

const (
    StatusHealthy   ClusterStatus = "healthy"
    StatusDegraded  ClusterStatus = "degraded"
    StatusUnhealthy ClusterStatus = "unhealthy"
)

// 도메인 규칙: 클러스터는 최소 3개 이상의 노드 필요
func (c *Cluster) ValidateMinimumNodes() error {
    if len(c.Nodes) < 3 {
        return ErrInsufficientNodes
    }
    return nil
}
```

#### 노드 (Node)
```go
// internal/domain/node/entity.go
type Node struct {
    ID           string
    ClusterID    string
    Name         string
    Status       NodeStatus
    CPU          *CPUInfo
    Memory       *MemoryInfo
    Storage      *StorageInfo
    UpdatedAt    time.Time
}

type NodeStatus string

const (
    StatusOnline  NodeStatus = "online"
    StatusOffline NodeStatus = "offline"
)
```

#### 가상 머신 (VirtualMachine)
```go
// internal/domain/virtualmachine/entity.go
type VirtualMachine struct {
    ID           string
    ClusterID    string
    NodeID       string
    Name         string
    Type         VMType        // KVM, LXC
    Status       VMStatus
    CPU          int
    Memory       int64         // MB
    Disk         int64         // GB
    Config       map[string]interface{}
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type VMType string
const (
    TypeKVM VMType = "kvm"
    TypeLXC VMType = "lxc"
)

type VMStatus string
const (
    StatusRunning VMStatus = "running"
    StatusStopped VMStatus = "stopped"
)
```

### 4.2 인프라 계층 (Infrastructure Layer)

#### Proxmox API 클라이언트
```go
// internal/infrastructure/proxmox/client.go
type Client struct {
    baseURL    string
    httpClient *http.Client
    token      *Token
    logger     Logger
}

type Token struct {
    Ticket string
    CSRF   string
}

// 연결 풀 관리, 재시도 로직, 타임아웃 처리
func (c *Client) Do(ctx context.Context, method, path string, ...) (*Response, error)

// 인증
func (c *Client) Authenticate(ctx context.Context, username, password string) error

// API 엔드포인트별 메서드
func (c *Client) ListClusters(ctx context.Context) ([]*ClusterInfo, error)
func (c *Client) ListNodes(ctx context.Context, clusterID string) ([]*NodeInfo, error)
func (c *Client) GetVirtualMachine(ctx context.Context, ...) (*VMInfo, error)
```

### 4.3 비즈니스 로직 계층 (Application Layer)

```go
// internal/application/services/cluster_service.go
type ClusterService struct {
    clusterRepo    domain.ClusterRepository
    proxmoxClient  *proxmox.Client
    logger         Logger
}

// 클러스터 정보 조회 및 동기화
func (s *ClusterService) SyncCluster(ctx context.Context, clusterID string) error {
    // 1. Proxmox에서 최신 정보 조회
    clusterInfo, err := s.proxmoxClient.GetCluster(ctx, clusterID)

    // 2. 도메인 엔티티로 변환
    cluster := s.mapToEntity(clusterInfo)

    // 3. 도메인 규칙 검증
    if err := cluster.ValidateMinimumNodes(); err != nil {
        return err
    }

    // 4. 저장소에 저장
    return s.clusterRepo.Save(ctx, cluster)
}
```

### 4.4 API 계층 (API Layer)

```go
// internal/api/http/handler/cluster_handler.go
type ClusterHandler struct {
    clusterService *application.ClusterService
    logger         Logger
}

func (h *ClusterHandler) GetCluster(w http.ResponseWriter, r *http.Request) {
    clusterID := mux.Vars(r)["id"]

    cluster, err := h.clusterService.GetCluster(r.Context(), clusterID)
    if err != nil {
        h.handleError(w, err)
        return
    }

    h.respondJSON(w, http.StatusOK, cluster)
}

func (h *ClusterHandler) ListClusters(w http.ResponseWriter, r *http.Request) {
    clusters, err := h.clusterService.ListClusters(r.Context())
    if err != nil {
        h.handleError(w, err)
        return
    }

    h.respondJSON(w, http.StatusOK, clusters)
}
```

---

## 5. Proxmox API 연동 전략

### 5.1 인증 메커니즘

```go
// internal/infrastructure/proxmox/auth.go
type AuthManager struct {
    baseURL  string
    username string
    password string
    token    *TokenCache
}

// 토큰 캐싱으로 인증 요청 최소화
type TokenCache struct {
    mu        sync.RWMutex
    ticket    string
    csrf      string
    expiresAt time.Time
}

// 자동 갱신 메커니즘
func (am *AuthManager) GetOrRefreshToken(ctx context.Context) (*Token, error) {
    if am.token.IsValid() {
        return am.token.Get(), nil
    }
    return am.RefreshToken(ctx)
}
```

### 5.2 API 호출 전략

#### 재시도 및 회로 차단기
```go
// pkg/retry/backoff.go
type RetryPolicy struct {
    MaxRetries int
    Backoff    BackoffStrategy
}

type BackoffStrategy interface {
    NextDelay(attempt int) time.Duration
}

// Exponential backoff 구현
type ExponentialBackoff struct {
    initialDelay time.Duration
    maxDelay     time.Duration
}

// 회로 차단기로 과도한 요청 방지
type CircuitBreaker struct {
    failures      int
    lastFailTime  time.Time
    state         State  // closed, open, half-open
    maxFailures   int
    resetTimeout  time.Duration
}
```

#### 동시성 제어
```go
// internal/infrastructure/proxmox/client.go
type Client struct {
    semaphore chan struct{}  // 최대 동시 요청 수 제한
    timeout   time.Duration  // 요청 타임아웃
}

func (c *Client) Do(ctx context.Context, ...) (*Response, error) {
    select {
    case c.semaphore <- struct{}{}:
        defer func() { <-c.semaphore }()
    case <-ctx.Done():
        return nil, ctx.Err()
    }

    // 실제 API 호출
}
```

### 5.3 데이터 매핑 전략

```go
// internal/infrastructure/proxmox/mapper.go
type Mapper struct{}

// Proxmox API 응답 -> 도메인 엔티티
func (m *Mapper) ToClusterEntity(apiResp *ProxmoxClusterResponse) *domain.Cluster {
    return &domain.Cluster{
        ID:        apiResp.Name,
        Name:      apiResp.Description,
        Version:   apiResp.Version,
        Status:    mapStatus(apiResp.Nodes),
        UpdatedAt: time.Now(),
    }
}

// 도메인 엔티티 -> API 응답 DTO
func (m *Mapper) ToClusterDTO(cluster *domain.Cluster) *dto.ClusterResponse {
    return &dto.ClusterResponse{
        ID:     cluster.ID,
        Name:   cluster.Name,
        Status: string(cluster.Status),
        Nodes:  len(cluster.Nodes),
    }
}
```

---

## 6. 데이터 모델 설계

### 6.1 데이터베이스 스키마 (PostgreSQL)

```sql
-- 클러스터 테이블
CREATE TABLE clusters (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    version VARCHAR(50),
    status VARCHAR(50) NOT NULL DEFAULT 'healthy',
    api_url VARCHAR(500) NOT NULL,
    auth_token_encrypted BYTEA,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 노드 테이블
CREATE TABLE nodes (
    id UUID PRIMARY KEY,
    cluster_id UUID NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
    proxmox_node_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    cpu_cores INT,
    cpu_usage FLOAT,
    memory_total BIGINT,
    memory_used BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(cluster_id, proxmox_node_id)
);

-- 가상 머신 테이블
CREATE TABLE virtual_machines (
    id UUID PRIMARY KEY,
    cluster_id UUID NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE SET NULL,
    proxmox_vm_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,  -- kvm, lxc
    status VARCHAR(50) NOT NULL,
    cpu_cores INT,
    memory_mb BIGINT,
    disk_gb BIGINT,
    config JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(cluster_id, proxmox_vm_id)
);

-- 메트릭 테이블 (모니터링용)
CREATE TABLE metrics (
    id BIGSERIAL PRIMARY KEY,
    cluster_id UUID NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
    metric_type VARCHAR(100) NOT NULL,  -- cpu, memory, disk, etc.
    value FLOAT NOT NULL,
    recorded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_metrics_cluster_type_time ON metrics(cluster_id, metric_type, recorded_at);

-- 감사 로그 테이블
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(255),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id UUID,
    old_value JSONB,
    new_value JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### 6.2 캐싱 전략

```go
// internal/infrastructure/cache/cache.go
type Cache interface {
    Get(ctx context.Context, key string, dest interface{}) error
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Clear(ctx context.Context) error
}

// 캐시 키 생성 전략
const (
    KeyCluster        = "cluster:%s"
    KeyNodes          = "cluster:%s:nodes"
    KeyVM             = "cluster:%s:vm:%s"
    CacheClusterTTL   = 5 * time.Minute
    CacheNodesTTL     = 1 * time.Minute
)
```

---

## 7. 보안 아키텍처

### 7.1 인증/인가

```go
// internal/application/services/auth_service.go
type AuthService struct {
    proxmoxClient *proxmox.Client
    sessionCache  *cache.SessionCache
}

// API 토큰 방식
type APIToken struct {
    UserID    string
    TokenID   string
    Secret    string
    CreatedAt time.Time
    ExpiresAt time.Time
}

// JWT를 통한 세션 관리
type SessionManager struct {
    jwtSecret string
    duration  time.Duration
}

func (sm *SessionManager) CreateSession(user *User) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": user.ID,
        "exp":     time.Now().Add(sm.duration).Unix(),
    })
    return token.SignedString([]byte(sm.jwtSecret))
}
```

### 7.2 권한 관리 (RBAC)

```go
// internal/domain/common/rbac.go
type Role string

const (
    RoleAdmin      Role = "admin"
    RoleOperator   Role = "operator"
    RoleViewer     Role = "viewer"
)

type Permission string

const (
    PermClusterRead   Permission = "cluster:read"
    PermClusterWrite  Permission = "cluster:write"
    PermVMCreate      Permission = "vm:create"
    PermVMDelete      Permission = "vm:delete"
)

// 권한 검증 미들웨어
func (m *AuthMiddleware) CheckPermission(permission Permission) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user, err := m.GetUser(r)
            if err != nil || !user.HasPermission(permission) {
                http.Error(w, "Unauthorized", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

### 7.3 민감 정보 관리

- 환경 변수를 통한 설정 관리
- 암호화된 토큰 저장
- HashiCorp Vault 또는 AWS Secrets Manager 연동 가능

---

## 8. 확장성 고려사항

### 8.1 수평 확장 전략

```
┌─────────────────────────────────────────────────┐
│         Load Balancer (Nginx/HAProxy)           │
└────────────┬────────────────┬───────────────────┘
             │                │
    ┌────────▼─────┐  ┌───────▼──────┐
    │ API Server 1 │  │ API Server 2 │  ...
    └────────┬─────┘  └───────┬──────┘
             │                │
    ┌────────▼────────────────▼──────┐
    │   Shared PostgreSQL Database   │
    │   (Primary/Standby Replication)│
    └────────────────────────────────┘
             │
    ┌────────▼──────────────┐
    │   Redis Cache Cluster │
    └───────────────────────┘
             │
    ┌────────▼──────────────────────────┐
    │  Message Queue (RabbitMQ/Kafka)   │
    │  (for async task processing)      │
    └───────────────────────────────────┘
```

### 8.2 마이크로서비스로의 진화 경로

```
Phase 1: 모놀리식 (현재)
├─ Single binary
├─ Shared database
└─ In-process communication

Phase 2: 모듈화된 모놀리식
├─ Clear domain boundaries
├─ Repository pattern for abstraction
└─ Event-driven internal communication

Phase 3: 분산 시스템
├─ ClusterService (별도 서비스)
├─ NodeService (별도 서비스)
├─ VMService (별도 서비스)
├─ Each with own database
└─ Event bus for inter-service communication
```

### 8.3 플러그인 시스템

```go
// pkg/plugin/plugin.go
type Plugin interface {
    Name() string
    Version() string
    Init(ctx context.Context) error
    Close() error
}

type NotificationPlugin interface {
    Plugin
    Send(ctx context.Context, notification *Notification) error
}

// Slack, Email, PagerDuty 플러그인 구현 가능
```

---

## 9. Go 언어 관용 패턴

### 9.1 에러 처리

```go
// 명시적 에러 처리 (Go way)
func (s *ClusterService) GetCluster(ctx context.Context, id string) (*domain.Cluster, error) {
    if id == "" {
        return nil, fmt.Errorf("cluster id cannot be empty: %w", ErrInvalidInput)
    }

    cluster, err := s.clusterRepo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to find cluster: %w", err)
    }

    return cluster, nil
}

// errors.Is 및 errors.As 활용
if errors.Is(err, ErrNotFound) {
    http.Error(w, "Not found", http.StatusNotFound)
}
```

### 9.2 컨텍스트 활용

```go
// 모든 I/O 작업에서 컨텍스트 사용
func (c *Client) GetCluster(ctx context.Context, id string) (*Cluster, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", c.url+"/clusters/"+id, nil)
    // ...
}

// 컨텍스트 타임아웃 설정
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

### 9.3 인터페이스 기반 설계

```go
// 작은 인터페이스 (Go idiom)
type ClusterRepository interface {
    Save(ctx context.Context, cluster *Cluster) error
    FindByID(ctx context.Context, id string) (*Cluster, error)
    Delete(ctx context.Context, id string) error
}

// 구현체는 인터페이스를 임포트하지 않음 (implicit satisfaction)
```

### 9.4 의존성 주입

```go
// 생성자를 통한 DI
func NewClusterService(
    repo domain.ClusterRepository,
    client *proxmox.Client,
    logger Logger,
) *ClusterService {
    return &ClusterService{
        repo:   repo,
        client: client,
        logger: logger,
    }
}
```

---

## 10. 개발 환경 설정

### 10.1 Makefile

```makefile
.PHONY: help build test clean run

help:
	@echo "Available targets:"
	@echo "  make build      - Build the application"
	@echo "  make test       - Run tests"
	@echo "  make lint       - Run linter"
	@echo "  make run        - Run the application"

build:
	go build -o bin/proxmoxer ./cmd/proxmoxer
	go build -o bin/proxmoxer-api ./cmd/proxmoxer-api

test:
	go test -v -cover ./...

lint:
	golangci-lint run ./...

run:
	go run ./cmd/proxmoxer-api

docker-build:
	docker build -t proxmoxer:latest -f docker/Dockerfile .
```

### 10.2 주요 의존성

```go
// go.mod
module github.com/your-org/proxmoxer

go 1.25

require (
    github.com/gorilla/mux v1.8.0
    github.com/lib/pq v1.10.0
    github.com/joho/godotenv v1.5.1
    github.com/google/uuid v1.3.0
    go.uber.org/zap v1.27.0
    github.com/golang-migrate/migrate/v4 v4.16.0
)
```

---

## 11. 구현 로드맵

### Phase 1: 기초 구축
- [ ] 프로젝트 구조 및 기본 의존성 설정
- [ ] Proxmox API 클라이언트 구현
- [ ] 도메인 엔티티 정의
- [ ] PostgreSQL 스키마 설계 및 마이그레이션

### Phase 2: 핵심 기능 개발
- [ ] Repository 패턴 구현
- [ ] 비즈니스 로직 계층 개발
- [ ] HTTP API 엔드포인트 구현
- [ ] 기본 인증/인가 시스템

### Phase 3: 안정성 및 운영성
- [ ] 로깅 및 모니터링 시스템
- [ ] 에러 처리 및 재시도 로직
- [ ] 단위 테스트 및 통합 테스트
- [ ] API 문서화

### Phase 4: 고급 기능
- [ ] 이벤트 기반 아키텍처
- [ ] 작업 큐 및 비동기 처리
- [ ] 대시보드 및 모니터링
- [ ] 플러그인 시스템

---

## 12. 위험 요소 및 완화 전략

| 위험 | 영향 | 완화 전략 |
|-----|------|---------|
| Proxmox API 변경 | API 호환성 깨짐 | API 버전 관리, 어댑터 패턴 |
| 높은 동시 요청 | 성능 저하 | 연결 풀, Rate limiting, 캐싱 |
| 데이터 일관성 | 상태 불일치 | 트랜잭션, 이벤트 소싱, 동기화 메커니즘 |
| 보안 위반 | 클러스터 제어 탈취 | RBAC, 암호화, 감사 로그, 네트워크 격리 |
| 확장성 한계 | 성능 저하 | 마이크로서비스 설계 준비, 캐싱 전략 |
