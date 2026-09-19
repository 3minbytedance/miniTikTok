# HTTP 接口文档

所有业务接口以 `/douyin` 为前缀；除标注外，写操作需登录（query 或表单携带 `token`）。
网关地址：`http://<host>:8080`。

## 用户（social 服务）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/douyin/user/register/` | 注册（username, password） |
| POST | `/douyin/user/login/` | 登录 |
| GET | `/douyin/user/` | 用户信息（user_id，token 可选） |

## 视频（video 服务）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/douyin/feed/` | 视频流（latest_time 可选，token 可选） |
| POST | `/douyin/publish/action/` | 发布视频（multipart：data, title, token），文件 1MB~50MB，走 Kafka 异步处理 |
| GET | `/douyin/publish/list/` | 用户作品列表（user_id） |

## 评论（video 服务）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/douyin/comment/action/` | 评论/删评论（video_id, action_type 1/2, comment_text） |
| GET | `/douyin/comment/list/` | 评论列表（video_id） |

## 点赞（video 服务）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/douyin/favorite/action/` | 点赞/取消（video_id, action_type 1/2） |
| GET | `/douyin/favorite/list/` | 用户点赞列表（user_id） |

## 关系链（social 服务）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/douyin/relation/action/` | 关注/取关（to_user_id, action_type 1/2） |
| GET | `/douyin/relation/follow/list/` | 关注列表（user_id） |
| GET | `/douyin/relation/follower/list/` | 粉丝列表（user_id） |
| GET | `/douyin/relation/friend/list/` | 好友列表（user_id，互关好友） |

## 私信（message 服务）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/douyin/message/action/` | 发送私信（to_user_id, content），需互关好友 |
| GET | `/douyin/message/chat/` | 私信记录（to_user_id, pre_msg_time） |

## 响应格式

业务响应统一为：

```json
{
  "status_code": 0,
  "status_msg": ""
}
```

`status_code=0` 表示成功；非 0 时 `status_msg` 为错误描述，完整错误码见
[common/response.go](../common/response.go)。

## 网关辅助接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/metrics` | Prometheus 指标（HTTP 计数/时延、缓存命中等） |
| GET | `/static/*` | OSS 未配置时的本地视频/封面访问 |

## 限流规则（按 IP，Redis 计数）

| 操作 | 限制 |
| --- | --- |
| 登录 / 注册 | 5 次 / 2 小时 |
| 发布视频 | 3 次 / 2 小时 |
| 评论 | 30 次 / 2 小时 |
| 发送私信 | 50 次 / 2 小时 |

超限返回 `status_code=18`（请求次数过多，已被限制）。
