package renderer

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"go-api-doc-gen/internal/models"
)

type Renderer struct{}

func New() *Renderer {
	return &Renderer{}
}

func (r *Renderer) RenderMarkdown(apis []models.API, versionName string) string {
	var sb strings.Builder

	apisByModule := groupByModule(apis)
	incomplete := findIncomplete(apis)

	if len(incomplete) > 0 {
		sb.WriteString("## 文档不完整的接口\n\n")
		sb.WriteString("以下接口的文档信息不完整，请补充：\n\n")
		for _, item := range incomplete {
			sb.WriteString(fmt.Sprintf("- `%s %s` (%s) - 缺失: %s\n",
				item.Method, item.Path, item.Module,
				strings.Join(item.MissingFields, ", ")))
		}
		sb.WriteString("\n---\n\n")
	}

	sb.WriteString(fmt.Sprintf("# API 文档 - %s\n\n", versionName))

	modules := make([]string, 0, len(apisByModule))
	for m := range apisByModule {
		modules = append(modules, m)
	}
	sort.Strings(modules)

	for _, module := range modules {
		moduleAPIs := apisByModule[module]
		sb.WriteString(fmt.Sprintf("## 模块: %s\n\n", module))

		for _, api := range moduleAPIs {
			statusBadge := ""
			if !api.IsComplete {
				statusBadge = " ⚠️ **文档不完整**"
			}

			sb.WriteString(fmt.Sprintf("### `%s %s`%s\n\n", api.Method, api.Path, statusBadge))

			if api.Description != "" {
				sb.WriteString(fmt.Sprintf("**描述**: %s\n\n", api.Description))
			} else {
				sb.WriteString("**描述**: 暂无描述\n\n")
			}

			if len(api.Params) > 0 {
				sb.WriteString("**参数**:\n\n")
				sb.WriteString("| 名称 | 类型 | 位置 | 必填 | 描述 |\n")
				sb.WriteString("|------|------|------|------|------|\n")
				for _, p := range api.Params {
					required := "否"
					if p.Required {
						required = "是"
					}
					sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
						p.Name, p.Type, p.In, required, p.Description))
				}
				sb.WriteString("\n")
			}

			if len(api.Returns) > 0 {
				sb.WriteString("**返回值**:\n\n")
				sb.WriteString("| 状态码 | 描述 |\n")
				sb.WriteString("|--------|------|\n")
				for _, ret := range api.Returns {
					sb.WriteString(fmt.Sprintf("| %s | %s |\n", ret.Code, ret.Description))
				}
				sb.WriteString("\n")
			}

			sb.WriteString("**示例**:\n\n")
			if api.Example.Request != "" || api.Example.Response != "" {
				if api.Example.Request != "" {
					sb.WriteString("**请求示例**:\n\n```json\n")
					sb.WriteString(api.Example.Request)
					sb.WriteString("\n```\n\n")
				}
				if api.Example.Response != "" {
					sb.WriteString("**响应示例**:\n\n```json\n")
					sb.WriteString(api.Example.Response)
					sb.WriteString("\n```\n\n")
				}
			} else {
				sb.WriteString("暂无示例\n\n")
			}

			if !api.IsComplete {
				sb.WriteString(fmt.Sprintf("⚠️ **文档缺失项**: %s\n\n",
					strings.Join(api.MissingFields, ", ")))
			}

			sb.WriteString("---\n\n")
		}
	}

	return sb.String()
}

