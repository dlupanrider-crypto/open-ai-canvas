# 接口定义

来源：[Grok Imagine 视频生成](https://docs.apib.ai/cn/api-reference/videos/grok-imagine/generation.md)、[任务查询](https://docs.apib.ai/cn/api-reference/tasks/status.md)。官方示例实际服务域名为 api.apimart.ai。

## 请求字段

- `apiKey`：必填，Bearer 请求头鉴权。
- `model`：渠道配置的上游模型名称，原样发送，不限制固定名称；`prompt`：视频描述。
- `duration`：整数 6–15 秒，未提供时默认 6。
- `aspectRatio` → `size`：16:9 / 9:16 / 1:1 / 3:2 / 2:3，默认 16:9。提供参考图后上游自动匹配参考图宽高比。
- `resolution`：480p / 720p，默认 480p。
- `images` → `image_urls`：最多 7 张，按 order 排序，只发送公网 HTTP(S) URL，不使用 dataUrl；网络可达性由上游确认，宿主负责资源准备和 SSRF 安全边界。拒绝独立 last_frame 输入，不把参考图列表宣称为首尾帧协议。
- `providerOptions.apimart-grok-imagine-video.nsfw_check`：可选布尔值，省略时采用上游默认 false。
- 不支持参考视频、参考音频；不发送未在文档声明的 generate_audio、水印或数量字段。

## 响应与生命周期

创建使用 `POST /v1/videos/generations`，读取 `data[0].task_id` 和 submitted 状态。查询使用 `GET /v1/tasks/{task_id}?language=zh`，读取 data.id、data.status：pending/submitted 为待处理、processing 为处理中、completed 为成功、failed 为失败、cancelled 为已取消。

上游业务 code=200 为成功，非 200 映射失败，错误信息优先读取 data.error.message、error.message。视频取自 `data.result.videos[].url`，接受 URL 字符串或 URL 数组（每个视频对象取第一个 URL）；查询文档只声明 videos 为对象数组，未提供视频对象完整示例，此字段仍需真实任务确认。临时视频 URL 标记 ephemeral，交给宿主下载持久化。

文档未提供取消端点，不声明远端取消能力。停止本地等待不代表上游停止生成或免计费。

## Manifest 完整接口定义

<!-- YINGCE_MANIFEST_CONTRACT_START -->
```json
{
  "apiVersion": "yingce.plugin/v2",
  "id": "apimart-grok-imagine-video",
  "name": "APIMart Grok Imagine 异步视频",
  "version": "1.0.1",
  "author": "影策",
  "description": "Grok Imagine 1.5 文生视频、参考图生视频与异步任务轮询。",
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
        "id": "apimart-grok-imagine-video",
        "label": "APIMart Grok Imagine Video",
        "capabilities": [
          "video"
        ],
        "scopes": [
          "admin.system-channel",
          "user.custom-channel",
          "canvas",
          "creation",
          "agent"
        ],
        "baseUrl": "https://api.apimart.ai",
        "requiresPublicMediaUrls": true,
        "auth": {
          "type": "bearer",
          "field": "apiKey"
        },
        "parameters": [
          {
            "name": "model",
            "type": "string",
            "required": true,
            "mapping": "model",
            "description": "渠道配置的上游模型名称，原样发送，不限制固定名称。"
          },
          {
            "name": "prompt",
            "type": "string",
            "required": true,
            "mapping": "prompt",
            "description": "视频内容描述。"
          },
          {
            "name": "images",
            "type": "media[]",
            "required": false,
            "mapping": "image_urls",
            "description": "最多 7 张公网参考图，按 order 排序，不支持 Base64 或尾帧语义。"
          },
          {
            "name": "duration",
            "type": "integer",
            "required": false,
            "mapping": "duration",
            "description": "6–15 秒，默认 6。"
          },
          {
            "name": "aspectRatio",
            "type": "string",
            "required": false,
            "mapping": "size",
            "description": "16:9、9:16、1:1、3:2、2:3，默认 16:9。"
          },
          {
            "name": "resolution",
            "type": "string",
            "required": false,
            "mapping": "resolution",
            "description": "480p 或 720p，默认 480p。"
          },
          {
            "name": "providerOptions",
            "type": "object",
            "required": false,
            "mapping": "nsfw_check",
            "description": "apimart-grok-imagine-video 命名空间中的可选布尔值。"
          }
        ],
        "validations": [
          {
            "assert": {
              "$and": [
                {
                  "$gte": [
                    {
                      "$if": {
                        "condition": {
                          "$eq": [
                            {
                              "$ref": "request.duration"
                            },
                            0
                          ]
                        },
                        "then": 6,
                        "else": {
                          "$ref": "request.duration"
                        }
                      }
                    },
                    6
                  ]
                },
                {
                  "$lte": [
                    {
                      "$if": {
                        "condition": {
                          "$eq": [
                            {
                              "$ref": "request.duration"
                            },
                            0
                          ]
                        },
                        "then": 6,
                        "else": {
                          "$ref": "request.duration"
                        }
                      }
                    },
                    15
                  ]
                }
              ]
            },
            "message": "APIMart Grok 视频时长必须为 6–15 秒"
          },
          {
            "assert": {
              "$in": [
                {
                  "$coalesce": [
                    {
                      "$ref": "request.aspectRatio"
                    },
                    "16:9"
                  ]
                },
                [
                  "16:9",
                  "9:16",
                  "1:1",
                  "3:2",
                  "2:3"
                ]
              ]
            },
            "message": "APIMart Grok 视频仅支持 16:9、9:16、1:1、3:2、2:3"
          },
          {
            "assert": {
              "$in": [
                {
                  "$coalesce": [
                    {
                      "$ref": "request.resolution"
                    },
                    "480p"
                  ]
                },
                [
                  "480p",
                  "720p"
                ]
              ]
            },
            "message": "APIMart Grok 视频仅支持 480p 或 720p"
          },
          {
            "assert": {
              "$lte": [
                {
                  "$len": {
                    "$ref": "request.images"
                  }
                },
                7
              ]
            },
            "message": "APIMart Grok 视频最多支持 7 张参考图"
          },
          {
            "assert": {
              "$eq": [
                {
                  "$len": {
                    "$filter": {
                      "from": {
                        "$ref": "request.images"
                      },
                      "as": "media",
                      "where": {
                        "$or": [
                          {
                            "$not": {
                              "$in": [
                                {
                                  "$first": {
                                    "$split": [
                                      {
                                        "$ref": "media.url"
                                      },
                                      ":"
                                    ]
                                  }
                                },
                                [
                                  "https",
                                  "http"
                                ]
                              ]
                            }
                          },
                          {
                            "$eq": [
                              {
                                "$ref": "media.role"
                              },
                              "last_frame"
                            ]
                          }
                        ]
                      }
                    }
                  }
                },
                0
              ]
            },
            "message": "参考图必须使用公网 HTTP(S) URL；此接口不支持 Base64 或独立尾帧"
          },
          {
            "assert": {
              "$and": [
                {
                  "$eq": [
                    {
                      "$len": {
                        "$ref": "request.videos"
                      }
                    },
                    0
                  ]
                },
                {
                  "$eq": [
                    {
                      "$len": {
                        "$ref": "request.audios"
                      }
                    },
                    0
                  ]
                }
              ]
            },
            "message": "此 Grok 接口不支持参考视频或参考音频"
          }
        ],
        "create": {
          "method": "POST",
          "path": "/v1/videos/generations",
          "originPath": true,
          "contentType": "application/json",
          "body": {
            "model": {
              "$ref": "request.model"
            },
            "prompt": {
              "$ref": "request.prompt"
            },
            "size": {
              "$coalesce": [
                {
                  "$ref": "request.aspectRatio"
                },
                "16:9"
              ]
            },
            "duration": {
              "$if": {
                "condition": {
                  "$eq": [
                    {
                      "$ref": "request.duration"
                    },
                    0
                  ]
                },
                "then": 6,
                "else": {
                  "$ref": "request.duration"
                }
              }
            },
            "resolution": {
              "$coalesce": [
                {
                  "$ref": "request.resolution"
                },
                "480p"
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
                    "$ref": "media.url"
                  }
                }
              }
            },
            "nsfw_check": {
              "$omitEmpty": {
                "$ref": "request.providerOptions.apimart-grok-imagine-video.nsfw_check"
              }
            }
          }
        },
        "poll": {
          "method": "GET",
          "path": "/v1/tasks/{{taskId}}",
          "originPath": true,
          "query": {
            "language": "zh"
          }
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
          "resultKind": "video",
          "resultEphemeral": true,
          "videos": {
            "$map": {
              "from": {
                "$ref": "response.data.result.videos"
              },
              "as": "video",
              "in": {
                "url": {
                  "$first": {
                    "$ref": "video.url"
                  }
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
