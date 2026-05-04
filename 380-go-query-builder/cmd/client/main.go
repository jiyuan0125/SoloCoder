package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"querybuilder/pkg/common"
	"querybuilder/pkg/querybuilder"
	"strconv"
	"strings"
)

type stringArray []string

func (s *stringArray) String() string {
	return strings.Join(*s, ",")
}

func (s *stringArray) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	var (
		serverURL = flag.String("server", "http://localhost:8080", "Server URL")
		equals    = flag.String("equals", "", "Equal conditions: field=value, multiple separated by semicolons")
		ranges    = flag.String("ranges", "", "Range conditions: field=start,end, multiple separated by semicolons")
		ins       = flag.String("ins", "", "In conditions: field=value1,value2,value3, multiple separated by semicolons")
		likes     = flag.String("likes", "", "Like conditions: field=value:type (type: both|prefix|suffix), multiple separated by semicolons")
		sorts     = flag.String("sorts", "", "Sort conditions: field=order (order: asc|desc), multiple separated by semicolons")
		page      = flag.Int("page", 1, "Page number")
		pageSize  = flag.Int("page-size", 10, "Page size (1-100)")
	)

	flag.Parse()

	req := common.QueryRequest{
		Page:      *page,
		PageSize:  *pageSize,
	}

	if *equals != "" {
		conditions := strings.Split(*equals, ";")
		for _, cond := range conditions {
			parts := strings.SplitN(cond, "=", 2)
			if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
				req.Conditions = append(req.Conditions, common.ConditionRequest{
					Type:  querybuilder.ConditionTypeEqual,
					Field: parts[0],
					Value: parseValue(parts[1]),
				})
			}
		}
	}

	if *ranges != "" {
		conditions := strings.Split(*ranges, ";")
		for _, cond := range conditions {
			parts := strings.SplitN(cond, "=", 2)
			if len(parts) == 2 && parts[0] != "" {
				rangeParts := strings.SplitN(parts[1], ",", 2)
				if len(rangeParts) == 2 {
					req.Conditions = append(req.Conditions, common.ConditionRequest{
						Type:  querybuilder.ConditionTypeRange,
						Field: parts[0],
						Start: parseValue(rangeParts[0]),
						End:   parseValue(rangeParts[1]),
					})
				}
			}
		}
	}

	if *ins != "" {
		conditions := strings.Split(*ins, ";")
		for _, cond := range conditions {
			parts := strings.SplitN(cond, "=", 2)
			if len(parts) == 2 && parts[0] != "" {
				values := strings.Split(parts[1], ",")
				if len(values) > 0 {
					interfaceValues := make([]interface{}, len(values))
					for i, v := range values {
						interfaceValues[i] = parseValue(v)
					}
					req.Conditions = append(req.Conditions, common.ConditionRequest{
						Type:   querybuilder.ConditionTypeIn,
						Field:  parts[0],
						Values: interfaceValues,
					})
				}
			}
		}
	}

	if *likes != "" {
		conditions := strings.Split(*likes, ";")
		for _, cond := range conditions {
			parts := strings.SplitN(cond, "=", 2)
			if len(parts) == 2 && parts[0] != "" {
				likeParts := strings.SplitN(parts[1], ":", 2)
				value := likeParts[0]
				likeType := querybuilder.LikeTypeBoth
				
				if len(likeParts) == 2 {
					switch likeParts[1] {
					case "prefix":
						likeType = querybuilder.LikeTypePrefix
					case "suffix":
						likeType = querybuilder.LikeTypeSuffix
					}
				}
				
				req.Conditions = append(req.Conditions, common.ConditionRequest{
					Type:     querybuilder.ConditionTypeLike,
					Field:    parts[0],
					Value:    value,
					LikeType: likeType,
				})
			}
		}
	}

	if *sorts != "" {
		sortConditions := strings.Split(*sorts, ";")
		for _, sort := range sortConditions {
			parts := strings.SplitN(sort, "=", 2)
			if len(parts) == 2 && parts[0] != "" {
				order := querybuilder.SortOrderAsc
				if parts[1] == "desc" {
					order = querybuilder.SortOrderDesc
				}
				
				req.SortOrders = append(req.SortOrders, common.SortRequest{
					Field: parts[0],
					Order: order,
				})
			}
		}
	}

	c := NewClient(*serverURL)
	resp, err := c.BuildQuery(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	output, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}

func parseValue(s string) interface{} {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}
	return s
}
