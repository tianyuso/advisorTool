// Package advisor provides SQL review rules definitions.
package advisor

import (
	"github.com/tianyuso/advisorTool/advisor"
)

// Rule types - re-exported from Bytebase advisor package.
// These are the rule type identifiers used in configuration files.
const (
	// Engine rules
	RuleMySQLEngine = string(advisor.SchemaRuleMySQLEngine)

	// Naming rules
	RuleFullyQualifiedObjectName  = string(advisor.SchemaRuleFullyQualifiedObjectName)
	RuleTableNaming               = string(advisor.SchemaRuleTableNaming)
	RuleColumnNaming              = string(advisor.SchemaRuleColumnNaming)
	RulePKNaming                  = string(advisor.SchemaRulePKNaming)
	RuleUKNaming                  = string(advisor.SchemaRuleUKNaming)
	RuleFKNaming                  = string(advisor.SchemaRuleFKNaming)
	RuleIDXNaming                 = string(advisor.SchemaRuleIDXNaming)
	RuleAutoIncrementColumnNaming = string(advisor.SchemaRuleAutoIncrementColumnNaming)
	RuleTableNameNoKeyword        = string(advisor.SchemaRuleTableNameNoKeyword)
	RuleIdentifierNoKeyword       = string(advisor.SchemaRuleIdentifierNoKeyword)
	RuleIdentifierCase            = string(advisor.SchemaRuleIdentifierCase)

	// Statement rules
	RuleStatementNoSelectAll                  = string(advisor.SchemaRuleStatementNoSelectAll)
	RuleStatementRequireWhereForSelect        = string(advisor.SchemaRuleStatementRequireWhereForSelect)
	RuleStatementRequireWhereForUpdateDelete  = string(advisor.SchemaRuleStatementRequireWhereForUpdateDelete)
	RuleStatementNoLeadingWildcardLike        = string(advisor.SchemaRuleStatementNoLeadingWildcardLike)
	RuleStatementDisallowOnDelCascade         = string(advisor.SchemaRuleStatementDisallowOnDelCascade)
	RuleStatementDisallowRemoveTblCascade     = string(advisor.SchemaRuleStatementDisallowRemoveTblCascade)
	RuleStatementDisallowCommit               = string(advisor.SchemaRuleStatementDisallowCommit)
	RuleStatementDisallowLimit                = string(advisor.SchemaRuleStatementDisallowLimit)
	RuleStatementDisallowOrderBy              = string(advisor.SchemaRuleStatementDisallowOrderBy)
	RuleStatementMergeAlterTable              = string(advisor.SchemaRuleStatementMergeAlterTable)
	RuleStatementInsertRowLimit               = string(advisor.SchemaRuleStatementInsertRowLimit)
	RuleStatementInsertMustSpecifyColumn      = string(advisor.SchemaRuleStatementInsertMustSpecifyColumn)
	RuleStatementInsertDisallowOrderByRand    = string(advisor.SchemaRuleStatementInsertDisallowOrderByRand)
	RuleStatementAffectedRowLimit             = string(advisor.SchemaRuleStatementAffectedRowLimit)
	RuleStatementDMLDryRun                    = string(advisor.SchemaRuleStatementDMLDryRun)
	RuleStatementDisallowAddColumnWithDefault = string(advisor.SchemaRuleStatementDisallowAddColumnWithDefault)
	RuleStatementAddCheckNotValid             = string(advisor.SchemaRuleStatementAddCheckNotValid)
	RuleStatementAddFKNotValid                = string(advisor.SchemaRuleStatementAddFKNotValid)
	RuleStatementDisallowAddNotNull           = string(advisor.SchemaRuleStatementDisallowAddNotNull)
	RuleStatementSelectFullTableScan          = string(advisor.SchemaRuleStatementSelectFullTableScan)
	RuleStatementCreateSpecifySchema          = string(advisor.SchemaRuleStatementCreateSpecifySchema)
	RuleStatementCheckSetRoleVariable         = string(advisor.SchemaRuleStatementCheckSetRoleVariable)
	RuleStatementDisallowUsingFilesort        = string(advisor.SchemaRuleStatementDisallowUsingFilesort)
	RuleStatementDisallowUsingTemporary       = string(advisor.SchemaRuleStatementDisallowUsingTemporary)
	RuleStatementWhereNoEqualNull             = string(advisor.SchemaRuleStatementWhereNoEqualNull)
	RuleStatementWhereDisallowFunctions       = string(advisor.SchemaRuleStatementWhereDisallowFunctionsAndCalculations)
	RuleStatementQueryMinimumPlanLevel        = string(advisor.SchemaRuleStatementQueryMinumumPlanLevel)
	RuleStatementWhereMaxLogicalOperatorCount = string(advisor.SchemaRuleStatementWhereMaximumLogicalOperatorCount)
	RuleStatementMaximumLimitValue            = string(advisor.SchemaRuleStatementMaximumLimitValue)
	RuleStatementMaximumJoinTableCount        = string(advisor.SchemaRuleStatementMaximumJoinTableCount)
	RuleStatementMaxStatementsInTransaction   = string(advisor.SchemaRuleStatementMaximumStatementsInTransaction)
	RuleStatementJoinStrictColumnAttrs        = string(advisor.SchemaRuleStatementJoinStrictColumnAttrs)
	RuleStatementPriorBackupCheck             = string(advisor.SchemaRuleStatementPriorBackupCheck)
	RuleStatementNonTransactional             = string(advisor.SchemaRuleStatementNonTransactional)
	RuleStatementAddColumnWithoutPosition     = string(advisor.SchemaRuleStatementAddColumnWithoutPosition)
	RuleStatementDisallowOfflineDDL           = string(advisor.SchemaRuleStatementDisallowOfflineDDL)
	RuleStatementDisallowCrossDBQueries       = string(advisor.SchemaRuleStatementDisallowCrossDBQueries)
	RuleStatementMaxExecutionTime             = string(advisor.SchemaRuleStatementMaxExecutionTime)
	RuleStatementRequireAlgorithmOption       = string(advisor.SchemaRuleStatementRequireAlgorithmOption)
	RuleStatementRequireLockOption            = string(advisor.SchemaRuleStatementRequireLockOption)
	RuleStatementObjectOwnerCheck             = string(advisor.SchemaRuleStatementObjectOwnerCheck)

	// Table rules
	RuleTableRequirePK             = string(advisor.SchemaRuleTableRequirePK)
	RuleTableNoFK                  = string(advisor.SchemaRuleTableNoFK)
	RuleTableDropNamingConvention  = string(advisor.SchemaRuleTableDropNamingConvention)
	RuleTableCommentConvention     = string(advisor.SchemaRuleTableCommentConvention)
	RuleTableDisallowPartition     = string(advisor.SchemaRuleTableDisallowPartition)
	RuleTableDisallowTrigger       = string(advisor.SchemaRuleTableDisallowTrigger)
	RuleTableNoDuplicateIndex      = string(advisor.SchemaRuleTableNoDuplicateIndex)
	RuleTableTextFieldsTotalLength = string(advisor.SchemaRuleTableTextFieldsTotalLength)
	RuleTableDisallowSetCharset    = string(advisor.SchemaRuleTableDisallowSetCharset)
	RuleTableDisallowDDL           = string(advisor.SchemaRuleTableDisallowDDL)
	RuleTableDisallowDML           = string(advisor.SchemaRuleTableDisallowDML)
	RuleTableLimitSize             = string(advisor.SchemaRuleTableLimitSize)
	RuleTableRequireCharset        = string(advisor.SchemaRuleTableRequireCharset)
	RuleTableRequireCollation      = string(advisor.SchemaRuleTableRequireCollation)

	// Column rules
	RuleRequiredColumn                  = string(advisor.SchemaRuleRequiredColumn)
	RuleColumnNotNull                   = string(advisor.SchemaRuleColumnNotNull)
	RuleColumnDisallowChangeType        = string(advisor.SchemaRuleColumnDisallowChangeType)
	RuleColumnSetDefaultForNotNull      = string(advisor.SchemaRuleColumnSetDefaultForNotNull)
	RuleColumnDisallowChange            = string(advisor.SchemaRuleColumnDisallowChange)
	RuleColumnDisallowChangingOrder     = string(advisor.SchemaRuleColumnDisallowChangingOrder)
	RuleColumnDisallowDrop              = string(advisor.SchemaRuleColumnDisallowDrop)
	RuleColumnDisallowDropInIndex       = string(advisor.SchemaRuleColumnDisallowDropInIndex)
	RuleColumnCommentConvention         = string(advisor.SchemaRuleColumnCommentConvention)
	RuleColumnAutoIncrementMustInteger  = string(advisor.SchemaRuleColumnAutoIncrementMustInteger)
	RuleColumnTypeDisallowList          = string(advisor.SchemaRuleColumnTypeDisallowList)
	RuleColumnDisallowSetCharset        = string(advisor.SchemaRuleColumnDisallowSetCharset)
	RuleColumnMaximumCharacterLength    = string(advisor.SchemaRuleColumnMaximumCharacterLength)
	RuleColumnMaximumVarcharLength      = string(advisor.SchemaRuleColumnMaximumVarcharLength)
	RuleColumnAutoIncrementInitialValue = string(advisor.SchemaRuleColumnAutoIncrementInitialValue)
	RuleColumnAutoIncrementMustUnsigned = string(advisor.SchemaRuleColumnAutoIncrementMustUnsigned)
	RuleCurrentTimeColumnCountLimit     = string(advisor.SchemaRuleCurrentTimeColumnCountLimit)
	RuleColumnRequireDefault            = string(advisor.SchemaRuleColumnRequireDefault)
	RuleColumnDefaultDisallowVolatile   = string(advisor.SchemaRuleColumnDefaultDisallowVolatile)
	RuleAddNotNullColumnRequireDefault  = string(advisor.SchemaRuleAddNotNullColumnRequireDefault)
	RuleColumnRequireCharset            = string(advisor.SchemaRuleColumnRequireCharset)
	RuleColumnRequireCollation          = string(advisor.SchemaRuleColumnRequireCollation)

	// Schema rules
	RuleSchemaBackwardCompatibility = string(advisor.SchemaRuleSchemaBackwardCompatibility)

	// Database rules
	RuleDropEmptyDatabase = string(advisor.SchemaRuleDropEmptyDatabase)

	// Index rules
	RuleIndexNoDuplicateColumn       = string(advisor.SchemaRuleIndexNoDuplicateColumn)
	RuleIndexKeyNumberLimit          = string(advisor.SchemaRuleIndexKeyNumberLimit)
	RuleIndexPKTypeLimit             = string(advisor.SchemaRuleIndexPKTypeLimit)
	RuleIndexTypeNoBlob              = string(advisor.SchemaRuleIndexTypeNoBlob)
	RuleIndexTotalNumberLimit        = string(advisor.SchemaRuleIndexTotalNumberLimit)
	RuleIndexPrimaryKeyTypeAllowlist = string(advisor.SchemaRuleIndexPrimaryKeyTypeAllowlist)
	RuleCreateIndexConcurrently      = string(advisor.SchemaRuleCreateIndexConcurrently)
	RuleIndexTypeAllowList           = string(advisor.SchemaRuleIndexTypeAllowList)
	RuleIndexNotRedundant            = string(advisor.SchemaRuleIndexNotRedundant)

	// System rules
	RuleCharsetAllowlist        = string(advisor.SchemaRuleCharsetAllowlist)
	RuleCollationAllowlist      = string(advisor.SchemaRuleCollationAllowlist)
	RuleCommentLength           = string(advisor.SchemaRuleCommentLength)
	RuleProcedureDisallowCreate = string(advisor.SchemaRuleProcedureDisallowCreate)
	RuleEventDisallowCreate     = string(advisor.SchemaRuleEventDisallowCreate)
	RuleViewDisallowCreate      = string(advisor.SchemaRuleViewDisallowCreate)
	RuleFunctionDisallowCreate  = string(advisor.SchemaRuleFunctionDisallowCreate)
	RuleFunctionDisallowList    = string(advisor.SchemaRuleFunctionDisallowList)

	// Advice rules
	RuleOnlineMigration = string(advisor.SchemaRuleOnlineMigration)

	// Builtin rules
	BuiltinRulePriorBackupCheck = string(advisor.BuiltinRulePriorBackupCheck)
)

