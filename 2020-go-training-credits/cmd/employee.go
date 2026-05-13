package cmd

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"credits/pkg/models"
)

var employeeCmd = &cobra.Command{
	Use:   "employee",
	Short: "员工管理",
	Long:  "员工管理命令 - 添加、查询、更新和删除员工",
}

var (
	employeeID         string
	employeeName       string
	employeeDepartment string
)

var addEmployeeCmd = &cobra.Command{
	Use:   "add",
	Short: "添加员工",
	Long:  "添加一个新的员工",
	Run: func(cmd *cobra.Command, args []string) {
		if employeeName == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 员工名称不能为空")
			return
		}
		if employeeDepartment == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 部门不能为空")
			return
		}

		id := employeeID
		if id == "" {
			id = uuid.New().String()[:8]
		}

		employee := &models.Employee{
			ID:                       id,
			Name:                     employeeName,
			Department:               employeeDepartment,
			CompletedRequiredCourses: make(map[string]bool),
			CarryOverCredits:         make(map[int]int),
			CreatedAt:                time.Now(),
			UpdatedAt:                time.Now(),
		}

		if err := store.AddEmployee(employee); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "添加员工失败: %v\n", err)
			return
		}

		printJSON(employee)
	},
}

var listEmployeesCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有员工",
	Long:  "列出所有已添加的员工",
	Run: func(cmd *cobra.Command, args []string) {
		employees := store.ListEmployees()
		if employees == nil {
			employees = []*models.Employee{}
		}
		printJSON(employees)
	},
}

var getEmployeeCmd = &cobra.Command{
	Use:   "get",
	Short: "获取员工详情",
	Long:  "根据员工 ID 获取员工详情",
	Run: func(cmd *cobra.Command, args []string) {
		if employeeID == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 请指定员工 ID")
			return
		}

		employee, err := store.GetEmployee(employeeID)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "获取员工失败: %v\n", err)
			return
		}

		printJSON(employee)
	},
}

var updateEmployeeCmd = &cobra.Command{
	Use:   "update",
	Short: "更新员工",
	Long:  "更新已存在的员工信息",
	Run: func(cmd *cobra.Command, args []string) {
		if employeeID == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 请指定员工 ID")
			return
		}

		existing, err := store.GetEmployee(employeeID)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "获取员工失败: %v\n", err)
			return
		}

		if employeeName != "" {
			existing.Name = employeeName
		}
		if employeeDepartment != "" {
			existing.Department = employeeDepartment
		}
		existing.UpdatedAt = time.Now()

		if err := store.UpdateEmployee(existing); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "更新员工失败: %v\n", err)
			return
		}

		printJSON(existing)
	},
}

var deleteEmployeeCmd = &cobra.Command{
	Use:   "delete",
	Short: "删除员工",
	Long:  "删除指定的员工",
	Run: func(cmd *cobra.Command, args []string) {
		if employeeID == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 请指定员工 ID")
			return
		}

		if err := store.DeleteEmployee(employeeID); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "删除员工失败: %v\n", err)
			return
		}

		fmt.Printf("员工 %s 已删除\n", employeeID)
	},
}

func init() {
	addEmployeeCmd.Flags().StringVar(&employeeID, "id", "", "员工 ID（自动生成）")
	addEmployeeCmd.Flags().StringVar(&employeeName, "name", "", "员工名称")
	addEmployeeCmd.Flags().StringVar(&employeeDepartment, "department", "", "部门名称")

	addEmployeeCmd.MarkFlagRequired("name")
	addEmployeeCmd.MarkFlagRequired("department")

	getEmployeeCmd.Flags().StringVar(&employeeID, "id", "", "员工 ID")
	getEmployeeCmd.MarkFlagRequired("id")

	updateEmployeeCmd.Flags().StringVar(&employeeID, "id", "", "员工 ID")
	updateEmployeeCmd.Flags().StringVar(&employeeName, "name", "", "员工名称")
	updateEmployeeCmd.Flags().StringVar(&employeeDepartment, "department", "", "部门名称")
	updateEmployeeCmd.MarkFlagRequired("id")

	deleteEmployeeCmd.Flags().StringVar(&employeeID, "id", "", "员工 ID")
	deleteEmployeeCmd.MarkFlagRequired("id")

	employeeCmd.AddCommand(addEmployeeCmd)
	employeeCmd.AddCommand(listEmployeesCmd)
	employeeCmd.AddCommand(getEmployeeCmd)
	employeeCmd.AddCommand(updateEmployeeCmd)
	employeeCmd.AddCommand(deleteEmployeeCmd)
}
