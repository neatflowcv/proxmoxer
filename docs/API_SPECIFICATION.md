# Proxmoxer Cluster Management API 명세

## 개요

Proxmoxer Cluster Management API는 Proxmox 클러스터를 등록, 제거, 조회하기 위한 RESTful API입니다.

**Base URL:** `http://localhost:8080/api/v1`

**API Version:** v1

**Content-Type:** `application/json`

---

## 엔드포인트

### 1. 클러스터 등록

#### 요청

```
POST /api/v1/clusters
Content-Type: application/json
```

**Request Body:**

| 필드 | 타입 | 필수 | 설명 | 예시 |
|------|------|------|------|------|
| name | string | O | 클러스터 이름 (max 255) | "prod-cluster" |
| api_endpoint | string | O | Proxmox API URL | "https://pve.example.com:8006" |
| username | string | O | Proxmox 사용자명 | "root@pam" |
| password | string | O | Proxmox 비밀번호 | "password123" |

**예시:**

```bash
curl -X POST http://localhost:8080/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "prod-cluster",
    "api_endpoint": "https://pve.example.com:8006",
    "username": "root@pam",
    "password": "your-password"
  }'
```

#### 응답

**성공 (201 Created):**

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

**에러 (400 Bad Request - 유효성 검증 실패):**

```json
{
  "code": "BadRequest",
  "message": "Cluster name is required",
  "details": null
}
```

**에러 (409 Conflict - 중복된 이름):**

```json
{
  "code": "Conflict",
  "message": "Cluster already exists",
  "details": null
}
```

**에러 (401 Unauthorized - 인증 실패):**

```json
{
  "code": "Unauthorized",
  "message": "Authentication failed",
  "details": null
}
```

**에러 (502 Bad Gateway - Proxmox 연결 실패):**

```json
{
  "code": "BadGateway",
  "message": "Failed to connect to Proxmox",
  "details": null
}
```

---

### 2. 클러스터 목록 조회

#### 요청

```
GET /api/v1/clusters
```

**쿼리 매개변수:** 없음

**예시:**

```bash
curl -X GET http://localhost:8080/api/v1/clusters
```

#### 응답

**성공 (200 OK):**

```json
{
  "clusters": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "prod-cluster",
      "api_endpoint": "https://pve.example.com:8006",
      "status": "healthy",
      "proxmox_version": "7.4-1",
      "node_count": 3,
      "created_at": "2024-01-11T10:30:00Z",
      "updated_at": "2024-01-11T10:30:00Z"
    },
    {
      "id": "660f9511-f40c-52e5-b827-557766551111",
      "name": "dev-cluster",
      "api_endpoint": "https://dev-pve.example.com:8006",
      "status": "healthy",
      "proxmox_version": "7.3-1",
      "node_count": 2,
      "created_at": "2024-01-10T15:20:00Z",
      "updated_at": "2024-01-10T15:20:00Z"
    }
  ],
  "total": 2
}
```

**빈 목록:**

```json
{
  "clusters": [],
  "total": 0
}
```

---

### 3. 클러스터 상세 조회

#### 요청

```
GET /api/v1/clusters/{id}
```

**Path 매개변수:**

| 매개변수 | 타입 | 설명 |
|---------|------|------|
| id | string | 클러스터 ID (UUID) |

**예시:**

```bash
curl -X GET http://localhost:8080/api/v1/clusters/550e8400-e29b-41d4-a716-446655440000
```

#### 응답

**성공 (200 OK):**

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

**에러 (404 Not Found):**

```json
{
  "code": "NotFound",
  "message": "Cluster not found",
  "details": null
}
```

---

### 4. 클러스터 제거 (Deregister)

#### 요청

```
DELETE /api/v1/clusters/{id}
```

**Path 매개변수:**

| 매개변수 | 타입 | 설명 |
|---------|------|------|
| id | string | 클러스터 ID (UUID) |

**예시:**

```bash
curl -X DELETE http://localhost:8080/api/v1/clusters/550e8400-e29b-41d4-a716-446655440000
```

#### 응답

**성공 (204 No Content):**

```
(Empty body)
```

**에러 (404 Not Found):**

```json
{
  "code": "NotFound",
  "message": "Cluster not found",
  "details": null
}
```

---

## HTTP 상태 코드

| 코드 | 설명 | 사용 상황 |
|------|------|---------|
| 200 | OK | 요청 성공 |
| 201 | Created | 리소스 생성 성공 |
| 204 | No Content | 리소스 삭제 성공 |
| 400 | Bad Request | 유효하지 않은 요청 |
| 401 | Unauthorized | 인증 실패 |
| 404 | Not Found | 리소스 미존재 |
| 409 | Conflict | 리소스 중복 |
| 500 | Internal Server Error | 서버 에러 |
| 502 | Bad Gateway | Proxmox 연결 실패 |

