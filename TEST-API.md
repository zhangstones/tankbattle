# TankBattle 测试 API 说明

## 1. 目的

`TEST-API.md` 用于说明游戏内调试 API 的接口契约和使用方式。  
这套 API 主要服务于两类自动化测试：

- 功能测试：通过动作序列驱动菜单与游戏状态流转
- 界面测试：切到命名场景并直接导出游戏内快照

它不是正式对外产品接口，只面向本地开发、自动化测试和 CI。

## 2. 启用方式

启动游戏时设置环境变量 `TANKBATTLE_DEBUG_API_ADDR`：

```powershell
$env:TANKBATTLE_DEBUG_API_ADDR = "127.0.0.1:18080"
.\tankbattle_gui.exe
```

也可以直接运行源码：

```powershell
$env:TANKBATTLE_DEBUG_API_ADDR = "127.0.0.1:18080"
go run .\cmd\tankbattle
```

启用后：

- 游戏会启动本地 HTTP 调试服务
- 游戏使用固定随机种子
- 不读取也不写入用户本地 `settings.json` / `history.json`
- 快照由游戏主循环内直接导出

## 3. 基本约束

- API 只监听本地地址，由调用方自行决定端口
- 当前无认证，不应暴露到外网
- 所有动作都在游戏主循环内执行，避免与渲染线程分叉
- `scene.*` 场景会进入调试冻结态，避免背景动画和计时器污染快照
- 一旦执行非 `scene.*` 的真实动作，冻结态会取消

## 4. 接口列表

当前提供 3 个接口：

- `GET /debug/state`
- `POST /debug/actions`
- `POST /debug/snapshot`

### 4.1 `GET /debug/state`

用途：

- 查询当前游戏调试状态

示例：

```powershell
Invoke-RestMethod -Method Get -Uri "http://127.0.0.1:18080/debug/state"
```

返回示例：

```json
{
  "game_state": "menu",
  "menu_index": 0,
  "difficulty": "normal",
  "total_waves": 4,
  "sound_enabled": true,
  "sound_volume": 75,
  "paused": false,
  "show_history": false,
  "wave": 0,
  "max_wave": 0,
  "score": 0,
  "enemy_count": 0,
  "win": false,
  "menu_resume_available": false,
  "menu_require_restart": false,
  "message": ""
}
```

字段说明：

- `game_state`
  - `menu` / `playing` / `ended`
- `menu_index`
  - 当前菜单选中项索引
- `difficulty`
  - `easy` / `normal` / `hard`
- `total_waves`
  - 当前配置总波数
- `sound_enabled`
  - 音效开关
- `sound_volume`
  - 音量百分比，范围 `0-100`
- `paused`
  - 是否暂停
- `show_history`
  - 是否显示历史面板
- `wave`
  - 当前波次
- `max_wave`
  - 当前对局最大波次
- `score`
  - 当前分数
- `enemy_count`
  - 当前敌人数量
- `win`
  - 结束状态下是否胜利
- `menu_resume_available`
  - 当前菜单是否可返回原对局
- `menu_require_restart`
  - 当前菜单修改是否要求重开
- `message`
  - 当前消息框文案

### 4.2 `POST /debug/actions`

用途：

- 批量执行调试动作
- 返回执行后的最新 `DebugState`

请求体示例：

```json
{
  "actions": ["menu.down", "menu.right", "menu.start"]
}
```

PowerShell 示例：

```powershell
$body = @{
  actions = @("menu.down", "menu.right", "menu.start")
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Post `
  -Uri "http://127.0.0.1:18080/debug/actions" `
  -ContentType "application/json" `
  -Body $body
