# Skills

本目录存放仓库内可复用的技能文档、脚本和模板。

## 目录约定

- 每个技能使用独立目录：`skills/<skill-name>/`
- 技能主文档固定为：`skills/<skill-name>/SKILL.md`
- 只有在技能确实需要脚本、模板或参考资产时，才增加子目录
- `README.md` 只负责导航，不承载执行细节

## 使用方式

1. 先根据任务类型定位最匹配的技能。
2. 进入对应目录阅读 `SKILL.md`。
3. 若技能引用脚本或模板，再按文档最小化读取相关文件。
4. 若多个技能同时适用，优先按“主任务技能 -> 配套技能”顺序组合使用。

## 当前技能

- `code-review`
  - 面向代码审查，优先识别缺陷、回归、测试缺口和高风险实现。
- `commit-hygiene`
  - 面向提交粒度、提交信息和历史整理，避免混杂提交和不可审查历史。
- `feature-change-sync`
  - 面向功能变更联动，确保代码、测试和职责匹配文档同步更新。
- `launchgui`
  - 面向 Windows GUI 程序拉起与桌面会话验证，默认复用 `guilauncher` 链路。
- `modular-maintainability`
  - 面向模块化设计、渐进式重构和高风险边界迁移。
- `pillow-image-draw`
  - 面向用 Pillow 代码绘制或微调图像资产。
- `qwen-image-assets`
  - 面向通过 Qwen 兼容图像接口生成 PNG 资产和迭代提示词。
- `ui-layout-stability`
  - 面向菜单、HUD、面板等布局稳定性与信息层级控制。
- `ui-testing`
  - 面向调试 API、功能流转、快照回归和 golden 维护。

## 选择规则

- 任务以“判断问题”为主时，优先用 `code-review`。
- 任务以“改功能并保持文档/测试一致”为主时，优先用 `feature-change-sync`。
- 任务包含提交整理时，再叠加 `commit-hygiene`。
- 任务涉及 GUI 启动、界面验证或 golden 回归时，优先看 `launchgui`、`ui-layout-stability`、`ui-testing`。
- 任务涉及图像资产时，根据产出方式选择 `pillow-image-draw` 或 `qwen-image-assets`。

## 维护原则

- 技能应描述稳定流程、边界和回退，不记录一次性任务细节。
- 若脚本、路径或命令发生变化，应优先更新对应技能，而不是只改任务说明。
- 若多个技能出现重复规则，应把共性保留在各自最相关的位置，避免相互拷贝漂移。