---

## 에러 응답 형식

모든 에러 응답은 다음 형식을 따릅니다:

```json
{
  "code": "ErrorCode",
  "message": "Human-readable error message",
  "details": {
    "key": "value"
  }
}
```

**필드:**
- `code`: HTTP 상태 코드 텍스트
- `message`: 에러 메시지
- `details`: 추가 정보 (선택사항)

---

## 요청/응답 데이터 타입

### Cluster 객체

```json
{
  "id": "string (UUID)",
  "name": "string",
  "api_endpoint": "string (URL)",
  "status": "string (healthy|degraded|unhealthy|unknown)",
  "proxmox_version": "string",
  "node_count": "integer",
  "created_at": "string (RFC3339)",
  "updated_at": "string (RFC3339)"
}
```

### ClusterStatus 열거형

```
healthy    - 클러스터 정상
degraded   - 클러스터 부분적 장애
unhealthy  - 클러스터 장애
unknown    - 상태 미확인
```

---

## 예제

### 예제 1: 새 클러스터 등록 및 조회

**Step 1: 클러스터 등록**

```bash
$ curl -X POST http://localhost:8080/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "prod-cluster",
    "api_endpoint": "https://pve.example.com:8006",
    "username": "root@pam",
    "password": "my-secure-password"
  }'

# 응답:
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

**Step 2: 생성된 클러스터 조회**

```bash
$ curl -X GET http://localhost:8080/api/v1/clusters/550e8400-e29b-41d4-a716-446655440000

# 응답:
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

### 예제 2: 모든 클러스터 나열

```bash
$ curl -X GET http://localhost:8080/api/v1/clusters

# 응답:
{
  "clusters": [
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
  ],
  "total": 1
}
```

### 예제 3: 클러스터 제거

```bash
$ curl -X DELETE http://localhost:8080/api/v1/clusters/550e8400-e29b-41d4-a716-446655440000

# 응답: 204 No Content (body 없음)
```

### 예제 4: 에러 처리 - 중복된 이름

```bash
$ curl -X POST http://localhost:8080/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "prod-cluster",
    "api_endpoint": "https://pve2.example.com:8006",
    "username": "root@pam",
    "password": "password"
  }'

# 응답: 409 Conflict
{
  "code": "Conflict",
  "message": "Cluster already exists",
  "details": null
}
```

### 예제 5: 에러 처리 - 인증 실패

```bash
$ curl -X POST http://localhost:8080/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-cluster",
    "api_endpoint": "https://pve.example.com:8006",
    "username": "root@pam",
    "password": "wrong-password"
  }'

# 응답: 401 Unauthorized
{
  "code": "Unauthorized",
  "message": "Authentication failed",
  "details": null
}
```

---

## 클라이언트 구현 예제

### Go 클라이언트

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RegisterClusterRequest struct {
	Name        string `json:"name"`
	APIEndpoint string `json:"api_endpoint"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

type ClusterResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	APIEndpoint    string `json:"api_endpoint"`
	Status         string `json:"status"`
	ProxmoxVersion string `json:"proxmox_version"`
	NodeCount      int    `json:"node_count"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

func registerCluster(client *http.Client, baseURL string, req *RegisterClusterRequest) (*ClusterResponse, error) {
	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", baseURL+"/clusters", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var cluster ClusterResponse
	json.Unmarshal(respBody, &cluster)

	return &cluster, nil
}

func main() {
	client := &http.Client{}
	req := &RegisterClusterRequest{
		Name:        "prod-cluster",
		APIEndpoint: "https://pve.example.com:8006",
		Username:    "root@pam",
		Password:    "password",
	}

	cluster, _ := registerCluster(client, "http://localhost:8080/api/v1", req)
	fmt.Printf("Registered cluster: %s (ID: %s)\n", cluster.Name, cluster.ID)
}
```

### cURL 명령어 모음

```bash
# 클러스터 등록
curl -X POST http://localhost:8080/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d @- <<EOF
{
  "name": "prod-cluster",
  "api_endpoint": "https://pve.example.com:8006",
  "username": "root@pam",
  "password": "password"
}
EOF

# 모든 클러스터 조회
curl -X GET http://localhost:8080/api/v1/clusters

# 특정 클러스터 조회
curl -X GET http://localhost:8080/api/v1/clusters/{id}

# 클러스터 제거
curl -X DELETE http://localhost:8080/api/v1/clusters/{id}

# 헬스 체크
curl -X GET http://localhost:8080/health
```

---

## 버전 히스토리

| 버전 | 날짜 | 변경사항 |
|------|------|---------|
| v1 | 2024-01-11 | 초기 릴리스 (MVP) |

---

## 지원 및 문의

문제나 질문이 있으시면 프로젝트 이슈 트래커를 통해 보고해주세요.
