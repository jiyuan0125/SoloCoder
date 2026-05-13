package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"contract-lifecycle/models"

	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "按条件查询合同",
	RunE: func(cmd *cobra.Command, args []string) error {
		status, _ := cmd.Flags().GetString("status")
		party, _ := cmd.Flags().GetString("party")
		startDateStr, _ := cmd.Flags().GetString("start-date")
		endDateStr, _ := cmd.Flags().GetString("end-date")

		var startDate, endDate *time.Time

		if startDateStr != "" {
			d, err := parseDate(startDateStr)
			if err != nil {
				return err
			}
			startDate = &d
		}

		if endDateStr != "" {
			d, err := parseDate(endDateStr)
			if err != nil {
				return err
			}
			endDate = &d
		}

		contracts, err := svc.QueryContracts(status, party, startDate, endDate)
		if err != nil {
			return err
		}

		if len(contracts) == 0 {
			fmt.Println("[]")
			return nil
		}

		data, _ := json.MarshalIndent(contracts, "", "  ")
		fmt.Println(string(data))
		return nil
	},
}

var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "资源管理 (create/list/show)",
}

var resourceCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建资源",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		resourceType, _ := cmd.Flags().GetString("type")

		resource := &models.Resource{
			Name: name,
			Type: models.ResourceType(resourceType),
		}

		if err := svc.CreateResource(resource); err != nil {
			return err
		}

		fmt.Printf("资源创建成功！\nID: %s\n名称: %s\n类型: %s\n", resource.ID, resource.Name, resource.Type)
		return nil
	},
}

var resourceListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有资源",
	RunE: func(cmd *cobra.Command, args []string) error {
		resources, err := svc.ListResources()
		if err != nil {
			return err
		}

		if len(resources) == 0 {
			fmt.Println("[]")
			return nil
		}

		fmt.Println("资源列表:")
		for _, r := range resources {
			fmt.Printf("  ID: %s\n    名称: %s\n    类型: %s\n\n", r.ID, r.Name, r.Type)
		}
		return nil
	},
}

var resourceShowCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "显示资源详情",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resource, err := svc.GetResource(args[0])
		if err != nil {
			return err
		}

		data, _ := json.MarshalIndent(resource, "", "  ")
		fmt.Println(string(data))
		return nil
	},
}

var linkCmd = &cobra.Command{
	Use:   "link [contract-id] [resource-id]",
	Short: "关联合同到资源",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := svc.LinkContractToResource(args[0], args[1]); err != nil {
			return err
		}

		fmt.Println("关联成功！")
		return nil
	},
}

var summaryCmd = &cobra.Command{
	Use:   "summary [resource-id]",
	Short: "查看资源关联的合同汇总",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resource, err := svc.GetResource(args[0])
		if err != nil {
			return err
		}

		contracts, err := svc.GetContractsByResource(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("资源: %s (%s)\n", resource.Name, resource.Type)
		fmt.Printf("关联合同数: %d\n", len(contracts))

		if len(contracts) == 0 {
			fmt.Println("[]")
			return nil
		}

		fmt.Println("\n关联合同列表:")
		totalAmount := 0.0
		for _, c := range contracts {
			fmt.Printf("  - %s (ID: %s)\n    状态: %s, 金额: %.2f\n", c.Title, c.ID, c.Status, c.Amount)
			totalAmount += c.Amount
		}
		fmt.Printf("\n关联合同总金额: %.2f\n", totalAmount)
		return nil
	},
}

func init() {
	queryCmd.Flags().String("status", "", "按状态过滤 (draft/legal_review/modifying/pending_sign/signed/performing/expiry_remind/expired/terminated)")
	queryCmd.Flags().String("party", "", "按签约方过滤")
	queryCmd.Flags().String("start-date", "", "到期日期开始 (YYYY-MM-DD)")
	queryCmd.Flags().String("end-date", "", "到期日期结束 (YYYY-MM-DD)")

	resourceCmd.AddCommand(resourceCreateCmd)
	resourceCmd.AddCommand(resourceListCmd)
	resourceCmd.AddCommand(resourceShowCmd)

	resourceCreateCmd.Flags().String("name", "", "资源名称")
	resourceCreateCmd.Flags().String("type", "", "资源类型 (project/customer/vendor/asset/other)")
	resourceCreateCmd.MarkFlagRequired("name")
	resourceCreateCmd.MarkFlagRequired("type")
}
