package store

import (
	"github.com/fireflycore/cli/pkg/buf"
	"github.com/fireflycore/cli/pkg/config"
)

type _CoreEntity struct {
	Buf    *buf.CoreEntity
	Config *config.CoreEntity
}

var Use = new(_CoreEntity)
