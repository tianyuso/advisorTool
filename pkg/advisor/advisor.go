// Package advisor provides a simplified wrapper around Bytebase's SQL advisor functionality.
package advisor

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"advisorTool/advisor"
	"advisorTool/component/sheet"
	storepb "advisorTool/generated-go/store"
	"advisorTool/store/model"

	// Import all advisors to register them
	_ "advisorTool/advisor/mssql"
	_ "advisorTool/advisor/mysql"
	_ "advisorTool/advisor/oceanbase"
	_ "advisorTool/advisor/oracle"
	_ "advisorTool/advisor/pg"
	_ "advisorTool/advisor/snowflake"
	_ "advisorTool/advisor/tidb"
)

// Engine represents the database engine type.
type Engine = storepb.Engine

// Engine constants for convenience.
const (
	EngineMySQL     = storepb.Engine_MYSQL
	EnginePostgres  = storepb.Engine_POSTGRES
	EngineTiDB      = storepb.Engine_TIDB
	EngineOracle    = storepb.Engine_ORACLE
	EngineMSSQL     = storepb.Engine_MSSQL
	EngineMariaDB   = storepb.Engine_MARIADB
	EngineSnowflake = storepb.Engine_SNOWFLAKE
	EngineOceanBase = storepb.Engine_OCEANBASE
)

// EngineFromString converts a string to Engine type.
func EngineFromString(s string) Engine {
	switch s {
	case "mysql", "MYSQL":
		return EngineMySQL
	case "postgres", "postgresql", "POSTGRES", "POSTGRESQL":
		return EnginePostgres
	case "tidb", "TIDB":
		return EngineTiDB
	case "oracle", "ORACLE":
		return EngineOracle
	case "mssql", "sqlserver", "MSSQL", "SQLSERVER":
		return EngineMSSQL
	case "mariadb", "MARIADB":
		return EngineMariaDB
	case "snowflake", "SNOWFLAKE":
		return EngineSnowflake
	case "oceanbase", "OCEANBASE":
		return EngineOceanBase
	default:
		return storepb.Engine_ENGINE_UNSPECIFIED
	}
}

// RuleLevel is the alias for SQLReviewRuleLevel.
type RuleLevel = storepb.SQLReviewRuleLevel

// Rule level constants.
const (
	RuleLevelError   = storepb.SQLReviewRuleLevel_ERROR
	RuleLevelWarning = storepb.SQLReviewRuleLevel_WARNING
	// RuleLevelDisabled uses LEVEL_UNSPECIFIED to indicate a disabled rule
	RuleLevelDisabled = storepb.SQLReviewRuleLevel_LEVEL_UNSPECIFIED
)

// RuleLevelFromString converts a string to RuleLevel.
func RuleLevelFromString(s string) RuleLevel {
	switch s {
	case "error", "ERROR":
		return RuleLevelError
	case "warning", "WARNING":
		return RuleLevelWarning
	case "disabled", "DISABLED":
		return RuleLevelDisabled
	default:
		return storepb.SQLReviewRuleLevel_LEVEL_UNSPECIFIED
	}
}

// AdviceStatus is the alias for Advice_Status.
type AdviceStatus = storepb.Advice_Status

// Advice status constants.
const (
	AdviceStatusSuccess = storepb.Advice_SUCCESS
	AdviceStatusWarning = storepb.Advice_WARNING
	AdviceStatusError   = storepb.Advice_ERROR
)

// Advice is the alias for storepb.Advice.
type Advice = storepb.Advice

// SQLReviewRule is the alias for storepb.SQLReviewRule.
type SQLReviewRule = storepb.SQLReviewRule

// Position is the alias for storepb.Position.
type Position = storepb.Position

// DatabaseSchemaMetadata is the alias for storepb.DatabaseSchemaMetadata.
type DatabaseSchemaMetadata = storepb.DatabaseSchemaMetadata

