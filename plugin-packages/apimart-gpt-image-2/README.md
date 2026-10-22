# APIMart GPT-Image-2 异步图片插件

在影策「插件」页面导入 `apimart-gpt-image-2.yingce-plugin`，启用后在系统渠道或自定义渠道选择此协议，填写 API Key，模型填写 `gpt-image-2`（或 `gpt-image-2-ext`）。默认 Base URL 为 `https://api.apimart.ai`。

支持文生图、最多 15 张参考图和 1k / 2k / 4k 分辨率；提交后由宿主轮询并保存图片。API Key 通过 Bearer 请求头发送。
