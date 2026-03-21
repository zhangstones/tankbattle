---
name: commit-hygiene
description: 用于规范提交粒度、提交信息与历史整理，保证提交记录清晰、可追溯、便于审查。
---

# 代码提交规范

## 目标

- 提交历史表达“问题-解决方案”的连续性。
- 减少同一问题的碎片提交，降低审查与回滚成本。
- 保持提交信息语义统一，便于检索与发布说明生成。

## 提交前缀规范

1. `fix:` 修复缺陷、回归、错位、错误行为。
2. `feat:` 新功能、新交互、新配置项。
3. `doc:` 文档或技能说明更新。

约束：

1. 只使用以上英文前缀。
2. 不使用中文前缀（如“修复:”“功能:”“文档:”）。

## 粒度与 amend 规则

1. 同一问题连续修正，优先 `git commit --amend`。
2. 只有在“问题域变化”时才拆新提交（例如从布局问题切到音频问题）。
3. 每个提交应可独立解释，不混入无关改动。

## 提交信息写法

1. 标题行：`<prefix>: <一句话结果>`。
2. 正文第 1 行写范围（Scope）。
3. 正文第 2 行写关键变更点（Key changes）。

示例：

```text
fix: refine menu layout and help text alignment
Scope: menu center section spacing and text alignment.
Key changes: adaptive layout computation, centered title/help text, geometry regression tests.
```

## 历史整理策略

1. 优先使用交互式 rebase 整理连续提交。
2. 若环境无法稳定使用交互式编辑器，可用等价流程：
- `git reset --soft <base-commit>`
- 重新提交为单一规范提交
3. 整理后检查 `git log --oneline`，确认历史符合“单问题单提交”。

## 交付前检查

1. `git status --short` 应无非预期文件。
2. 提交前执行项目要求的测试与构建命令。
3. 若为同一问题追加修复，确认是否应 `--amend` 而非新增提交。
