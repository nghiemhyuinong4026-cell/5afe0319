# 企业用车申请与车辆调度系统 (MVP)

## 项目简介

这是一个基于 Go 1.23 + Gin + GORM + PostgreSQL 构建的企业用车申请与车辆调度系统 MVP 版本。

### 核心功能闭环

1. **员工提交用车申请**
2. **主管审批通过/驳回**
3. **行政派车**（含冲突检测）
4. **司机开始行程**
5. **司机完成行程并回填里程**

### 技术栈

| 技术 | 版本 | 说明 |
|------|------|------|
| Go | 1.23 | 编程语言 |
| Gin | 1.9.1 | Web 框架 |
| GORM | 1.9.16 | ORM 框架 |
| PostgreSQL | 15 | 数据库 |
| JWT | - | 身份认证 |

### 企业级特征

1. **集中状态机** - 统一管理所有状态转换，确保业务逻辑一致性
2. **派车冲突检测** - 使用数据库事务检查车辆在申请时间段内是否已有未完成派车单

---

## 快速开始

### Docker 单文件启动

```bash
# 构建镜像
docker build -t vehicle-management-system .

# 运行容器（使用分配的端口）
docker run -d -p 18102:8080 --name docker-question-102 vehicle-management-system

# 查看日志
docker logs -f docker-question-102
```

服务启动后访问：`http://127.0.0.1:18102`

---

## 数据表设计 (6张)

### 1. users (用户表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| username | string | 用户名（唯一） |
| password | string | 密码（加密存储） |
| name | string | 姓名 |
| role | string | 角色 |
| phone | string | 电话 |

### 2. vehicles (车辆表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| plate_number | string | 车牌号（唯一） |
| brand | string | 品牌 |
| vehicle_model | string | 型号 |
| status | string | 状态 |
| seat_capacity | int | 座位数 |
| current_mileage | float64 | 当前里程 |

### 3. drivers (司机表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| user_id | uint | 关联用户ID |
| name | string | 姓名 |
| license_id | string | 驾照号（唯一） |
| phone | string | 电话 |
| is_available | bool | 是否可用 |

### 4. vehicle_requests (用车申请表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| requester_id | uint | 申请人ID |
| status | string | 状态 |
| start_location | string | 出发地 |
| end_location | string | 目的地 |
| purpose | string | 用途 |
| passengers | int | 乘客数 |
| departure_time | time | 出发时间 |
| return_time | time | 返回时间 |
| remark | string | 备注 |
| approver_id | uint | 审批人ID |

### 5. dispatch_orders (派车单表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| request_id | uint | 申请ID（唯一） |
| vehicle_id | uint | 车辆ID |
| driver_id | uint | 司机ID |
| dispatcher_id | uint | 调度员ID |
| start_mileage | float64 | 开始里程 |
| end_mileage | float64 | 结束里程 |
| departure_time | time | 实际出发时间 |
| return_time | time | 实际返回时间 |

### 6. audit_logs (审计日志表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| user_id | uint | 操作人ID |
| action | string | 操作类型 |
| resource_type | string | 资源类型 |
| resource_id | uint | 资源ID |
| detail | string | 详情（JSON） |

---

## 状态机设计

### 车辆状态

```
available (可用)
    ↓ 派车
dispatched (已派车)
    ↓ 完成行程
available (可用)
    ↓ 维护
maintenance (维护中)
```

### 申请状态

```
pending_approval (待审批)
    ↓ 审批通过          ↓ 审批驳回
approved (已通过)    rejected (已驳回)
    ↓ 派车              ↓ 取消
dispatched (已派车)    cancelled (已取消)
    ↓ 开始行程
in_progress (行程中)
    ↓ 完成行程
completed (已完成)
```

---

## API 接口

### 基础URL

`http://127.0.0.1:18102/api`

### 认证方式

除登录接口外，其他接口需要在 Header 中携带 JWT Token：

```
Authorization: Bearer <token>
```

### 接口列表

| # | 方法 | 路径 | 角色 | 说明 |
|---|------|------|------|------|
| 1 | POST | `/login` | 公开 | 用户登录 |
| 2 | GET | `/vehicles` | 所有登录用户 | 车辆列表 |
| 3 | GET | `/drivers` | 所有登录用户 | 司机列表 |
| 4 | POST | `/requests` | employee | 创建用车申请 |
| 5 | GET | `/requests/my` | employee | 我的申请 |
| 6 | GET | `/approvals/pending` | manager | 待审批列表 |
| 7 | POST | `/approvals/:id/approve` | manager | 审批通过 |
| 8 | POST | `/approvals/:id/reject` | manager | 审批驳回 |
| 9 | GET | `/dispatches/pending` | dispatcher | 待调度列表 |
| 10 | POST | `/dispatches` | dispatcher | 派车 |
| 11 | GET | `/trips/my` | driver | 我的行程 |
| 12 | POST | `/trips/:id/start` | driver | 开始行程 |
| 13 | POST | `/trips/:id/complete` | driver | 完成行程 |
| 14 | GET | `/audit-logs` | 所有登录用户 | 审计日志 |

