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
- 菜单框已扩大并优化上下留白，避免顶部说明与选项区域拥挤/遮挡
- 菜单说明区与选项区间距改为按可用高度自动计算，避免“过远”或“挤压”
- 菜单标题与说明文本按区域居中计算，修复对齐偏移
- 标题在标题栏内按上下居中计算，`FIRE J/Space` 已合并到 Combat 说明行

## 主要功能

- 堡垒防守玩法：清空敌方波次并保护堡垒
- 菜单支持难度、波数、音效开关和音量配置
- 支持暂停、即时重开、菜单切换与对局恢复
- 包含敌方 AI、射击碰撞、道具增益与胜负结算
- HUD 显示核心战斗信息，历史战绩可通过 `H` 面板查看
- 音效资源内置打包：用户配置持久化到 `~/.tankbattle/settings.json`，历史战绩持久化到 `~/.tankbattle/history.json`
- Windows 版本运行时窗口使用自定义程序图标

## 测试

```powershell
go test ./...
```

当前包含单元测试与功能测试，覆盖菜单、状态切换、敌军生成、寻路/防抖、战斗结算、道具生命周期等主要逻辑。

## FAQ

- 常见问题见 [FAQ.md](FAQ.md)






