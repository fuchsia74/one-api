package common

import (
	"github.com/songquanpeng/one-api/common/config"
)

var UsingSQLite = false
var UsingPostgreSQL = false
var UsingMySQL = false

var SQLitePath = config.SQLitePath
var SQLiteBusyTimeout = config.SQLiteBusyTimeout