// AllRules returns all available rule types.
func AllRules() []string {
	return []string{
		RuleMySQLEngine,
		RuleFullyQualifiedObjectName,
		RuleTableNaming,
		RuleColumnNaming,
		RulePKNaming,
		RuleUKNaming,
		RuleFKNaming,
		RuleIDXNaming,
		RuleAutoIncrementColumnNaming,
		RuleTableNameNoKeyword,
		RuleIdentifierNoKeyword,
		RuleIdentifierCase,
		RuleStatementNoSelectAll,
		RuleStatementRequireWhereForSelect,
		RuleStatementRequireWhereForUpdateDelete,
		RuleStatementNoLeadingWildcardLike,
		RuleStatementDisallowOnDelCascade,
		RuleStatementDisallowRemoveTblCascade,
		RuleStatementDisallowCommit,
		RuleStatementDisallowLimit,
		RuleStatementDisallowOrderBy,
		RuleStatementMergeAlterTable,
		RuleStatementInsertRowLimit,
		RuleStatementInsertMustSpecifyColumn,
		RuleStatementInsertDisallowOrderByRand,
		RuleStatementAffectedRowLimit,
		RuleStatementDMLDryRun,
		RuleStatementDisallowAddColumnWithDefault,
		RuleStatementAddCheckNotValid,
		RuleStatementAddFKNotValid,
		RuleStatementDisallowAddNotNull,
		RuleStatementSelectFullTableScan,
		RuleStatementCreateSpecifySchema,
		RuleStatementCheckSetRoleVariable,
		RuleStatementDisallowUsingFilesort,
		RuleStatementDisallowUsingTemporary,
		RuleStatementWhereNoEqualNull,
		RuleStatementWhereDisallowFunctions,
		RuleStatementQueryMinimumPlanLevel,
		RuleStatementWhereMaxLogicalOperatorCount,
		RuleStatementMaximumLimitValue,
		RuleStatementMaximumJoinTableCount,
		RuleStatementMaxStatementsInTransaction,
		RuleStatementJoinStrictColumnAttrs,
		RuleStatementPriorBackupCheck,
		RuleStatementNonTransactional,
		RuleStatementAddColumnWithoutPosition,
		RuleStatementDisallowOfflineDDL,
		RuleStatementDisallowCrossDBQueries,
		RuleStatementMaxExecutionTime,
		RuleStatementRequireAlgorithmOption,
		RuleStatementRequireLockOption,
		RuleStatementObjectOwnerCheck,
		RuleTableRequirePK,
		RuleTableNoFK,
		RuleTableDropNamingConvention,
		RuleTableCommentConvention,
		RuleTableDisallowPartition,
		RuleTableDisallowTrigger,
		RuleTableNoDuplicateIndex,
		RuleTableTextFieldsTotalLength,
		RuleTableDisallowSetCharset,
		RuleTableDisallowDDL,
		RuleTableDisallowDML,
		RuleTableLimitSize,
		RuleTableRequireCharset,
		RuleTableRequireCollation,
		RuleRequiredColumn,
		RuleColumnNotNull,
		RuleColumnDisallowChangeType,
		RuleColumnSetDefaultForNotNull,
		RuleColumnDisallowChange,
		RuleColumnDisallowChangingOrder,
		RuleColumnDisallowDrop,
		RuleColumnDisallowDropInIndex,
		RuleColumnCommentConvention,
		RuleColumnAutoIncrementMustInteger,
		RuleColumnTypeDisallowList,
		RuleColumnDisallowSetCharset,
		RuleColumnMaximumCharacterLength,
		RuleColumnMaximumVarcharLength,
		RuleColumnAutoIncrementInitialValue,
		RuleColumnAutoIncrementMustUnsigned,
		RuleCurrentTimeColumnCountLimit,
		RuleColumnRequireDefault,
		RuleColumnDefaultDisallowVolatile,
		RuleAddNotNullColumnRequireDefault,
		RuleColumnRequireCharset,
		RuleColumnRequireCollation,
		RuleSchemaBackwardCompatibility,
		RuleDropEmptyDatabase,
		RuleIndexNoDuplicateColumn,
		RuleIndexKeyNumberLimit,
		RuleIndexPKTypeLimit,
		RuleIndexTypeNoBlob,
		RuleIndexTotalNumberLimit,
		RuleIndexPrimaryKeyTypeAllowlist,
		RuleCreateIndexConcurrently,
		RuleIndexTypeAllowList,
		RuleIndexNotRedundant,
		RuleCharsetAllowlist,
		RuleCollationAllowlist,
		RuleCommentLength,
		RuleProcedureDisallowCreate,
		RuleEventDisallowCreate,
		RuleViewDisallowCreate,
		RuleFunctionDisallowCreate,
		RuleFunctionDisallowList,
		RuleOnlineMigration,
		BuiltinRulePriorBackupCheck,
	}
}

