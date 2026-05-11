package checker

import (
	"generic-checker/pkg/inference"
	"generic-checker/pkg/types"
)

func CheckCallSite(sig *types.FunctionSignature, callSite *types.CallSite) types.CheckResult {
	result := types.CheckResult{
		CallSiteID:    callSite.ID,
		InferredTypes: make(map[string]types.Type),
		Errors:        []types.CheckError{},
	}

	infResult := inference.InferTypeParams(sig, callSite)

	if len(infResult.Conflicts) > 0 {
		result.Status = types.StatusInferenceFailed
		for tpName, msg := range infResult.Conflicts {
			result.Errors = append(result.Errors, types.CheckError{
				TypeParamName: tpName,
				Message:       msg,
			})
		}
		return result
	}

	if len(infResult.Missing) > 0 {
		result.Status = types.StatusInferenceFailed
		for _, tpName := range infResult.Missing {
			result.Errors = append(result.Errors, types.CheckError{
				TypeParamName: tpName,
				Message:       "cannot infer type parameter from arguments",
			})
		}
		for name, t := range infResult.InferredTypes {
			result.InferredTypes[name] = *t
		}
		return result
	}

	for name, t := range infResult.InferredTypes {
		result.InferredTypes[name] = *t
	}

	result.Status = types.StatusPassed
	for _, tp := range sig.TypeParams {
		inferredType, ok := infResult.InferredTypes[tp.Name]
		if !ok {
			continue
		}
		if err := CheckConstraint(inferredType, &tp.Constraint); err != nil {
			result.Status = types.StatusConstraintViolation
			result.Errors = append(result.Errors, types.CheckError{
				TypeParamName: tp.Name,
				Message:       err.Error(),
			})
		}
	}

	return result
}

func CheckAll(sig *types.FunctionSignature, callSites []types.CallSite) []types.CheckResult {
	results := make([]types.CheckResult, len(callSites))
	for i, cs := range callSites {
		results[i] = CheckCallSite(sig, &cs)
	}
	return results
}
