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
