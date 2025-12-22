package doris

import (
	"context"

	storepb "github.com/tianyuso/advisorTool/generated-go/store"
	"github.com/tianyuso/advisorTool/parser/base"
)

func init() {
	base.RegisterGetQuerySpan(storepb.Engine_DORIS, GetQuerySpan)
	base.RegisterGetQuerySpan(storepb.Engine_STARROCKS, GetQuerySpan)
}

func GetQuerySpan(
	ctx context.Context,
	gCtx base.GetQuerySpanContext,
	statement, database, _ string,
	ignoreCaseSensitive bool,
) (*base.QuerySpan, error) {
	q := newQuerySpanExtractor(database, gCtx, ignoreCaseSensitive)
	querySpan, err := q.getQuerySpan(ctx, statement)
	if err != nil {
		return nil, err
	}
	return querySpan, nil
}
