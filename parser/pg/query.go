package pg

import (
	storepb "advisorTool/generated-go/store"
	"advisorTool/parser/base"
)

func init() {
	base.RegisterQueryValidator(storepb.Engine_POSTGRES, validateQueryANTLR)
	// Redshift has its own implementation in the redshift package
	base.RegisterQueryValidator(storepb.Engine_COCKROACHDB, validateQueryANTLR)
}
