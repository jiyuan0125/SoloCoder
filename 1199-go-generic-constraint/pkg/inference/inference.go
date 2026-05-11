package inference

import (
	"generic-checker/pkg/types"
)

type InferenceResult struct {
	InferredTypes map[string]*types.Type
	Missing       []string
	Conflicts     map[string]string
}

func InferTypeParams(sig *types.FunctionSignature, callSite *types.CallSite) *InferenceResult {
	result := &InferenceResult{
		InferredTypes: make(map[string]*types.Type),
		Missing:       []string{},
		Conflicts:     make(map[string]string),
	}

	for name, t := range callSite.ExplicitTypeArgs {
		tCopy := t
		result.InferredTypes[name] = &tCopy
	}

	for i, param := range sig.Params {
		if i >= len(callSite.Args) {
			continue
		}
		arg := &callSite.Args[i]
		unify(param, arg, result)
	}

	for _, tp := range sig.TypeParams {
		if _, ok := result.InferredTypes[tp.Name]; !ok {
			result.Missing = append(result.Missing, tp.Name)
		}
	}

	return result
}

func unify(paramType types.Type, argType *types.Type, result *InferenceResult) {
	if isTypeParamReference(&paramType) {
		tpName := getTypeParamName(&paramType)
		if existing, ok := result.InferredTypes[tpName]; ok {
			if !typesCompatible(existing, argType) {
				result.Conflicts[tpName] = "conflicting type inference"
			}
		} else {
			result.InferredTypes[tpName] = argType
		}
		return
	}

	if paramType.Kind == types.KindSlice && argType.Kind == types.KindSlice {
		if paramType.ElementType != nil && argType.ElementType != nil {
			unify(*paramType.ElementType, argType.ElementType, result)
		}
		return
	}

	if paramType.Kind == types.KindPointer && argType.Kind == types.KindPointer {
		if paramType.ElementType != nil && argType.ElementType != nil {
			unify(*paramType.ElementType, argType.ElementType, result)
		}
		return
	}

	if paramType.Kind == types.KindMap && argType.Kind == types.KindMap {
		if paramType.KeyType != nil && argType.KeyType != nil {
			unify(*paramType.KeyType, argType.KeyType, result)
		}
		if paramType.ElementType != nil && argType.ElementType != nil {
			unify(*paramType.ElementType, argType.ElementType, result)
		}
		return
	}
}

func isTypeParamReference(t *types.Type) bool {
	return t.Kind == types.KindNamed && isTypeParamMarker(t.Name)
}

const (
	typeParamPrefix    = "TParam:"
	typeParamPrefixLen = len(typeParamPrefix)
)

func getTypeParamName(t *types.Type) string {
	if len(t.Name) > typeParamPrefixLen && t.Name[:typeParamPrefixLen] == typeParamPrefix {
		return t.Name[typeParamPrefixLen:]
	}
	return t.Name
}

func isTypeParamMarker(name string) bool {
	return len(name) > typeParamPrefixLen && name[:typeParamPrefixLen] == typeParamPrefix
}

func typesCompatible(t1, t2 *types.Type) bool {
	if t1.Kind != t2.Kind {
		return false
	}
	if t1.Kind == types.KindBasic {
		return t1.Name == t2.Name
	}
	if t1.Kind == types.KindNamed {
		return t1.Name == t2.Name
	}
	if t1.Kind == types.KindSlice {
		return typesCompatible(t1.ElementType, t2.ElementType)
	}
	if t1.Kind == types.KindPointer {
		return typesCompatible(t1.ElementType, t2.ElementType)
	}
	if t1.Kind == types.KindMap {
		return typesCompatible(t1.KeyType, t2.KeyType) &&
			typesCompatible(t1.ElementType, t2.ElementType)
	}
	return t1.Name == t2.Name
}

func MakeTypeParamRef(name string) types.Type {
	return types.Type{
		Kind: types.KindNamed,
		Name: "TParam:" + name,
	}
}
