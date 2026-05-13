package cmd

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"credits/pkg/models"
)

var courseCmd = &cobra.Command{
	Use:   "course",
	Short: "课程管理",
	Long:  "课程管理命令 - 添加、查询、更新和删除课程",
}

var (
	courseID        string
	courseName      string
	courseCredits   int
	courseType      string
	courseValidDays int
)

var addCourseCmd = &cobra.Command{
	Use:   "add",
	Short: "添加课程",
	Long:  "添加一个新的培训课程",
	Run: func(cmd *cobra.Command, args []string) {
		if courseName == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 课程名称不能为空")
			return
		}
		if courseCredits <= 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 学分必须大于 0")
			return
		}
		if courseType != "required" && courseType != "elective" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 课程类型必须是 required 或 elective")
			return
		}

		id := courseID
		if id == "" {
			id = uuid.New().String()[:8]
		}

		validUntil := time.Now().AddDate(0, 0, courseValidDays)

		course := &models.Course{
			ID:         id,
			Name:       courseName,
			Credits:    courseCredits,
			ValidUntil: validUntil,
			Type:       models.CourseType(courseType),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if err := store.AddCourse(course); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "添加课程失败: %v\n", err)
			return
		}

		printJSON(course)
	},
}

var listCoursesCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有课程",
	Long:  "列出所有已添加的培训课程",
	Run: func(cmd *cobra.Command, args []string) {
		courses := store.ListCourses()
		if courses == nil {
			courses = []*models.Course{}
		}
		printJSON(courses)
	},
}

var getCourseCmd = &cobra.Command{
	Use:   "get",
	Short: "获取课程详情",
	Long:  "根据课程 ID 获取课程详情",
	Run: func(cmd *cobra.Command, args []string) {
		if courseID == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 请指定课程 ID")
			return
		}

		course, err := store.GetCourse(courseID)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "获取课程失败: %v\n", err)
			return
		}

		printJSON(course)
	},
}

var updateCourseCmd = &cobra.Command{
	Use:   "update",
	Short: "更新课程",
	Long:  "更新已存在的课程信息",
	Run: func(cmd *cobra.Command, args []string) {
		if courseID == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 请指定课程 ID")
			return
		}

		existing, err := store.GetCourse(courseID)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "获取课程失败: %v\n", err)
			return
		}

		if courseName != "" {
			existing.Name = courseName
		}
		if courseCredits > 0 {
			existing.Credits = courseCredits
		}
		if courseType != "" {
			if courseType != "required" && courseType != "elective" {
				fmt.Fprintln(cmd.ErrOrStderr(), "错误: 课程类型必须是 required 或 elective")
				return
			}
			existing.Type = models.CourseType(courseType)
		}
		if courseValidDays > 0 {
			existing.ValidUntil = time.Now().AddDate(0, 0, courseValidDays)
		}
		existing.UpdatedAt = time.Now()

		if err := store.UpdateCourse(existing); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "更新课程失败: %v\n", err)
			return
		}

		printJSON(existing)
	},
}

var deleteCourseCmd = &cobra.Command{
	Use:   "delete",
	Short: "删除课程",
	Long:  "删除指定的课程",
	Run: func(cmd *cobra.Command, args []string) {
		if courseID == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 请指定课程 ID")
			return
		}

		if err := store.DeleteCourse(courseID); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "删除课程失败: %v\n", err)
			return
		}

		fmt.Printf("课程 %s 已删除\n", courseID)
	},
}

func init() {
	addCourseCmd.Flags().StringVar(&courseID, "id", "", "课程 ID（自动生成）")
	addCourseCmd.Flags().StringVar(&courseName, "name", "", "课程名称")
	addCourseCmd.Flags().IntVar(&courseCredits, "credits", 0, "学分数（必须大于 0）")
	addCourseCmd.Flags().StringVar(&courseType, "type", "elective", "课程类型：required 必修，elective 选修")
	addCourseCmd.Flags().IntVar(&courseValidDays, "valid-days", 365, "有效期天数")

	addCourseCmd.MarkFlagRequired("name")
	addCourseCmd.MarkFlagRequired("credits")

	getCourseCmd.Flags().StringVar(&courseID, "id", "", "课程 ID")
	getCourseCmd.MarkFlagRequired("id")

	updateCourseCmd.Flags().StringVar(&courseID, "id", "", "课程 ID")
	updateCourseCmd.Flags().StringVar(&courseName, "name", "", "课程名称")
	updateCourseCmd.Flags().IntVar(&courseCredits, "credits", 0, "学分数（必须大于 0）")
	updateCourseCmd.Flags().StringVar(&courseType, "type", "", "课程类型：required 必修，elective 选修")
	updateCourseCmd.Flags().IntVar(&courseValidDays, "valid-days", 0, "有效期天数")
	updateCourseCmd.MarkFlagRequired("id")

	deleteCourseCmd.Flags().StringVar(&courseID, "id", "", "课程 ID")
	deleteCourseCmd.MarkFlagRequired("id")

	courseCmd.AddCommand(addCourseCmd)
	courseCmd.AddCommand(listCoursesCmd)
	courseCmd.AddCommand(getCourseCmd)
	courseCmd.AddCommand(updateCourseCmd)
	courseCmd.AddCommand(deleteCourseCmd)
}