func (r *Renderer) RenderOpenAPI(apis []models.API, versionName string, title string) string {
	incomplete := findIncomplete(apis)

	openapi := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":   title,
			"version": versionName,
		},
		"paths": make(map[string]interface{}),
	}

	if len(incomplete) > 0 {
		var missingList []string
		for _, item := range incomplete {
			missingList = append(missingList,
				fmt.Sprintf("%s %s (%s) - 缺失: %s",
					item.Method, item.Path, item.Module,
					strings.Join(item.MissingFields, ", ")))
		}
		openapi["x-incomplete-apis"] = missingList
	}

	for _, api := range apis {
		pathItem, ok := openapi["paths"].(map[string]interface{})[api.Path]
		if !ok {
			pathItem = make(map[string]interface{})
			openapi["paths"].(map[string]interface{})[api.Path] = pathItem
		}

		method := strings.ToLower(api.Method)
		if method == "all" {
			method = "get"
		}

		operation := map[string]interface{}{
			"summary":     api.Description,
			"operationId": api.HandlerName,
			"tags":        []string{api.Module},
		}

		if !api.IsComplete {
			operation["x-documentation-incomplete"] = true
			operation["x-missing-fields"] = api.MissingFields
		}

		if len(api.Params) > 0 {
			var params []map[string]interface{}
			for _, p := range api.Params {
				param := map[string]interface{}{
					"name":     p.Name,
					"in":       p.In,
					"required": p.Required,
					"schema": map[string]interface{}{
						"type": p.Type,
					},
				}
				if p.Description != "" {
					param["description"] = p.Description
				}
				params = append(params, param)
			}
			operation["parameters"] = params
		}

		responses := make(map[string]interface{})
		for _, ret := range api.Returns {
			responses[ret.Code] = map[string]interface{}{
				"description": ret.Description,
			}
		}
		if len(responses) == 0 {
			responses["200"] = map[string]interface{}{
				"description": "成功",
			}
		}
		operation["responses"] = responses

		if api.Example.Request != "" || api.Example.Response != "" {
			examples := make(map[string]interface{})
			if api.Example.Request != "" {
				examples["request"] = api.Example.Request
			}
			if api.Example.Response != "" {
				examples["response"] = api.Example.Response
			}
			operation["x-examples"] = examples
		}

		pathItem.(map[string]interface{})[method] = operation
	}

	data, _ := json.MarshalIndent(openapi, "", "  ")
	return string(data)
}

func (r *Renderer) CompareVersions(oldAPIs, newAPIs []models.API) models.DiffResult {
	oldMap := make(map[string]models.API)
	for _, api := range oldAPIs {
		key := fmt.Sprintf("%s:%s:%s", api.Module, api.Method, api.Path)
		oldMap[key] = api
	}

	newMap := make(map[string]models.API)
	for _, api := range newAPIs {
		key := fmt.Sprintf("%s:%s:%s", api.Module, api.Method, api.Path)
		newMap[key] = api
	}

	var added, removed, modified []models.DiffItem

	for key, newAPI := range newMap {
		if _, exists := oldMap[key]; !exists {
			added = append(added, models.DiffItem{
				Module:      newAPI.Module,
				Path:        newAPI.Path,
				Method:      newAPI.Method,
				HandlerName: newAPI.HandlerName,
			})
		} else {
			oldAPI := oldMap[key]
			if !apiEqual(oldAPI, newAPI) {
				modified = append(modified, models.DiffItem{
					Module:      newAPI.Module,
					Path:        newAPI.Path,
					Method:      newAPI.Method,
					HandlerName: newAPI.HandlerName,
				})
			}
		}
	}

	for key, oldAPI := range oldMap {
		if _, exists := newMap[key]; !exists {
			removed = append(removed, models.DiffItem{
				Module:      oldAPI.Module,
				Path:        oldAPI.Path,
				Method:      oldAPI.Method,
				HandlerName: oldAPI.HandlerName,
			})
		}
	}

	return models.DiffResult{
		Added:    added,
		Removed:  removed,
		Modified: modified,
	}
}

func groupByModule(apis []models.API) map[string][]models.API {
	result := make(map[string][]models.API)
	for _, api := range apis {
		module := api.Module
		if module == "" {
			module = "默认"
		}
		result[module] = append(result[module], api)
	}
	return result
}

func findIncomplete(apis []models.API) []models.API {
	var result []models.API
	for _, api := range apis {
		if !api.IsComplete {
			result = append(result, api)
		}
	}
	return result
}

func apiEqual(a, b models.API) bool {
	if a.Description != b.Description {
		return false
	}
	if a.HandlerName != b.HandlerName {
		return false
	}
	if a.Example.Request != b.Example.Request {
		return false
	}
	if a.Example.Response != b.Example.Response {
		return false
	}
	if len(a.Params) != len(b.Params) {
		return false
	}
	if len(a.Returns) != len(b.Returns) {
		return false
	}
	return true
}