```

支持的动作分 3 类。

菜单动作：

- `menu.up`
- `menu.down`
- `menu.left`
- `menu.decrease`
- `menu.right`
- `menu.increase`
- `menu.start`
- `menu.easy`
- `menu.set_easy`
- `menu.normal`
- `menu.set_normal`
- `menu.hard`
- `menu.set_hard`

游戏动作：

- `game.enter_menu`
- `game.leave_menu`
- `game.start_match`
- `game.restart`
- `game.pause`
- `game.resume`
- `game.toggle_history`

场景动作：

- `scene.menu.default`
- `scene.menu.hard`
- `scene.menu.resume`
- `scene.hud.playing`
- `scene.hud.progressed`
- `scene.hud.shield`
- `scene.hud.history`
- `scene.pause`
- `scene.victory`
- `scene.defeat`

场景说明：

- `scene.menu.default`
  - 默认菜单基线
- `scene.menu.hard`
  - 难度切到 `hard` 的菜单
- `scene.menu.resume`
  - 从对局进入、可恢复原对局的菜单
- `scene.hud.playing`
  - 基础 HUD 场景
- `scene.hud.progressed`
  - 已推进波次和分数的 HUD 场景
- `scene.hud.shield`
  - 带 Shield / Rapid buff 的 HUD 场景
- `scene.hud.history`
  - 打开历史面板的 HUD 场景
- `scene.pause`
  - 暂停弹层
- `scene.victory`
  - 胜利结算弹层
- `scene.defeat`
  - 失败结算弹层

错误处理：

- 非法 JSON：返回 `400`
- 空动作列表：返回 `400`
- 当前状态不允许的动作：返回 `400`
- 不支持的动作名：返回 `400`

### 4.3 `POST /debug/snapshot`

用途：

- 直接从游戏渲染结果导出 PNG

请求体示例：

```json
{
  "dir": "D:\\Workspace\\tankbattle\\.tmp\\testing\\manual",
  "name": "menu-default.png"
}
```

PowerShell 示例：

```powershell
$body = @{
  dir  = "D:\Workspace\tankbattle\.tmp\testing\manual"
  name = "menu-default.png"
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Post `
  -Uri "http://127.0.0.1:18080/debug/snapshot" `
  -ContentType "application/json" `
  -Body $body
```

返回示例：

```json
{
  "path": "D:\\Workspace\\tankbattle\\.tmp\\testing\\manual\\menu-default.png"
}
```

路径规则：

- `dir` 必填
- `name` 必填
- `name` 不能包含路径分隔符
- 只允许 `.png` 扩展名
- 如果 `name` 无扩展名，会自动补 `.png`

## 5. 推荐使用方式

### 5.1 手工验证当前界面

```powershell
$env:TANKBATTLE_DEBUG_API_ADDR = "127.0.0.1:18080"
.\tankbattle_gui.exe
```

切到菜单场景并截图：

```powershell
$body = @{ actions = @("scene.menu.default") } | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:18080/debug/actions" -ContentType "application/json" -Body $body

$snap = @{
  dir  = "D:\Workspace\tankbattle\.tmp\testing\manual"
  name = "menu-default.png"
} | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:18080/debug/snapshot" -ContentType "application/json" -Body $snap
```

### 5.2 跑真实 GUI 功能与界面测试

```powershell
$env:TANKBATTLE_E2E = "1"
$env:TANKBATTLE_E2E_BINARY = "tankbattle_gui.exe"
go test ./testing/functional ./testing/ui
```

### 5.3 更新 golden 基线

```powershell
$env:TANKBATTLE_E2E = "1"
$env:TANKBATTLE_E2E_BINARY = "tankbattle_gui.exe"
$env:TANKBATTLE_UPDATE_GOLDEN = "1"
go test ./testing/ui
```

## 6. 与测试目录的关系

当前测试目录结构：

```text
/testing
  /testkit
  /functional
  /ui
  /testdata/golden

/.tmp/testing
```

说明：

- `testing/testkit` 负责 API client、E2E 启动器、diff 和 golden 管理
- `testing/functional` 通过动作和状态断言验证功能正确性
- `testing/ui` 通过命名场景和快照比对验证界面回归
- `.tmp/testing` 放失败产物和临时快照，不提交

## 7. 当前未覆盖能力

当前调试 API 还没有这些接口：

- `POST /debug/reset`
- `POST /debug/advance`
- 单独的 `POST /debug/scene`

现在的替代方案是：

- 通过 `scene.*` 动作完成快速场景切换
- 通过 `POST /debug/actions` 执行批量动作

如果后续测试需要更细粒度动画推进或状态复位，再单独扩展接口。
