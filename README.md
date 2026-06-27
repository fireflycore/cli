# Firefly CLI

Firefly CLI 是 Firefly 工程侧开发辅助工具。它负责创建服务项目、维护本地项目元信息，以及在 proto 仓库中生成、上传、发布 api-gateway descriptor。

当前 CLI 版本：`v0.1.3`。

代码结构：

```text
cmd/firefly         firefly 二进制入口
internal/cli         命令树和命令输出
internal/template    基于 github.com/fireflycore/go-layout 创建服务
internal/project     .firefly/project.yaml、版本解析和本地检查
internal/descriptor  descriptor build / push / publish
internal/fsutil      文件系统与文本替换工具
```

主要命令：

```text
firefly create
firefly project init
firefly project info
firefly project check
firefly descriptor build
firefly descriptor push
firefly descriptor publish
```

## 项目类型

`.firefly/project.yaml` 支持两类项目：

```yaml
project:
  type: service
```

```yaml
project:
  type: proto
```

`service` 是默认值，兼容现有业务服务仓库。`proto` 表示按 namespace 发布 whole-repo descriptor 的 proto 项目。项目类型值直接叫 `proto`，不要写成 `proto_repo`。

proto 项目不定义 `service` 或 `bootstrap` 配置块；descriptor 路径模板按 namespace/repo/version 推导，不使用 `${service}` 或 `${app_id}`。

## 初始化

业务服务仓库：

```bash
firefly project init \
  --service app \
  --app-id app \
  --module github.com/fireflycore/app
```

业务服务项目默认不生成、上传或展示 api-gateway descriptor 引用；只保留 gateway manifest 和 route document facts。
业务服务项目不写入 `descriptor` 或 `s3` 配置块；这两个配置只属于 `project.type: proto` 的 proto 仓库。

proto 仓库：

```bash
firefly project init \
  --type proto \
  --namespace lhdht \
  --proto-repo lhdht/backend/proto \
  --proto-module buf.build/lhdht/grpc \
  --proto-source . \
  --proto-version v0.0.1 \
  --consul-address http://127.0.0.1:18500 \
  --s3-bucket descriptor \
  --s3-endpoint https://minio.example.com \
  --s3-force-path-style
```

proto 项目默认 descriptor 规则：

```text
file            = dep/protobuf/gen/{namespace}/{version}.pb
current_file    = dep/protobuf/gen/{namespace}/current.pb
bucket          = descriptor
key             = {namespace}/{version}.pb
current_key     = {namespace}/current.pb
consul_kv       = {namespace}/api-gateway/descriptor/current
```

查看配置：

```bash
firefly project info
```

本地检查：

```bash
firefly project check
```

## Descriptor 发布

proto 仓库发布推荐使用：

```bash
firefly descriptor publish
```

它会执行：

1. `buf build . --as-file-descriptor-set --exclude-source-info -o dep/protobuf/gen/{namespace}/{version}.pb`
2. 复制本地 `current.pb`
3. 上传 versioned pb 到 S3/MinIO
4. 上传/覆盖 current pb 到 S3/MinIO
5. 写入 Consul KV `{namespace}/api-gateway/descriptor/current`

只生成本地 pb：

```bash
firefly descriptor build
```

只上传已有 pb：

```bash
firefly descriptor push
```

`descriptor build/push/publish` 都要求当前项目为 `project.type: proto`。普通业务服务项目只生成 gateway manifest 和 route document facts，不生成或上传 api-gateway descriptor pb。

只解析路径和计算摘要，不写远端状态：

```bash
firefly descriptor publish --dry-run
firefly descriptor push --dry-run
```

常用参数：

```bash
firefly descriptor publish \
  --version v0.0.2 \
  --source . \
  --source-revision "$(git rev-parse HEAD)"
```

跳过某些阶段：

```bash
firefly descriptor publish --skip-build
firefly descriptor publish --skip-consul
firefly descriptor publish --skip-current-object
```

descriptor current JSON 形态：

```json
{
  "schema": "firefly.api_gateway.descriptor.v1",
  "namespace": "lhdht",
  "version": "v0.0.1",
  "ref": "s3://descriptor/lhdht/v0.0.1.pb",
  "current_ref": "s3://descriptor/lhdht/current.pb",
  "sha256": "hex-encoded-sha256",
  "proto_repo": "lhdht/backend/proto",
  "source_revision": "git-sha-or-tag",
  "published_at": "2026-06-24T00:00:00Z"
}
```

这里的 `proto_repo` 是来源元数据字段，不是 project type。

## S3 凭证

`descriptor push/publish` 使用 AWS SDK 默认凭证链，并支持 STS 临时凭证：

```bash
AWS_ACCESS_KEY_ID=...
AWS_SECRET_ACCESS_KEY=...
AWS_SESSION_TOKEN=...
AWS_REGION=us-east-1
FIREFLY_S3_ENDPOINT=https://minio.example.com
FIREFLY_S3_BUCKET=descriptor
FIREFLY_S3_FORCE_PATH_STYLE=true
```

兼容目标包括 AWS S3、阿里云 OSS S3 兼容接口、腾讯云 COS S3 兼容接口、MinIO，以及其他支持 S3 协议的对象存储。

## 构建 CLI

```bash
go build -o firefly -ldflags "-s -w" ./cmd/firefly
```

安装：

```bash
go install github.com/fireflycore/cli/cmd/firefly@v0.1.3
```
