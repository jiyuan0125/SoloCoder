package cmd

import (
	"encoding/json"
	"fmt"

	"protoconv/internal/resource"

	"github.com/spf13/cobra"
)

var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Manage resources",
	Long:  `Manage resources and their relationships`,
}

var resourceAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new resource",
	RunE:  runResourceAdd,
}

var resourceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all resources",
	RunE:  runResourceList,
}

var resourceSummaryCmd = &cobra.Command{
	Use:   "summary [resource-id]",
	Short: "Get summary for a resource or all resources",
	RunE:  runResourceSummary,
}

var resourceRelateCmd = &cobra.Command{
	Use:   "relate",
	Short: "Create a relation between two resources",
	RunE:  runResourceRelate,
}

func init() {
	rootCmd.AddCommand(resourceCmd)
	resourceCmd.AddCommand(resourceAddCmd)
	resourceCmd.AddCommand(resourceListCmd)
	resourceCmd.AddCommand(resourceSummaryCmd)
	resourceCmd.AddCommand(resourceRelateCmd)

	resourceAddCmd.Flags().String("type", "", "Resource type (proto_file, proto_binary, json_output, json_input)")
	resourceAddCmd.Flags().String("name", "", "Resource name")
	resourceAddCmd.Flags().String("path", "", "Resource file path")

	resourceAddCmd.MarkFlagRequired("type")
	resourceAddCmd.MarkFlagRequired("name")

	resourceListCmd.Flags().String("type", "", "Filter by resource type")

	resourceRelateCmd.Flags().String("source", "", "Source resource ID")
	resourceRelateCmd.Flags().String("target", "", "Target resource ID")
	resourceRelateCmd.Flags().String("relation", "", "Relation type (defines, generated_from, converted_to, part_of)")

	resourceRelateCmd.MarkFlagRequired("source")
	resourceRelateCmd.MarkFlagRequired("target")
	resourceRelateCmd.MarkFlagRequired("relation")
}

func runResourceAdd(cmd *cobra.Command, args []string) error {
	storagePath, _ := cmd.Flags().GetString("storage")
	resType, _ := cmd.Flags().GetString("type")
	name, _ := cmd.Flags().GetString("name")
	path, _ := cmd.Flags().GetString("path")

	rm, err := resource.NewManager(storagePath)
	if err != nil {
		return err
	}

	res := resource.Resource{
		Type: resource.ResourceType(resType),
		Name: name,
		Path: path,
	}

	if err := rm.AddResource(res); err != nil {
		return err
	}

	fmt.Printf("Resource added: %s\n", res.ID)
	return nil
}

func runResourceList(cmd *cobra.Command, args []string) error {
	storagePath, _ := cmd.Flags().GetString("storage")
	filterType, _ := cmd.Flags().GetString("type")

	rm, err := resource.NewManager(storagePath)
	if err != nil {
		return err
	}

	resources := rm.ListResources(resource.ResourceType(filterType))

	if len(resources) == 0 {
		fmt.Println("No resources found")
		return nil
	}

	fmt.Printf("%-20s %-20s %-30s %s\n", "ID", "TYPE", "NAME", "PATH")
	fmt.Println("---------------------------------------------------------------------------------------------")
	for _, r := range resources {
		fmt.Printf("%-20s %-20s %-30s %s\n", r.ID, r.Type, r.Name, r.Path)
	}

	return nil
}

func runResourceSummary(cmd *cobra.Command, args []string) error {
	storagePath, _ := cmd.Flags().GetString("storage")

	rm, err := resource.NewManager(storagePath)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		resourceID := args[0]
		summary, err := rm.GetSummary(resourceID)
		if err != nil {
			return err
		}
		printSummary(summary)
		return nil
	}

	summaries := rm.ListAllSummaries()
	if len(summaries) == 0 {
		fmt.Println("No resources found")
		return nil
	}

	for _, s := range summaries {
		printSummary(&s)
		fmt.Println("---")
	}

	return nil
}

func printSummary(s *resource.ResourceSummary) {
	data, _ := json.MarshalIndent(s, "", "  ")
	fmt.Println(string(data))
}

func runResourceRelate(cmd *cobra.Command, args []string) error {
	storagePath, _ := cmd.Flags().GetString("storage")
	sourceID, _ := cmd.Flags().GetString("source")
	targetID, _ := cmd.Flags().GetString("target")
	relType, _ := cmd.Flags().GetString("relation")

	rm, err := resource.NewManager(storagePath)
	if err != nil {
		return err
	}

	rel := resource.Relation{
		SourceID: sourceID,
		TargetID: targetID,
		Type:     resource.RelationType(relType),
	}

	if err := rm.AddRelation(rel); err != nil {
		return err
	}

	fmt.Printf("Relation created: %s\n", rel.ID)
	return nil
}
