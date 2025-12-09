package snowflake

import (
	"context"

	"github.com/antlr4-go/antlr/v4"

	"advisorTool/parser/base"

	parser "github.com/bytebase/parser/snowflake"

	storepb "advisorTool/generated-go/store"
)

func init() {
	base.RegisterStatementRangesFunc(storepb.Engine_SNOWFLAKE, GetStatementRanges)
}

func GetStatementRanges(_ context.Context, _ base.StatementRangeContext, statement string) ([]base.Range, error) {
	createLexer := func(input antlr.CharStream) antlr.Lexer {
		return parser.NewSnowflakeLexer(input)
	}
	stream := base.PrepareANTLRTokenStream(statement, createLexer)
	ranges := base.GetANTLRStatementRangesUTF16Position(stream, parser.SnowflakeParserEOF, parser.SnowflakeParserSEMI)
	return ranges, nil
}
