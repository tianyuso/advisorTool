package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	extractor "github.com/tianyuso/advisorTool/extractObject"
)

var (
	dbType     = flag.String("dbtype", "mysql", "数据库类型 (mysql, postgres, oracle, sqlserver, tidb, mariadb, oceanbase, snowflake)")
	sqlFile    = flag.String("file", "", "SQL文件路径")
	sqlText    = flag.String("sql", "", "SQL语句")
	outputJSON = flag.Bool("json", false, "以JSON格式输出")
	version    = flag.Bool("version", false, "显示版本信息")
)

const VERSION = "1.0.0"

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("extractObject v%s\n", VERSION)
		return
	}

	// 获取SQL语句
	var sql string
	if *sqlFile != "" {
		content, err := ioutil.ReadFile(*sqlFile)
		if err != nil {
			log.Fatalf("读取文件失败: %v", err)
		}
		sql = string(content)
	} else if *sqlText != "" {
		sql = *sqlText
	} else {
		flag.Usage()
		fmt.Println("\n错误: 必须指定 -file 或 -sql 参数")
		os.Exit(1)
	}

	// 转换数据库类型
	dbTypeEnum, err := extractor.ParseDBType(*dbType)
	if err != nil {
		log.Fatalf("不支持的数据库类型: %v", err)
	}

	// 提取表名
	tables, err := extractor.ExtractTables(dbTypeEnum, sql)
	if err != nil {
		log.Fatalf("提取表名失败: %v", err)
	}

	// 输出结果
	if *outputJSON {
		outputAsJSON(tables)
	} else {
		outputAsText(tables)
	}
}

func outputAsJSON(tables []extractor.TableInfo) {
	data, err := json.MarshalIndent(tables, "", "  ")
	if err != nil {
		log.Fatalf("JSON序列化失败: %v", err)
	}
	fmt.Println(string(data))
}

func outputAsText(tables []extractor.TableInfo) {
	if len(tables) == 0 {
		fmt.Println("未找到任何表")
		return
	}

	fmt.Printf("找到 %d 个表:\n\n", len(tables))

	// 打印表头
	fmt.Printf("%-20s %-20s %-30s %-20s %-10s\n", "数据库名", "模式名", "表名", "别名", "类型")
	fmt.Println("--------------------------------------------------------------------------------------------")

	// 打印表信息
	for _, table := range tables {
		dbName := table.DBName
		if dbName == "" {
			dbName = "-"
		}
		schema := table.Schema
		if schema == "" {
			schema = "-"
		}
		alias := table.Alias
		if alias == "" {
			alias = "-"
		}

		tableType := "物理表"
		if table.IsCTE {
			tableType = "CTE临时表"
		}

		fmt.Printf("%-20s %-20s %-30s %-20s %-10s\n", dbName, schema, table.TBName, alias, tableType)
	}
}
