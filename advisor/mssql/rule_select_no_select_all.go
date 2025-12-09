package mssql

import (
	"context"

	"github.com/antlr4-go/antlr/v4"
	parser "github.com/bytebase/parser/tsql"

	"advisorTool/advisor"
	"advisorTool/advisor/code"
	"advisorTool/common"
	storepb "advisorTool/generated-go/store"
)

var (
	_ advisor.Advisor = (*SelectNoSelectAllAdvisor)(nil)
)

func init() {
	advisor.Register(storepb.Engine_MSSQL, advisor.SchemaRuleStatementNoSelectAll, &SelectNoSelectAllAdvisor{})
}

// SelectNoSelectAllAdvisor is the advisor checking for no select all.
type SelectNoSelectAllAdvisor struct {
}

// Check checks for no select all.
func (*SelectNoSelectAllAdvisor) Check(_ context.Context, checkCtx advisor.Context) ([]*storepb.Advice, error) {
	parseResults, err := getANTLRTree(checkCtx)

	if err != nil {
		return nil, err
	}

	level, err := advisor.NewStatusBySQLReviewRuleLevel(checkCtx.Rule.Level)
	if err != nil {
		return nil, err
	}

	// Create the rule
	rule := NewSelectNoSelectAllRule(level, string(checkCtx.Rule.Type))

	// Create the generic checker with the rule
	checker := NewGenericChecker([]Rule{rule})

	for _, parseResult := range parseResults {
		rule.SetBaseLine(parseResult.BaseLine)
		antlr.ParseTreeWalkerDefault.Walk(checker, parseResult.Tree)
	}

	return checker.GetAdviceList(), nil
}

// SelectNoSelectAllRule checks for no select all.
type SelectNoSelectAllRule struct {
	BaseRule
}

// NewSelectNoSelectAllRule creates a new SelectNoSelectAllRule.
func NewSelectNoSelectAllRule(level storepb.Advice_Status, title string) *SelectNoSelectAllRule {
	return &SelectNoSelectAllRule{
		BaseRule: BaseRule{
			level: level,
			title: title,
		},
	}
}

// Name returns the rule name.
func (*SelectNoSelectAllRule) Name() string {
	return "SelectNoSelectAllRule"
}

// OnEnter is called when entering a parse tree node.
func (r *SelectNoSelectAllRule) OnEnter(ctx antlr.ParserRuleContext, nodeType string) error {
	if nodeType == NodeTypeSelectListElem {
		r.enterSelectListElem(ctx.(*parser.Select_list_elemContext))
	}
	return nil
}

// OnExit is called when exiting a parse tree node.
func (*SelectNoSelectAllRule) OnExit(_ antlr.ParserRuleContext, _ string) error {
	// This rule doesn't need exit processing
	return nil
}

func (r *SelectNoSelectAllRule) enterSelectListElem(ctx *parser.Select_list_elemContext) {
	if v := ctx.Asterisk(); v != nil {
		if v.STAR() != nil {
			r.AddAdvice(&storepb.Advice{
				Status:        r.level,
				Code:          code.StatementSelectAll.Int32(),
				Title:         r.title,
				Content:       "Avoid using SELECT *.",
				StartPosition: common.ConvertANTLRLineToPosition(v.STAR().GetSymbol().GetLine()),
			})
		}
	}
}
