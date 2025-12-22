package snowflake

import (
	"context"

	storepb "github.com/tianyuso/advisorTool/generated-go/store"
	"github.com/tianyuso/advisorTool/parser/base"
)

func init() {
	base.RegisterGetQuerySpan(storepb.Engine_SNOWFLAKE, GetQuerySpan)
}

// GetQuerySpan returns the query span for the given statement.
func GetQuerySpan(
	ctx context.Context,
	gCtx base.GetQuerySpanContext,
	statement, database, schema string,
	ignoreCaseSensitive bool,
) (*base.QuerySpan, error) {
	q := newQuerySpanExtractor(database, schema, gCtx, ignoreCaseSensitive)
	querySpan, err := q.getQuerySpan(ctx, statement)
	if err != nil {
		return nil, err
	}
	return querySpan, nil
}
