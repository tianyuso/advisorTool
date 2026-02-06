package extractobject

import (
	"context"
)

// TiDBExtractor TiDB表名提取器
// TiDB使用MySQL兼容语法，所以直接复用MySQLExtractor
type TiDBExtractor struct {
	MySQLExtractor
}

// Extract 从TiDB SQL语句中提取表名
func (e *TiDBExtractor) Extract(ctx context.Context, sql string) ([]TableInfo, error) {
	// TiDB使用MySQL解析器
	return e.MySQLExtractor.Extract(ctx, sql)
}


