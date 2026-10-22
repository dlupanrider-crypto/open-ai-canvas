# APIMart Grok Imagine 异步视频插件

在影策「插件」页面导入 `apimart-grok-imagine-video.yingce-plugin` 并启用，在系统渠道或自定义渠道选择此协议，配置 API Key，模型填写渠道提供的模型名称（例如 `grok-imagine-1.5-video-ext`），插件原样发送。默认地址 `https://api.apimart.ai`。

支持文生视频、最多 7 张公网参考图、6–15 秒和 480p/720p。任务由宿主异步轮询并下载持久化。参考图不支持 Base64；本地图片须先经宿主转换为公网可访问资源。
