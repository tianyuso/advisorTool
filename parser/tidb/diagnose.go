package tidb

import (
	"context"
	"errors"

	"github.com/tianyuso/advisorTool/generated-go/store"
	"github.com/tianyuso/advisorTool/parser/base"
)

func init() {
	base.RegisterDiagnoseFunc(store.Engine_TIDB, Diagnose)
}

func Diagnose(_ context.Context, _ base.DiagnoseContext, statement string) ([]base.Diagnostic, error) {
	diagnostics := make([]base.Diagnostic, 0)
	_, err := ParseTiDB(statement, "", "")
	var syntaxError *base.SyntaxError
	if err != nil {
		if errors.As(err, &syntaxError) {
			diagnostics = append(diagnostics, base.ConvertSyntaxErrorToDiagnostic(syntaxError, statement))
		} else {
			return nil, err
		}
	}

	return diagnostics, nil
}
