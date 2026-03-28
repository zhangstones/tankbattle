---
name: launchgui
description: 通过 `guilauncher` 链路把 Windows GUI 程序启动到用户桌面会话；优先复用已有部署，不重复创建环境。
---

# Launch GUI

## 适用场景

- 当前代理会话无法直接展示 GUI，需要把程序拉起到用户桌面。
- 需要验证 GUI 可执行文件是否能真正启动，而不是只构建成功。
- 需要在不阻塞当前终端前台的前提下启动程序。

## 不适用场景

- 只需要执行 CLI 程序或后台服务。
- 当前环境本来就能直接显示 GUI，且无需桌面会话中转。
- 用户未授权的高风险 GUI 启动或重启操作。

## 默认实现

本仓库默认复用 `guilauncher`：

- 默认基础目录：`D:\Workspace\guilauncher`
- 默认任务名：`guilauncher`
- 技能内置模板：
  - `skills/launchgui/templates/launchgui.ps1`
  - `skills/launchgui/templates/config.json`
  - `skills/launchgui/templates/run.json`
- 初始化脚本：
  - `skills/launchgui/scripts/bootstrap.ps1`

若新环境路径不同，应通过 `bootstrap.ps1` 参数或实际部署路径调整，而不是硬编码假设仍是默认目录。

## 开始前确认

- 目标程序的可执行文件路径。
- 程序的稳定应用标识，例如 `tankbattle`。
- 若需要重启旧实例，是否已配置对应 `processNames`。
- 当前环境是否已有 `guilauncher` 部署与计划任务。

## 关键约束

- `run.json` 只传受控的 `app` 标识，不直接暴露任意 `exe` 路径。
- 白名单配置中不要加入 `cmd.exe`、`powershell.exe` 等通用解释器。
- 已存在的 `guilauncher` 目录、脚本、任务要优先复用，不重复创建平行实现。
- 只有在用户明确需要重启时，才使用 `restart: true`。

## 标准流程

1. 先检查部署是否存在：
   - `launchgui.ps1`
   - `config.json`
   - 可选的计划任务 `guilauncher`
2. 若缺失，用 `bootstrap.ps1` 补齐最小必要项。
3. 确认 `config.json` 的 `apps` 中已注册目标应用。
4. 写入 `run.json`，至少包含：

```json
{
  "app": "tankbattle"
}
```

5. 优先用计划任务触发：

```powershell
schtasks /run /tn "guilauncher"
```

6. 若当前环境不可用 `schtasks`，回退到直接执行：

```powershell
powershell.exe -ExecutionPolicy Bypass -File D:\Workspace\guilauncher\launchgui.ps1
```

7. 查看日志确认实际结果。

## 新环境初始化

首次部署可执行：

```powershell
powershell -ExecutionPolicy Bypass -File .\skills\launchgui\scripts\bootstrap.ps1
```

可选参数：

- `-BaseDir`：自定义部署目录
- `-TaskName`：自定义计划任务名称

## 验证方式

- 日志中出现 `launch success app=<appId>`。
- 若使用 `restart: true`，日志中应先出现重启记录，再出现成功启动。
- 用户桌面会话中能看到目标 GUI。
- 启动流程不应长时间阻塞当前代理前台。

## 常见问题

- 触发成功但无窗口：先检查 `config.json` 中的实际路径是否有效。
- 日志无记录：先检查 `run.json` 是否写入到了正确目录。
- `schtasks` 可查询但无法启动：检查任务名、权限与当前会话环境。
- 重启失败：检查 `processNames` 是否配置正确，是否误用了过宽的进程名。

## 完成标准

- 目标 GUI 已通过 `guilauncher` 或明确回退路径成功拉起。
- 启动过程有日志证据可追踪。
- 配置、脚本和任务复用关系清晰，没有重复部署平行链路。
