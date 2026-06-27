package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServiceConfigResolveDoesNotInferDescriptorObjectRef(t *testing.T) {
	// 构造一个最小业务项目目录。
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module github.com/fireflycore/app\n")
	writeFile(t, root, "conf/bootstrap.json", `{"app":{"version":"v1.2.3"}}`)

	// 生成并保存项目配置。
	cfg, err := NewConfig(InitOptions{
		Root:        root,
		ServiceName: "app",
		Module:      "github.com/fireflycore/app",
		S3Endpoint:  "https://minio.example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	configPath, err := Save(root, cfg, false)
	if err != nil {
		t.Fatal(err)
	}

	// 读取配置后只解析服务版本；service 项目不再解析 descriptor 路径。
	loaded, path, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if path != configPath {
		t.Fatalf("config path mismatch: %s != %s", path, configPath)
	}
	resolved, err := loaded.Resolve(root, path)
	if err != nil {
		t.Fatal(err)
	}

	if resolved.Version != "v1.2.3" {
		t.Fatalf("version mismatch: %s", resolved.Version)
	}
	if resolved.VersionedRef != "" {
		t.Fatalf("service project must not infer descriptor object ref: %s", resolved.VersionedRef)
	}
	if resolved.DescriptorFile != "" {
		t.Fatalf("service project must not resolve descriptor file: %s", resolved.DescriptorFile)
	}
	if resolved.ObjectKey != "" {
		t.Fatalf("service project must not resolve descriptor object key: %s", resolved.ObjectKey)
	}
}

func TestCheckAcceptsLowercaseMakefile(t *testing.T) {
	// go-layout 当前使用小写 makefile，检查逻辑需要接受它。
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module github.com/fireflycore/app\n")
	writeFile(t, root, "makefile", "generate:\n\tbuf generate\n")
	writeFile(t, root, "buf.gen.yaml", "version: v2\n")
	writeFile(t, root, "conf/bootstrap.json", `{"app":{"version":"v0.0.1"}}`)
	writeFile(t, root, "dep/protobuf/gen/v0.0.1.pb", "descriptor")

	cfg, err := NewConfig(InitOptions{
		Root:        root,
		ServiceName: "app",
		Module:      "github.com/fireflycore/app",
		S3Endpoint:  "https://minio.example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Save(root, cfg, false); err != nil {
		t.Fatal(err)
	}

	results, _, _ := Check(root)
	status := findStatus(results, "makefile")
	if status != CheckOK {
		t.Fatalf("makefile status = %s, want %s", status, CheckOK)
	}
}

func TestCheckServiceProjectDoesNotRequireDescriptorPublishing(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module github.com/fireflycore/app\n")
	writeFile(t, root, "makefile", "proto:\n\tbuf generate\n")
	writeFile(t, root, "buf.yaml", "version: v2\n")
	writeFile(t, root, "buf.gen.yaml", "version: v2\n")
	writeFile(t, root, "conf/bootstrap.json", `{"app":{"version":"v0.0.1"}}`)

	cfg, err := NewConfig(InitOptions{
		Root:        root,
		ServiceName: "app",
		Module:      "github.com/fireflycore/app",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Save(root, cfg, false); err != nil {
		t.Fatal(err)
	}

	results, _, _ := Check(root)
	for _, name := range []string{"descriptor file", "current descriptor file", "descriptor current key", "s3 endpoint", "s3 bucket", "s3 credentials"} {
		if status := findStatus(results, name); status != "" {
			t.Fatalf("%s status = %s, want no service-project check", name, status)
		}
	}
	if status := findStatus(results, "bootstrap version"); status != CheckOK {
		t.Fatalf("bootstrap version status = %s, want %s", status, CheckOK)
	}
}

func TestProtoConfigResolveUsesProjectTypeProto(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module github.com/lhdht/proto\n")

	cfg, err := NewConfig(InitOptions{
		Root:          root,
		ProjectType:   ProjectTypeProto,
		Namespace:     "lhdht",
		ProtoRepo:     "lhdht/backend/proto",
		ProtoModule:   "buf.build/lhdht/grpc",
		ProtoVersion:  "v0.0.2",
		ConsulAddress: "http://127.0.0.1:18500",
	})
	if err != nil {
		t.Fatal(err)
	}
	configPath, err := Save(root, cfg, false)
	if err != nil {
		t.Fatal(err)
	}

	loaded, path, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if path != configPath {
		t.Fatalf("config path mismatch: %s != %s", path, configPath)
	}
	if !loaded.IsProtoProject() {
		t.Fatalf("project type = %s, want %s", loaded.Project.Type, ProjectTypeProto)
	}
	resolved, err := loaded.Resolve(root, path)
	if err != nil {
		t.Fatal(err)
	}

	wantFile := filepath.Join(root, "dep", "protobuf", "gen", "lhdht", "v0.0.2.pb")
	if resolved.DescriptorFile != wantFile {
		t.Fatalf("descriptor file mismatch: %s", resolved.DescriptorFile)
	}
	if resolved.CurrentDescriptorFile != filepath.Join(root, "dep", "protobuf", "gen", "lhdht", "current.pb") {
		t.Fatalf("current descriptor file mismatch: %s", resolved.CurrentDescriptorFile)
	}
	if resolved.ObjectKey != "lhdht/v0.0.2.pb" {
		t.Fatalf("object key mismatch: %s", resolved.ObjectKey)
	}
	if resolved.CurrentObjectKey != "lhdht/current.pb" {
		t.Fatalf("current object key mismatch: %s", resolved.CurrentObjectKey)
	}
	if resolved.DescriptorCurrentKey != "lhdht/api-gateway/descriptor/current" {
		t.Fatalf("descriptor current key mismatch: %s", resolved.DescriptorCurrentKey)
	}
	if resolved.VersionedRef != "s3://descriptor/lhdht/v0.0.2.pb" {
		t.Fatalf("descriptor ref mismatch: %s", resolved.VersionedRef)
	}
}

func TestCheckProtoProjectValidatesDescriptorCurrentInputs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module github.com/lhdht/proto\n")
	writeFile(t, root, "buf.yaml", "version: v2\n")
	writeFile(t, root, "buf.gen.yaml", "version: v2\n")
	writeFile(t, root, filepath.Join("dep", "protobuf", "gen", "lhdht", "v0.0.1.pb"), "descriptor")
	writeFile(t, root, filepath.Join("dep", "protobuf", "gen", "lhdht", "current.pb"), "descriptor")

	cfg, err := NewConfig(InitOptions{
		Root:          root,
		ProjectType:   ProjectTypeProto,
		Namespace:     "lhdht",
		ProtoRepo:     "lhdht/backend/proto",
		ProtoModule:   "buf.build/lhdht/grpc",
		ProtoVersion:  "v0.0.1",
		S3Endpoint:    "https://minio.example.com",
		ConsulAddress: "http://127.0.0.1:18500",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Save(root, cfg, false); err != nil {
		t.Fatal(err)
	}

	results, _, _ := Check(root)
	for _, name := range []string{
		"project config",
		"project type",
		"buf.yaml",
		"proto version",
		"descriptor file",
		"current descriptor file",
		"descriptor current key",
		"s3 endpoint",
		"s3 bucket",
	} {
		if status := findStatus(results, name); status != CheckOK {
			t.Fatalf("%s status = %s, want %s", name, status, CheckOK)
		}
	}
}

func TestCheckProtoProjectAcceptsS3EndpointFromEnv(t *testing.T) {
	root := t.TempDir()
	t.Setenv("FIREFLY_S3_ENDPOINT", "https://minio.example.com")
	writeFile(t, root, "go.mod", "module github.com/lhdht/proto\n")
	writeFile(t, root, "buf.yaml", "version: v2\n")
	writeFile(t, root, filepath.Join("dep", "protobuf", "gen", "lhdht", "v0.0.1.pb"), "descriptor")
	writeFile(t, root, filepath.Join("dep", "protobuf", "gen", "lhdht", "current.pb"), "descriptor")

	cfg, err := NewConfig(InitOptions{
		Root:          root,
		ProjectType:   ProjectTypeProto,
		Namespace:     "lhdht",
		ProtoRepo:     "lhdht/backend/proto",
		ProtoModule:   "buf.build/lhdht/grpc",
		ProtoVersion:  "v0.0.1",
		ConsulAddress: "http://127.0.0.1:18500",
	})
	if err != nil {
		t.Fatal(err)
	}
	cfg.S3.Endpoint = ""
	if _, err = Save(root, cfg, false); err != nil {
		t.Fatal(err)
	}

	results, _, _ := Check(root)
	if status := findStatus(results, "s3 endpoint"); status != CheckOK {
		t.Fatalf("s3 endpoint status = %s, want %s", status, CheckOK)
	}
}

func TestProtoRepoProjectTypeIsRejectedWithProtoHint(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".firefly/project.yaml", `schema: firefly.project.v1
project:
  type: proto_repo
`)

	_, _, err := Load(root)
	if err == nil {
		t.Fatal("Load succeeded, want unsupported project type error")
	}
	if !strings.Contains(err.Error(), `use "proto"`) {
		t.Fatalf("error = %q, want proto hint", err.Error())
	}
}

func TestServiceProjectConfigRejectsDescriptorAndS3Sections(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".firefly/project.yaml", `schema: firefly.project.v1
project:
  type: service
service:
  name: app
  app_id: app
  namespace: lhdht
  language: go
  module: github.com/fireflycore/app
bootstrap:
  file: conf/bootstrap.json
  version_path: app.version
descriptor:
  dir: dep/protobuf/gen
s3:
  bucket: descriptor
`)

	_, _, err := Load(root)
	if err == nil {
		t.Fatal("Load succeeded, want descriptor/s3 rejection")
	}
	if !strings.Contains(err.Error(), "service project must not define descriptor config") {
		t.Fatalf("error = %q, want descriptor config rejection", err.Error())
	}
}

func writeFile(t *testing.T, root, name, content string) {
	// 写入测试文件前确保目录存在。
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func findStatus(results []CheckResult, name string) CheckStatus {
	// 从检查结果中查找指定项。
	for _, result := range results {
		if result.Name == name {
			return result.Status
		}
	}
	return ""
}
