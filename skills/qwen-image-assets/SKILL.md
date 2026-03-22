---
name: qwen-image-assets
description: 用于通过千问图像模型生成游戏图片资源，例如图标、贴图、纹理、UI 元素和特效草图。当用户通过环境变量提供 API URL、API Key 和模型名，希望稳定地产出 PNG 资源、迭代提示词、控制风格一致性时使用。
---

# Qwen Image Assets

## 目标

- 用固定脚本调用千问图像接口生成图片资源。
- 把“提示词编写、请求发送、文件落盘”收敛成稳定流程。
- 先产出可筛选的基础图，再决定是否进入正式资源链。

## 环境变量

- `QWEN_IMAGE_API_URL`：必填。可为完整图片生成接口，也可为基础地址；若只给基础地址，脚本默认补到 `/images/generations`。
- `QWEN_IMAGE_API_KEY`：必填。用于 Bearer 鉴权。
- `QWEN_IMAGE_MODEL`：可选。默认 `qwen-image`。
- `QWEN_IMAGE_TIMEOUT`：可选。默认 `120` 秒。
- `QWEN_IMAGE_API_MODE`：可选。支持 `auto`、`openai-images`、`dashscope-qwen-image`；默认 `auto`。
- `QWEN_IMAGE_PROMPT_FIELD`：可选。默认 `prompt`；若网关要求 `input`，可在环境变量中覆盖。
- `QWEN_IMAGE_EXTRA_JSON`：可选。附加到请求体中的 JSON 对象，用于私有网关扩展字段。

说明：
- 若是 OpenAI 风格图片接口，脚本走 `openai-images`。
- 若是阿里云百炼官方 `qwen-image` 接口，脚本走 `dashscope-qwen-image`，并自动改用同步图像生成地址与请求体结构。

## 使用方式

脚本路径：
- `skills/qwen-image-assets/scripts/generate_image.py`

常用命令：

```powershell
python .\skills\qwen-image-assets\scripts\generate_image.py `
  --prompt "top-down tank fortress icon, clean vector style, centered composition, no text" `
  --out .\assets\generated\fortress_icon.png
```

可选参数：

- `--negative-prompt`：负向提示词。
- `--size`：默认 `1024x1024`。
- `--model`：覆盖环境变量中的模型名。
- `--style`、`--quality`：仅在网关支持时传递。
- `--dry-run`：只校验接口地址、环境变量映射和请求体，不发网络请求。

## 提示词原则

1. 先写资源角色，再写视角和风格
- 例如：`top-down fortress`, `game HUD icon`, `tileable ground texture`

2. 一次只控制一个视觉目标
- 先控构图，再控风格，再控材质，不要一轮同时大改。

3. 明确排除项
- 常用排除：`no text, no watermark, no UI, no extra objects, no perspective distortion`

4. 小尺寸资源优先低噪声
- 图标、HUD、小贴图优先：简洁轮廓、少细节、高对比主形。

5. 原始生成图不要直接进正式资源链
- 先输出到临时目录筛选。
- 确认方向后，再做裁切、缩放、透明处理或二次重绘。

## 推荐流程

1. 先判断资源类型：
- 图标 / HUD 元素：强调轮廓和中心主体
- 地表 / 墙体纹理：强调平铺与低噪声
- 特效草图：强调形状和能量走向，不追求成品

2. 先生成 1 张验证方向
- 不要上来批量生成多张，先确认风格和构图。

3. 发现问题时只改一个维度
- 构图不对就改构图
- 风格太花就降噪
- 主体太小就放大主体

4. 定稿前不要覆盖正式资源
- 正式资源更新前，先保留生成路径和提示词，确保可复现。

## 自检

先做无网络自检，再做真实生成：

```powershell
$env:QWEN_IMAGE_API_URL="https://example.com/v1"
$env:QWEN_IMAGE_MODEL="qwen-image"
python .\skills\qwen-image-assets\scripts\generate_image.py `
  --prompt "top-down fortress icon, clean vector style" `
  --out .\tmp\fortress.png `
  --dry-run
```

预期结果：
- 输出识别到的 `mode`
- 输出解析后的 `endpoint`
- 输出最终 `payload`
- 不访问网络，不要求 `QWEN_IMAGE_API_KEY`

无网络自检通过后，再设置真实 `QWEN_IMAGE_API_KEY` 执行正式生成。

## 失败处理

- 返回 `401/403`：先检查 `QWEN_IMAGE_API_KEY`。
- 返回 `404`：先检查 `QWEN_IMAGE_API_URL` 是否是完整接口路径。
- DashScope 返回 `404`：优先检查当前是否错误走到了 `openai-images`；`qwen-image` 正确接口应为官方同步图像生成接口。
- 返回字段错误：优先检查 `QWEN_IMAGE_PROMPT_FIELD` 是否应改成 `input`。
- 返回模型错误：优先检查 `QWEN_IMAGE_MODEL` 是否为图像模型（如 `qwen-image-2.0-pro`、`qwen-image-2.0`、`qwen-image`），而不是通用对话模型。
- 返回成功但没有图片：检查网关返回的是 `b64_json`、`url`、`image_url` 还是其他字段，必要时扩展脚本解析逻辑。
