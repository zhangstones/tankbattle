---
name: launchgui
description: 通过 guilauncher 将 GUI 程序启动到用户桌面会话；优先复用现有配置，不重复创建。
---

# Launch GUI 到用户会话

## 目标

当代理会话无法直接显示 GUI 时，通过 `D:\Workspace\guilauncher` 的触发链路，把程序启动到用户当前桌面会话，并保持非阻塞。

## 必做检查

1. 目录存在：`D:\Workspace\guilauncher`
2. 文件存在：
- `D:\Workspace\guilauncher\launchgui.ps1`
- `D:\Workspace\guilauncher\config.json`
3. 计划任务可查询（可选）：`schtasks /query /tn guilauncher`

规则：
- 已存在则复用，不得重复创建同名目录/脚本/任务。
- 仅在缺失时补齐最小必要项。

## 技能归档内容

本技能已内置可迁移模板，供新环境快速部署：

- 脚本模板：`skills/launchgui/templates/launchgui.ps1`
- 配置模板：`skills/launchgui/templates/config.json`
- 触发模板：`skills/launchgui/templates/run.json`
- 一键部署：`skills/launchgui/scripts/bootstrap.ps1`

## 关键约束

1. `run.json` 只允许传 `app`，不允许传任意 `exe` 路径。
2. `config.json` 必须是合法 JSON，Windows 路径要写成双反斜杠，例如：
- `D:\\Workspace\\tankbattle\\tankbattle_gui.exe`
3. 白名单里不要加入 `cmd.exe`、`powershell.exe` 等通用解释器。

## 标准流程

1. 确认 `config.json` 的 `apps` 中存在目标应用（例如 `tankbattle`）。
2. 写入 `run.json`：

```powershell
@'
{
  "app": "tankbattle"
}
'@ | Set-Content -Path D:\Workspace\guilauncher\run.json -Encoding UTF8
```

3. 优先触发任务：

```powershell
schtasks /run /tn "guilauncher"
```

4. 若当前环境 `schtasks` 不可用，则回退到直调脚本：

```powershell
powershell.exe -ExecutionPolicy Bypass -File D:\Workspace\guilauncher\launchgui.ps1
```

5. 查看日志：

```powershell
Get-Content D:\Workspace\guilauncher\logs\launcher.log -Tail 30
```

## 新环境快速部署

在新环境第一次使用时：

```powershell
powershell -ExecutionPolicy Bypass -File .\skills\launchgui\scripts\bootstrap.ps1
```

说明：
- `bootstrap.ps1` 只在缺失时创建 `launchgui.ps1/config.json`，已有文件会保留（复用原则）。
- 会确保 `guilauncher` 任务存在。

## 重启模式（restart）

当需要“先关闭旧实例再拉起新实例”时，在 `run.json` 中增加：

```json
{
  "app": "tankbattle",
  "restart": true
}
```

要求：
- `config.json` 中应为该应用配置 `processNames`，用于精确匹配待关闭进程。
- 日志中应出现 `restart requested` 与 `restart: ...` 记录。

## 验证标准

- 日志出现 `launch success app=<appId>`。
- 若使用 `restart:true`，日志应出现重启流程记录后再 `launch success`。
- 触发过程不阻塞当前代理前台操作。
- 用户桌面会话中可见目标 GUI（以用户可见结果为准）。

## 完备性检查（新环境）

以下检查全部通过，视为技能在新环境“完备可用”：

1. `bootstrap.ps1` 可执行且无报错。
2. `D:\Workspace\guilauncher\launchgui.ps1` 与 `config.json` 存在。
3. `schtasks /query /tn guilauncher` 可返回任务。
4. 写入 `run.json` 后，触发一次可在日志看到 `launch success`。

