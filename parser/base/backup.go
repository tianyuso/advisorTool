package base

import (
	storebp "advisorTool/generated-go/store"
)

type BackupStatement struct {
	Statement       string
	SourceSchema    string
	SourceTableName string
	TargetTableName string

	StartPosition *storebp.Position
	EndPosition   *storebp.Position
}
