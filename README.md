# Firefly CLI

Firefly CLI 是 Firefly 工程侧开发辅助工具。它的定位很小：创建服务项目、维护本地项目元信息、推送已经构建好的 gateway descriptor。

当前 CLI 版本：`v0.1.0`。

代码结构已经收敛到 `internal` 包：

```text
cmd/firefly         firefly 二进制入口
internal/cli         命令树和命令输出
internal/template    基于 github.com/fireflycore/go-layout 创建服务
internal/project     .firefly/project.yaml、版本解析和本地检查
internal/descriptor  descriptor push 和 S3 上传
internal/fsutil      文件系统与文本替换工具
```

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

当前交互模式是轻量 stdin prompt：如果没有传项目名，CLI 只询问 `Project name`。更推荐日常直接传位置参数：

```bash
firefly create cms
```

非交互式创建：

```bash
firefly create \
  --name app \
  --language go \
  --template-version v0.3.3 \
  --module github.com/fireflycore/app \
  --app-id app \
  --service app \
  --non-interactive
```

也可以把项目名作为位置参数：

```bash
firefly create app \
  --module github.com/fireflycore/app \
  --app-id app \
  --service app \
  --non-interactive
```

模板固定来自 `github.com/fireflycore/go-layout`。`--template-version` 为空或为 `latest` 时，CLI 会读取模板仓库 tag 并选择最新语义化版本。

### TUI 后续扩展

当前不引入 Charmbracelet/Bubble Tea TUI，原因是 `create` 仍以脚本化、Makefile、CI 友好为主，参数数量也还不需要完整表单。

后续如果 `create` 需要同时编辑模板版本、module、app id、service、namespace、descriptor/S3 预设等更多字段，可以新增独立展示层：

```text
internal/prompt
```

约束：

- TUI 只负责收集和确认输入，不负责拉取模板、替换文件或写项目配置。
- TUI 输出统一转换为 `template.CreateOptions` 或 `project.InitOptions`。
- 非交互式 flag 和位置参数必须继续可用，不能因为 TUI 影响脚本和 CI。
- CLI 输出仍保持英文；代码注释保持中文。

## 项目元信息

在业务服务仓库中初始化本地项目配置：

```bash
firefly project init \
  --service app \
  --app-id app \
  --module github.com/fireflycore/app \
  --s3-endpoint https://minio.example.com \
  --s3-bucket descriptor \
  --s3-force-path-style
```

该命令会生成 `.firefly/project.yaml`。descriptor 默认规则为：

```text
version = conf/bootstrap.json 中的 app.version，也就是 bootstrapConf.app.version
file    = dep/protobuf/gen/{version}.pb
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

`project check` 只检查本地文件和静态配置，例如 `.firefly/project.yaml`、`go.mod`、`Makefile` / `makefile`、`buf.yaml`、`buf.gen.yaml`、`conf/bootstrap.json`、`dep/protobuf/gen/{version}.pb` 和 S3 配置。它不会连接 sidecar、gateway、authz、token 服务、配置中心或观测性系统。

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

`descriptor push` 会读取 `.firefly/project.yaml`，再从 `conf/bootstrap.json` 读取服务版本 `app.version`，解析本地文件 `dep/protobuf/gen/{version}.pb`，并通过 S3 PutObject 上传到：

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
FIREFLY_S3_ENDPOINT=https://minio.example.com
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
	VERSION=$$(jq -r '.app.version' conf/bootstrap.json); \
	TYPE_FLAGS=$$(awk ' \
		function add(value) { if (value != "" && !(value in seen)) { seen[value]=1; selected[++selected_count]=value } } \
		/^[[:space:]]*- include_package=/ { split($$0, item, "="); add(item[2]); next } \
		/^[[:space:]]*- include_service=/ { split($$0, item, "="); add(item[2]); next } \
		/^[[:space:]]*- include_package_prefix=/ { split($$0, item, "="); prefixes[++prefix_count]=item[2]; next } \
		/^[[:space:]]*types:/ { in_types=1; next } \
		in_types && /^[[:space:]]*- [A-Za-z0-9_.]+$$/ { type=$$0; gsub(/^[[:space:]]*- /, "", type); for (i=1; i<=prefix_count; i++) { if (index(type, prefixes[i]) == 1) add(type) } } \
		END { for (i=1; i<=selected_count; i++) printf "--type %s ", selected[i] } \
	' buf.gen.yaml); \
	test -n "$$VERSION" && test "$$VERSION" != "null"; \
	test -n "$$TYPE_FLAGS"; \
	mkdir -p dep/protobuf/gen; \
	buf build buf.build/lhdht/grpc:main $$TYPE_FLAGS \
	  --as-file-descriptor-set \
	  --exclude-source-info \
	  -o dep/protobuf/gen/$$VERSION.pb

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

Go 的二进制名来自 main package 路径最后一段，所以入口放在 `cmd/firefly`。构建时需要显式构建这个入口：

```bash
go build -o firefly -ldflags "-s -w" ./cmd/firefly
```

安装时也需要使用 `cmd/firefly` 路径，否则根路径 `github.com/fireflycore/cli` 不会生成名为 `firefly` 的二进制：

```bash
go install github.com/fireflycore/cli/cmd/firefly@v0.1.1
```
