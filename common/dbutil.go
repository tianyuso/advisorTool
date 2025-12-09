package common

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	// PostgreSQL non-transactional statement patterns
	dropDatabaseReg = regexp.MustCompile(`(?i)DROP DATABASE`)
	createIndexReg  = regexp.MustCompile(`(?i)CREATE(\s+(UNIQUE\s+)?)INDEX(\s+)CONCURRENTLY`)
	dropIndexReg    = regexp.MustCompile(`(?i)DROP(\s+)INDEX(\s+)CONCURRENTLY`)
	vacuumReg       = regexp.MustCompile(`(?i)^\s*VACUUM`)
)

// TrimStatement trims the unused characters from the statement.
func TrimStatement(statement string) string {
	return strings.TrimLeftFunc(strings.TrimRightFunc(statement, IsSpaceOrSemicolon), unicode.IsSpace)
}

// IsSpaceOrSemicolon checks if the rune is a space or a semicolon.
func IsSpaceOrSemicolon(r rune) bool {
	if ok := unicode.IsSpace(r); ok {
		return true
	}
	return r == ';'
}

// IsNonTransactionStatement checks if a PostgreSQL statement cannot run inside a transaction block.
func IsNonTransactionStatement(stmt string) bool {
	if len(dropDatabaseReg.FindString(stmt)) > 0 {
		return true
	}
	if len(createIndexReg.FindString(stmt)) > 0 {
		return true
	}
	if len(dropIndexReg.FindString(stmt)) > 0 {
		return true
	}
	return len(vacuumReg.FindString(stmt)) > 0
}
