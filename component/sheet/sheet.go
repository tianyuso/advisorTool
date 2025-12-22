package sheet

import (
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/zeebo/xxh3"

	"github.com/tianyuso/advisorTool/common"
	storepb "github.com/tianyuso/advisorTool/generated-go/store"
	"github.com/tianyuso/advisorTool/parser/base"
	tsqlbatch "github.com/tianyuso/advisorTool/parser/tsql/batch"

	// Import parsers to register their parse functions.
	_ "github.com/tianyuso/advisorTool/parser/cockroachdb"
	_ "github.com/tianyuso/advisorTool/parser/doris"
	_ "github.com/tianyuso/advisorTool/parser/mysql"
	_ "github.com/tianyuso/advisorTool/parser/partiql"
	_ "github.com/tianyuso/advisorTool/parser/plsql"
	_ "github.com/tianyuso/advisorTool/parser/redshift"
	_ "github.com/tianyuso/advisorTool/parser/snowflake"
	_ "github.com/tianyuso/advisorTool/parser/tidb"
	_ "github.com/tianyuso/advisorTool/parser/tsql"
)

const (
	// SyntaxErrorTitle is the error title for syntax error.
	SyntaxErrorTitle         string = "Syntax error"
	StatementSyntaxErrorCode int32  = 201
	InternalErrorCode        int32  = 1
)

// Manager is the coordinator for all sheets and SQL statements.
// For advisorTool, we only need AST caching functionality.
type Manager struct {
	sync.Mutex
	astCache *lru.LRU[astHashKey, *Result]
}

type astHashKey struct {
	hash   uint64
	engine storepb.Engine
}

// NewManager creates a new sheet manager.
func NewManager() *Manager {
	return &Manager{
		astCache: lru.NewLRU[astHashKey, *Result](8, nil, 3*time.Minute),
	}
}

func getSheetCommands(engine storepb.Engine, statement string) []*storepb.Range {
	// Burnout for large SQL.
	if len(statement) > common.MaxSheetCheckSize {
		return nil
	}

	switch engine {
	case
		storepb.Engine_TIDB,
		storepb.Engine_ORACLE:
		return getSheetCommandsFromByteOffset(engine, statement)
	case storepb.Engine_MSSQL:
		return getSheetCommandsForMSSQL(statement)
	default:
		return getSheetCommandsGeneral(engine, statement)
	}
}

func getSheetCommandsGeneral(engine storepb.Engine, statement string) []*storepb.Range {
	singleSQLs, err := base.SplitMultiSQL(engine, statement)
	if err != nil {
		if !strings.Contains(err.Error(), "not supported") {
			slog.Warn("failed to split multi sql", "engine", engine.String(), "statement", statement)
		}
		return nil
	}
	// HACK(p0ny): always split for pg
	if len(singleSQLs) > common.MaximumCommands && engine != storepb.Engine_POSTGRES {
		return nil
	}

	var sheetCommands []*storepb.Range
	p := 0
	for _, s := range singleSQLs {
		np := p + len(s.Text)
		sheetCommands = append(sheetCommands, &storepb.Range{
			Start: int32(p),
			End:   int32(np),
		})
		p = np
	}
	return sheetCommands
}

func getSheetCommandsFromByteOffset(engine storepb.Engine, statement string) []*storepb.Range {
	singleSQLs, err := base.SplitMultiSQL(engine, statement)
	if err != nil {
		if !strings.Contains(err.Error(), "not supported") {
			slog.Warn("failed to get sheet command from byte offset", "engine", engine.String(), "statement", statement)
		}
		return nil
	}
	if len(singleSQLs) > common.MaximumCommands {
		return nil
	}

	var sheetCommands []*storepb.Range
	for _, s := range singleSQLs {
		sheetCommands = append(sheetCommands, &storepb.Range{
			Start: int32(s.ByteOffsetStart),
			End:   int32(s.ByteOffsetEnd),
		})
	}
	return sheetCommands
}

func getSheetCommandsForMSSQL(statement string) []*storepb.Range {
	var sheetCommands []*storepb.Range

	batch := tsqlbatch.NewBatcher(statement)
	for {
		command, err := batch.Next()
		if err == io.EOF {
			b := batch.Batch()
			sheetCommands = append(sheetCommands, &storepb.Range{
				Start: int32(b.Start),
				End:   int32(b.End),
			})
			batch.Reset(nil)
			break
		}
		if err != nil {
			slog.Warn("failed to get sheet commands for mssql", "statement", statement)
			return nil
		}
		if command == nil {
			continue
		}
		switch command.(type) {
		case *tsqlbatch.GoCommand:
			b := batch.Batch()
			sheetCommands = append(sheetCommands, &storepb.Range{
				Start: int32(b.Start),
				End:   int32(b.End),
			})
			batch.Reset(nil)
		default:
		}
		// No command count limit for MSSQL to ensure consistency between sheet payload
		// and actual execution in mssql.go which splits and executes all batches
	}
	return sheetCommands
}

type Result struct {
	sync.Mutex
	ast     []base.AST
	advices []*storepb.Advice
}

// GetASTsForChecks gets the ASTs of statement with caching, and it should only be used
// for plan checks because it involves some truncating.
func (sm *Manager) GetASTsForChecks(dbType storepb.Engine, statement string) ([]base.AST, []*storepb.Advice) {
	var result *Result
	h := xxh3.HashString(statement)
	key := astHashKey{hash: h, engine: dbType}
	sm.Lock()
	if v, ok := sm.astCache.Get(key); ok {
		result = v
	} else {
		result = &Result{}
		sm.astCache.Add(key, result)
	}
	sm.Unlock()

	result.Lock()
	defer result.Unlock()
	if result.ast != nil || result.advices != nil {
		return result.ast, result.advices
	}
	ast, err := base.Parse(dbType, statement)
	if err != nil {
		result.advices = convertErrorToAdvice(err)
	} else {
		result.ast = ast
	}
	return result.ast, result.advices
}

func convertErrorToAdvice(err error) []*storepb.Advice {
	if syntaxErr, ok := err.(*base.SyntaxError); ok {
		return []*storepb.Advice{
			{
				Status:        storepb.Advice_ERROR,
				Code:          StatementSyntaxErrorCode,
				Title:         SyntaxErrorTitle,
				Content:       syntaxErr.Message,
				StartPosition: syntaxErr.Position,
			},
		}
	}
	return []*storepb.Advice{
		{
			Status:        storepb.Advice_ERROR,
			Code:          InternalErrorCode,
			Title:         SyntaxErrorTitle,
			Content:       err.Error(),
			StartPosition: nil,
		},
	}
}
