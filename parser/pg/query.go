package pg

import (
	storepb "github.com/tianyuso/advisorTool/generated-go/store"
	"github.com/tianyuso/advisorTool/parser/base"
)

func init() {
	base.RegisterQueryValidator(storepb.Engine_POSTGRES, validateQueryANTLR)
	// Redshift has its own implementation in the redshift package
	base.RegisterQueryValidator(storepb.Engine_COCKROACHDB, validateQueryANTLR)
}
