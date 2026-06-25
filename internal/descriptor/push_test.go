package descriptor

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/fireflycore/cli/internal/project"
)

func TestBuildProtoDescriptorWithFakeBuf(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module github.com/lhdht/proto\n")
	saveProtoProject(t, root)
	buf := fakeBuf(t, root)

	result, err := Build(context.Background(), BuildOptions{
		Root: root,
		Buf:  buf,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Namespace != "lhdht" {
		t.Fatalf("namespace = %s", result.Namespace)
	}
	if result.Version != "v0.0.1" {
		t.Fatalf("version = %s", result.Version)
	}
	if result.File != filepath.Join(root, "dep", "protobuf", "gen", "lhdht", "v0.0.1.pb") {
		t.Fatalf("file = %s", result.File)
	}
	if data := readFile(t, result.CurrentFile); string(data) != "fake descriptor" {
		t.Fatalf("current descriptor = %q", data)
	}
	if result.SHA256 == "" || result.Size == 0 {
		t.Fatalf("digest not populated: sha=%s size=%d", result.SHA256, result.Size)
	}
}

func TestPublishDryRunBuildsCurrentJSON(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module github.com/lhdht/proto\n")
	saveProtoProject(t, root)
	buf := fakeBuf(t, root)

	result, err := Publish(context.Background(), PublishOptions{
		PushOptions: PushOptions{
			Root:   root,
			DryRun: true,
		},
		Buf:            buf,
		SourceRevision: "abc123",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.CurrentKey != "lhdht/api-gateway/descriptor/current" {
		t.Fatalf("current key = %s", result.CurrentKey)
	}
	var doc map[string]string
	if err := json.Unmarshal(result.CurrentJSON, &doc); err != nil {
		t.Fatal(err)
	}
	if doc["schema"] != "firefly.api_gateway.descriptor.v1" {
		t.Fatalf("schema = %s", doc["schema"])
	}
	if doc["namespace"] != "lhdht" {
		t.Fatalf("namespace = %s", doc["namespace"])
	}
	if doc["version"] != "v0.0.1" {
		t.Fatalf("version = %s", doc["version"])
	}
	if doc["ref"] != "s3://descriptor/lhdht/v0.0.1.pb" {
		t.Fatalf("ref = %s", doc["ref"])
	}
	if doc["current_ref"] != "s3://descriptor/lhdht/current.pb" {
		t.Fatalf("current_ref = %s", doc["current_ref"])
	}
	if doc["proto_repo"] != "lhdht/backend/proto" {
		t.Fatalf("proto_repo = %s", doc["proto_repo"])
	}
	if doc["source_revision"] != "abc123" {
		t.Fatalf("source_revision = %s", doc["source_revision"])
	}
	if doc["sha256"] == "" {
		t.Fatal("sha256 is empty")
	}
}

func saveProtoProject(t *testing.T, root string) {
	t.Helper()
	cfg, err := project.NewConfig(project.InitOptions{
		Root:          root,
		ProjectType:   project.ProjectTypeProto,
		Namespace:     "lhdht",
		ProtoRepo:     "lhdht/backend/proto",
		ProtoModule:   "buf.build/lhdht/grpc",
		ProtoVersion:  "v0.0.1",
		ConsulAddress: "http://127.0.0.1:18500",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = project.Save(root, cfg, false); err != nil {
		t.Fatal(err)
	}
}

func fakeBuf(t *testing.T, root string) string {
	t.Helper()
	path := filepath.Join(root, "fake-buf.sh")
	script := "#!/bin/sh\nout=\"\"\nwhile [ \"$#\" -gt 0 ]; do\n  if [ \"$1\" = \"-o\" ]; then\n    shift\n    out=\"$1\"\n  fi\n  shift\ndone\nmkdir -p \"$(dirname \"$out\")\"\nprintf 'fake descriptor' > \"$out\"\n"
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
