---
name: qwen-image-assets
description: 用于通过 Qwen 兼容图像模型生成游戏图片资源，例如图标、贴图、纹理、UI 元素和特效草图；当用户通过环境变量提供 API URL、API Key 和模型名，希望稳定地产出 PNG 资源、迭代提示词、控制风格一致性时使用。
---

# Qwen Image Assets

## 适用场景

- 需要快速生成图标、贴图、纹理、UI 元素、特效草图等位图资产。
- 需要用脚本稳定调用 Qwen 兼容图像接口，而不是手工点网页。
- 需要多轮迭代提示词并把结果落盘为本地 PNG。

## 不适用场景

- 当前环境无法提供网络或图片接口凭证，且不能只做 `--dry-run`。
- 任务更适合代码绘制、SVG 编辑或直接修改已有资产。
- 需要严格确定性像素输出，而模型生成本身会带来随机性。

## 默认工具

- 脚本路径：`skills/qwen-image-assets/scripts/generate_image.py`
- 输出能力：调用接口、解析返回图片、写入本地 PNG
- 自检能力：`--dry-run` 可验证 URL、模式识别和 payload

## 环境变量

- `QWEN_IMAGE_API_URL`：必填，可为完整接口或基础地址
- `QWEN_IMAGE_API_KEY`：正式请求必填
- `QWEN_IMAGE_MODEL`：可选，默认 `qwen-image`
- `QWEN_IMAGE_TIMEOUT`：可选，默认 `120`
- `QWEN_IMAGE_API_MODE`：可选，支持 `auto`、`openai-images`、`dashscope-qwen-image`
- `QWEN_IMAGE_PROMPT_FIELD`：可选，默认 `prompt`
- `QWEN_IMAGE_EXTRA_JSON`：可选，附加 JSON 对象

## 开始前确认

- 资源类型：图标、HUD 元素、纹理、草图还是宣传图。
- 使用场景：小尺寸显示、平铺、透明背景、中心构图等。
- 输出路径是否应先写到临时目录，而不是直接覆盖正式资源。
- 当前网关是 OpenAI 风格还是 DashScope 风格。

## 使用原则

- 先写资源角色，再写视角、风格和材质。
- 一次只控制一个视觉目标，例如先控构图，再控风格。
- 明确排除项，例如 `no text`、`no watermark`、`no extra objects`。
- 小尺寸资源优先低噪声和清晰轮廓。
- 生成图先筛选，不直接进入正式资源链。

## 标准流程

1. 先用 `--dry-run` 验证接口模式、endpoint 和 payload。
2. 先生成 1 张确认方向，不要一开始就批量出图。
3. 如果问题明确，只调整一个维度：构图、风格、噪声或主体占比。
4. 结果可用后，再落到正式候选目录。
5. 如需进入正式资源链，再做裁切、缩放、透明处理或二次绘制。

## 常用命令

```powershell
python .\skills\qwen-image-assets\scripts\generate_image.py `
  --prompt "top-down fortress icon, clean vector style, centered composition, no text" `
  --out .\assets\generated\fortress_icon.png
```

```powershell
python .\skills\qwen-image-assets\scripts\generate_image.py `
  --prompt "top-down fortress icon, clean vector style" `
  --out .\tmp\fortress.png `
  --dry-run
```

## 失败处理

- `401/403`：先检查 `QWEN_IMAGE_API_KEY`
- `404`：先检查 `QWEN_IMAGE_API_URL` 是否正确，是否走错接口模式
- payload 字段错误：检查 `QWEN_IMAGE_PROMPT_FIELD` 与 `QWEN_IMAGE_EXTRA_JSON`
- 模型错误：确认 `QWEN_IMAGE_MODEL` 是图像模型，而不是通用对话模型
- 返回成功但无图：检查返回字段是否为 `b64_json`、`image_base64`、`url` 或 `image_url`

## 交付约束

- 未确认前不要覆盖正式资产。
- 生成结果应保留可复现线索，至少能回溯提示词和输出路径。
- 若最终放弃模型生成路线，应说明原因，不把低质量候选图混入正式资源目录。

## 完成标准

- 能稳定跑通一次 `--dry-run`。
- 至少有一张符合目标方向的候选 PNG。
- 输出路径、提示词方向和后续处理边界清晰。
