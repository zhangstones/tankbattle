---
name: ui-testing
description: 为桌面游戏和界面密集型功能建立、维护和排查可回归的 UI 测试流程。用于新增或修改菜单、HUD、弹层、历史面板、调试场景、golden 快照、调试 API，或处理界面回归时；优先使用游戏内调试 API 和快照导出，不依赖桌面截图或键盘模拟。
---

# UI 测试

## 目标

把界面验证固定成三层：

1. 布局/约束测试：快速发现对齐、列宽、重叠和边界问题。
2. 功能验收测试：验证“动作序列 -> 游戏状态”是否正确。
3. 快照回归测试：验证菜单、HUD、弹层和历史面板的视觉输出。

原则：

- 正式回归优先走游戏内调试 API。
- 正式快照优先走游戏内 PNG 导出。
- 桌面整屏截图只用于临时人工排查，不作为正式回归入口。

## 先做检查

开始前先确认当前仓库布局，不要假设旧路径仍然有效：

- 运行时代码：`internal/tankbattle/`
- 测试目录：`testing/testkit/`、`testing/functional/`、`testing/ui/`、`testing/testdata/golden/`
- 测试文档：`docs/TEST-API.md`、`docs/TEST-DESIGN.md`、`docs/TEST-CASES.md`、`docs/ISSUE.md`

优先阅读这些文件：

- `internal/tankbattle/debug_api.go`
- `internal/tankbattle/debug_api_test.go`
- `testing/functional/debug_api_flow_test.go`
- `testing/ui/snapshot_test.go`
- `docs/TEST-API.md`
- `docs/TEST-CASES.md`

如果用户说“界面测一下”或“看图验证”，先判断要做的是哪一类：

- 视觉回归：补 `scene.*` 和 snapshot / golden。
- 交互流转：补 `debug_api_flow_test.go` 一类的功能验收。
- 单纯对齐/溢出风险：先补布局约束测试，再决定是否需要 golden。

## 核心经验

1. 内部快照比外部截图可靠。
- 外部整屏截图依赖窗口前台、焦点和遮挡。
- 游戏内导出的是渲染结果本身，稳定且可重复。

2. `scene.*` 是 UI 测试的稳定入口。
- 命名场景适合直接切到菜单、HUD、暂停、胜负等固定画面。
- 切到 `scene.*` 后应进入调试冻结态，避免动画、pulse 和计时器污染快照。

3. 功能测试和 UI 测试要分层。
- 功能测试只断言状态，不比图片。
- UI 测试只构造场景、导出 PNG、比对 golden。
- 不要把大量状态断言和图片断言混在同一个测试里。

4. `go test ./...` 不等于真实 GUI 回归。
- 默认回归会覆盖单元测试、约束测试和非 E2E 路径。
- 真实 GUI 功能 / 快照回归必须显式启用 `TANKBATTLE_E2E=1`。

5. golden 更新不是“修测试”，而是确认设计。
- 如果视觉变化是预期设计调整，先人工确认，再更新 golden。
- 如果视觉变化并非预期，就先修代码，不要直接覆盖基线。

6. 文档要按职责联动，不要到处重复。
- 动作名、场景名、接口字段变化：更新 `docs/TEST-API.md`
- 分层策略、命令、目录变化：更新 `docs/TEST-DESIGN.md`
- 案例矩阵变化：更新 `docs/TEST-CASES.md`
- 缺口和待跟进事项：更新 `docs/ISSUE.md`

## 标准流程

### 1. 明确变更类型

先判断本次改动属于哪几类：

- 视觉样式变化
- 菜单 / HUD / 弹层布局变化
- 菜单操作或状态切换变化
- 调试 API / 场景能力变化
- golden 基线变化

### 2. 补稳定入口

如果新界面状态没有稳定入口，优先补到调试 API：

- 在 `internal/tankbattle/debug_api.go` 新增或调整 `scene.*`
- 在 `internal/tankbattle/debug_api_test.go` 补对应基础测试

要求：

- 场景命名清晰，直接对应用户可识别画面
- 场景构造应尽量复用真实状态流转，而不是维护第二套 UI 逻辑

### 3. 按层补测试

功能变化：

- 更新 `testing/functional/debug_api_flow_test.go`
- 断言状态字段，例如 `game_state`、`difficulty`、`wave`、`score`、`paused`

视觉变化：

- 更新 `testing/ui/snapshot_test.go`
- 新增或调整场景到 golden 的映射
- 只在确认预期视觉变化后更新 `testing/testdata/golden/`

布局稳定性变化：

- 优先补对应约束测试
- 典型问题：文本超框、右侧 HUD 重叠、label/value 漂移、列宽变化

### 4. 更新文档

最小职责更新规则：

- 只改接口契约，不改 `README.md`
- 只改测试策略，不改功能文档
- 只改测试案例时，优先更新 `docs/TEST-CASES.md`
- 用户可见的运行 / 测试入口变化时，才更新 `README.md`

### 5. 验证命令

代码变化后，至少执行：

```powershell
go test ./...
```

如果这次改动涉及真实 GUI UI 回归，再执行：

```powershell
$env:TANKBATTLE_E2E = "1"
$env:TANKBATTLE_E2E_BINARY = "tankbattle_gui.exe"
go test ./testing/functional ./testing/ui
```

如果确认要同步新的视觉基线，再执行：

```powershell
$env:TANKBATTLE_E2E = "1"
$env:TANKBATTLE_E2E_BINARY = "tankbattle_gui.exe"
$env:TANKBATTLE_UPDATE_GOLDEN = "1"
go test ./testing/ui
```

注意：

- 不要手动覆盖用户已经配置好的 `GOCACHE`
- 不要跳过功能测试只跑 snapshot

## 快速判定规则

出现 snapshot mismatch 时，按下面顺序判断：

1. 这是预期设计调整吗？
- 是：更新 golden，并在提交说明里写明原因。
- 否：继续下一步。

2. 这是命名场景本身不稳定吗？
- 检查是否进入了调试冻结态。
- 检查是否混入了非 `scene.*` 动作导致场景继续推进。

3. 这是布局约束先坏了，还是纯视觉样式变化？
- 如果是重叠、错位、溢出，先补或修约束测试。
- 如果只是样式调整，按设计评审后更新 golden。

## 常见坑

- 继续使用旧的根目录文档路径，而不是 `docs/` 下的新路径。
- 把桌面截图当正式测试证据。
- 只更新 golden，不更新 `docs/TEST-CASES.md`。
- 新增了用户可见界面状态，但没有新增 `scene.*`，导致无法稳定复现。
- 只跑 `go test ./...` 就宣称真实 GUI 回归通过。
- 在 snapshot 用例里塞太多状态断言，导致定位失败原因困难。

## 完成清单

- 已确认本次是视觉、功能、布局还是基础设施变更
- 已提供稳定的 `scene.*` 或动作入口
- 已按层更新功能测试 / UI 测试 / 约束测试
- 已按职责更新 `docs/TEST-API.md`、`docs/TEST-DESIGN.md`、`docs/TEST-CASES.md`、`docs/ISSUE.md`
- 已区分“预期设计变化”与“非预期回归”
- 已执行对应验证命令
