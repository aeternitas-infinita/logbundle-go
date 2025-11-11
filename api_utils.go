package logbundle

import (
	"log/slog"

	"github.com/aeternitas-infinita/logbundle-go/pkg/core"
)

func ErrAttr(err error) slog.Attr {
	return core.ErrAttr(err)
}

func GetLvlFromStr(s string) slog.Level {
	return core.GetLvlFromStr(s)
}

func GetBoolFromStr(s string) bool {
	return core.GetBoolFromStr(s)
}