---

## 测试账号

| 用户名 | 密码 | 角色 | 说明 |
|--------|------|------|------|
| employee | 123456 | 员工 | 提交用车申请 |
| manager | 123456 | 主管 | 审批申请 |
| dispatcher | 123456 | 调度员 | 派车 |
| driver | 123456 | 司机 | 执行行程 |

---

## 验收测试流程

### 运行测试脚本

```bash
chmod +x test_api.sh
./test_api.sh
```

### 手动测试步骤

#### 步骤 1：员工登录并提交申请

```bash
curl -X POST http://127.0.0.1:18102/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"employee","password":"123456"}'

curl -X POST http://127.0.0.1:18102/api/requests \
  -H "Authorization: Bearer <employee_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "start_location": "公司总部",
    "end_location": "机场",
    "purpose": "接送客户",
    "passengers": 3,
    "departure_time": "2026-05-10T09:00:00",
    "return_time": "2026-05-10T18:00:00",
    "remark": "需要商务车"
  }'
```

#### 步骤 2：主管审批通过

```bash
curl -X POST http://127.0.0.1:18102/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"manager","password":"123456"}'

curl -X GET http://127.0.0.1:18102/api/approvals/pending \
  -H "Authorization: Bearer <manager_token>"

curl -X POST http://127.0.0.1:18102/api/approvals/1/approve \
  -H "Authorization: Bearer <manager_token>" \
  -H "Content-Type: application/json" \
  -d '{"remark":"同意"}'
```

#### 步骤 3：行政派车

```bash
curl -X POST http://127.0.0.1:18102/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"dispatcher","password":"123456"}'

curl -X GET http://127.0.0.1:18102/api/dispatches/pending \
  -H "Authorization: Bearer <dispatcher_token>"

curl -X POST http://127.0.0.1:18102/api/dispatches \
  -H "Authorization: Bearer <dispatcher_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": 1,
    "vehicle_id": 1,
    "driver_id": 1
  }'
```

#### 步骤 4：司机开始行程

```bash
curl -X POST http://127.0.0.1:18102/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"driver","password":"123456"}'

curl -X POST http://127.0.0.1:18102/api/trips/1/start \
  -H "Authorization: Bearer <driver_token>"
```

#### 步骤 5：司机完成行程

```bash
curl -X POST http://127.0.0.1:18102/api/trips/1/complete \
  -H "Authorization: Bearer <driver_token>" \
  -H "Content-Type: application/json" \
  -d '{"end_mileage": 15100.5}'
```

---

## 项目结构

```
├── config/           # 配置管理
├── controllers/      # 控制器层
├── database/         # 数据库连接
├── middleware/       # 中间件
├── models/           # 数据模型
├── routes/           # 路由配置
├── seed/             # 测试数据初始化
├── services/         # 业务逻辑层
├── statemachine/     # 集中状态机
├── utils/            # 工具函数
├── .dockerignore     # Docker 忽略文件
├── .env              # 环境变量
├── Dockerfile        # Docker 构建文件
├── go.mod            # Go 模块依赖
├── main.go           # 入口文件
├── README.md         # 项目说明
└── test_api.sh       # API 测试脚本
```

---

## 环境变量配置

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| DB_HOST | 数据库地址 | localhost |
| DB_PORT | 数据库端口 | 5432 |
| DB_USER | 数据库用户名 | postgres |
| DB_PASSWORD | 数据库密码 | postgres |
| DB_NAME | 数据库名 | vehicle_management |
| JWT_SECRET | JWT 密钥 | default_secret |
| JWT_EXPIRES_IN | Token 有效期（小时） | 24 |
| SERVER_PORT | 服务端口 | 8080 |

---

## 注意事项

1. **生产环境**：请修改 `JWT_SECRET` 为强密钥
2. **数据库**：GORM 会自动迁移表结构，首次启动会自动创建表
3. **测试数据**：首次启动会自动创建测试账号和初始数据
4. **冲突检测**：派车时会检查车辆在申请时间段内是否已有未完成的派车单

---

## 许可证

MIT License