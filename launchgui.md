下面给你一版**按你的要求重写**的完整落地方案：

* 目录放到 `D:\Workspace` 下
* 目录名、文件名、任务名尽量使用**小写**
* 方案仍然基于 **Windows 任务计划程序**
* 目标仍然是：**后台触发，桌面可见，非阻塞，白名单控制**

我这里统一采用以下路径与命名：

```text
D:\Workspace\guilauncher\
```

目录结构如下：

```text
D:\Workspace\guilauncher\
    launchgui.ps1
    config.json
    run.json
    logs\
        launcher.log
```

---

# 一、方案目标

这套方案要解决的问题是：

1. 你的后台环境或代理环境**不能直接在当前桌面会话中显示 GUI**
2. 通过 **Task Scheduler**，让 GUI 在**当前登录用户的桌面会话**中启动
3. 后续只需要：

   * 写入一个触发参数文件
   * 调用一次计划任务
4. GUI 就会在桌面中弹出，且不会阻塞调用端

---

# 二、创建目录

先在 PowerShell 中创建目录：

```powershell
New-Item -ItemType Directory -Force -Path D:\Workspace\guilauncher
New-Item -ItemType Directory -Force -Path D:\Workspace\guilauncher\logs
```

如果目录已经存在，这两个命令也可以重复执行，不会有问题。

---

# 三、创建白名单配置文件

文件路径：

```text
D:\Workspace\guilauncher\config.json
```

内容示例：

```json
{
  "apps": {
    "notepad": {
      "exe": "C:\\Windows\\System32\\notepad.exe",
      "workdir": "C:\\Windows\\System32"
    },
    "calc": {
      "exe": "C:\\Windows\\System32\\calc.exe",
      "workdir": "C:\\Windows\\System32"
    },
    "game1": {
      "exe": "D:\\Games\\Game1\\game.exe",
      "workdir": "D:\\Games\\Game1"
    }
  }
}
```

这里的设计原则是：

* 外部只能传一个 `app id`
* 实际可执行文件路径由 `config.json` 映射
* 不允许直接传任意 exe 路径
* 这样可以避免被滥用成“任意命令执行器”

你后续要启动什么 GUI，就往这里加白名单项即可。

---

# 四、创建启动脚本

文件路径：

```text
D:\Workspace\guilauncher\launchgui.ps1
```

建议使用下面这份脚本。

```powershell
$ErrorActionPreference = "Stop"

$baseDir   = "D:\Workspace\guilauncher"
$configFile = Join-Path $baseDir "config.json"
$paramFile  = Join-Path $baseDir "run.json"
$logDir     = Join-Path $baseDir "logs"
$logFile    = Join-Path $logDir "launcher.log"

function Write-Log {
    param(
        [string]$Message
    )

    $time = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Add-Content -Path $logFile -Value "$time $Message"
}

try {
    if (-not (Test-Path $logDir)) {
        New-Item -ItemType Directory -Force -Path $logDir | Out-Null
    }

    if (-not (Test-Path $configFile)) {
        throw "config.json not found: $configFile"
    }

    if (-not (Test-Path $paramFile)) {
        Write-Log "run.json not found, nothing to do"
        exit 0
    }

    $config = Get-Content -Path $configFile -Raw | ConvertFrom-Json
    $param  = Get-Content -Path $paramFile -Raw | ConvertFrom-Json

    if (-not $param.app) {
        throw "missing field: app"
    }

    $appId = [string]$param.app

    if (-not $config.apps.PSObject.Properties.Name.Contains($appId)) {
        throw "app id not allowed: $appId"
    }

    $appConfig = $config.apps.$appId
    $exe = [string]$appConfig.exe
    $workdir = [string]$appConfig.workdir

    if ([string]::IsNullOrWhiteSpace($exe)) {
        throw "exe is empty for app id: $appId"
    }

    if (-not (Test-Path $exe)) {
        throw "executable not found: $exe"
    }

    if ([string]::IsNullOrWhiteSpace($workdir)) {
        $workdir = Split-Path -Path $exe -Parent
    }

    if (-not (Test-Path $workdir)) {
        throw "working directory not found: $workdir"
    }

    Write-Log "launching app=$appId exe=$exe workdir=$workdir"

    Start-Process -FilePath $exe -WorkingDirectory $workdir -WindowStyle Normal

    Remove-Item -Path $paramFile -Force -ErrorAction SilentlyContinue

    Write-Log "launch success app=$appId"
}
catch {
    Write-Log "ERROR: $($_.Exception.Message)"
    exit 1
}
```

---

# 五、脚本说明

这份脚本做了几件关键的事情：

## 1. 从固定目录读取配置

只读取：

* `D:\Workspace\guilauncher\config.json`
* `D:\Workspace\guilauncher\run.json`

不会从外部接受任意 exe 路径。

## 2. 用 `app` 做白名单映射

例如 `run.json` 写入：

