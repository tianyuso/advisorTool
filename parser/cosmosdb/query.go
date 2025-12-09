package cosmosdb

import (
	storepb "advisorTool/generated-go/store"
	"advisorTool/parser/base"
)

func init() {
	base.RegisterQueryValidator(storepb.Engine_COSMOSDB, validateQuery)
}

func validateQuery(statement string) (bool, bool, error) {
	_, err := ParseCosmosDBQuery(statement)
	if err != nil {
		return false, false, err
	}

	return true, true, nil
}
