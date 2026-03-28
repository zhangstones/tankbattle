# TankBattle 测试方案设计

## 1. 目标

本文档定义 TankBattle 后续功能测试与界面测试的统一方案，目标是：

- 提高菜单、状态切换、战斗流程等核心行为的可回归性
- 提高菜单、HUD、弹层等界面的可验证性
- 让测试不依赖人工截图、桌面焦点或键盘模拟
- 让测试结果可重复、可在 CI 中稳定运行

本文档只关注测试方案，不重复用户操作说明与运行手册。

## 2. 设计结论

测试框架主方案：

- 语言：`Go`
- 主测试入口：`go test`
- 驱动方式：游戏内调试 API
- 快照方式：由游戏主循环内直接导出 PNG
- 测试分层：功能测试 + 界面快照测试 + 布局约束测试

不采用主方案：

- 不以桌面截图作为正式测试入口
- 不以键盘模拟作为主要自动化手段
- 当前阶段不以 `Python/pytest` 作为主测试框架

## 3. 为什么主选 Go 而不是 Python/pytest

### 3.1 选择 Go 的原因

- 项目本身是 Go，测试可直接复用常量、状态枚举、调试结构与 helper。
- 现有测试已经基于 `go test`，沿用成本最低，集成最自然。
- HTTP 调试 API、JSON、PNG 读写、文件比对都可直接用标准库完成。
- 功能测试与界面测试可以共用一套 `testkit`，无需额外维护 Python 运行时与依赖。
- CI 中只要有 Go 和 Windows GUI 运行环境即可，不必再维护多语言工具链。

### 3.2 为什么当前不主选 Python/pytest

`Python/pytest` 并不是不能用，但当前收益不高：

- 会引入第二套运行时、依赖安装和环境排障成本。
- 需要在 Go 项目外再维护一套 API client、快照工具和断言体系。
- 对当前桌面游戏来说，真正的稳定入口是“游戏内调试 API”，不是 Python 本身。
- 如果只为了写测试更快，短期可能有利；但长期会增加协议变更和工具链维护成本。

### 3.3 Python 适合的补充角色

后续如有需要，可让 Python 作为辅助工具，而不是主测试框架：

- 批量生成 diff 报告
- 汇总快照结果生成 HTML
- 本地测试数据整理脚本

结论：

- 主测试框架仍应留在 Go
- Python 只作为附属工具可选引入

## 4. 测试入口设计

### 4.1 游戏内调试 API

测试应通过调试 API 驱动游戏，而不是走真实键盘模拟。

已有/计划中的接口能力：

- `GET /debug/state`
  - 查询当前状态，例如 `game_state`、`menu_index`、`difficulty`、`wave`、`score`、`paused`
- `POST /debug/actions`
  - 批量执行菜单动作与状态动作
- `POST /debug/snapshot`
  - 由游戏自身导出当前界面 PNG 到指定目录

建议继续补充：

- `POST /debug/advance`
  - 推进指定帧数，用于等待消息框、波次切换、动画稳定
- `POST /debug/reset`
  - 将游戏重置到确定性初始状态
- `POST /debug/scene`
  - 直接切换到命名测试场景，例如 `menu-default`、`hud-playing`、`pause-panel`

### 4.2 为什么不能把系统截图作为正式入口

- 桌面截图依赖窗口焦点
- 会被其他窗口遮挡
- 不同缩放、前后台状态会干扰结果
- 无法保证 CI 环境稳定复现

因此：

- 系统截图只可用于临时人工排查
- 正式自动化测试必须走游戏内快照导出

## 5. 确定性要求

为了让快照和功能测试结果可回归，调试模式必须满足：

- 固定随机种子
- 不读取用户本地 `settings.json`
- 不写入用户本地 `settings.json` / `history.json`
- 快照在主循环内导出
- 同一组动作序列可重复得到相同状态和近似一致的图像结果

这是功能测试与界面测试能稳定落地的前提。

## 6. 测试分层

### 6.1 单元/约束测试

目标：

- 快速验证小范围逻辑与布局规则

当前应继续保留并扩展：

- 菜单边界行为
- 音量/难度上下界
- 状态切换规则
- HUD / 菜单布局常量约束
- 文本列宽与面板不重叠约束
- 调试 API 的动作解析、路径校验、确定性初始化

特点：

- 运行快
- 定位精确
- 不依赖 GUI 窗口

### 6.2 功能测试

目标：

- 验证“动作序列 -> 游戏状态”是否正确

核心原则：

- 只断言状态和业务结果
- 不直接比对图像
- 所有动作都通过调试 API 发送

优先覆盖的功能测试：

