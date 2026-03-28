# TankBattle 测试案例清单

## 1. 目的

`TEST-CASES.md` 用于维护当前项目的“功能验收案例”和“界面回归案例”清单，作为测试范围的单点索引。

职责边界：

- 这里记录“测什么、怎么触发、验什么、当前自动化落点在哪”。
- 接口契约、动作名称、快照导出规则统一维护在 `TEST-API.md`。
- 测试分层、golden 策略、目录设计统一维护在 `TEST-DESIGN.md`。
- 临时问题、待处理缺口统一记录在 `ISSUE.md`。

维护规则：

- 只要新增或修改用户可见行为、调试场景、菜单流程、HUD/弹层视觉，就必须同步更新本文件。
- 优先更新已有案例；只有出现新的用户场景时才新增案例编号。
- 案例编号长期稳定，避免因顺序调整频繁改号。

## 2. 执行入口

默认回归：

```powershell
go test ./...
```

真实 GUI 功能 / 界面回归：

```powershell
$env:TANKBATTLE_E2E = "1"
$env:TANKBATTLE_E2E_BINARY = "tankbattle_gui.exe"
go test ./testing/functional ./testing/ui
```

更新界面 golden：

```powershell
$env:TANKBATTLE_E2E = "1"
$env:TANKBATTLE_E2E_BINARY = "tankbattle_gui.exe"
$env:TANKBATTLE_UPDATE_GOLDEN = "1"
go test ./testing/ui
```

## 3. 功能验收案例

| 编号 | 场景 | 触发方式 | 关键断言 | 自动化落点 |
| --- | --- | --- | --- | --- |
| FA-001 | 菜单默认配置基线 | `scene.menu.default` | 状态为 `menu`；难度 `normal`；总波数 `4`；音效开启；音量 `75` | `testing/functional/debug_api_flow_test.go` `TestDebugAPIMenuConfigurationFlow` |
| FA-002 | 菜单难度预设切换 | `menu.hard` | 难度切到 `hard`；总波数联动到 `5` | `testing/functional/debug_api_flow_test.go` `TestDebugAPIMenuConfigurationFlow` |
| FA-003 | 菜单音效开关 | `menu.down` x2, `menu.right` | 光标位于音效项；音效从开变关 | `testing/functional/debug_api_flow_test.go` `TestDebugAPIMenuConfigurationFlow` |
| FA-004 | 菜单音量调整 | 在音效项基础上 `menu.down`, `menu.left` | 光标位于音量项；音量从 `75` 降到 `50` | `testing/functional/debug_api_flow_test.go` `TestDebugAPIMenuConfigurationFlow` |
| FA-005 | 对局中间态基线 | `scene.hud.progressed` | 状态为 `playing`；波次 `3`；分数 `275` | `testing/functional/debug_api_flow_test.go` `TestDebugAPIPauseHistoryAndOutcomeFlow` |
| FA-006 | 暂停与恢复 | `game.pause`, `game.resume` | `paused` 可正确切换为 `true/false` | `testing/functional/debug_api_flow_test.go` `TestDebugAPIPauseHistoryAndOutcomeFlow` |
| FA-007 | 历史面板开关 | `game.toggle_history` 两次 | `show_history` 可正确切换为 `true/false` | `testing/functional/debug_api_flow_test.go` `TestDebugAPIPauseHistoryAndOutcomeFlow` |
| FA-008 | 胜利/失败结算状态 | `scene.victory`, `scene.defeat` | 状态为 `ended`；`win` 分别为 `true/false` | `testing/functional/debug_api_flow_test.go` `TestDebugAPIPauseHistoryAndOutcomeFlow` |
| FA-009 | 菜单返回原对局 | `scene.hud.progressed`, `game.enter_menu`，只改音频项后 `game.leave_menu` | 回到 `playing`；保留波次 `3` 和分数 `275`；清除 resume 标记 | `testing/functional/debug_api_flow_test.go` `TestDebugAPIMenuResumeVsRestart` |
| FA-010 | 菜单触发重开 | `scene.hud.progressed`, `game.enter_menu`, `menu.hard`, `game.leave_menu` | 回到 `playing`；难度保留 `hard`；波次重置为 `1`；分数重置为 `0` | `testing/functional/debug_api_flow_test.go` `TestDebugAPIMenuResumeVsRestart` |

## 4. 界面回归案例

| 编号 | 场景 | 调试场景 | 视觉检查点 | golden |
| --- | --- | --- | --- | --- |
| UI-001 | 菜单默认态 | `scene.menu.default` | 标题区、配置区、右侧说明区、选中态、边框对齐 | `testing/testdata/golden/menu/menu-default.png` |
| UI-002 | 菜单高难度态 | `scene.menu.hard` | 难度/波数联动后的菜单视觉、状态标签、侧栏摘要 | `testing/testdata/golden/menu/menu-hard.png` |
| UI-003 | 菜单可恢复态 | `scene.menu.resume` | 从对局进入菜单后的恢复提示与菜单布局 | `testing/testdata/golden/menu/menu-resume.png` |
| UI-004 | HUD 基础战斗态 | `scene.hud.playing` | 顶部状态区、右侧状态卡、底部提示、战场主体 | `testing/testdata/golden/hud/hud-playing.png` |
| UI-005 | HUD Buff 态 | `scene.hud.shield` | 护盾/连发状态卡、玩家状态区和右侧信息对齐 | `testing/testdata/golden/hud/hud-shield.png` |
| UI-006 | HUD 历史面板态 | `scene.hud.history` | HUD 与历史面板共存时的层级、遮罩、文本列宽 | `testing/testdata/golden/hud/hud-history.png` |
| UI-007 | 暂停弹层 | `scene.pause` | 暂停面板、背景压暗、边框和按钮文案布局 | `testing/testdata/golden/panels/pause-panel.png` |
| UI-008 | 胜利结算弹层 | `scene.victory` | 胜利标题、统计信息、底部操作提示 | `testing/testdata/golden/panels/victory-panel.png` |
| UI-009 | 失败结算弹层 | `scene.defeat` | 失败标题、统计信息、底部操作提示 | `testing/testdata/golden/panels/defeat-panel.png` |

当前自动化入口：

- `testing/ui/snapshot_test.go` `TestDebugUISnapshots`

## 5. 维护约定

新增或修改测试时，按下面规则同步维护：

1. 改了调试动作、场景名、状态字段：更新 `TEST-API.md`，并检查本文件中的触发方式和断言描述。
2. 改了测试分层、目录、golden 更新流程：更新 `TEST-DESIGN.md`，并检查本文件中的执行入口和自动化落点。
3. 改了用户可见菜单/HUD/弹层行为：更新本文件对应案例；若视觉发生预期变化，再更新 golden。
4. 新增一个真实用户场景但没有案例编号时：先补本文件，再补自动化测试。

## 6. 当前覆盖缺口

- 真实 GUI E2E 仍不是默认 `go test ./...` 的强制路径；这属于流程缺口，当前状态见 `ISSUE.md`。
- 当前功能验收主要覆盖调试 API 驱动下的菜单流转、暂停/历史、结算与恢复策略；真实战斗过程中的碰撞、AI、掉落等仍主要依赖现有单元测试和约束测试。
