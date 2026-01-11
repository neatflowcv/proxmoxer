# 클러스터 관리 기능 설계 문서

## 개요

이 문서는 Proxmoxer의 클러스터 관리 기능 (Register, Deregister, List)의 상세 설계를 설명합니다.

---

## 1. 요구사항 분석

### 1.1 기능 요구사항

| 기능 | 설명 | 입력 | 출력 |
|------|------|------|------|
| **Register Cluster** | Proxmox 클러스터 등록 | name, api_endpoint, username, password | cluster_id, status |
| **Deregister Cluster** | 등록된 클러스터 제거 | cluster_id | success/error |
| **List Clusters** | 등록된 모든 클러스터 조회 | - | clusters[], total |

### 1.2 비기능 요구사항

| 속성 | 요구사항 |
|------|---------|
| **성능** | 클러스터 목록 조회: < 100ms (메모리 저장소) |
| **보안** | Proxmox API 인증 검증, 비밀번호 메모리 저장 (MVP) |
| **신뢰성** | Proxmox API 연결 실패 시 명확한 에러 메시지 |
| **확장성** | 향후 데이터베이스 교체 용이한 설계 |
| **테스트성** | 단위 테스트 가능한 구조 |

---

## 2. 시스템 아키텍처

### 2.1 전체 흐름도

```
┌─────────────────────────────────────────────────────────────┐
│                         HTTP Client                          │
│                    (REST API Consumer)                       │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           │ HTTP Request
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    API Layer (HTTP)                          │
│  - Cluster Handler                                           │
│  - Request validation, Response formatting                  │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           │ Use Case Call
                           ▼
┌─────────────────────────────────────────────────────────────┐
│               Application Layer (Services)                   │
│  - ClusterService                                            │
│  - Business logic, DTO transformation                        │
│  - Orchestration                                             │
└──────────────────────────┬──────────────────────────────────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
        ▼                  ▼                  ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│Domain Layer  │  │Proxmox Client│  │Repository    │
│(Entities)    │  │(Auth Check)  │  │(Persistence) │
└──────────────┘  └──────────────┘  └──────────────┘
        │                  │                  │
        └──────────────────┴──────────────────┘
```

### 2.2 계층별 책임

#### API Layer (`internal/api/http/`)
- HTTP 요청 수신
- Request/Response 직렬화 및 검증
- HTTP 상태 코드 설정
- 에러 메시지 포맷팅

#### Application Layer (`internal/application/`)
- 비즈니스 로직 구현
- 도메인 엔티티 생성/조회
- DTO 변환
- 서비스 간 조율

#### Domain Layer (`internal/domain/`)
- Cluster 엔티티 정의
- Repository 인터페이스 정의
- 도메인 규칙 및 제약 조건
- 에러 정의

#### Infrastructure Layer (`internal/infrastructure/`)
- 저장소 구현 (메모리)
- Proxmox API 클라이언트
- 외부 시스템과의 통신

---

## 3. 상세 설계

### 3.1 Domain Layer

#### 3.1.1 Cluster 엔티티

```go
type Cluster struct {
    ID              string          // 고유 식별자 (UUID)
    Name            string          // 클러스터 이름
    APIEndpoint     string          // Proxmox API URL
    Username        string          // Proxmox 사용자명
    Password        string          // Proxmox 비밀번호 (MVP: 메모리 저장)
    Status          ClusterStatus   // 상태 (healthy/degraded/unhealthy/unknown)
    ProxmoxVersion  string          // Proxmox 버전
    NodeCount       int             // 노드 개수
    CreatedAt       time.Time       // 생성 시간
    UpdatedAt       time.Time       // 수정 시간
}
```

#### 3.1.2 Repository 인터페이스

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
- Context 기반 취소 지원
- 명시적 에러 처리

### 3.2 Application Layer

#### 3.2.1 ClusterService

**메서드:**

```go
func (s *ClusterService) RegisterCluster(
    ctx context.Context,
    req *dto.RegisterClusterRequest,
) (*dto.ClusterResponse, error)

func (s *ClusterService) DeregisterCluster(
    ctx context.Context,
    clusterID string,
) error

func (s *ClusterService) ListClusters(
    ctx context.Context,
) (*dto.ListClustersResponse, error)

func (s *ClusterService) GetCluster(
    ctx context.Context,
    clusterID string,
) (*dto.ClusterResponse, error)
```

#### 3.2.2 DTO 정의

**RegisterClusterRequest:**
```go
type RegisterClusterRequest struct {
    Name        string `json:"name" binding:"required,max=255"`
    APIEndpoint string `json:"api_endpoint" binding:"required,url"`
    Username    string `json:"username" binding:"required,max=255"`
    Password    string `json:"password" binding:"required,min=1"`
}
```

**ClusterResponse:**
```go
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

### 3.3 Infrastructure Layer

#### 3.3.1 In-Memory Repository

**특징:**
- 스레드 안전 (sync.RWMutex)
- MVP용 빠른 개발
- 향후 데이터베이스로 교체 용이

**구현:**
```go
type MemoryRepository struct {
    mu       sync.RWMutex
    clusters map[string]*cluster.Cluster
}
```

#### 3.3.2 Proxmox API 클라이언트

**주요 메서드:**

```go
// 인증 (자격증명 검증)
func (c *Client) Authenticate(
    ctx context.Context,
    username string,
    password string,
) (ticket string, csrf string, err error)

