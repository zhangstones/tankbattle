# TankBattle Design

本文档只描述当前已经落地的代码设计、模块边界和实现取舍，不重复用户操作说明、测试命令或调试 API 字段细节。

- 运行与用户可见行为：见 `README.md`
- 测试分层与回归流程：见 `TEST-DESIGN.md`
- 调试 API 契约与调用示例：见 `TEST-API.md`

## 1. 设计目标

当前设计的目标有三点：

- 保持一个可直接运行的 Ebiten 桌面游戏主循环。
- 在不打碎核心状态聚合的前提下，逐步收敛模块边界。
- 让功能测试和界面快照测试都基于游戏内稳定接口，而不是依赖系统级截图或桌面自动化。

## 2. 当前目录结构

当前仓库按“薄入口 + 运行时模块 + 测试资产 + 文档”组织：

- `cmd/tankbattle`
  - GUI 入口，只负责调用运行时包。
- `cmd/sfxgen`
  - 音效资源生成工具，不参与主游戏状态机。
- `internal/tankbattle`
  - 运行时编排层和核心聚合根。
- `internal/audio`
  - 音频资源与音效播放实现。
- `internal/storage`
  - 设置和战绩历史的文件读写与清洗规则。
- `internal/debugapi`
  - 调试 API 控制器与本地 HTTP 请求收发。
- `internal/ui`
  - 主题原语、界面布局和最终渲染实现。
- `assets`
  - 嵌入式资源入口。
- `testing`
  - 功能测试、界面快照测试、golden 基线和测试工具。
- `docs`
  - 设计、测试、需求和问题跟踪文档。

## 3. 运行时装配

顶层运行从 `internal/tankbattle.RunWithOptions` 开始。

默认模式下：

- 加载用户设置和历史战绩。
- 注入真实音频管理器。
- 初始化 Ebiten 窗口。

调试模式下：

- 启用本地调试 API。
- 禁用用户本地配置和战绩读写。
- 关闭真实音频依赖。
- 固定随机种子，保证场景和快照稳定。
- 允许失焦时继续运行，便于自动化测试。

这个切换点的目的不是引入第二套游戏逻辑，而是在同一套运行时上切换“用户运行模式”和“可回归测试模式”。

## 4. 核心聚合根

`internal/tankbattle` 里仍然保留单一聚合根 `game`。它统一持有：

- 顶层状态：`state`、`paused`、`win`
- 实体集合：玩家、敌人、子弹、墙体、堡垒、爆炸、道具
- 对局数据：分数、波次、消息、buff 计时
- 菜单配置：难度、总波数、音效开关、音量、当前菜单项
- 历史数据：排行榜、滚动位置、历史面板开关
- 菜单恢复控制：`menuResumeAvailable`、`menuReturnState`、`menuReturnPaused`、`menuRequireRestart`
- 调试状态：`debug`、`debugFreeze`

当前只有三个顶层主状态：

- `stateMenu`
- `statePlaying`
- `stateEnded`

暂停、历史面板、消息框都作为叠加状态或表现层开关处理，而不是继续拆出更多主状态。

## 5. 业务逻辑拆分

`internal/tankbattle` 内部仍按职责拆文件，但不急于拆成大量小包：

- `setup.go`
  - 对局初始化、地图与障碍布局、敌军生成。
- `state.go`
  - `Update` 主时序、暂停、历史面板、胜负推进。
- `menu.go`
  - 菜单导航、配置变更、开始/恢复逻辑。
- `movement.go`
  - 玩家移动、双击转向、敌人移动和 AI 节奏。
- `combat.go`
  - 子弹推进、碰撞和伤害结算。
- `powerup.go`
  - 道具刷新、拾取、生效与过期。
- `settings.go`、`score_history.go`
  - 把 `internal/storage` 能力接回聚合根。
- `debug_api.go`
  - 把 `internal/debugapi` 请求映射成游戏动作和快照导出。

这层拆分的原则是：先把“时序”“规则”“渲染”“基础设施”分开，再考虑进一步拆包，而不是为了目录整洁过早引入跨包噪声。

## 6. UI 边界

这一轮迁移后，渲染层已经从 `internal/tankbattle` 抽到 `internal/ui`。

当前边界是：

- `internal/ui/render_theme.go`
  - 提供色板、描边、辉光、面板、胶囊标签、进度条等主题原语。
- `internal/ui/render.go`
  - 使用这些原语完成菜单、HUD、历史面板、消息框、暂停框、胜负框以及战场对象的最终绘制。
- `internal/ui/types.go`
  - 定义稳定的 UI 侧数据结构 `ui.Snapshot` 以及相关枚举和常量。
