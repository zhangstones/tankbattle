---
name: launchgui
description: 通过 Windows 任务计划程序把 GUI 程序启动到用户桌面会话；优先复用已有 guilauncher 配置，避免重复创建。
---

# Launch GUI To User Session

## 目标

在代理会话无法直接显示窗口时，通过 `guilauncher` 任务把 GUI 启动到用户当前桌面会话，且保持非阻塞。

## 前置检查（必须先做）

1. 检查文档是否存在：`launchgui.md`。
2. 检查目录是否存在：`D:\Workspace\guilauncher`。
3. 检查关键文件是否存在：
- `D:\Workspace\guilauncher\launchgui.ps1`
- `D:\Workspace\guilauncher\config.json`
4. 检查任务是否存在：`schtasks /query /tn guilauncher`

若以上项已存在：直接复用，不得重复创建同名目录/任务/脚本。  
仅当缺失时，按 `launchgui.md` 补齐最小必要项。

## 标准执行流程

1. 确认目标应用已加入白名单（`config.json` 的 `apps`）。
2. 写入触发文件 `run.json`，仅写 `app` 字段，不接受任意 exe 路径入参。
3. 触发任务：`schtasks /run /tn "guilauncher"`。
4. 检查日志：`D:\Workspace\guilauncher\logs\launcher.log`。

## 推荐命令模板

```powershell
@'
{
  "app": "tankbattle"
}
'@ | Set-Content -Path D:\Workspace\guilauncher\run.json -Encoding UTF8

schtasks /run /tn "guilauncher"
Get-Content D:\Workspace\guilauncher\logs\launcher.log -Tail 30
```

## 约束

- 不要在技能中引入“任意命令执行”能力。
- 不要把 `cmd.exe`、`powershell.exe` 等通用解释器加入白名单。
- 仅维护固定应用 `app id -> exe/workdir` 映射。

## 验证口径

- 任务触发成功且日志出现 `launch success`。
- 用户可在桌面会话看到目标 GUI 窗口。
- 触发过程不阻塞当前代理前台操作。
