#!/bin/bash

# 企业用车申请与车辆调度系统 - API 测试示例

set -e

# 基础URL
BASE_URL="http://localhost:8080/api"

# 临时文件存储响应
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# ============================================
# 辅助函数
# ============================================
log_info() {
    echo -e "\033[1;34m[INFO]\033[0m $1"
}

log_success() {
    echo -e "\033[1;32m[SUCCESS]\033[0m $1"
}

log_error() {
    echo -e "\033[1;31m[ERROR]\033[0m $1"
}

# ============================================
# 1. 登录获取Token
# ============================================

log_info "=== 1. 登录各角色用户 ==="

# 员工登录
log_info "员工登录 (employee / 123456)"
curl -s -X POST "$BASE_URL/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"employee","password":"123456"}' > $TMP_DIR/emp_login.json
EMPLOYEE_TOKEN=$(cat $TMP_DIR/emp_login.json | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('token',''))")
if [ -z "$EMPLOYEE_TOKEN" ]; then
    log_error "员工登录失败"
    cat $TMP_DIR/emp_login.json
    exit 1
fi
log_success "员工登录成功, Token: ${EMPLOYEE_TOKEN:0:30}..."

# 主管登录
log_info "主管登录 (manager / 123456)"
curl -s -X POST "$BASE_URL/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"manager","password":"123456"}' > $TMP_DIR/mgr_login.json
MANAGER_TOKEN=$(cat $TMP_DIR/mgr_login.json | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('token',''))")
if [ -z "$MANAGER_TOKEN" ]; then
    log_error "主管登录失败"
    cat $TMP_DIR/mgr_login.json
    exit 1
fi
log_success "主管登录成功, Token: ${MANAGER_TOKEN:0:30}..."

# 调度员登录
log_info "调度员登录 (dispatcher / 123456)"
curl -s -X POST "$BASE_URL/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"dispatcher","password":"123456"}' > $TMP_DIR/disp_login.json
DISPATCHER_TOKEN=$(cat $TMP_DIR/disp_login.json | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('token',''))")
if [ -z "$DISPATCHER_TOKEN" ]; then
    log_error "调度员登录失败"
    cat $TMP_DIR/disp_login.json
    exit 1
fi
log_success "调度员登录成功, Token: ${DISPATCHER_TOKEN:0:30}..."

# 司机登录
log_info "司机登录 (driver / 123456)"
curl -s -X POST "$BASE_URL/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"driver","password":"123456"}' > $TMP_DIR/drv_login.json
DRIVER_TOKEN=$(cat $TMP_DIR/drv_login.json | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('token',''))")
if [ -z "$DRIVER_TOKEN" ]; then
    log_error "司机登录失败"
    cat $TMP_DIR/drv_login.json
    exit 1
fi
log_success "司机登录成功, Token: ${DRIVER_TOKEN:0:30}..."

# ============================================
# 2. 获取车辆和司机列表
# ============================================

log_info "=== 2. 获取车辆和司机列表 ==="

# 获取车辆列表
log_info "获取车辆列表"
curl -s -X GET "$BASE_URL/vehicles" \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" | python3 -m json.tool

# 获取司机列表
log_info "获取司机列表"
curl -s -X GET "$BASE_URL/drivers" \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" | python3 -m json.tool

# ============================================
# 3. 员工提交用车申请
# ============================================

log_info "=== 3. 员工提交用车申请 ==="

log_info "创建用车申请"
TOMORROW=$(date -v+1d +%Y-%m-%d)
curl -s -X POST "$BASE_URL/requests" \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"start_location\": \"公司总部\",
    \"end_location\": \"机场\",
    \"purpose\": \"接送客户\",
    \"passengers\": 3,
    \"departure_time\": \"${TOMORROW}T09:00:00\",
    \"return_time\": \"${TOMORROW}T18:00:00\",
    \"remark\": \"需要商务车\"
  }" > $TMP_DIR/create_req.json

REQUEST_ID=$(cat $TMP_DIR/create_req.json | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('id',''))")
if [ -z "$REQUEST_ID" ]; then
    log_error "创建申请失败"
    cat $TMP_DIR/create_req.json
    exit 1
fi
log_success "申请创建成功, 申请ID: $REQUEST_ID"
cat $TMP_DIR/create_req.json | python3 -m json.tool

# 查看我的申请
log_info "查看我的申请"
curl -s -X GET "$BASE_URL/requests/my" \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" | python3 -m json.tool

# ============================================
# 4. 主管审批通过
# ============================================

log_info "=== 4. 主管审批通过 ==="

log_info "查看待审批列表"
curl -s -X GET "$BASE_URL/approvals/pending" \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -H "Content-Type: application/json" | python3 -m json.tool

