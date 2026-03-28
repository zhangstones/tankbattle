# Build Guide

## 常规构建

开发运行：

```powershell
go run .\cmd\tankbattle
```

构建普通可执行文件：

```powershell
go build -o tankbattle.exe .\cmd\tankbattle
```

构建 GUI 版本（无命令行黑窗）：

```powershell
go build -ldflags="-H windowsgui" -o tankbattle_gui.exe .\cmd\tankbattle
```

## Windows 图标嵌入

- Windows 可执行文件图标通过 `cmd/tankbattle/rsrc_windows_amd64.syso` 自动嵌入。
- 当前图标源文件为 `assets/icons/icon_final.ico`。
- 只要 `.syso` 文件位于 `cmd/tankbattle` 目录下，`go build` 时就会自动带入图标。

如果更换图标，需要重新生成 `.syso`：

```powershell
go install github.com/akavel/rsrc@latest
rsrc -ico assets\icons\icon_final.ico -o cmd\tankbattle\rsrc_windows_amd64.syso
```
