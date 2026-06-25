package cli

import (
	"bytes"
	"context"
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