log_info "审批通过申请ID $REQUEST_ID"
curl -s -X POST "$BASE_URL/approvals/$REQUEST_ID/approve" \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"remark": "同意，请注意安全"}' > $TMP_DIR/approve.json

log_success "审批结果:"
cat $TMP_DIR/approve.json | python3 -m json.tool

# 查看申请详情
log_info "查看申请详情"
curl -s -X GET "$BASE_URL/requests/$REQUEST_ID" \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -H "Content-Type: application/json" | python3 -m json.tool

# ============================================
# 5. 行政派车
# ============================================

log_info "=== 5. 行政派车 ==="

log_info "查看待调度列表"
curl -s -X GET "$BASE_URL/dispatches/pending" \
  -H "Authorization: Bearer $DISPATCHER_TOKEN" \
  -H "Content-Type: application/json" | python3 -m json.tool

# 查看可用车辆
log_info "查看可用车辆"
curl -s -X GET "$BASE_URL/vehicles?status=available" \
  -H "Authorization: Bearer $DISPATCHER_TOKEN" \
  -H "Content-Type: application/json" | python3 -m json.tool

log_info "派车 (车辆ID: 1, 司机ID: 1)"
curl -s -X POST "$BASE_URL/dispatches" \
  -H "Authorization: Bearer $DISPATCHER_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"request_id\": $REQUEST_ID,
    \"vehicle_id\": 1,
    \"driver_id\": 1
  }" > $TMP_DIR/dispatch.json

DISPATCH_ID=$(cat $TMP_DIR/dispatch.json | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('id',''))")
if [ -z "$DISPATCH_ID" ]; then
    log_error "派车失败"
    cat $TMP_DIR/dispatch.json
    exit 1
fi
log_success "派车成功, 派车单ID: $DISPATCH_ID"
cat $TMP_DIR/dispatch.json | python3 -m json.tool

# ============================================
# 6. 重复派车测试 (应该失败)
# ============================================

log_info "=== 6. 重复派车测试 (应该失败) ==="

log_info "尝试对同一申请重复派车 (应该失败)"
# 暂时禁用 set -e，因为这个请求预期会失败
set +e
curl -s -X POST "$BASE_URL/dispatches" \
  -H "Authorization: Bearer $DISPATCHER_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"request_id\": $REQUEST_ID,
    \"vehicle_id\": 1,
    \"driver_id\": 1
  }" > $TMP_DIR/dup_dispatch.json
set -e

log_info "重复派车结果 (预期是错误):"
cat $TMP_DIR/dup_dispatch.json | python3 -m json.tool

# ============================================
# 7. 司机开始行程
# ============================================

log_info "=== 7. 司机开始行程 ==="

log_info "查看我的行程"
curl -s -X GET "$BASE_URL/trips/my" \
  -H "Authorization: Bearer $DRIVER_TOKEN" \
  -H "Content-Type: application/json" | python3 -m json.tool

log_info "开始行程 (派车单ID: $DISPATCH_ID"
curl -s -X POST "$BASE_URL/trips/$DISPATCH_ID/start" \
  -H "Authorization: Bearer $DRIVER_TOKEN" \
  -H "Content-Type: application/json" > $TMP_DIR/start_trip.json

log_success "开始行程结果:"
cat $TMP_DIR/start_trip.json | python3 -m json.tool

# ============================================
# 8. 司机完成行程并回填里程
# ============================================

log_info "=== 8. 司机完成行程并回填里程 ==="

log_info "完成行程 (结束里程: 15100.5)"
curl -s -X POST "$BASE_URL/trips/$DISPATCH_ID/complete" \
  -H "Authorization: Bearer $DRIVER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"end_mileage": 15100.5}' > $TMP_DIR/complete_trip.json

log_success "完成行程结果:"
cat $TMP_DIR/complete_trip.json | python3 -m json.tool

# ============================================
# 9. 查看审计日志
# ============================================

log_info "=== 9. 查看审计日志 ==="

log_info "获取审计日志"
curl -s -X GET "$BASE_URL/audit-logs?page=1&page_size=20" \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" | python3 -m json.tool

# ============================================
# 测试完成
# ============================================
echo " "
log_success "==========================================="
log_success "      所有测试流程完成！"
log_success "==========================================="
log_success "测试流程回顾:"
log_success "1. 员工登录 -> 提交用车申请"
log_success "2. 主管登录 -> 审批通过"
log_success "3. 调度员登录 -> 派车"
log_success "4. 重复派车测试 -> 预期失败（冲突检测）"
log_success "5. 司机登录 -> 开始行程"
log_success "6. 司机完成行程 -> 回填里程"
log_success "7. 查看审计日志 -> 所有操作可查"
log_success "==========================================="
