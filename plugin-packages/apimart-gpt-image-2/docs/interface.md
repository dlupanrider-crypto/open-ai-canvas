# 接口定义

来源：[图片生成](https://docs.apib.ai/cn/api-reference/images/gpt-image-2/generation)、[官方文本](https://docs.apib.ai/cn/api-reference/images/gpt-image-2/generation.md)。该文档给出的实际服务域名是 api.apimart.ai。

## 请求

- `apiKey`：必填密钥，Bearer 鉴权。
- 创建：`POST /v1/images/generations`；查询：`GET /v1/tasks/{task_id}`，均使用域名根路径。
- `model`：gpt-image-2 或 gpt-image-2-ext；`prompt`：图片描述。
- `aspectRatio` → `size`：auto、1:1、3:2、2:3、4:3、3:4、5:4、4:5、16:9、9:16、2:1、1:2、3:1、1:3、21:9、9:21 或 WxH。空值省略，保留图生图的上游原尺寸行为。
- `resolution`：1k / 2k / 4k，默认 1k，参与计价。
- `imageCount` → `n`：固定数值 1。官方参数表限制为 1，但多图示例写了 2；本插件遵循参数表，不采用冲突示例。
- `images` → `image_urls`：最多 15 张，按 order 排序，每张优先 dataUrl，其次 url；未提供参考图时省略。官方限制单图 20M、总计 256M，由上游校验。
- `providerOptions.apimart-gpt-image-2.nsfw_check`、`providerOptions.apimart-gpt-image-2.official_fallback`：可选布尔值。未设置时省略，采用上游默认值。
- 不发送 quality、style、response_format；该端点只返回 URL。

## 响应与生命周期

创建从 `data[0].task_id` 读取任务 ID，`submitted` 为待处理；查询从 `data.id` 读取 ID，processing 为处理中、completed 为成功、failed 为失败。顶层 code=200 是上游业务成功，不按影策业务信封的 code=0 处理；非 200 映射失败。错误信息优先读取 data.error.message、error.message。

图片从 `data.result.images[].url[0]` 读取，单次 n=1；标记为 ephemeral，交由宿主下载持久化。文档未提供取消端点，不声明上游取消能力。

## Manifest 完整接口定义

<!-- YINGCE_MANIFEST_CONTRACT_START -->
```json
{
  "apiVersion": "yingce.plugin/v2",
  "id": "apimart-gpt-image-2",
  "name": "APIMart GPT-Image-2 异步图片",
  "version": "1.0.0",
  "author": "影策",
  "description": "APIMart GPT-Image-2 文生图、参考图生图和异步任务轮询。",
  "permissions": [
    "generation.run",
    "media.read"
  ],
  "configuration": {
    "fields": [
      {
        "name": "apiKey",
        "type": "secret",
        "label": "API Key",
        "required": true
      }
    ]
  },
  "contributes": {
    "providers": [
      {
        "id": "apimart-gpt-image-2",
        "label": "APIMart GPT-Image-2",
        "capabilities": [
          "image"
        ],
        "scopes": [
          "admin.system-channel",
          "user.custom-channel",
          "canvas",
          "creation",
          "agent"
        ],
        "baseUrl": "https://api.apimart.ai",
        "requiresPublicMediaUrls": false,
        "auth": {
          "type": "bearer",
          "field": "apiKey"
        },
        "imageParameters": [
          {
            "name": "size",
            "label": "比例 / 像素尺寸",
            "type": "string",
            "binding": "aspectRatio"
          },
          {
            "name": "imageResolution",
            "label": "分辨率",
            "type": "string",
            "binding": "resolution",
            "values": [
              "1k",
              "2k",
              "4k"
            ],
            "default": "1k",
            "pricing": true
          },
          {
            "name": "count",
            "label": "生成数量",
            "type": "integer",
            "binding": "imageCount",
            "min": 1,
            "max": 1,
            "step": 1,
            "default": 1
          }
        ],
        "parameters": [
          {
            "name": "model",
            "type": "string",
            "required": true,
            "mapping": "model",
            "description": "gpt-image-2 或 gpt-image-2-ext。"
          },
          {
            "name": "prompt",
            "type": "string",
            "required": true,
            "mapping": "prompt",
            "description": "图片提示词。"
          },
          {
            "name": "images",
            "type": "media[]",
            "required": false,
            "mapping": "image_urls",
            "description": "最多 15 张参考图，按 order 排序，支持 URL 和 Data URI。"
          },
          {
            "name": "aspectRatio",
            "type": "string",
            "required": false,
            "mapping": "size",
            "description": "15 种比例、auto 或 WxH 像素尺寸；空值时不发送。"
          },
          {
            "name": "resolution",
            "type": "string",
            "required": false,
            "mapping": "resolution",
            "description": "1k / 2k / 4k。"
          },
          {
            "name": "imageCount",
            "type": "integer",
            "required": false,
            "mapping": "n",
            "description": "单次仅支持 1。"
          },
          {
            "name": "providerOptions",
            "type": "object",
            "required": false,
            "mapping": "nsfw_check / official_fallback",
            "description": "apimart-gpt-image-2 命名空间内的布尔扩展。"
          }
        ],
        "validations": [
          {
            "assert": {
              "$in": [
                {
                  "$ref": "request.model"
                },
                [
                  "gpt-image-2",
                  "gpt-image-2-ext"
                ]
              ]
            },
            "message": "APIMart 仅支持 gpt-image-2 或 gpt-image-2-ext"
          },
          {
            "assert": {
              "$lte": [
                {
                  "$len": {
                    "$ref": "request.images"
                  }
                },
                15
              ]
            },
            "message": "APIMart 最多支持 15 张参考图"
          }
        ],
        "create": {
          "method": "POST",
          "path": "/v1/images/generations",
          "originPath": true,
          "contentType": "application/json",
          "body": {
            "model": {
              "$ref": "request.model"
            },
            "prompt": {
              "$ref": "request.prompt"
            },
            "n": 1,
            "size": {
              "$omitEmpty": {
                "$ref": "request.aspectRatio"
              }
            },
            "resolution": {
              "$coalesce": [
                {
                  "$ref": "request.resolution"
                },
                "1k"
              ]
            },
            "image_urls": {
              "$omitEmpty": {
                "$map": {
                  "from": {
                    "$sortByOrder": {
                      "$ref": "request.images"
                    }
                  },
                  "as": "media",
                  "in": {
                    "$coalesce": [
                      {
                        "$ref": "media.dataUrl"
                      },
                      {
                        "$ref": "media.url"
                      }
                    ]
                  }
                }
              }
            },
            "nsfw_check": {
              "$omitEmpty": {
                "$ref": "request.providerOptions.apimart-gpt-image-2.nsfw_check"
              }
            },
            "official_fallback": {
              "$omitEmpty": {
                "$ref": "request.providerOptions.apimart-gpt-image-2.official_fallback"
              }
            }
          }
        },
        "poll": {
          "method": "GET",
          "path": "/v1/tasks/{{taskId}}",
          "originPath": true
        },
        "response": {
          "taskId": {
            "$coalesce": [
              {
                "$ref": "response.data.0.task_id"
              },
              {
                "$ref": "response.data.id"
              },
              {
                "$ref": "taskId"
              }
            ]
          },
          "status": {
            "$if": {
              "condition": {
                "$and": [
                  {
                    "$ref": "response.code"
                  },
                  {
                    "$ne": [
                      {
                        "$ref": "response.code"
                      },
                      200
                    ]
                  }
                ]
              },
              "then": "failed",
              "else": {
                "$coalesce": [
                  {
                    "$ref": "response.data.0.status"
                  },
                  {
                    "$ref": "response.data.status"
                  }
                ]
              }
            }
          },
          "messagePaths": [
            "data.error.message",
            "error.message",
            "message",
            "msg"
          ],
          "errorPaths": [
            "error.code",
            "data.error.code"
          ],
          "resultKind": "image",
          "resultEphemeral": true,
          "images": {
            "$map": {
              "from": {
                "$ref": "response.data.result.images"
              },
              "as": "image",
              "in": {
                "url": {
                  "$ref": "image.url.0"
                }
              }
            }
          }
        }
      }
    ]
  },
  "documentation": "<当前插件的完整 documentation，由 README.md 与 docs/interface.md 拼接而成；为避免 JSON 递归，此处不重复展开正文。>"
}
```
<!-- YINGCE_MANIFEST_CONTRACT_END -->
