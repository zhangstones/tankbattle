# Skills

本目录存放可复用技能。

## 结构约定

- 每个技能使用单独目录。
- 技能定义文件固定为 `skills/<skill-name>/SKILL.md`。
- 若技能需要脚本、模板或参考资料，再放到各自技能子目录下。

## 当前技能

- `code-review`
  - 用于代码审查，优先找缺陷、风险、回归和缺失测试。
- `commit-hygiene`
  - 用于控制提交粒度、提交信息和历史整理。
- `feature-change-sync`
  - 用于功能变更时同步文档与测试。
- `launchgui`
  - 用于将 GUI 程序稳定拉起到用户桌面会话。
- `modular-maintainability`
  - 用于模块化设计、渐进式重构和高风险目录迁移。
- `pillow-image-draw`
  - 用于基于 Pillow 的图像绘制与处理。
- `qwen-image-assets`
  - 用于图像资产生成相关流程。
- `ui-layout-stability`
  - 用于界面布局稳定性、对齐和防重叠约束。
- `ui-testing`
  - 用于功能/UI 快照测试、golden 回归和调试 API 驱动验证。

## 使用规则

- 需要具体技能时，直接阅读对应目录下的 `SKILL.md`。
- `README.md` 只做导航，不承载技能细节。
- 是否会被技能机制识别，取决于各技能自己的 `SKILL.md`，不是这个总览文件。
