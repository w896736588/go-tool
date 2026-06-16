# dtool-notify 详细说明

## 必要约束

- 调用前，先向用户确认 `webhook`、消息内容，以及是否需要 `secret`
- 该 skill 默认发送纯文本消息，不负责 Markdown 富文本
- 单条消息不宜过长，超长内容应先压缩
- 需要具体调用参数或脚本入口时，再去看 `scripts/send_dingtalk.py`

## 文件索引

- 钉钉通知脚本：`scripts/send_dingtalk.py`
