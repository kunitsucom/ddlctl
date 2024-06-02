package util

import "regexp"

type StmtRegex struct {
	Regex *regexp.Regexp
	Index int
}

//nolint:gochecknoglobals
var (
	StmtRegexCreateTable = StmtRegex{
		Regex: regexp.MustCompile(`^\s*(//+\s*|/\*\s*)?\S+\s*:\s*table(s)?\s*[: ]\s*((CREATE\s+TABLE\s+)?\S+.*)`),
		Index: 3, //nolint:mnd // Index 3 is CREATE TABLE name
	}
	StmtRegexCreateTableConstraint = StmtRegex{
		Regex: regexp.MustCompile(`^\s*(//+\s*|/\*\s*)?\S+\s*:\s*constraint(s)?\s*[: ]\s*(\S+.*)`),
		Index: 3, //nolint:mnd // Index 3 is CONSTRAINT
	}
	StmtRegexCreateTableOptions = StmtRegex{
		Regex: regexp.MustCompile(`^\s*(//+\s*|/\*\s*)?\S+\s*:\s*option(s)?\s*[: ]\s*(\S+.*)`),
		Index: 3, //nolint:mnd // Index 3 is OPTION
	}
	StmtRegexCreateIndex = StmtRegex{
		Regex: regexp.MustCompile(`^\s*(//+\s*|/\*\s*)?\S+\s*:\s*index(es)?\s*[: ]\s*(\S+.*)`),
		Index: 3, //nolint:mnd // Index 3 is INDEX name
	}
)
