package services

import (
	"strings"
	"testing"

	"github.com/tianyuso/advisorTool/pkg/advisor"
)

func TestSplitSQLWithMetadata_MSSQLSemicolonInString(t *testing.T) {
	statement := "INSERT INTO demo_table(id, body) VALUES (1, 'a;b;c');"

	sqlStatements := splitSQLWithMetadata(statement, advisor.EngineMSSQL)
	if len(sqlStatements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(sqlStatements))
	}

	if sqlStatements[0].Text != strings.TrimSpace(statement) {
		t.Fatalf("unexpected statement split result: %q", sqlStatements[0].Text)
	}
}

func TestSplitSQLWithMetadata_MSSQLMultiStatement(t *testing.T) {
	statement := "INSERT INTO demo_table(id, body) VALUES (1, 'a;b;c');\nUPDATE demo_table SET body = 'ok' WHERE id = 1;"

	sqlStatements := splitSQLWithMetadata(statement, advisor.EngineMSSQL)
	if len(sqlStatements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(sqlStatements))
	}
}

func TestConvertToReviewResults_MSSQLSemicolonInString(t *testing.T) {
	statement := "INSERT INTO demo_table(id, body) VALUES (1, 'a;b;c');"

	resp := &advisor.ReviewResponse{}
	results := ConvertToReviewResults(resp, statement, advisor.EngineMSSQL, nil)
	if len(results) != 1 {
		t.Fatalf("expected 1 review result, got %d", len(results))
	}

	if results[0].SQL != strings.TrimSpace(statement) {
		t.Fatalf("unexpected SQL in result: %q", results[0].SQL)
	}
}