- `internal/tankbattle/ui_bridge.go`
  - 负责把运行时内部状态映射成 `ui.Snapshot`，并调用 `ui.Draw` / `ui.Layout`。

这里最关键的设计取舍是：`internal/ui` 不直接依赖 `game` 聚合根，而只消费一个稳定快照。

这样做的收益：

- 渲染可以独立演进，不必暴露大量运行时内部字段。
- 快照测试和布局测试可以围绕 `ui` 包单独收口。
- 后续继续拆 `internal/game` 时，UI 不需要再跟着一起大改。

## 7. 菜单与状态流转

菜单不是独立子系统，而是主状态机的一个入口。

关键流转如下：

- 菜单开始游戏：`menuStart -> startMatch() -> statePlaying`
- 对局中进菜单：`enterMenuForConfig() -> stateMenu`
- 菜单离开：`leaveMenuByToggle()`
- 对局结束：`finishMatch(win) -> stateEnded`

当前菜单恢复策略是明确的：

- 只修改音效开关或音量时，允许离开菜单后恢复原对局。
- 修改难度或总波数时，离开菜单后强制重开新对局。

这套策略集中由菜单恢复相关字段表达，避免把“恢复旧局”和“强制重开”的判断散落到渲染层或输入层。

## 8. 主循环时序

`Update()` 的时序保持以下顺序：

1. 处理调试请求和调试冻结。
2. 处理历史面板开关与收起。
3. 处理重开和菜单切换。
4. 按主状态分发菜单更新或结算态热键。
5. 处理暂停。
6. 推进消息和 buff 计时。
7. 更新玩家。
8. 更新敌人。
9. 更新子弹。
10. 更新道具。
11. 更新爆炸和墙体清理。
12. 处理随机道具生成。
13. 处理胜负和波次推进。

这样做的目的是让输入优先于战斗推进，同时保证一帧内尽量完成主要结算，减少状态延迟感。

## 9. 音频、存储与调试 API

三类基础能力都已经独立成包：

- `internal/audio`
  - 对外暴露音效播放器接口和真实实现。
- `internal/storage`
  - 负责设置与历史战绩的本地文件读写。
- `internal/debugapi`
  - 负责本地 HTTP 控制器和请求队列。

`internal/tankbattle` 只保留桥接逻辑：

- 音频通过 `audio_bridge.go` 接入聚合根。
- 存储通过 `settings.go` / `score_history.go` 回接。
- 调试请求通过 `debug_api.go` 在游戏主循环内消费。

这个边界保证了：

- 游戏主逻辑不关心文件格式和音频设备细节。
- 调试 API 不会直接跨线程改动游戏状态。
- 功能测试和界面测试都可以通过统一入口驱动。

## 10. 可测试性设计

当前测试能力建立在两层基础上：

- 游戏内调试 API
  - 负责状态查询、动作注入、命名场景切换和快照导出。
- UI 快照边界
  - 通过 `ui.Snapshot` 保持渲染输入稳定。

快照导出使用游戏内部 `Draw + ReadPixels`，而不是桌面截图。这样可以避免：

- 窗口被遮挡
- 焦点问题
- 系统缩放和桌面环境干扰

`scene.*` 命名场景配合 `debugFreeze` 使用，用于冻结帧推进，保证 golden 基线稳定。

## 11. 当前迁移状态

目前已经完成的重构阶段：

1. 根目录运行时代码和文档收口到 `internal/tankbattle` 与 `docs`。
2. 独立基础模块拆出：`internal/audio`、`internal/storage`、`internal/debugapi`。
3. UI 渲染层拆出到 `internal/ui`，通过 `ui.Snapshot` 与运行时解耦。

尚未完成的下一阶段是：

- 继续把核心玩法逻辑从 `internal/tankbattle` 收敛成更清晰的子模块边界。

但当前不会为了“看起来更整洁”而立即把玩法逻辑拆成大量细包。后续继续拆分的前提是：边界稳定、依赖清晰、回归成本可控。

## 12. 文档职责边界

- `README.md`
  - 用户如何运行、如何操作、有哪些用户可见行为。
- `TEST-DESIGN.md`
  - 测试分层、目录结构、回归策略、golden 流程。
- `TEST-API.md`
  - 调试 API 端点、字段、动作名和调用示例。
- `DESIGN.md`
  - 当前代码结构、模块边界和设计取舍。

如果只是接口字段变化，优先更新 `TEST-API.md`；如果只是回归流程变化，优先更新 `TEST-DESIGN.md`；只有代码结构或实现边界变化时，才更新 `DESIGN.md`。