// GetRuleDescription returns a description for the given rule type.
func GetRuleDescription(ruleType string) string {
	descriptions := map[string]string{
		RuleMySQLEngine:                           "Require InnoDB storage engine",
		RuleFullyQualifiedObjectName:              "Require fully qualified object names",
		RuleTableNaming:                           "Table naming convention",
		RuleColumnNaming:                          "Column naming convention",
		RulePKNaming:                              "Primary key naming convention",
		RuleUKNaming:                              "Unique key naming convention",
		RuleFKNaming:                              "Foreign key naming convention",
		RuleIDXNaming:                             "Index naming convention",
		RuleAutoIncrementColumnNaming:             "Auto-increment column naming convention",
		RuleTableNameNoKeyword:                    "Disallow keywords as table names",
		RuleIdentifierNoKeyword:                   "Disallow keywords as identifiers",
		RuleIdentifierCase:                        "Identifier case convention",
		RuleStatementNoSelectAll:                  "Disallow SELECT *",
		RuleStatementRequireWhereForSelect:        "Require WHERE clause for SELECT",
		RuleStatementRequireWhereForUpdateDelete:  "Require WHERE clause for UPDATE/DELETE",
		RuleStatementNoLeadingWildcardLike:        "Disallow leading % in LIKE",
		RuleStatementDisallowOnDelCascade:         "Disallow ON DELETE CASCADE",
		RuleStatementDisallowRemoveTblCascade:     "Disallow CASCADE when removing table",
		RuleStatementDisallowCommit:               "Disallow COMMIT statement",
		RuleStatementDisallowLimit:                "Disallow LIMIT clause",
		RuleStatementDisallowOrderBy:              "Disallow ORDER BY clause",
		RuleStatementMergeAlterTable:              "Merge ALTER TABLE statements",
		RuleStatementInsertRowLimit:               "Limit inserted rows",
		RuleStatementInsertMustSpecifyColumn:      "INSERT must specify columns",
		RuleStatementInsertDisallowOrderByRand:    "Disallow ORDER BY RAND in INSERT",
		RuleStatementAffectedRowLimit:             "Limit affected rows",
		RuleStatementDMLDryRun:                    "Dry run DML statements",
		RuleStatementDisallowAddColumnWithDefault: "Disallow ADD COLUMN with DEFAULT",
		RuleStatementAddCheckNotValid:             "Add CHECK with NOT VALID",
		RuleStatementAddFKNotValid:                "Add FK with NOT VALID",
		RuleStatementDisallowAddNotNull:           "Disallow ADD NOT NULL",
		RuleStatementSelectFullTableScan:          "Disallow full table scan",
		RuleStatementCreateSpecifySchema:          "Require schema name in CREATE",
		RuleStatementCheckSetRoleVariable:         "Check SET ROLE variable",
		RuleStatementDisallowUsingFilesort:        "Disallow using filesort",
		RuleStatementDisallowUsingTemporary:       "Disallow using temporary",
		RuleStatementWhereNoEqualNull:             "Disallow WHERE = NULL",
		RuleStatementWhereDisallowFunctions:       "Disallow functions in WHERE",
		RuleStatementQueryMinimumPlanLevel:        "Minimum query plan level",
		RuleStatementWhereMaxLogicalOperatorCount: "Maximum logical operators in WHERE",
		RuleStatementMaximumLimitValue:            "Maximum LIMIT value",
		RuleStatementMaximumJoinTableCount:        "Maximum JOIN table count",
		RuleStatementMaxStatementsInTransaction:   "Maximum statements in transaction",
		RuleStatementJoinStrictColumnAttrs:        "Strict JOIN column attributes",
		RuleStatementPriorBackupCheck:             "Prior backup check",
		RuleStatementNonTransactional:             "Non-transactional statement check",
		RuleStatementAddColumnWithoutPosition:     "ADD COLUMN without position",
		RuleStatementDisallowOfflineDDL:           "Disallow offline DDL",
		RuleStatementDisallowCrossDBQueries:       "Disallow cross-database queries",
		RuleStatementMaxExecutionTime:             "Maximum execution time",
		RuleStatementRequireAlgorithmOption:       "Require ALGORITHM option",
		RuleStatementRequireLockOption:            "Require LOCK option",
		RuleStatementObjectOwnerCheck:             "Object owner check",
		RuleTableRequirePK:                        "Require primary key",
		RuleTableNoFK:                             "Disallow foreign key",
		RuleTableDropNamingConvention:             "Drop naming convention",
		RuleTableCommentConvention:                "Table comment convention",
		RuleTableDisallowPartition:                "Disallow partition table",
		RuleTableDisallowTrigger:                  "Disallow trigger",
		RuleTableNoDuplicateIndex:                 "Disallow duplicate index",
		RuleTableTextFieldsTotalLength:            "Total text fields length limit",
		RuleTableDisallowSetCharset:               "Disallow set table charset",
		RuleTableDisallowDDL:                      "Disallow DDL on specific tables",
		RuleTableDisallowDML:                      "Disallow DML on specific tables",
		RuleTableLimitSize:                        "Limit table size for DDL",
		RuleTableRequireCharset:                   "Require table charset",
		RuleTableRequireCollation:                 "Require table collation",
		RuleRequiredColumn:                        "Required columns",
		RuleColumnNotNull:                         "Columns no NULL value",
		RuleColumnDisallowChangeType:              "Disallow changing column type",
		RuleColumnSetDefaultForNotNull:            "Set DEFAULT for NOT NULL columns",
		RuleColumnDisallowChange:                  "Disallow CHANGE COLUMN",
		RuleColumnDisallowChangingOrder:           "Disallow changing column order",
		RuleColumnDisallowDrop:                    "Disallow DROP COLUMN",
		RuleColumnDisallowDropInIndex:             "Disallow dropping columns in indexes",
		RuleColumnCommentConvention:               "Column comment convention",
		RuleColumnAutoIncrementMustInteger:        "Auto-increment must be integer",
		RuleColumnTypeDisallowList:                "Column type disallow list",
		RuleColumnDisallowSetCharset:              "Disallow set column charset",
		RuleColumnMaximumCharacterLength:          "Maximum CHAR length",
		RuleColumnMaximumVarcharLength:            "Maximum VARCHAR length",
		RuleColumnAutoIncrementInitialValue:       "Auto-increment initial value",
		RuleColumnAutoIncrementMustUnsigned:       "Auto-increment must be unsigned",
		RuleCurrentTimeColumnCountLimit:           "Current time column count limit",
		RuleColumnRequireDefault:                  "Require column default value",
		RuleColumnDefaultDisallowVolatile:         "Disallow volatile default value",
		RuleAddNotNullColumnRequireDefault:        "NOT NULL column requires default",
		RuleColumnRequireCharset:                  "Require column charset",
		RuleColumnRequireCollation:                "Require column collation",
		RuleSchemaBackwardCompatibility:           "Backward compatible schema change",
		RuleDropEmptyDatabase:                     "Drop database restriction",
		RuleIndexNoDuplicateColumn:                "No duplicate column in index",
		RuleIndexKeyNumberLimit:                   "Index key number limit",
		RuleIndexPKTypeLimit:                      "Primary key type limit",
		RuleIndexTypeNoBlob:                       "Disallow BLOB/TEXT in index",
		RuleIndexTotalNumberLimit:                 "Index total number limit",
		RuleIndexPrimaryKeyTypeAllowlist:          "Primary key type allowlist",
		RuleCreateIndexConcurrently:               "Create index concurrently",
		RuleIndexTypeAllowList:                    "Index type allowlist",
		RuleIndexNotRedundant:                     "No redundant index",
		RuleCharsetAllowlist:                      "Charset allowlist",
		RuleCollationAllowlist:                    "Collation allowlist",
		RuleCommentLength:                         "Comment length limit",
		RuleProcedureDisallowCreate:               "Disallow create procedure",
		RuleEventDisallowCreate:                   "Disallow create event",
		RuleViewDisallowCreate:                    "Disallow create view",
		RuleFunctionDisallowCreate:                "Disallow create function",
		RuleFunctionDisallowList:                  "Function disallow list",
		RuleOnlineMigration:                       "Online migration advice",
		BuiltinRulePriorBackupCheck:               "Prior backup check",
	}

	if desc, ok := descriptions[ruleType]; ok {
		return desc
	}
	return ruleType
}
