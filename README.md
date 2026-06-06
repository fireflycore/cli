# Firefly CLI

Firefly CLI 是 Firefly 工程侧开发辅助工具。它的定位很小：创建服务项目、维护本地项目元信息、推送已经构建好的 gateway descriptor。

当前只保留这些命令：

```text
firefly create
firefly project init
firefly project info
firefly project check
firefly descriptor push
```

CLI 不接管 Buf 生成、descriptor 构建、sidecar、gateway、authz、token、config、观测性或运行时联调。这些能力分别交给业务仓库 Makefile、Buf CLI、部署系统、管理后台或 AI 辅助检查。

## 创建项目

交互式创建：

```bash
firefly create
```

非交互式创建：

```bash
firefly create \
  --name app \
  --language go \
  --template-version v0.3.5 \
  --module github.com/fireflycore/app \
  --app-id app \
  --service app \
  --non-interactive
```

`--version` 保留为兼容参数，语义等同于 `--template-version`。

## 项目元信息

在业务服务仓库中初始化本地项目配置：

```bash
firefly project init \
  --service app \
  --app-id app \
  --module github.com/fireflycore/app \
  --s3-endpoint https://minio.lhdht.cn \
  --s3-bucket descriptor \
  --s3-force-path-style
```

该命令会生成 `.firefly/project.yaml`。descriptor 默认规则为：

```text
version = conf/bootstrap.json 中的 app.version，也就是 bootstrapConf.app.version
file    = dist/descriptors/{version}.pb
bucket  = descriptor
key     = {service}/{version}.pb
```

查看项目元信息：

```bash
firefly project info
```

执行本地检查：

```bash
firefly project check
```

`project check` 只检查本地文件和静态配置，例如 `.firefly/project.yaml`、`go.mod`、`Makefile`、`buf.yaml`、`buf.gen.yaml`、`conf/bootstrap.json`、`dist/descriptors/{version}.pb` 和 S3 配置。它不会连接 sidecar、gateway、authz、token 服务、配置中心或观测性系统。

## 推送 Descriptor

descriptor 由业务仓库的 Makefile 构建，CLI 只负责推送：

```bash
make descriptor
firefly descriptor push
```

只解析路径和计算摘要，不执行上传：

```bash
firefly descriptor push --dry-run
```

`descriptor push` 会读取 `.firefly/project.yaml`，再从 `conf/bootstrap.json` 读取服务版本 `app.version`，解析本地文件 `dist/descriptors/{version}.pb`，并通过 S3 PutObject 上传到：

```text
bucket = descriptor
key    = {service}/{version}.pb
```

命令输出包括：

```text
file
size
sha256
bucket
key
descriptor_ref
```

## S3 凭证

`descriptor push` 使用 AWS SDK 的默认凭证链，并支持 STS 临时凭证。常用环境变量如下：

```bash
AWS_ACCESS_KEY_ID=...
AWS_SECRET_ACCESS_KEY=...
AWS_SESSION_TOKEN=...
AWS_REGION=us-east-1
FIREFLY_S3_ENDPOINT=https://minio.lhdht.cn
FIREFLY_S3_BUCKET=descriptor
FIREFLY_S3_FORCE_PATH_STYLE=true
```

兼容目标包括 AWS S3、阿里云 OSS S3 兼容接口、腾讯云 COS S3 兼容接口、MinIO，以及其他支持 S3 协议的对象存储。

CLI 不负责申请 STS。它只消费环境变量、AWS profile、项目配置或 AWS SDK 默认链中已经存在的凭证。

## Makefile 分工

推荐业务仓库提供以下 Makefile 目标：

```makefile
proto:
	buf generate

descriptor:
	mkdir -p dist/descriptors
	VERSION=$$(jq -r '.app.version' conf/bootstrap.json); \
	buf build buf.build/lhdht/grpc:main \
	  --as-file-descriptor-set \
	  --exclude-source-info \
	  -o dist/descriptors/$$VERSION.pb

descriptor-push:
	firefly descriptor push
```

分工边界：

| 能力 | 归属 |
| --- | --- |
| proto 生成 | Makefile + Buf CLI |
| descriptor build | Makefile + Buf CLI |
| descriptor push | Firefly CLI |
| 服务创建 | Firefly CLI |
| 项目元信息 | Firefly CLI |
| 运行时联调 | 部署系统、管理后台、测试流水线或 AI 辅助 |

## 构建 CLI

```bash
go build -ldflags "-s -w"
```
