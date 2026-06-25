package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigResolveUsesBootstrapVersion(t *testing.T) {
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

	// 读取配置后解析 descriptor 路径。
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

	// descriptor 文件名必须使用版本号，而不是服务名。
	wantFile := filepath.Join(root, "dep", "protobuf", "gen", "v1.2.3.pb")
	if resolved.DescriptorFile != wantFile {
		t.Fatalf("descriptor file mismatch: %s", resolved.DescriptorFile)
	}
	if resolved.ObjectKey != "app/v1.2.3.pb" {
		t.Fatalf("object key mismatch: %s", resolved.ObjectKey)
	}
	if resolved.DescriptorRef != "https://minio.example.com/descriptor/app/v1.2.3.pb" {
		t.Fatalf("descriptor ref mismatch: %s", resolved.DescriptorRef)
	}
}

func TestCheckAcceptsLowercaseMakefile(t *testing.T) {
	// go-layout 当前使用小写 makefile，检查逻辑需要接受它。
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module github.com/fireflycore/app\n")
	writeFile(t, root, "makefile", "descriptor:\n\tbuf build\n")
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
	if resolved.DescriptorRef != "s3://descriptor/lhdht/v0.0.2.pb" {
		t.Fatalf("descriptor ref mismatch: %s", resolved.DescriptorRef)
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
