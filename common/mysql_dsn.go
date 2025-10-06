package common

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	gosqlmysql "github.com/go-sql-driver/mysql"

	"github.com/Laisky/errors/v2"
)

// NormalizeMySQLDSN converts mysql:// URLs to go-sql-driver compatible DSNs and enforces parseTime=true
// so DATETIME/TIMESTAMP fields map to time.Time in scans. When no explicit loc parameter is provided,
// the location defaults to UTC to honour the repository-wide UTC requirement.
func NormalizeMySQLDSN(dsn string) (string, error) {
	normalized, err := convertMySQLURLToDSN(dsn)
	if err != nil {
		return "", errors.Wrap(err, "convert MySQL DSN")
	}

	cfg, err := gosqlmysql.ParseDSN(normalized)
	if err != nil {
		return "", errors.Wrap(err, "parse MySQL DSN")
	}

	cfg.ParseTime = true

	if !containsMySQLLocOption(normalized) {
		cfg.Loc = time.UTC
	}

	return cfg.FormatDSN(), nil
}

func convertMySQLURLToDSN(dsn string) (string, error) {
	if !strings.HasPrefix(strings.ToLower(dsn), "mysql://") {
		return dsn, nil
	}

	parsed, err := url.Parse(dsn)
	if err != nil {
		return "", errors.Wrap(err, "parse mysql:// DSN")
	}

	if parsed.Host == "" {
		return "", errors.New("mysql DSN missing host")
	}

	userInfo := ""
	if parsed.User != nil {
		userInfo = parsed.User.Username()
		if pwd, ok := parsed.User.Password(); ok {
			userInfo = fmt.Sprintf("%s:%s", userInfo, pwd)
		}
	}

	dbName := strings.TrimPrefix(parsed.Path, "/")
	base := ""
	if userInfo != "" {
		base = fmt.Sprintf("%s@", userInfo)
	}
	base += fmt.Sprintf("tcp(%s)/%s", parsed.Host, dbName)

	if parsed.RawQuery != "" {
		base = fmt.Sprintf("%s?%s", base, parsed.RawQuery)
	}

	return base, nil
}

func containsMySQLLocOption(dsn string) bool {
	idx := strings.Index(dsn, "?")
	if idx == -1 {
		return false
	}

	query := dsn[idx+1:]
	values, err := url.ParseQuery(query)
	if err != nil {
		return false
	}

	_, ok := values["loc"]
	return ok
}
