package polynomial

import (
	"fmt"
	"math"
)

func Factor(p Polynomial) (string, error) {
	degree := p.Degree()
	
	if degree == -1 {
		return "", fmt.Errorf("零多项式无法分解")
	}
	
	if degree == 0 {
		return fmt.Sprintf("%v", p[0]), nil
	}
	
	if degree == 1 {
		a := p[1]
		b := p[0]
		
		if math.Abs(a-1.0) < 1e-10 {
			if math.Abs(b) < 1e-10 {
				return "x", nil
			}
			if b > 0 {
				return fmt.Sprintf("(x+%v)", formatFloat(b)), nil
			}
			return fmt.Sprintf("(x%v)", formatFloat(b)), nil
		}
		
		if math.Abs(b) < 1e-10 {
			return fmt.Sprintf("(%v x)", formatFloat(a)), nil
		}
		
		if b > 0 {
			return fmt.Sprintf("(%v x+%v)", formatFloat(a), formatFloat(b)), nil
		}
		return fmt.Sprintf("(%v x%v)", formatFloat(a), formatFloat(b)), nil
	}
	
	if degree == 2 {
		a := p[2]
		b := p[1]
		c := p[0]
		
		discriminant := b*b - 4*a*c
		
		if discriminant < 0 {
			return "无法分解", nil
		}
		
		sqrtDisc := math.Sqrt(discriminant)
		
		root1 := (-b + sqrtDisc) / (2 * a)
		root2 := (-b - sqrtDisc) / (2 * a)
		
		if math.Abs(root1-root2) < 1e-10 {
			root := root1
			
			if math.Abs(root) < 1e-10 {
				if math.Abs(a-1.0) < 1e-10 {
					return "x²", nil
				}
				return fmt.Sprintf("%v x²", formatFloat(a)), nil
			}
			
			if math.Abs(a-1.0) < 1e-10 {
				if root > 0 {
					return fmt.Sprintf("(x-%v)²", formatFloat(root)), nil
				}
				return fmt.Sprintf("(x+%v)²", formatFloat(-root)), nil
			}
			
			if root > 0 {
				return fmt.Sprintf("%v (x-%v)²", formatFloat(a), formatFloat(root)), nil
			}
			return fmt.Sprintf("%v (x+%v)²", formatFloat(a), formatFloat(-root)), nil
		}
		
		buildFactor := func(root float64) string {
			if math.Abs(root) < 1e-10 {
				return "x"
			}
			if root > 0 {
				return fmt.Sprintf("(x-%v)", formatFloat(root))
			}
			return fmt.Sprintf("(x+%v)", formatFloat(-root))
		}
		
		result := buildFactor(root1) + buildFactor(root2)
		
		if math.Abs(a-1.0) > 1e-10 {
			result = fmt.Sprintf("%v ", formatFloat(a)) + result
		}
		
		return result, nil
	}
	
	return "无法分解", nil
}

func formatFloat(f float64) string {
	if math.Abs(f-math.Round(f)) < 1e-10 {
		return fmt.Sprintf("%v", int(math.Round(f)))
	}
	return fmt.Sprintf("%.4g", f)
}
