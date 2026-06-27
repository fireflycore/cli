package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fireflycore/cli/internal/project"
)

func TestProjectInitTypeProtoWritesProtoProject(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	runtime := Runtime{
		WorkDir:  root,
		CacheDir: t.TempDir(),
		Out:      &out,
		Err:      &out,
	}

	err := Execute(context.Background(), runtime, []string{
		"project", "init",
		"--type", "proto",
		"--namespace", "lhdht",
		"--proto-repo", "lhdht/backend/proto",
		"--proto-module", "buf.build/lhdht/grpc",
		"--proto-version", "v0.0.1",
		"--consul-address", "http://127.0.0.1:18500",
	})
	if err != nil {
		t.Fatal(err)
	}

	cfg, _, err := project.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Project.Type != project.ProjectTypeProto {
		t.Fatalf("project.type = %s, want %s", cfg.Project.Type, project.ProjectTypeProto)
	}
	if cfg.Proto.Namespace != "lhdht" {
		t.Fatalf("proto.namespace = %s", cfg.Proto.Namespace)
	}
	if cfg.Proto.Repo != "lhdht/backend/proto" {
		t.Fatalf("proto.repo = %s", cfg.Proto.Repo)
	}
	resolved, err := cfg.Resolve(root, project.ConfigPath(root))
	if err != nil {
		t.Fatal(err)
	}
	if resolved.DescriptorCurrentKey != "lhdht/api-gateway/descriptor/current" {
		t.Fatalf("descriptor current key = %s", resolved.DescriptorCurrentKey)
	}
}

func TestProjectInitRejectsProtoRepoTypeWithProtoHint(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	runtime := Runtime{
		WorkDir:  root,
		CacheDir: t.TempDir(),
		Out:      &out,
		Err:      &out,
	}

	err := Execute(context.Background(), runtime, []string{
		"project", "init",
		"--type", "proto_repo",
	})
	if err == nil {
		t.Fatal("project init succeeded, want unsupported project type error")
	}
	if !strings.Contains(err.Error(), `use "proto"`) {
		t.Fatalf("error = %q, want proto hint", err.Error())
	}
}

func TestProjectInfoTypeProtoUsesDescriptorCurrentLabels(t *testing.T) {
	root := t.TempDir()
	runtime := Runtime{
		WorkDir:  root,
		CacheDir: t.TempDir(),
		Out:      &bytes.Buffer{},
		Err:      &bytes.Buffer{},
	}

	err := Execute(context.Background(), runtime, []string{
		"project", "init",
		"--type", "proto",
		"--namespace", "lhdht",
		"--proto-repo", "lhdht/backend/proto",
		"--proto-module", "buf.build/lhdht/grpc",
		"--proto-version", "v0.0.1",
		"--consul-address", "http://127.0.0.1:18500",
	})
	if err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	runtime.Out = &out
	err = Execute(context.Background(), runtime, []string{"project", "info"})
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	if !strings.Contains(text, "descriptor.versioned_ref: s3://descriptor/lhdht/v0.0.1.pb") {
		t.Fatalf("project info missing versioned ref label:\n%s", text)
	}
	legacyLabel := "\ndescriptor" + "_ref:"
	if strings.Contains(text, legacyLabel) {
		t.Fatalf("proto project info must not use legacy descriptor label:\n%s", text)
	}
	if !strings.Contains(text, "descriptor.current_ref: s3://descriptor/lhdht/current.pb") {
		t.Fatalf("project info missing current ref:\n%s", text)
	}
}

func TestProjectInfoTypeServiceDoesNotShowDescriptorPublishingFields(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "conf"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module github.com/fireflycore/app\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "conf", "bootstrap.json"), []byte(`{"app":{"version":"v1.2.3"}}`), 0644); err != nil {
		t.Fatal(err)
	}

	runtime := Runtime{
		WorkDir:  root,
		CacheDir: t.TempDir(),
		Out:      &bytes.Buffer{},
		Err:      &bytes.Buffer{},
	}
	err := Execute(context.Background(), runtime, []string{
		"project", "init",
		"--service", "app",
		"--app-id", "app",
		"--module", "github.com/fireflycore/app",
		"--s3-endpoint", "https://minio.example.com",
	})
	if err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	runtime.Out = &out
	err = Execute(context.Background(), runtime, []string{"project", "info"})
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	if !strings.Contains(text, "version: v1.2.3") {
		t.Fatalf("project info missing service version:\n%s", text)
	}
	for _, disallowed := range []string{
		"descriptor.file:",
		"descriptor.object_key:",
		"descriptor.versioned_ref:",
		"descriptor" + "_ref:",
	} {
		if strings.Contains(text, disallowed) {
			t.Fatalf("service project info must not show %s\n%s", disallowed, text)
		}
	}
	configBytes, err := os.ReadFile(filepath.Join(root, ".firefly", "project.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	configText := string(configBytes)
	for _, disallowed := range []string{"\ndescriptor:", "\ns3:"} {
		if strings.Contains(configText, disallowed) {
			t.Fatalf("service project config must not contain %s\n%s", disallowed, configText)
		}
	}
}

func TestDescriptorPublishTypeProtoUsesDescriptorCurrentLabels(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	runtime := Runtime{
		WorkDir:  root,
		CacheDir: t.TempDir(),
		Out:      &out,
		Err:      &out,
	}

	err := Execute(context.Background(), runtime, []string{
		"project", "init",
		"--type", "proto",
		"--namespace", "lhdht",
		"--proto-repo", "lhdht/backend/proto",
		"--proto-module", "buf.build/lhdht/grpc",
		"--proto-version", "v0.0.1",
		"--consul-address", "http://127.0.0.1:18500",
	})
	if err != nil {
		t.Fatal(err)
	}

	out.Reset()
	err = Execute(context.Background(), runtime, []string{"descriptor", "publish", "--dry-run", "--buf", fakeBuf(t, root)})
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	if !strings.Contains(text, "descriptor.versioned_ref: s3://descriptor/lhdht/v0.0.1.pb") {
		t.Fatalf("publish output missing versioned ref label:\n%s", text)
	}
	legacyVersionedLabel := "\ndescriptor" + "_ref:"
	legacyCurrentLabel := "\ncurrent_descriptor" + "_ref:"
	if strings.Contains(text, legacyVersionedLabel) || strings.Contains(text, legacyCurrentLabel) {
		t.Fatalf("publish output must not use legacy descriptor labels:\n%s", text)
	}
	if !strings.Contains(text, "descriptor.current_ref: s3://descriptor/lhdht/current.pb") {
		t.Fatalf("publish output missing current ref label:\n%s", text)
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
