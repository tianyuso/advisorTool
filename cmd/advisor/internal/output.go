package internal

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"

	"advisorTool/pkg/advisor"
)

// OutputResults outputs the review results in the specified format.
func OutputResults(resp *advisor.ReviewResponse, statement string, engineType advisor.Engine, format string, dbParams *DBConnectionParams) error {
	switch format {
	case "json":
		results := ConvertToReviewResults(resp, statement, engineType, dbParams)
		data, err := json.Marshal(results)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "yaml":
		data, err := yaml.Marshal(resp)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		// Text format
		if len(resp.Advices) == 0 {
			fmt.Println("✅ No issues found!")
			return nil
		}

		fmt.Printf("Found %d issue(s):\n\n", len(resp.Advices))
		for i, advice := range resp.Advices {
			icon := "⚠️"
			statusStr := "WARNING"
			if advice.Status == advisor.AdviceStatusError {
				icon = "❌"
				statusStr = "ERROR"
			}
			fmt.Printf("%d. %s [%s] %s\n", i+1, icon, statusStr, advice.Title)
			fmt.Printf("   %s\n", advice.Content)
			if advice.StartPosition != nil {
				fmt.Printf("   Location: line %d, column %d\n", advice.StartPosition.Line, advice.StartPosition.Column)
			}
			fmt.Println()
		}
	}
	return nil
}

// ListAvailableRules lists all available SQL review rules.
func ListAvailableRules() {
	fmt.Println("Available SQL Review Rules:")
	fmt.Println("===========================")
	fmt.Println()

	rules := advisor.AllRules()
	for _, ruleType := range rules {
		desc := advisor.GetRuleDescription(ruleType)
		fmt.Printf("  %s\n    %s\n\n", ruleType, desc)
	}

	fmt.Printf("\nTotal: %d rules\n", len(rules))
}
