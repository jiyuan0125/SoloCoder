package checker

import (
	"fmt"
	"strings"

	"generic-checker/pkg/types"
)

func CheckConstraint(t *types.Type, c *types.Constraint) error {
	switch c.Kind {
	case types.ConstraintAny:
		return nil
	case types.ConstraintTypes:
		return checkTypeList(t, c.Types, c.Approximate)
	case types.ConstraintUnion:
		for _, term := range c.UnionTerms {
			if err := CheckConstraint(t, &term); err == nil {
				return nil
			}
		}
		return fmt.Errorf("type %q does not satisfy any term in union constraint", describeType(t))
	case types.ConstraintInterface:
		return checkInterfaceConstraint(t, c)
	}
	return fmt.Errorf("unknown constraint kind")
}

func checkTypeList(t *types.Type, typeList []types.Type, approximate bool) error {
	for _, ct := range typeList {
		if matchesType(t, &ct, approximate) {
			return nil
		}
	}
	typeNames := make([]string, len(typeList))
	for i, ct := range typeList {
		if approximate {
			typeNames[i] = "~" + describeType(&ct)
		} else {
			typeNames[i] = describeType(&ct)
		}
	}
	return fmt.Errorf("type %q does not match any of [%s]", describeType(t), strings.Join(typeNames, " | "))
}

func matchesType(t *types.Type, constraintType *types.Type, approximate bool) bool {
	if t.Kind == types.KindNamed {
		if !approximate {
			if t.Name == constraintType.Name {
				return true
			}
		}
		underlying := t.UnderlyingType()
		if underlying == t {
			return t.Name == constraintType.Name
		}
		return matchesType(underlying, constraintType, true)
	}
	if t.Kind == constraintType.Kind {
		if t.Kind == types.KindBasic {
			return t.Name == constraintType.Name
		}
		if t.Kind == types.KindSlice {
			if constraintType.Kind == types.KindSlice {
				return matchesType(t.ElementType, constraintType.ElementType, approximate)
			}
		}
		if t.Kind == types.KindPointer {
			if constraintType.Kind == types.KindPointer {
				return matchesType(t.ElementType, constraintType.ElementType, approximate)
			}
		}
		if t.Kind == types.KindMap {
			if constraintType.Kind == types.KindMap {
				return matchesType(t.KeyType, constraintType.KeyType, approximate) &&
					matchesType(t.ElementType, constraintType.ElementType, approximate)
			}
		}
	}
	return false
}

func checkInterfaceConstraint(t *types.Type, c *types.Constraint) error {
	for _, m := range c.Methods {
		if !hasMethod(t, &m) {
			return fmt.Errorf("type %q missing method %q", describeType(t), formatMethodSignature(&m))
		}
	}
	if len(c.Types) > 0 || len(c.UnionTerms) > 0 {
		if len(c.UnionTerms) > 0 {
			for _, term := range c.UnionTerms {
				if CheckConstraint(t, &term) == nil {
					return nil
				}
			}
			return fmt.Errorf("type %q does not satisfy interface constraint union", describeType(t))
		}
		if err := checkTypeList(t, c.Types, c.Approximate); err != nil {
			return err
		}
	}
	return nil
}

func hasMethod(t *types.Type, method *types.Method) bool {
	for _, m := range t.Methods {
		if m.Name == method.Name {
			if methodsMatch(&m, method) {
				return true
			}
		}
	}
	if t.Kind == types.KindNamed && t.Underlying != nil {
		return hasMethod(t.Underlying, method)
	}
	return false
}

func methodsMatch(m1, m2 *types.Method) bool {
	if m1.Name != m2.Name {
		return false
	}
	if len(m1.Params) != len(m2.Params) {
		return false
	}
	if len(m1.Results) != len(m2.Results) {
		return false
	}
	for i := range m1.Params {
		if !typesEqual(&m1.Params[i], &m2.Params[i]) {
			return false
		}
	}
	for i := range m1.Results {
		if !typesEqual(&m1.Results[i], &m2.Results[i]) {
			return false
		}
	}
	return true
}

func typesEqual(t1, t2 *types.Type) bool {
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
		return typesEqual(t1.ElementType, t2.ElementType)
	}
	if t1.Kind == types.KindPointer {
		return typesEqual(t1.ElementType, t2.ElementType)
	}
	if t1.Kind == types.KindMap {
		return typesEqual(t1.KeyType, t2.KeyType) && typesEqual(t1.ElementType, t2.ElementType)
	}
	return t1.Name == t2.Name
}

func describeType(t *types.Type) string {
	if t.Kind == types.KindNamed {
		return t.Name
	}
	if t.Kind == types.KindPointer {
		return "*" + describeType(t.ElementType)
	}
	if t.Kind == types.KindSlice {
		return "[]" + describeType(t.ElementType)
	}
	if t.Kind == types.KindMap {
		return "map[" + describeType(t.KeyType) + "]" + describeType(t.ElementType)
	}
	return t.Name
}

func formatMethodSignature(m *types.Method) string {
	params := make([]string, len(m.Params))
	for i, p := range m.Params {
		params[i] = describeType(&p)
	}
	results := make([]string, len(m.Results))
	for i, r := range m.Results {
		results[i] = describeType(&r)
	}
	resultStr := strings.Join(results, ", ")
	if len(results) > 1 {
		resultStr = "(" + resultStr + ")"
	}
	return fmt.Sprintf("%s(%s) %s", m.Name, strings.Join(params, ", "), resultStr)
}
