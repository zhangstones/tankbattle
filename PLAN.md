# 美术资源生成与集成计划

## 1. 目标

- 基于 `skills/qwen-image-assets` 持续生成游戏图片资源。
- 先产出候选图，再筛选、裁切、集成到正式资源目录。
- 用本文件持续记录每类资源的状态、落盘路径和验收结果。

## 2. 执行规则

1. 统一使用脚本生成候选图：

```powershell
python .\skills\qwen-image-assets\scripts\generate_image.py `
  --model qwen-image-2.0 `
  --prompt "<prompt>" `
  --out .\assets\generated\<module>\<name>.png
```

2. 生成路径与正式路径分离：
- 候选图：`assets/generated/...`
- 正式资源：`assets/tiles`、`assets/sprites`、`assets/vfx`、`assets/ui`

3. 单次只推进一个资源项：
- 一次只生成、筛选并确认一个资源项，避免并行扩散导致风格失控。

4. 生成图不直接进正式链路：
- 必须先人工确认，再决定是否裁切、缩放、透明处理或二次编辑。

5. 小尺寸资源优先低噪声：
- 图标、HUD、小贴图优先明确轮廓和主体，不追求复杂纹理。

6. 每完成一个资源项，更新本文件：
- 至少记录：状态、候选路径、正式路径、备注。

## 3. 状态定义

- `pending`：尚未开始
- `generating`：正在生成候选图
- `selected`：已选中候选图，待集成
- `integrated`：已接入程序
- `verified`：已完成运行验证
- `rejected`：方向错误，已放弃

## 4. 当前约束

- 当前环境中的默认模型名不是图像模型，执行生成时默认显式传 `--model qwen-image-2.0`。
- 先从静态资源开始，不先做复杂动态特效。
- 当前视觉方向以“低噪声、俯视、可读性优先”为准，不走复杂贴图或高写实路线。

## 5. 分阶段计划

### Phase 1：战场静态资源

| ID | 资源项 | 目标 | 候选输出目录 | 正式目录 | 状态 | 备注 |
|---|---|---|---|---|---|---|
| P1-01 | 地表基础纹理 | 提升地表质感，但不抢主体视线 | `assets/generated/tiles/ground/` | `assets/tiles/` | pending | 优先低对比、低重复感 |
| P1-02 | 可破坏砖墙 | 让砖墙一眼可识别，并保留可破坏语义 | `assets/generated/tiles/brick/` | `assets/tiles/` | pending | 不走写实砖块堆砌风格 |
| P1-03 | 不可破坏钢墙 | 与砖墙风格统一，但材质区分明确 | `assets/generated/tiles/steel/` | `assets/tiles/` | pending | 重点是板材感和稳固感 |
| P1-04 | 堡垒主体 | 成为战场中的明确核心目标 | `assets/generated/tiles/fortress/` | `assets/tiles/` | pending | 保持俯视简洁，不做背景装饰感 |

### Phase 2：战斗反馈资源

| ID | 资源项 | 目标 | 候选输出目录 | 正式目录 | 状态 | 备注 |
|---|---|---|---|---|---|---|
| P2-01 | 子弹命中火花 | 提升命中反馈 | `assets/generated/vfx/hit-spark/` | `assets/vfx/` | pending | 优先图形化火花，不做复杂烟尘 |
| P2-02 | 小爆炸 | 表示命中或轻度摧毁 | `assets/generated/vfx/explosion-small/` | `assets/vfx/` | pending | 保持清晰轮廓 |
| P2-03 | 大爆炸 | 表示单位摧毁 | `assets/generated/vfx/explosion-large/` | `assets/vfx/` | pending | 强调能量释放层级 |
| P2-04 | 护盾/拾取反馈 | 强化 buff 与拾取事件 | `assets/generated/vfx/powerup/` | `assets/vfx/` | pending | 先做护盾，再做拾取闪光 |

### Phase 3：单位与道具资源

| ID | 资源项 | 目标 | 候选输出目录 | 正式目录 | 状态 | 备注 |
|---|---|---|---|---|---|---|
| P3-01 | 玩家坦克 | 提升识别度与造型完整性 | `assets/generated/sprites/player-tank/` | `assets/sprites/` | pending | 小尺寸读图优先 |
| P3-02 | 敌方坦克 | 与玩家同语言，但能快速区分 | `assets/generated/sprites/enemy-tank/` | `assets/sprites/` | pending | 可按敌人层级做轻微变体 |
| P3-03 | 道具图标 | 统一护盾/连发/修复视觉语言 | `assets/generated/sprites/powerups/` | `assets/sprites/` | pending | 先确保轮廓明确 |

### Phase 4：界面资源

| ID | 资源项 | 目标 | 候选输出目录 | 正式目录 | 状态 | 备注 |
|---|---|---|---|---|---|---|
| P4-01 | HUD 小图形元素 | 提升信息区精致度 | `assets/generated/ui/hud/` | `assets/ui/` | pending | 不降低可读性 |
| P4-02 | 状态弹框背景 | 统一暂停/胜利/失败面板风格 | `assets/generated/ui/panels/` | `assets/ui/` | pending | 保持轻量，不喧宾夺主 |
| P4-03 | 菜单背景元素 | 增强主菜单完成度 | `assets/generated/ui/menu/` | `assets/ui/` | pending | 避免复杂背景影响阅读 |

## 6. 记录模板

后续每推进一个资源项，在对应条目备注中至少补充：

- 使用的 prompt 版本
- 候选图路径
- 选中的最终候选
- 是否已裁切/缩放/透明处理
- 是否已接入代码
- 是否已运行验证

## 7. 下一步

- 默认先从 `P1-04 堡垒主体` 开始。
- 原因：堡垒是核心目标物，单体资源最容易先建立风格，不会像地表平铺那样一开始就暴露重复和噪声问题。