- 菜单导航：`up/down/left/right/start`
- 菜单配置：难度、总波数、音效、音量
- 状态流转：菜单 -> 对局 -> 暂停 -> 恢复 -> 结束
- 菜单恢复策略：
  - 只改音频时恢复原对局
  - 改难度/波数时重开
- 历史面板打开/关闭/滚动
- `R` 重开与 `M` 返回菜单行为

### 6.3 界面快照测试

目标：

- 验证菜单、HUD、状态弹层、历史面板的视觉回归

核心原则：

- 通过调试 API 进入指定场景
- 通过游戏内快照导出 PNG
- 与 golden image 比较

第一批建议覆盖的场景：

- `menu-default`
- `menu-hard`
- `menu-resume-available`
- `hud-playing-default`
- `hud-with-shield`
- `hud-history-open`
- `pause-panel`
- `victory-panel`
- `defeat-panel`

## 7. 图像回归策略

不建议一开始做“全屏严格逐像素一致”。

推荐分级策略：

### 7.1 区域优先

优先比较关键区域：

- 菜单主区
- HUD 区
- 状态弹层区
- 历史面板区

这样可以减少背景细节波动带来的噪声。

### 7.2 允许轻微阈值

若后续引入更明显的动态效果，可设置轻微容差，例如：

- 单像素颜色偏差阈值
- 总差异像素比例阈值

但当前 UI 仍以静态绘制为主，应尽量保持低波动。

### 7.3 失败产物

快照测试失败时应输出：

- 实际图
- golden 图
- diff 图

这样便于快速判断是设计变更还是回归问题。

## 8. 目录建议

建议新增如下结构：

```text
/testing/testkit
  client.go
  launcher.go
  snapshot.go
  diff.go

/testing/functional
  menu_flow_test.go
  state_flow_test.go
  pause_history_test.go

/testing/ui
  menu_snapshot_test.go
  hud_snapshot_test.go
  panel_snapshot_test.go

/testing/testdata/golden
  /menu
  /hud
  /panels
```

说明：

- `testkit` 负责公共能力，不承载具体业务断言
- `testing/functional` 只看状态，不看图片
- `testing/ui` 只负责场景构建、快照导出和回归比较

## 9. testkit 责任划分

建议 `testkit` 提供这些能力：

- 启动带调试 API 的游戏实例
- 等待 API 可用
- 执行动作序列
- 查询状态
- 导出快照
- 进行图片 diff
- 失败时输出调试工件路径

这样每个测试文件只需要表达：

- 进入什么场景
- 执行什么动作
- 断言什么状态
- 或者与哪张 golden 比较

## 10. 典型测试流程

### 10.1 功能测试流程

1. 启动游戏调试实例
2. 重置到固定初始状态
3. 发送动作序列
4. 查询状态
5. 断言状态与数值

### 10.2 UI 快照测试流程

1. 启动游戏调试实例
2. 切换到命名场景或执行动作序列
3. 必要时推进若干帧
4. 导出快照到临时目录
5. 与 golden image 比对
6. 失败时输出 diff 图

## 11. CI 策略

建议分层执行：

### 11.1 常规提交

每次提交默认执行：

- `go test ./...`
- 布局约束测试
- 调试 API 功能测试

### 11.2 Windows UI 回归

在 Windows runner 上执行：

- 快照测试
- golden 对比
- 失败工件上传

原因：

- UI 快照导出与桌面窗口环境更贴近 Windows 目标平台

## 12. golden 更新策略

需要提供明确流程，避免把回归误当作“正常更新”：

建议规则：

- 只有明确界面设计变更时才更新 golden
- golden 更新应单独提交或单独列出
- PR 中应说明：
  - 为什么 golden 变化
  - 变化是否为预期设计调整

建议后续补一个辅助命令：

- `go test ./testing/ui -update`

作用：

- 在确认设计变更后重写 golden 图

## 13. 分阶段落地顺序

建议按以下顺序推进：

### Phase 1

- 稳定调试 API：`state/actions/snapshot`
- 完善确定性初始化
- 补基础 API 单元测试

### Phase 2

- 增加 `testkit`
- 落地第一批功能测试

### Phase 3

- 增加 UI 快照测试
- 建立第一批 golden 图

### Phase 4

- 补 `advance/reset/scene`
- 扩大覆盖面并接入 CI 工件回传

## 14. 风险与控制

### 14.1 风险：调试接口和真实逻辑分叉

控制方式：

- 调试动作尽量复用已有状态机和菜单逻辑
- 不单独维护一套测试专用菜单实现

### 14.2 风险：快照结果漂移

