package common

import (
	"sync/atomic"

	"github.com/songquanpeng/one-api/common/config"
)

var UsingSQLite atomic.Bool
var UsingPostgreSQL atomic.Bool
var UsingMySQL atomic.Bool

var SQLitePath = config.SQLitePath
var SQLiteBusyTimeout = config.SQLiteBusyTimeout
