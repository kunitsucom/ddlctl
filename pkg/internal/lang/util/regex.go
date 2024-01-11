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
		Index: 3,
	}
	StmtRegexCreateTableConstraint = StmtRegex{
		Regex: regexp.MustCompile(`^\s*(//+\s*|/\*\s*)?\S+\s*:\s*constraint(s)?\s*[: ]\s*(\S+.*)`),
		Index: 3,
	}
	StmtRegexCreateTableOptions = StmtRegex{
		Regex: regexp.MustCompile(`^\s*(//+\s*|/\*\s*)?\S+\s*:\s*option(s)?\s*[: ]\s*(\S+.*)`),
		Index: 3,
	}
	StmtRegexCreateIndex = StmtRegex{
		Regex: regexp.MustCompile(`^\s*(//+\s*|/\*\s*)?\S+\s*:\s*index(es)?\s*[: ]\s*(\S+.*)`),
		Index: 3,
	}
)
