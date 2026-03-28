# ISSUE

## 2026-03-28 测试记录

已执行：

- `go test ./...`
- `go build ./...`
- `go build -ldflags="-H windowsgui" -o tankbattle_gui.exe .\cmd\tankbattle`
- GUI 启动验证：`tankbattle_gui.exe` 可正常拉起
- 调试 API 功能套件：`$env:TANKBATTLE_E2E='1'; $env:TANKBATTLE_E2E_BINARY='tankbattle_gui.exe'; go test ./testing/functional`
- 调试 API 界面快照套件：`$env:TANKBATTLE_E2E='1'; $env:TANKBATTLE_E2E_BINARY='tankbattle_gui.exe'; go test ./testing/ui`

结果概览：

- 逻辑单测、构建、GUI 启动验证通过
- 调试 API 功能套件通过
- 调试 API 界面快照套件失败，当前只影响菜单场景快照

## 待办事项

### 1. 菜单界面快照与 golden 基线不一致

现象：

- `go test ./testing/ui` 失败
- 失败场景：
  - `menu/menu-default.png`
  - `menu/menu-hard.png`
  - `menu/menu-resume.png`

失败产物：

- `.tmp/testing/failures/menu_menu-default.actual.png`
- `.tmp/testing/failures/menu_menu-default.diff.png`
- `.tmp/testing/failures/menu_menu-hard.actual.png`
- `.tmp/testing/failures/menu_menu-hard.diff.png`
- `.tmp/testing/failures/menu_menu-resume.actual.png`
- `.tmp/testing/failures/menu_menu-resume.diff.png`

判断：

- 近期提交包含菜单文案/侧栏重构，当前更像是菜单视觉输出发生了变化
- 需要确认这是“预期设计变更”还是“未预期回归”

处理建议：

- 若菜单改动是预期的：重新审核菜单视觉并更新 `testing/testdata/golden/menu/`
- 若菜单改动不是预期的：按 diff 图回查菜单渲染，修复后再跑 `go test ./testing/ui`

### 2. 将真实 GUI 快照套件纳入固定回归流程

现象：

- `go test ./...` 当前是通过的，但默认不会覆盖真实 GUI 快照回归
- 只有显式设置 `TANKBATTLE_E2E=1` 后，菜单快照问题才会暴露

处理建议：

- 在本地验收或 CI 中固定增加：
  - `$env:TANKBATTLE_E2E='1'; $env:TANKBATTLE_E2E_BINARY='tankbattle_gui.exe'; go test ./testing/functional ./testing/ui`
- 避免仅凭默认 `go test ./...` 误判界面回归状态
