# 供应商注册流程改造计划（修订版）

## 目标

1. 供应商注册：填邮箱+密码+主体资料 → **直接可以登录**，进入供应商看板
2. 提交账号功能：需管理员审核通过后才开放（后端已有此限制，无需改动）
3. 移除独立的"供应商入驻"页面（ApplyView.vue），因为注册时已经填完了

## 当前状态（上次改动后的残留）

| 文件 | 当前状态 | 需要改为 |
|------|---------|---------|
| RegisterView.vue | 供应商主体资料表单**已被移除** | **恢复**回来 |
| auth_handler.go applySupplierSignupProfile | company_name 为空时不报错 | **恢复报错**（必填） |
| ApplyView.vue | 仍存在 | **删除** |
| router: /supplier/apply | 仍存在 | **删除** |
| AppSidebar: 供应商入驻 | 仍存在 | **删除** |
| Login | 无阻断 | **不改** |
| CreateSupplierAccount | 已限制未审核不可提交 | **不改**（已有） |

## 改造步骤

### 步骤 1：恢复 RegisterView.vue 供应商主体资料表单

**文件**: `frontend/src/views/auth/RegisterView.vue`

恢复之前删除的供应商主体资料表单区块（`v-if="isSupplierSignup"`）：
- 主体名称(必填)、联系人、联系邮箱、联系电话、资源说明
- 恢复 `formData.supplier_profile` 字段
- 恢复 `supplierProfilePayload` computed
- 恢复 `errors.supplier_company_name` 错误字段
- 恢复 `validateForm()` 中主体名称的校验
- 恢复 `oauthSignupContext` 中的 `supplier_profile` 参数

**注册后行为不变**：选择 supplier 后，先调用 `authStore.register`（会传入 supplier_profile），后端返回 token，自动登录跳转。

### 步骤 2：后端 auth_handler.go — 恢复 company_name 必填校验

**文件**: `backend/internal/handler/auth_handler.go`

`applySupplierSignupProfile` 函数：
- 将 company_name 为空时的逻辑**恢复为返回错误**
- 不允许跳过 company_name 注册供应商

```go
if strings.TrimSpace(profile.CompanyName) == "" {
    return infraerrors.BadRequest("SUPPLIER_COMPANY_REQUIRED", "company name is required")
}
```

### 步骤 3：后端 Register — 保持返回 Token（不改）

注册接口已正确：供应商注册成功后调用 `respondWithTokenPair` 发 token，供应商可立即登录。**此步骤无需改动。**

### 步骤 4：后端 Login — 保持不拦截（不改）

Login 接口无 SupplierStatus 校验，pending 供应商可正常登录。用户 profile 接口中 `enrichSupplierAccess` 仅设 `has_supplier_access = false`，前端据此隐藏账号提交功能。**此步骤无需改动。**

### 步骤 5：删除供应商入驻页面

1. **router/index.ts** — 删除 `/supplier/apply` 路由定义
2. **ApplyView.vue** — 删除文件
3. **AppSidebar.vue** — 移除侧边栏中"供应商入驻"菜单项（普通用户可见的那个）

### 步骤 6（验证）：确认账号提交流程的阻断点

后端已有阻断，无需修改：

- `CreateSupplierAccount`（supplier.go:470）→ 检查 `profile.Status != SupplierStatusApproved` → 返回 Forbidden
- `TestSupplierAccount`（supplier.go:376）→ 同上
- 前端 `isSupplier = user.has_supplier_access === true`，未审核时为 false → 供应商页面 UI 层面也可控

## 验证点

| # | 验证场景 | 预期结果 |
|---|---------|---------|
| 1 | 供应商注册未填公司名 | 前端提示"请输入供应商主体名称" |
| 2 | 供应商正常注册（填完所有字段） | 注册成功，自动登录，进入 `/supplier/dashboard` |
| 3 | 供应商登录后页面 | 显示供应商看板，侧边栏是供应商视图 |
| 4 | 供应商登录后尝试创建账号 | 后端返回 "supplier profile is not approved" |
| 5 | 管理员审核通过后 | 供应商可正常创建账号 |
| 6 | 普通用户注册 | 不受影响 |
| 7 | 侧边栏无"供应商入驻" | 普通用户看不到此菜单项 |