// ReviewRequest represents a request to review SQL statements.
type ReviewRequest struct {
	// Engine is the database engine type.
	Engine Engine
	// Statement is the SQL statement to review.
	Statement string
	// Rules is the list of review rules to apply.
	Rules []*SQLReviewRule
	// CurrentDatabase is the current database context (optional).
	CurrentDatabase string
	// DBSchema is the database schema metadata (optional, needed for some rules).
	// If not provided, rules that require metadata will be skipped or may fail.
	DBSchema *DatabaseSchemaMetadata
}

// ReviewResponse represents the response from SQL review.
type ReviewResponse struct {
	// Advices is the list of advice generated from the review.
	Advices []*Advice
	// HasError indicates if there are any error-level advices.
	HasError bool
	// HasWarning indicates if there are any warning-level advices.
	HasWarning bool
}

// SQLReviewCheck performs SQL review on the given statement with the specified rules.
// This is the main entry point for SQL review.
func SQLReviewCheck(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error) {
	if req == nil {
		return nil, errors.New("review request is nil")
	}

	if req.Engine == storepb.Engine_ENGINE_UNSPECIFIED {
		return nil, errors.New("engine is not specified")
	}

	if req.Statement == "" {
		return &ReviewResponse{}, nil
	}

	// Create a sheet manager for parsing (standalone mode, no database store needed)
	sheetManager := sheet.NewManager()

	// Determine case sensitivity based on engine
	isCaseSensitive := false
	if req.Engine == storepb.Engine_POSTGRES {
		isCaseSensitive = true
	}

	// Build the check context
	checkContext := advisor.Context{
		DBType:          req.Engine,
		CurrentDatabase: req.CurrentDatabase,
		DBSchema:        req.DBSchema,
		NoAppendBuiltin: true, // Don't append builtin rules
	}

	// If DBSchema is provided, create metadata objects for rules that need them
	if req.DBSchema != nil {
		originalMetadata := model.NewDatabaseMetadata(req.DBSchema, nil, nil, req.Engine, isCaseSensitive)
		finalMetadata := model.NewDatabaseMetadata(req.DBSchema, nil, nil, req.Engine, isCaseSensitive)
		checkContext.OriginalMetadata = originalMetadata
		checkContext.FinalMetadata = finalMetadata
	}

	// Perform the SQL review
	advices, err := advisor.SQLReviewCheck(
		ctx,
		sheetManager,
		req.Statement,
		req.Rules,
		checkContext,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to perform SQL review")
	}

	// Build response
	resp := &ReviewResponse{
		Advices: advices,
	}

	for _, advice := range advices {
		switch advice.Status {
		case storepb.Advice_ERROR:
			resp.HasError = true
		case storepb.Advice_WARNING:
			resp.HasWarning = true
		}
	}

	return resp, nil
}

// NewRule creates a new SQL review rule.
func NewRule(ruleType string, level RuleLevel) *SQLReviewRule {
	return &SQLReviewRule{
		Type:  ruleType,
		Level: level,
	}
}

// NewRuleWithPayload creates a new SQL review rule with payload.
func NewRuleWithPayload(ruleType string, level RuleLevel, payload interface{}) (*SQLReviewRule, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal payload")
	}

	return &SQLReviewRule{
		Type:    ruleType,
		Level:   level,
		Payload: string(payloadBytes),
	}, nil
}

// NewRuleForEngine creates a new SQL review rule for a specific engine.
func NewRuleForEngine(ruleType string, level RuleLevel, engine Engine) *SQLReviewRule {
	return &SQLReviewRule{
		Type:   ruleType,
		Level:  level,
		Engine: engine,
	}
}

// NamingRulePayload is the payload for naming rules.
type NamingRulePayload = advisor.NamingRulePayload

// NumberTypeRulePayload is the payload for number type rules.
type NumberTypeRulePayload = advisor.NumberTypeRulePayload

// StringArrayTypeRulePayload is the payload for string array rules.
type StringArrayTypeRulePayload = advisor.StringArrayTypeRulePayload

// CommentConventionRulePayload is the payload for comment convention rules.
type CommentConventionRulePayload = advisor.CommentConventionRulePayload