// Proxmox 버전 조회
func (c *Client) GetVersion(
    ctx context.Context,
    ticket string,
) (version string, err error)

// 노드 개수 조회
func (c *Client) GetNodeCount(
    ctx context.Context,
    ticket string,
) (count int, err error)
```

**Proxmox API 인증 흐름:**

1. **자격증명으로 인증 토큰 요청**
   ```
   POST /api2/json/access/ticket
   Body: username=xxx&password=xxx
   ```

2. **응답에서 ticket 추출**
   ```json
   {
     "data": {
       "ticket": "PVE:...",
       "csrf": "..."
     }
   }
   ```

3. **이후 요청에 인증 헤더 추가**
   ```
   Authorization: PVEAPIToken={ticket}
   ```

### 3.4 API Layer

#### 3.4.1 HTTP 엔드포인트

| Method | Path | 설명 |
|--------|------|------|
| POST | `/api/v1/clusters` | 클러스터 등록 |
| GET | `/api/v1/clusters` | 모든 클러스터 조회 |
| GET | `/api/v1/clusters/{id}` | 특정 클러스터 조회 |
| DELETE | `/api/v1/clusters/{id}` | 클러스터 제거 |

#### 3.4.2 HTTP 응답 형식

**성공 응답 (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "prod-cluster",
  "api_endpoint": "https://pve.example.com:8006",
  "status": "healthy",
  "proxmox_version": "7.4-1",
  "node_count": 3,
  "created_at": "2024-01-11T10:30:00Z",
  "updated_at": "2024-01-11T10:30:00Z"
}
```

**에러 응답 (400/401/404/409):**
```json
{
  "code": "BadRequest",
  "message": "Cluster name is required",
  "details": null
}
```

#### 3.4.3 상태 코드 매핑

| 상황 | 상태 코드 | 메시지 |
|------|---------|--------|
| 등록 성공 | 201 | Created |
| 조회 성공 | 200 | OK |
| 제거 성공 | 204 | No Content |
| 잘못된 요청 | 400 | Bad Request |
| 인증 실패 | 401 | Unauthorized |
| 클러스터 미존재 | 404 | Not Found |
| 중복 이름 | 409 | Conflict |
| Proxmox 연결 실패 | 502 | Bad Gateway |
| 서버 에러 | 500 | Internal Server Error |

---

## 4. 데이터 흐름

### 4.1 Register Cluster Flow

```
1. HTTP POST /api/v1/clusters
   ├─ Request Body: {name, api_endpoint, username, password}
   │
2. ClusterHandler.RegisterCluster()
   ├─ JSON 파싱 및 Request 검증
   │
3. ClusterService.RegisterCluster()
   ├─ Request 유효성 검증
   ├─ 중복 이름 확인 (Repository.FindByName)
   ├─ Proxmox API 인증 (Proxmox Client.Authenticate)
   ├─ Proxmox 버전 조회 (선택사항)
   ├─ 노드 개수 조회 (선택사항)
   ├─ Cluster 엔티티 생성 (UUID 생성)
   ├─ Cluster 저장 (Repository.Save)
   │
4. ClusterResponse 반환
   └─ HTTP 201 Created
```

### 4.2 Deregister Cluster Flow

```
1. HTTP DELETE /api/v1/clusters/{id}
   │
2. ClusterHandler.DeregisterCluster()
   ├─ URL에서 클러스터 ID 추출
   │
3. ClusterService.DeregisterCluster()
   ├─ 클러스터 ID 유효성 검증
   ├─ 클러스터 존재 여부 확인 (Repository.FindByID)
   ├─ 클러스터 삭제 (Repository.Delete)
   │
4. HTTP 204 No Content
```

### 4.3 List Clusters Flow

```
1. HTTP GET /api/v1/clusters
   │
2. ClusterHandler.ListClusters()
   │
3. ClusterService.ListClusters()
   ├─ 모든 클러스터 조회 (Repository.List)
   ├─ 각 클러스터를 DTO로 변환
   │
4. ListClustersResponse 반환
   └─ HTTP 200 OK
```

---

## 5. 에러 처리 전략

### 5.1 계층별 에러 처리

**Domain Layer:**
- 도메인 규칙 위반 시 명시적 에러 반환
- 예: `ErrClusterNotFound`, `ErrInvalidClusterID`

**Application Layer:**
- 도메인 에러를 래핑하여 컨텍스트 추가
- 예: `fmt.Errorf("failed to find cluster: %w", err)`

**API Layer:**
- 도메인 에러를 HTTP 상태 코드로 매핑
- `errors.Is()` 사용하여 에러 타입 확인

### 5.2 Proxmox API 에러 처리

**인증 실패:**
```go
// Proxmox 응답: HTTP 401 or 400
return common.ErrAuthenticationFailed
```