```json
{
  "app": "notepad"
}
```

脚本不会直接执行 `"notepad"`，而是先去 `config.json` 里查：

```json
"notepad": {
  "exe": "C:\\Windows\\System32\\notepad.exe",
  "workdir": "C:\\Windows\\System32"
}
```

然后再启动对应程序。

## 3. 启动成功后删除 `run.json`

这样可以避免同一个触发文件被重复使用。

## 4. 写日志

日志会记录到：

```text
D:\Workspace\guilauncher\logs\launcher.log
```

方便排查是否真的执行了，以及失败原因是什么。

---

# 六、创建计划任务

这一步建议在**管理员 PowerShell**中执行。

先创建任务：

```powershell
schtasks /create /tn "guilauncher" /tr "powershell.exe -ExecutionPolicy Bypass -File D:\Workspace\guilauncher\launchgui.ps1" /sc ONDEMAND /rl LIMITED /f
```

这条命令的含义如下：

* `/tn "guilauncher"`
  任务名为小写 `guilauncher`

* `/tr "powershell.exe -ExecutionPolicy Bypass -File D:\Workspace\guilauncher\launchgui.ps1"`
  任务执行 PowerShell 脚本

* `/sc ONDEMAND`
  只允许手动触发，不按时间自动运行

* `/rl LIMITED`
  以普通权限运行，通常更安全

* `/f`
  如果任务已存在则覆盖

---

# 七、在任务计划程序里手工调整关键选项

仅通过 `schtasks /create` 创建后，还需要进入图形界面检查几个关键项，否则 GUI 可能无法显示到桌面。

打开任务计划程序：

```powershell
taskschd.msc
```

找到任务：

```text
任务计划程序库 → guilauncher
```

然后检查并调整以下内容。

---

## 1. “常规”页签

最关键的是这一项：

```text
仅当用户登录时运行
```

必须选择这个，而不要选：

```text
无论用户是否登录都要运行
```

原因很简单：

* “无论用户是否登录都要运行” 通常会在非交互环境中执行
* 程序可能启动了，但不会显示在你的当前桌面上
* 你需要的是 GUI **出现在当前桌面会话**

因此，必须确保任务运行在**当前登录用户的交互 Session** 中。

---

## 2. “常规”页签中的用户账户

建议该任务的运行账户就是**你当前登录 Windows 的那个用户**。

也就是说，谁的桌面要显示 GUI，就用谁的账户创建或修改这个任务。

---

## 3. “设置”页签

建议把：

```text
如果任务已经在运行，则以下规则适用：
```

改为：

```text
并行运行新实例
```

或者至少不要用“拒绝新实例”。

否则可能出现这种情况：

* 上一次脚本还没彻底退出
* 你第二次触发计划任务
* 新请求被直接忽略

对于这种“短时启动器”，通常用“并行运行新实例”更省心。

---

# 八、创建触发参数文件

后续后台调用时，需要先写入：

```text
D:\Workspace\guilauncher\run.json
```

示例内容：

```json
{
  "app": "notepad"
}
```

或者：

```json
{
  "app": "calc"
}
```

或者：

```json
{
  "app": "game1"
}
```

---

# 九、后台触发方式

后续每次启动 GUI，只需要执行两步。

## 第一步：写入 `run.json`

例如启动记事本：

```powershell
@'
{
  "app": "notepad"
}
'@ | Set-Content -Path D:\Workspace\guilauncher\run.json -Encoding UTF8
```

例如启动计算器：

```powershell
@'
{
  "app": "calc"
}
'@ | Set-Content -Path D:\Workspace\guilauncher\run.json -Encoding UTF8
```

例如启动游戏：

```powershell
@'
{
  "app": "game1"
}
'@ | Set-Content -Path D:\Workspace\guilauncher\run.json -Encoding UTF8
```

---

## 第二步：运行计划任务

```powershell
schtasks /run /tn "guilauncher"
```

执行后，计划任务会在当前登录用户桌面会话中运行 `launchgui.ps1`，再由脚本根据 `run.json` 和 `config.json` 启动对应 GUI。

---

# 十、完整测试流程

下面给你一套最小可验证流程。

## 1. 写入 `config.json`

内容至少包含 `notepad`：

```json
{
  "apps": {
    "notepad": {
      "exe": "C:\\Windows\\System32\\notepad.exe",
      "workdir": "C:\\Windows\\System32"
    }
  }
}
```

## 2. 写入 `run.json`

```powershell
@'
{
  "app": "notepad"
}
'@ | Set-Content -Path D:\Workspace\guilauncher\run.json -Encoding UTF8
```

## 3. 触发任务

```powershell
schtasks /run /tn "guilauncher"
```

## 4. 预期结果

* 记事本出现在当前桌面
* `D:\Workspace\guilauncher\run.json` 被删除
* `D:\Workspace\guilauncher\logs\launcher.log` 中记录成功日志

---

