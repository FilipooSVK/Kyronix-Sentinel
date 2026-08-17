package analyzer

import (
	"kyronix/sentinel/internal/analyzer/rules"
)

// NewDefaultAnalyzer creates the production analyzer configuration.
//
// All built-in health rules are enabled by default.
func NewDefaultAnalyzer() *Analyzer {
	return NewAnalyzer(
		rules.NewMemoryRule(),
		rules.NewPressureRule(),
		rules.NewDiskRule(),
		rules.NewKernelRule(),
	)
}
