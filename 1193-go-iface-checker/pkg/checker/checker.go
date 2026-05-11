package checker

import (
	"iface-checker/pkg/models"
	"strings"
)

type TypeRegistry struct {
	structs   map[string]models.StructType
	interfaces map[string]models.InterfaceType
}

func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		structs:    make(map[string]models.StructType),
		interfaces: make(map[string]models.InterfaceType),
	}
}

func (r *TypeRegistry) RegisterStruct(s models.StructType) {
	r.structs[s.Name] = s
}

func (r *TypeRegistry) RegisterInterface(i models.InterfaceType) {
	r.interfaces[i.Name] = i
}

func (r *TypeRegistry) GetStruct(name string) (models.StructType, bool) {
	s, ok := r.structs[name]
	return s, ok
}

func (r *TypeRegistry) GetInterface(name string) (models.InterfaceType, bool) {
	i, ok := r.interfaces[name]
	return i, ok
}

func (r *TypeRegistry) CollectMethods(structName string, useValueReceiver bool, visited map[string]bool) map[string]models.Method {
	if visited == nil {
		visited = make(map[string]bool)
	}
	
	if visited[structName] {
		return make(map[string]models.Method)
	}
	visited[structName] = true
	
	result := make(map[string]models.Method)
	
	s, ok := r.structs[structName]
	if !ok {
		return result
	}
	
	for _, m := range s.Methods {
		if useValueReceiver && m.IsPointerReceiver {
			continue
		}
		result[m.Name] = m
	}
	
	for _, embedded := range s.EmbeddedFields {
		var embeddedMethods map[string]models.Method
		
		if embedded.IsInterface {
			if iface, ok := r.interfaces[embedded.Type]; ok {
				embeddedMethods = make(map[string]models.Method)
				for _, m := range iface.Methods {
					embeddedMethods[m.Name] = m
				}
			}
		} else {
			embeddedTypeName := embedded.Type
			if strings.HasPrefix(embeddedTypeName, "*") {
				embeddedTypeName = embeddedTypeName[1:]
			}
			
			useEmbeddedValueReceiver := useValueReceiver
			if embedded.IsPointer {
				useEmbeddedValueReceiver = false
			}
			
			embeddedMethods = r.CollectMethods(embeddedTypeName, useEmbeddedValueReceiver, visited)
		}
		
		for name, m := range embeddedMethods {
			if _, exists := result[name]; !exists {
				result[name] = m
			}
		}
	}
	
	return result
}

func CompareMethods(expected, actual models.Method) []models.MismatchDetail {
	var details []models.MismatchDetail
	
	if len(expected.Params) != len(actual.Params) {
		details = append(details, models.MismatchDetail{
			IssueType: "param_count",
			Expected:  string(rune(len(expected.Params))),
			Actual:    string(rune(len(actual.Params))),
		})
		return details
	}
	
	for i, expParam := range expected.Params {
		actParam := actual.Params[i]
		
		if expParam.Type != actParam.Type {
			details = append(details, models.MismatchDetail{
				IssueType: "param_type",
				Expected:  expParam.Type,
				Actual:    actParam.Type,
			})
		}
		
		if expParam.IsVariadic != actParam.IsVariadic {
			details = append(details, models.MismatchDetail{
				IssueType: "variadic",
				Expected:  boolToString(expParam.IsVariadic),
				Actual:    boolToString(actParam.IsVariadic),
			})
		}
	}
	
	if len(expected.Returns) != len(actual.Returns) {
		details = append(details, models.MismatchDetail{
			IssueType: "return_count",
			Expected:  string(rune(len(expected.Returns))),
			Actual:    string(rune(len(actual.Returns))),
		})
		return details
	}
	
	for i, expRet := range expected.Returns {
		actRet := actual.Returns[i]
		
		if expRet != actRet {
			details = append(details, models.MismatchDetail{
				IssueType: "return_type",
				Expected:  expRet,
				Actual:    actRet,
			})
		}
	}
	
	return details
}

func Check(registry *TypeRegistry, structName, interfaceName string, useValueReceiver bool) models.CheckResponse {
	iface, ok := registry.GetInterface(interfaceName)
	if !ok {
		return models.CheckResponse{
			Satisfies: false,
			Message:   "interface not found: " + interfaceName,
		}
	}
	
	if _, ok := registry.GetStruct(structName); !ok {
		return models.CheckResponse{
			Satisfies: false,
			Message:   "struct not found: " + structName,
		}
	}
	
	structMethods := registry.CollectMethods(structName, useValueReceiver, nil)
	
	var missing []models.MethodCheckResult
	var mismatched []models.MethodCheckResult
	
	for _, expectedMethod := range iface.Methods {
		actualMethod, found := structMethods[expectedMethod.Name]
		
		if !found {
			missing = append(missing, models.MethodCheckResult{
				MethodName: expectedMethod.Name,
				IsMissing:  true,
			})
			continue
		}
		
		mismatches := CompareMethods(expectedMethod, actualMethod)
		if len(mismatches) > 0 {
			mismatched = append(mismatched, models.MethodCheckResult{
				MethodName: expectedMethod.Name,
				IsMismatch: true,
				Details:    mismatches,
			})
		}
	}
	
	satisfies := len(missing) == 0 && len(mismatched) == 0
	
	message := "struct " + structName + " satisfies interface " + interfaceName
	if !satisfies {
		message = "struct " + structName + " does NOT satisfy interface " + interfaceName
		if !useValueReceiver {
			message += " (using pointer receiver)"
		} else {
			message += " (using value receiver)"
		}
	}
	
	return models.CheckResponse{
		Satisfies:         satisfies,
		MissingMethods:    missing,
		MismatchedMethods: mismatched,
		Message:           message,
	}
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
