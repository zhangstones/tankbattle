# Tank Battle (Windows, Golang)

基于 `Go + Ebiten` 的 Windows 坦克大战，核心目标是守住堡垒并清空敌方波次。

## 环境要求

- Windows
- Go 1.23+

## 运行方式

开发运行：

```powershell
go run .\cmd\tankbattle
```

构建普通可执行文件：

```powershell
go build -o tankbattle.exe .\cmd\tankbattle
.\tankbattle.exe
```

构建 GUI 版本（无命令行弹窗，推荐）：

```powershell
go build -ldflags="-H windowsgui" -o tankbattle_gui.exe .\cmd\tankbattle
.\tankbattle_gui.exe
```

## 操作说明

- `WASD` / `方向键` 长按：平移移动（不自动转向）
- `WASD` / `方向键` 双击同方向：转向并同步炮塔朝向
- `J` / `Space`：开火
- `P`：暂停 / 继续
- `M`：进入菜单；在菜单中再次按 `M` 返回上一状态
- `R`：即时重开
- `H`：显示 / 隐藏历史战绩面板（滚轮或 `PgUp/PgDn` 滚动）
- `H` 切换历史面板时不弹出提示框
- 历史面板打开后，按任意功能键会自动隐藏（如 `P/M/R/J/方向键/WASD/Enter/Space`）
- 历史面板默认展示 10 条记录，包含分数、对局时长与本地时间

## 菜单说明

- `↑/↓`：选择菜单项
- `←/→`：调整当前项
- `Enter` / `Space`：开始游戏
- `1/2/3`：快捷设置难度
- `Total Waves`：通过菜单 `←/→` 调整总波数（`1~5`）
- 菜单中的 `Sound Effects` 项可通过 `←/→` 开关音效
- 菜单中的 `SFX Volume` 项可通过 `←/→` 调整音量（0~100%，步进 25%）
- 对局中通过 `M` 进入菜单后，若只改音量/声效，可按 `M` 直接恢复原对局（含暂停状态）
- 若改了 `Difficulty` 或 `Total Waves`，按 `M` 返回时会自动重开新对局
- 主菜单已重做为任务控制台风格，包含标题主视觉、任务摘要、快捷键区和更明显的选中态
- 每个菜单项都带状态标签或进度条，方便快速确认当前难度、波数和音量设定
- 标题与说明文本仍按区域居中计算，`FIRE J/Space` 已保留在战斗说明行中

## 主要功能

- 堡垒防守玩法：清空敌方波次并保护堡垒
- 菜单支持难度、波数、音效开关和音量配置
- 支持暂停、即时重开、菜单切换与对局恢复
- 包含敌方 AI、射击碰撞、道具增益与胜负结算
- HUD 已重做为指挥面板风格，强化波次、分数、敌人数、堡垒血量与玩家状态的层次
- 战场背景、墙体、堡垒、坦克、道具、子弹和爆炸特效均已增强，提升整体质感但不改变玩法
- 历史战绩可通过 `H` 面板查看，面板样式已与主菜单/HUD 统一
- 音效资源内置打包：用户配置持久化到 `~/.tankbattle/settings.json`，历史战绩持久化到 `~/.tankbattle/history.json`
- Windows 版本运行时窗口使用自定义程序图标

## 测试

```powershell
go test ./...
```

默认测试覆盖单元测试、布局约束测试和调试 API 基础测试，包含菜单、状态切换、敌军生成、寻路/防抖、战斗结算、道具生命周期与调试动作解析等主要逻辑。

如需运行基于调试 API 的真实功能 / 界面测试套件：

```powershell
$env:TANKBATTLE_E2E = "1"
go test ./tests/functional ./tests/ui
```

如需更新界面 golden 快照：

```powershell
$env:TANKBATTLE_E2E = "1"
$env:TANKBATTLE_UPDATE_GOLDEN = "1"
go test ./tests/ui
```

测试临时产物会写入 `.tmp_test_artifacts/`，快照对比失败时会输出到 `testdata/failures/`。

## 调试 API / 功能测试

可通过环境变量启动本地调试 API，供功能测试脚本直接驱动菜单操作并导出界面快照：

```powershell
$env:TANKBATTLE_DEBUG_API_ADDR = "127.0.0.1:18080"
.\tankbattle_gui.exe
```

启用后：

- 游戏使用固定随机种子启动，避免菜单/HUD/地图快照漂移
- 不读取用户本地设置，也不写入 `settings.json` / `history.json`
- 调试请求在游戏主循环内执行，快照直接由游戏自身导出 PNG

可用接口：

- `GET /debug/state`
  - 返回当前 `game_state`、`menu_index`、`difficulty`、`wave`、`score` 等调试状态
- `POST /debug/actions`
  - JSON 示例：

```json
{
  "actions": ["menu.down", "menu.right", "menu.start"]
}
```

- `POST /debug/snapshot`
  - JSON 示例：

```json
{
  "dir": "D:\\Workspace\\tankbattle\\snapshots",
  "name": "menu-after-toggle.png"
}
```

当前支持的动作包括：

- `menu.up`
- `menu.down`
- `menu.left` / `menu.decrease`
- `menu.right` / `menu.increase`
- `menu.start`
- `menu.easy` / `menu.set_easy`
- `menu.normal` / `menu.set_normal`
- `menu.hard` / `menu.set_hard`
- `game.enter_menu`
- `game.leave_menu`
- `game.start_match`
- `game.restart`
- `game.pause`
- `game.resume`
- `game.toggle_history`
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

## FAQ

- 常见问题见 [FAQ.md](FAQ.md)






