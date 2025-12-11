// Package main provides the CLI entry point for the SQL advisor tool.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"advisorTool/cmd/advisor/internal"
	"advisorTool/pkg/advisor"
)

var (
	configFile     = flag.String("config", "", "Path to the review config file (YAML or JSON)")
	engine         = flag.String("engine", "", "Database engine: mysql, postgres, tidb, oracle, mssql, snowflake, mariadb, oceanbase")
	sqlFile        = flag.String("file", "", "Path to the SQL file to review")
	sqlStatement   = flag.String("sql", "", "SQL statement to review (use - to read from stdin)")
	outputFormat   = flag.String("format", "text", "Output format: text, json, yaml")
	listRules      = flag.Bool("list-rules", false, "List all available rules")
	generateConfig = flag.Bool("generate-config", false, "Generate a sample config file for the specified engine")
	version        = flag.Bool("version", false, "Print version information")

	// Database connection parameters
	dbHost        = flag.String("host", "", "Database host address")
	dbPort        = flag.Int("port", 0, "Database port")
	dbUser        = flag.String("user", "", "Database username")
	dbPassword    = flag.String("password", "", "Database password")
	dbName        = flag.String("dbname", "", "Database name")
	dbCharset     = flag.String("charset", "", "Database charset (default: utf8mb4 for MySQL)")
	dbServiceName = flag.String("service-name", "", "Oracle service name")
	dbSid         = flag.String("sid", "", "Oracle SID")
	dbSSLMode     = flag.String("sslmode", "disable", "PostgreSQL SSL mode")
	dbTimeout     = flag.Int("timeout", 5, "Database connection timeout in seconds")
)

const toolVersion = "1.0.0"

func main() {
	flag.Parse()

	// Handle version flag
	if *version {
		printVersion()
		os.Exit(0)
	}

	// Handle list-rules flag
	if *listRules {
		internal.ListAvailableRules()
		os.Exit(0)
	}

	// Validate engine flag
	if *engine == "" {
		fmt.Fprintln(os.Stderr, "Error: -engine flag is required")
		flag.Usage()
		os.Exit(1)
	}

	engineType := advisor.EngineFromString(*engine)
	if engineType == 0 {
		fmt.Fprintf(os.Stderr, "Error: unsupported engine: %s\n", *engine)
		fmt.Fprintln(os.Stderr, "Supported engines: mysql, postgres, tidb, oracle, mssql, snowflake, mariadb, oceanbase")
		os.Exit(1)
	}

	// Handle generate-config flag
	if *generateConfig {
		config := internal.GenerateSampleConfig(engineType)
		fmt.Println(config)
		os.Exit(0)
	}

	// Get SQL statement
	statement, err := getStatement()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading SQL: %v\n", err)
		os.Exit(1)
	}

	if statement == "" {
		fmt.Fprintln(os.Stderr, "Error: no SQL statement provided. Use -sql or -file flag")
		flag.Usage()
		os.Exit(1)
	}

	// Prepare database connection parameters
	dbParams := buildDBParams()

	// Prepare review request
	req := &advisor.ReviewRequest{
		Engine:          engineType,
		Statement:       statement,
		CurrentDatabase: *dbName,
	}

	// Check if database connection parameters are provided
	hasMetadata := false
	if dbParams.Host != "" && dbParams.Port > 0 {
		metadata, err := internal.FetchDatabaseMetadata(engineType, dbParams)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to fetch database metadata: %v\n", err)
			fmt.Fprintf(os.Stderr, "Some rules that require metadata will be skipped.\n")
		} else {
			req.DBSchema = metadata
			hasMetadata = true
		}
	}

	// Load review rules
	rules, err := internal.LoadRules(*configFile, engineType, hasMetadata)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading rules: %v\n", err)
		os.Exit(1)
	}
	req.Rules = rules

	// Perform review
	resp, err := advisor.SQLReviewCheck(context.Background(), req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during review: %v\n", err)
		os.Exit(1)
	}

	// Output results
	if err := internal.OutputResults(resp, statement, engineType, *outputFormat, dbParams); err != nil {
		fmt.Fprintf(os.Stderr, "Error outputting results: %v\n", err)
		os.Exit(1)
	}

	// Exit with error code if there are errors
	if resp.HasError {
		os.Exit(2)
	}
	if resp.HasWarning {
		os.Exit(1)
	}
}

// getStatement reads SQL statement from command line or file.
func getStatement() (string, error) {
	if *sqlStatement != "" {
		if *sqlStatement == "-" {
			// Read from stdin
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return "", fmt.Errorf("failed to read from stdin: %w", err)
			}
			return string(data), nil
		}
		return *sqlStatement, nil
	}

	if *sqlFile != "" {
		data, err := os.ReadFile(*sqlFile)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", *sqlFile, err)
		}
		return string(data), nil
	}

	return "", nil
}

// buildDBParams builds database connection parameters from command line flags.
func buildDBParams() *internal.DBConnectionParams {
	return &internal.DBConnectionParams{
		Host:        *dbHost,
		Port:        *dbPort,
		User:        *dbUser,
		Password:    *dbPassword,
		DbName:      *dbName,
		Charset:     *dbCharset,
		ServiceName: *dbServiceName,
		Sid:         *dbSid,
		SSLMode:     *dbSSLMode,
		Timeout:     *dbTimeout,
	}
}

// printVersion prints version information.
func printVersion() {
	fmt.Printf("SQL Advisor Tool v%s\n", toolVersion)
	fmt.Println("Based on Bytebase SQL Review Engine")
	fmt.Println("Supported engines: mysql, postgres, tidb, oracle, mssql, snowflake, mariadb, oceanbase")
}
