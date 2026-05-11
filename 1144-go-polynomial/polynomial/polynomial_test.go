package polynomial

import (
	"testing"
)

func TestAdd(t *testing.T) {
	a, _ := New([]float64{1, 2})
	b, _ := New([]float64{3, 4, 5})
	result := Add(a, b)
	expected := []float64{4, 6, 5}
	
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("Add 错误: 期望 %v, 得到 %v", expected, result)
			break
		}
	}
}

func TestSub(t *testing.T) {
	a, _ := New([]float64{5, 4, 3})
	b, _ := New([]float64{3, 2, 1})
	result := Sub(a, b)
	expected := []float64{2, 2, 2}
	
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("Sub 错误: 期望 %v, 得到 %v", expected, result)
			break
		}
	}
}

func TestMul(t *testing.T) {
	a, _ := New([]float64{0, 1})
	b, _ := New([]float64{-1, 1})
	result := Mul(a, b)
	expected := []float64{0, -1, 1}
	
	if len(result) != len(expected) {
		t.Errorf("Mul 长度错误: 期望 %d, 得到 %d", len(expected), len(result))
		return
	}
	
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("Mul 错误: 期望 %v, 得到 %v", expected, result)
			break
		}
	}
}

func TestDiv(t *testing.T) {
	dividend, _ := New([]float64{0, 1, 1})
	divisor, _ := New([]float64{0, 1})
	quotient, remainder, err := Div(dividend, divisor)
	
	if err != nil {
		t.Errorf("Div 错误: %v", err)
		return
	}
	
	expectedQuotient := []float64{1, 1}
	expectedRemainder := []float64{0}
	
	for i, v := range expectedQuotient {
		if quotient[i] != v {
			t.Errorf("Div 商错误: 期望 %v, 得到 %v", expectedQuotient, quotient)
			break
		}
	}
	
	if len(remainder) != 1 || remainder[0] != 0 {
		t.Errorf("Div 余数错误: 期望 %v, 得到 %v", expectedRemainder, remainder)
	}
}

func TestEval(t *testing.T) {
	p, _ := New([]float64{-2, 0, 1})
	result := Eval(p, 1.414)
	expected := 1.414*1.414 - 2
	
	if result-expected > 0.001 || result-expected < -0.001 {
		t.Errorf("Eval 错误: 期望 %v, 得到 %v", expected, result)
	}
}

func TestDerivative(t *testing.T) {
	p, _ := New([]float64{2, 3, 1})
	result := Derivative(p)
	expected := []float64{3, 2}
	
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("Derivative 错误: 期望 %v, 得到 %v", expected, result)
			break
		}
	}
}

func TestFactor(t *testing.T) {
	p, _ := New([]float64{-4, 0, 1})
	result, err := Factor(p)
	
	if err != nil {
		t.Errorf("Factor 错误: %v", err)
		return
	}
	
	if result != "(x-2)(x+2)" {
		t.Errorf("Factor 错误: 期望 (x-2)(x+2), 得到 %s", result)
	}
	
	p2, _ := New([]float64{1, 2, 1})
	result2, _ := Factor(p2)
	if result2 != "(x+1)²" {
		t.Errorf("Factor 错误: 期望 (x+1)², 得到 %s", result2)
	}
}

func TestZeroPolynomial(t *testing.T) {
	p, _ := New([]float64{0, 0, 0})
	if !p.IsZero() {
		t.Error("零多项式识别错误")
	}
	
	if Eval(p, 5) != 0 {
		t.Error("零多项式求值错误")
	}
	
	deriv := Derivative(p)
	if !deriv.IsZero() {
		t.Error("零多项式求导错误")
	}
}
