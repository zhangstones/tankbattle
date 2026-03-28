---
name: ui-testing
description: 为桌面游戏和界面密集型功能建立、维护和排查可回归的 UI 测试流程；优先使用游戏内调试 API 和快照导出，不依赖桌面截图或键盘模拟。
---

# UI 测试

## 适用场景

- 新增或修改菜单、HUD、弹层、历史面板、调试场景或快照用例。
- 需要补功能流转测试、黄金图回归或布局约束测试。
- 需要判断一次视觉变化是预期设计还是非预期回归。

## 不适用场景

- 纯逻辑改动且完全不影响 UI、调试 API 或测试入口。
- 只想人工看一眼界面，不需要建立可回归验证链路。
- 当前问题属于单纯代码解释或设计讨论，还没进入实现和验证。

## 核心目标

- 把界面验证拆成稳定分层，而不是所有问题都靠人工截图。
- 优先使用游戏内稳定入口复现状态，减少桌面环境噪声。
- 让功能回归、布局回归和视觉回归各有清晰责任边界。

## 本仓库默认入口

- 调试 API：
  - `internal/game/debug_api.go`
  - `internal/game/debug_api_test.go`
- 功能流转测试：
  - `testing/functional/debug_api_flow_test.go`
- 快照测试：
  - `testing/ui/snapshot_test.go`
- 文档：
  - `docs/TEST-API.md`
  - `docs/TEST-DESIGN.md`
  - `docs/TEST-CASES.md`

## 分层原则

1. 布局/约束测试
   - 用于发现对齐、溢出、重叠、列宽漂移等问题
2. 功能验收测试
   - 用于验证动作序列到状态结果是否正确
3. 快照回归测试
   - 用于验证固定场景下的视觉输出

规则：

- 功能测试断言状态，不比图片。
- 快照测试比图片，不塞大量状态断言。
- 人工桌面截图只用于排查，不作为正式回归依据。

## 开始前确认

- 这次改动影响的是布局、视觉样式、交互流转，还是测试基础设施。
- 当前是否已有稳定 `scene.*` 或等价调试入口。
- 现有 golden 是否仍代表正确设计。

## 标准流程

1. 先判断变更类型：布局、功能、视觉、基础设施。
2. 若缺少稳定入口，先补调试 API 场景或动作入口。
3. 按分层补测试：
   - 交互流转改动：补功能测试
   - 视觉输出改动：补 snapshot / golden
   - 对齐和边界问题：补布局约束测试
4. 按职责更新测试文档，而不是把所有文档都改一遍。
5. 运行对应验证命令。

## 验证命令

基础回归：

```powershell
go test ./...
```

真实 GUI 功能 / 快照回归：

```powershell
$env:TANKBATTLE_E2E = "1"
$env:TANKBATTLE_E2E_BINARY = "tankbattle_gui.exe"
go test ./testing/functional ./testing/ui
```

更新 golden：

```powershell
$env:TANKBATTLE_E2E = "1"
$env:TANKBATTLE_E2E_BINARY = "tankbattle_gui.exe"
$env:TANKBATTLE_UPDATE_GOLDEN = "1"
go test ./testing/ui
```

## 文档联动

- 接口字段、动作名、场景名变化：更新 `docs/TEST-API.md`
- 测试分层、目录、命令变化：更新 `docs/TEST-DESIGN.md`
- 测试案例矩阵变化：更新 `docs/TEST-CASES.md`
- 用户可见运行方式变化时，再考虑更新 `README.md`

## 快速判定

- snapshot mismatch：
  - 先判断是否为预期设计变化
  - 再判断场景是否不稳定
  - 最后区分是布局约束先坏了，还是只是样式变化
- 功能测试失败：
  - 先看动作入口和状态断言是否仍与契约一致
  - 再看是否因场景构造不稳定导致误报

## 常见错误

- 只更新 golden，不更新对应测试或文档。
- 只跑 `go test ./...` 就宣称真实 GUI 回归通过。
- 把桌面截图当正式证据。
- 新增用户可见状态，却没有稳定调试入口。
- 在快照测试里塞过多行为断言，导致失败定位困难。

## 完成标准

- 已明确本次改动对应的测试分层。
- 已提供稳定复现场景或动作入口。
- 已按职责更新测试、golden 和相关测试文档。
- 已区分预期设计变化与真实回归，并完成对应验证。
