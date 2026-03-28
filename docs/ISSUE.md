# ISSUE

## 2026-03-28 测试复核

已执行：

- `go test ./...`
- `go build ./...`
- `go build -ldflags="-H windowsgui" -o tankbattle_gui.exe .\cmd\tankbattle`
- `$env:TANKBATTLE_E2E='1'; $env:TANKBATTLE_E2E_BINARY='tankbattle_gui.exe'; go test ./testing/functional`
- `$env:TANKBATTLE_E2E='1'; $env:TANKBATTLE_E2E_BINARY='tankbattle_gui.exe'; go test ./testing/ui`
- `$env:TANKBATTLE_E2E='1'; $env:TANKBATTLE_E2E_BINARY='tankbattle_gui.exe'; $env:TANKBATTLE_UPDATE_GOLDEN='1'; go test ./testing/ui`
- `$env:TANKBATTLE_E2E='1'; $env:TANKBATTLE_E2E_BINARY='tankbattle_gui.exe'; go test ./testing/functional ./testing/ui`

复核结果：

- 逻辑单测、构建、GUI 构建通过
- 调试 API 功能套件通过
- 菜单 golden 已同步到当前预期视觉输出
- 更新后 `./testing/ui` 与组合 GUI E2E 套件均已通过

## 问题状态

### 1. 菜单快照与 golden 基线不一致

状态：`已完成 golden 同步`

处理结论：

- 开发已确认当前菜单视觉调整是预期变更
- `testing/testdata/golden/menu/` 下的 3 张菜单 golden 已同步更新
- 更新后已验证：
  - `go test ./testing/ui`
  - `go test ./testing/functional ./testing/ui`

结论：

- 该项不再是待处理缺陷
- 本质是测试基线随预期设计变更同步，而不是运行时菜单逻辑问题

### 2. 真实 GUI 快照套件未纳入默认回归流程

状态：`已确认，仍待处理`

确认结论：

- 该项仍是有效问题
- `go test ./...` 虽然会包含 `testing/ui` 和 `testing/functional` 包，但未设置 `TANKBATTLE_E2E=1` 时，真实 GUI 调试 API 回归不会执行
- 这意味着默认回归只能覆盖单元测试和非 E2E 路径，无法自动暴露真实 GUI 快照差异

下一步：

- 在本地验收脚本或 CI 中显式加入：

```powershell
$env:TANKBATTLE_E2E='1'
$env:TANKBATTLE_E2E_BINARY='tankbattle_gui.exe'
go test ./testing/functional ./testing/ui
```

- 在 CI 接入前，不应仅凭默认 `go test ./...` 判断界面回归已完全覆盖

## 总结

当前 issue 状态如下：

- 菜单 golden 差异：已同步完成
- GUI E2E 未进默认回归：已确认，属于流程缺口

## 2026-03-28 Review 待办补充

来源：

- reviewer 对新增 E2E / UI bridge 相关改动的审查结论

### 3. E2E 会话可能误用仓库根目录旧可执行文件

状态：`已完成`

处理结论：

- `testing/testkit/launcher.go` 已改为只接受显式指定的 `TANKBATTLE_E2E_BINARY`
- 未指定 `TANKBATTLE_E2E_BINARY` 时，E2E 会话会统一构建并复用 `.tmp/testing/bin/tankbattle-e2e.exe`
- 不再默认复用仓库根目录已有的 `tankbattle_gui.exe` / `tankbattle.exe`
- 已补充 `testing/testkit/launcher_test.go`，覆盖相对路径解析、缺失路径报错和默认不复用仓库产物

验证：

- `go test ./...`
- `$env:TANKBATTLE_E2E='1'; $env:TANKBATTLE_E2E_BINARY='tankbattle_gui.exe'; go test ./testing/functional ./testing/ui`

### 4. E2E 清理阶段会把子进程异常退出当作正常结束

状态：`已完成`

处理结论：

- `Session.Close()` 现在会在子进程已经自行退出时直接返回原始错误，不再吞掉异常退出
- 仅在 `Close()` 主动 `Kill()` 子进程后的清理路径上，才会归一化 kill 产生的退出错误
- 已补充 `testing/testkit/launcher_test.go`，覆盖“异常退出必须报错”和“主动 kill 不应报错”两条回归用例

验证：

- `go test ./...`
- `$env:TANKBATTLE_E2E='1'; $env:TANKBATTLE_E2E_BINARY='tankbattle_gui.exe'; go test ./testing/functional ./testing/ui`

### 5. UI bridge 每帧重建完整渲染模型，存在额外分配压力

状态：`已完成`

确认结论：

- `internal/game/ui_bridge.go` 已改为复用 `uiSnapshot` 缓冲区，避免每帧重新分配敌人、子弹、墙体、爆炸、道具、历史等切片
- `internal/ui` 兼容层已改为对象池 + 值切片复用，不再每帧重建独立指针切片
- 当前仍保留快照边界，继续满足调试 API 与 UI 回归快照的稳定性要求

结论：

- 该项已按“可复用缓冲区 + 减少重复映射”完成收敛
- 不再作为待处理代码缺陷保留