**연결 실패:**
```go
// Network error or timeout
return common.ErrProxmoxConnectionFailed
```

---

## 6. 보안 고려사항

### 6.1 비밀번호 관리 (MVP)

**현재 (MVP):**
- 메모리에 평문 저장
- HTTPS 필수 (운영 환경)

**향후 개선:**
- 암호화된 저장소 (데이터베이스)
- HashiCorp Vault 통합
- 토큰 기반 인증

### 6.2 API 보안

**필수 사항:**
- HTTPS/TLS 암호화
- 요청 유효성 검증
- 레이트 리미팅
- 감사 로깅

---

## 7. 테스트 전략

### 7.1 단위 테스트

**ClusterService 테스트:**
```
✓ TestRegisterCluster_Success
✓ TestRegisterCluster_DuplicateName
✓ TestRegisterCluster_InvalidRequest
✓ TestListClusters
✓ TestDeregisterCluster_Success
✓ TestDeregisterCluster_NotFound
✓ TestGetCluster
✓ TestAuthenticationFailure
```

**MemoryRepository 테스트:**
```
✓ TestMemoryRepository_Save
✓ TestMemoryRepository_FindByID
✓ TestMemoryRepository_FindByName
✓ TestMemoryRepository_List
✓ TestMemoryRepository_Delete
✓ TestMemoryRepository_Exists
```

### 7.2 통합 테스트

```go
// internal/api/http/handler/cluster_handler_test.go
func TestClusterHandler_RegisterCluster_Integration(t *testing.T)
func TestClusterHandler_ListClusters_Integration(t *testing.T)
func TestClusterHandler_DeregisterCluster_Integration(t *testing.T)
```

### 7.3 E2E 테스트

```bash
# 실제 Proxmox 클러스터와의 통합 테스트
go test -v -tags=e2e ./test/e2e
```

---

## 8. 성능 고려사항

### 8.1 메모리 저장소 성능

| 작업 | 시간 복잡도 | 예상 시간 |
|------|-----------|---------|
| Save | O(1) | < 1μs |
| FindByID | O(1) | < 1μs |
| FindByName | O(n) | < 10μs (n=100) |
| List | O(n) | < 100μs (n=100) |
| Delete | O(1) | < 1μs |

### 8.2 Proxmox API 성능

- 인증: 200-500ms (보안 검사)
- 버전 조회: 50-100ms
- 노드 개수 조회: 50-100ms

**최적화:**
- 토큰 캐싱 (향후)
- 병렬 API 호출 (향후)

---

## 9. 확장성 고려사항

### 9.1 데이터베이스 마이그레이션

**현재:** `MemoryRepository` (MVP)

**향후:** PostgreSQL 기반 `DatabaseRepository`

**변경 영역:**
```
internal/infrastructure/persistence/
├── memory_repository.go      # MVP용 (삭제 예정)
└── database_repository.go    # 새 구현
```

**변경 불필요:**
- Domain layer (Repository 인터페이스 동일)
- Application layer
- API layer

### 9.2 Proxmox 클라이언트 확장

**향후 추가 메서드:**
```go
func (c *Client) ListNodes(ctx context.Context, ticket string) ([]*NodeInfo, error)
func (c *Client) ListVMs(ctx context.Context, ticket string) ([]*VMInfo, error)
func (c *Client) CreateVM(ctx context.Context, ticket string, config *VMConfig) error
```

---

## 10. 배포 및 운영

### 10.1 환경 변수

```bash
SERVER_PORT=8080                    # API 서버 포트
```

### 10.2 로깅

```
[INFO] Cluster registered successfully, cluster_id=xxx, name=xxx
[ERROR] Proxmox authentication failed, error=...
[WARN] Cluster name already exists, name=...
```

### 10.3 헬스 체크

```bash
curl http://localhost:8080/health
# {"status":"healthy"}
```

---

## 11. 향후 개선 사항

| 우선순위 | 항목 | 설명 |
|---------|------|------|
| P1 | 데이터베이스 통합 | PostgreSQL 기반 영속성 |
| P1 | 비밀번호 암호화 | 저장된 비밀번호 암호화 |
| P2 | 토큰 캐싱 | Proxmox 인증 토큰 캐싱 |
| P2 | API 속도 향상 | 병렬 API 호출 |
| P3 | 모니터링 | Prometheus 메트릭 |
| P3 | 감사 로깅 | 모든 작업 추적 |

---

## 12. 용어 정의

| 용어 | 설명 |
|------|------|
| **Cluster** | Proxmox 하이퍼바이저 클러스터 |
| **Node** | Proxmox 클러스터의 개별 호스트 |
| **Repository** | 데이터 영속성 추상화 계층 |
| **DTO** | Data Transfer Object - 계층 간 데이터 전달 |
| **Ticket** | Proxmox API 인증 토큰 |
| **MVP** | Minimum Viable Product |

---

## 참고 자료

- [Proxmox API 문서](https://pve.proxmox.com/pve-docs/api-viewer/)
- [Go 에러 처리](https://go.dev/blog/error-handling-and-go)
- [Repository 패턴](https://martinfowler.com/eaaCatalog/repository.html)