# 十一、查看日志

查看日志命令：

```powershell
Get-Content D:\Workspace\guilauncher\logs\launcher.log -Tail 50
```

你可能会看到类似内容：

```text
2026-03-21 13:20:01 launching app=notepad exe=C:\Windows\System32\notepad.exe workdir=C:\Windows\System32
2026-03-21 13:20:01 launch success app=notepad
```

如果失败，可能看到：

```text
2026-03-21 13:22:10 ERROR: app id not allowed: xxx
```

或者：

```text
2026-03-21 13:22:15 ERROR: executable not found: D:\Games\Game1\game.exe
```

---

# 十二、安全建议

这部分很重要，建议不要省略。

## 1. 只允许白名单程序

不要把脚本写成这样：

```powershell
Start-Process $param.exe
```

或者这样：

```powershell
cmd /c $userInput
```

这种写法会把整个系统变成一个远程任意执行入口，风险非常高。

正确方式是：

* 外部只传 `app`
* 本地配置文件决定具体 exe 路径

---

## 2. 控制目录写权限

建议检查 `D:\Workspace\guilauncher` 的 ACL，确保：

* 只有你自己的用户有写权限
* 不要给无关用户写权限
* 不要把它放在共享目录或临时目录下

因为只要别人能改 `run.json` 或 `config.json`，就能影响启动行为。

---

## 3. 保留日志

不要去掉日志。
哪怕最终产品化，也建议保留最基本的日志，否则出了问题很难判断：

* 是任务没跑
* 还是脚本没读到参数
* 还是白名单没配
* 还是 exe 路径错了
* 还是工作目录不存在

---

## 4. 不要让 `config.json` 暴露过多系统能力

建议只放你明确需要启动的 GUI，不要把任何系统工具都放进去，更不要放：

* `cmd.exe`
* `powershell.exe`
* `wscript.exe`
* `mshta.exe`
* 其他通用执行器

否则白名单机制就失去意义了。

---

# 十三、常见问题排查

## 问题 1：任务执行了，但桌面没有看到窗口

优先检查：

1. 任务是否设置为
   **“仅当用户登录时运行”**

2. 运行账户是否是**当前正在登录桌面的那个用户**

3. 程序是否本身真的会弹出 GUI，而不是后台静默启动

4. 日志里是否记录了 `launch success`

---

## 问题 2：日志显示成功，但没有看到窗口

这通常有几种可能：

* 程序启动后最小化了
* 程序已经在后台运行，未创建新窗口
* 程序是单实例，激活的是旧实例
* 启动器调用没问题，但目标程序自身行为不是“新弹一个窗口”

这时建议先用 `notepad` 或 `calc` 做验证，确认计划任务链路本身没有问题。

---

## 问题 3：第二次触发无效

检查：

* 任务设置里是否阻止新实例
* `run.json` 是否已被旧进程删除
* 是否有旧脚本还没退出
* 是否写入了新的 `run.json`

---

## 问题 4：中文路径或特殊字符路径问题

你当前要求使用的是：

```text
D:\Workspace\guilauncher
```

这是很好的路径选择，纯英文、无空格，能减少不少脚本解析和转义问题。

---

# 十四、推荐的最终命令清单

下面是一组可以直接参考的命令。

## 创建目录

```powershell
New-Item -ItemType Directory -Force -Path D:\Workspace\guilauncher
New-Item -ItemType Directory -Force -Path D:\Workspace\guilauncher\logs
```

## 创建任务

```powershell
schtasks /create /tn "guilauncher" /tr "powershell.exe -ExecutionPolicy Bypass -File D:\Workspace\guilauncher\launchgui.ps1" /sc ONDEMAND /rl LIMITED /f
```

## 写入触发参数并启动 notepad

```powershell
@'
{
  "app": "notepad"
}
'@ | Set-Content -Path D:\Workspace\guilauncher\run.json -Encoding UTF8

schtasks /run /tn "guilauncher"
```

## 查看日志

```powershell
Get-Content D:\Workspace\guilauncher\logs\launcher.log -Tail 50
```

---

# 十五、我对这版方案的建议

如果你的目标只是：

* 让后台环境以后能够触发桌面里可见的 GUI
* 且不需要复杂回传
* 只需要“启动几个固定应用”

那么这一版已经足够实用，而且比“常驻轮询脚本 + 触发文件”更稳定，因为：

* 任务调度由 Windows 原生组件负责
* 交互 Session 由计划任务明确控制
* 脚本逻辑简单，故障面较小
* 白名单机制清晰

如果你下一步需要，我可以继续直接给你一版：

**“可直接复制使用”的完整初始化脚本**，一次性帮你自动生成：

* `D:\Workspace\guilauncher\config.json`
* `D:\Workspace\guilauncher\launchgui.ps1`
* `logs` 目录
* `guilauncher` 计划任务

这样你基本只要执行一段 PowerShell 就能完成部署。