控制方式：

- 固定随机种子
- 禁止用户本地配置干扰
- 优先测试静态稳定场景

### 14.3 风险：测试过于依赖整屏图像

控制方式：

- 先做状态断言
- 再做关键区域快照
- 最后才扩大到更全面的图像回归

## 15. 当前文档职责说明

本次只新增 `TEST-DESIGN.md`：

- 不更新 `README.md`，因为这里不是用户操作手册
- 不扩写 `DESIGN.md` 的测试细节，避免与整体架构文档重复

后续若测试框架正式落地并形成稳定命令，再将“如何运行测试”补充到 `README.md`。
## 16. 2026-03-28 落地总结

当前测试方案已经完成首轮落地，仓库中的测试目录统一为：

```text
/testing
  /testkit
  /functional
  /ui
  /testdata/golden

/.tmp/testing
```

目录职责：

- `testing/testkit`
  - 放调试 API client、E2E 启动器、快照 diff 和 golden 更新辅助逻辑
- `testing/functional`
  - 放通过调试 API 驱动真实游戏进程的功能测试
- `testing/ui`
  - 放命名场景快照和 golden 回归测试
- `testing/testdata/golden`
  - 放需要提交的快照基线
- `.tmp/testing`
  - 放运行期临时产物、失败 diff 图和临时快照，不提交

### 16.1 命令分层

当前建议固定区分三层命令：

1. 默认回归
   - `go test ./...`
   - 覆盖单元测试、布局约束测试和调试 API 基础测试
2. 真实 GUI 功能与界面回归
   - `$env:TANKBATTLE_E2E='1'; $env:TANKBATTLE_E2E_BINARY='tankbattle_gui.exe'; go test ./testing/functional ./testing/ui`
3. golden 更新
   - `$env:TANKBATTLE_E2E='1'; $env:TANKBATTLE_E2E_BINARY='tankbattle_gui.exe'; $env:TANKBATTLE_UPDATE_GOLDEN='1'; go test ./testing/ui`

这样分层的原因是：

- `go test ./...` 需要保持在无 GUI 会话时也能运行
- 真实 GUI 快照回归必须显式启用 `TANKBATTLE_E2E=1`
- golden 更新必须和普通回归分离，避免误把回归问题直接覆盖成新基线

### 16.2 场景稳定性策略

当前 UI 快照依赖命名场景 `scene.*`。实践结论是，命名场景必须进入“调试冻结态”：

- 切到 `scene.*` 后，不继续推进 `frame` / `audioFrame`
- 防止背景动画、pulse 效果、buff 倒计时等随帧变化的元素污染快照
- 一旦执行非 `scene.*` 的真实动作，冻结态取消，不影响功能测试

这条策略已经被证明是 UI 快照稳定回归的前提。

### 16.3 当前回归经验

这次落地后得到两个明确经验：

- 仅跑 `go test ./...` 不等于完成真实界面回归
- 真实 GUI 套件要单独作为验收入口执行，否则菜单 / HUD 视觉变化可能不会被及时发现

当前已知的菜单快照基线问题已经单独记录在 `ISSUE.md`，后续处理时应先判断：

- 是预期设计变化，需要审核后更新 golden
- 还是非预期回归，需要先修复渲染再回归

### 16.4 文档边界

当前文档边界约定如下：

- `README.md`
  - 只保留测试入口、运行命令和面向使用者的最小说明
- `TEST-DESIGN.md`
  - 只描述测试目标、分层策略、目录结构、回归流程和设计取舍
- `TEST-API.md`
  - 统一维护调试 API 的接口契约、请求示例、动作列表、场景列表和使用说明

后续如果调试 API 的字段、动作名、场景名或调用方式发生变化，应优先更新 `TEST-API.md`；如果测试分层、目录组织或回归流程发生变化，再更新 `TEST-DESIGN.md`。

### 16.5 仓库布局（2026-03-28）

本轮目录整理后，测试方案依赖的仓库布局补充如下：

- 运行时代码集中到 `internal/tankbattle/`，继续保持单个运行时包，避免目录收口时同时引入多包耦合。
- 测试代码和 golden 基线继续集中在 `testing/`，对外测试命令不变。
- 设计、需求、测试和问题跟踪文档集中到 `docs/`，`README.md` 只保留面向使用者的最小入口说明。

这次调整的影响边界是：

- `go test ./...`、`go build ./...` 和 `go build ./cmd/tankbattle` 仍然是标准入口。
- 测试框架不依赖运行时代码位于仓库根目录，只依赖模块根目录下的 `go.mod`、`cmd/tankbattle` 和 `testing/` 结构。
