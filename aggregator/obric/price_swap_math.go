package obric

import (
	"math/big"
)

var (
	BILLION           = big.NewInt(1000000000)
	PRECISION_FACTOR  = big.NewInt(1000000)
	PRECISION_FACTOR2 = big.NewInt(1000000000000)
)

func getSwapXToYOut(currentX, currentY, inputX, k, k2, xa, xb, m, n *big.Int) *big.Int {
	var temp1 *big.Int
	maxXY := Max(currentX, currentY)
	numerator := big.NewInt(1)
	denominator := big.NewInt(1)
	if maxXY.Cmp(BILLION) > 0 {
		maxXY = new(big.Int).Div(maxXY, big.NewInt(10))
		denominator = new(big.Int).Mul(denominator, big.NewInt(10))
	}
	if maxXY.Cmp(new(big.Int).Div(BILLION, big.NewInt(10))) < 0 {
		maxXY = new(big.Int).Mul(maxXY, big.NewInt(10))
		numerator = new(big.Int).Mul(numerator, big.NewInt(10))
	}
	preprocessedInputX := new(big.Int).Div(new(big.Int).Mul(inputX, numerator), denominator)
	currentX_ := new(big.Int).Div(new(big.Int).Mul(currentX, numerator), denominator)
	currentY_ := new(big.Int).Div(new(big.Int).Mul(currentY, numerator), denominator)
	if preprocessedInputX.Cmp(big.NewInt(10000)) < 0 {
		temp1 = getSwapXToYOutPreprocessedInner(currentX_, currentY_, inputX, numerator, denominator, k, k2, xa, xb, m, n)
	} else {
		temp1 = getSwapXToYOutPreprocessedInner(currentX_, currentY_, preprocessedInputX, big.NewInt(1), big.NewInt(1), k, k2, xa, xb, m, n)
	}

	return new(big.Int).Div(new(big.Int).Div(new(big.Int).Mul(temp1, denominator), PRECISION_FACTOR), numerator)
}

func getSwapYToXOut(currentX, currentY, inputY, k, k2, xa, xb, m, n *big.Int) *big.Int {
	return getSwapXToYOut(currentY, currentX, inputY, k, k2, xa, xb, m, n)
}

func getSwapXToYOutPreprocessed(currentX, currentY, inputX, k, k2, xa, xb, m, n *big.Int) *big.Int {
	return getSwapXToYOutPreprocessedInner(currentX, currentY, inputX, big.NewInt(1), big.NewInt(1), k, k2, xa, xb, m, n)
}

func getSwapXToYOutPreprocessedInner(currentX, currentY, inputX, preprocessingNumerator, preprocessingDenominator, k, k2, xa, xb, m, n *big.Int) *big.Int {
	var temp1, temp20, temp28, temp3, temp31, temp32,
		dydx_denominator, dydx_denominator__7, dydx_numerator, dydx_numerator__6,
		f_denominator, f_denominator__22, f_denominator__5,
		f_numerator, f_numerator__21, f_numerator__4,
		input_xF_next_stage, input_xF_next_stage__14,
		p_current_xF, p_current_xF__23, p_current_xF__8, p_current_yF,
		p_current_yF__24, p_current_yF__9, p_delta_yF, p_delta_yF__18,
		p_delta_yF__29, p_delta_yF_this_stage, p_delta_yF_this_stage__13,
		p_input_xF, p_input_xF__10, p_input_xF__25,
		p_k, p_k2, p_m, p_n, p_new_xF,
		p_new_xF__11, p_new_xF__26, p_new_yF, p_new_yF__17,
		p_new_yF__27, p_output_y, p_output_y__16, p_output_y__19,
		p_output_y__2, p_output_y__30, p_output_yF_next_stage,
		p_output_yF_next_stage__15, p_output_y_max, p_output_y_max__12,
		p_xa, p_xb *big.Int
	p_xa = new(big.Int).Mul(xa, PRECISION_FACTOR)
	p_xb = new(big.Int).Mul(xb, PRECISION_FACTOR)
	p_m = new(big.Int).Mul(m, PRECISION_FACTOR)
	p_n = new(big.Int).Mul(n, PRECISION_FACTOR)
	p_k = new(big.Int).Mul(k, PRECISION_FACTOR2)
	p_k2 = new(big.Int).Mul(k2, PRECISION_FACTOR2)
	if compareFraction(currentX, currentY, xa, xb) {
		f_numerator, f_denominator, dydx_numerator, dydx_denominator = solveFUpperLeft(currentX, currentY, n, k2)
		p_current_xF = new(big.Int).Div(new(big.Int).Mul(new(big.Int).Mul(currentX, f_numerator), PRECISION_FACTOR), f_denominator)
		p_current_yF = new(big.Int).Div(new(big.Int).Mul(new(big.Int).Mul(currentY, f_numerator), PRECISION_FACTOR), f_denominator)
		p_input_xF = new(big.Int).Div(new(big.Int).Mul(new(big.Int).Div(new(big.Int).Mul(new(big.Int).Mul(inputX, f_numerator), PRECISION_FACTOR), f_denominator), preprocessingNumerator), preprocessingDenominator)
		p_new_xF = new(big.Int).Add(p_current_xF, p_input_xF)
		if p_new_xF.Cmp(p_xa) > 0 {
			p_output_y_max = mulW(new(big.Int).Div(new(big.Int).Mul(new(big.Int).Mul(inputX, PRECISION_FACTOR), preprocessingNumerator), preprocessingDenominator), dydx_numerator, dydx_denominator)
			p_delta_yF_this_stage = new(big.Int).Sub(p_current_yF, p_xb)
			input_xF_next_stage = new(big.Int).Div(new(big.Int).Sub(p_new_xF, p_xa), PRECISION_FACTOR)
			p_output_yF_next_stage = getSwapXToYOutPreprocessed(xa, xb, input_xF_next_stage, k, k2, xa, xb, m, n)
			p_output_y = new(big.Int).Div(new(big.Int).Mul(new(big.Int).Add(p_delta_yF_this_stage, p_output_yF_next_stage), f_denominator), f_numerator)
			temp3 = Min(p_output_y, p_output_y_max)
		} else {
			p_new_yF = new(big.Int).Add(new(big.Int).Div(p_k2, p_new_xF), p_n)
			if p_current_yF.Cmp(p_new_yF) > 0 {
				temp1 = new(big.Int).Sub(p_current_yF, p_new_yF)
			} else {
				temp1 = big.NewInt(0)
			}
			p_delta_yF = temp1
			p_output_y__2 = new(big.Int).Div(new(big.Int).Mul(p_delta_yF, f_denominator), f_numerator)
			temp3 = new(big.Int).Mul(p_output_y__2, big.NewInt(1))
		}
		temp32 = temp3
	} else {
		if compareFraction(currentX, currentY, xb, xa) {
			f_numerator__4, f_denominator__5, dydx_numerator__6, dydx_denominator__7 = solveFMiddle(currentX, currentY, m, k)
			p_current_xF__8 = new(big.Int).Div(new(big.Int).Mul(new(big.Int).Mul(currentX, f_numerator__4), PRECISION_FACTOR), f_denominator__5)
			p_current_yF__9 = new(big.Int).Div(new(big.Int).Mul(new(big.Int).Mul(currentY, f_numerator__4), PRECISION_FACTOR), f_denominator__5)
			p_input_xF__10 = new(big.Int).Div(new(big.Int).Mul(new(big.Int).Div(new(big.Int).Mul(new(big.Int).Mul(inputX, f_numerator__4), PRECISION_FACTOR), f_denominator__5), preprocessingNumerator), preprocessingDenominator)
			p_new_xF__11 = new(big.Int).Add(p_current_xF__8, p_input_xF__10)
			if p_new_xF__11.Cmp(p_xb) > 0 {
				p_output_y_max__12 = mulW(new(big.Int).Div(new(big.Int).Mul(new(big.Int).Mul(inputX, PRECISION_FACTOR), preprocessingNumerator), preprocessingDenominator), dydx_numerator__6, dydx_denominator__7)
				p_delta_yF_this_stage__13 = new(big.Int).Sub(p_current_yF__9, p_xa)
				input_xF_next_stage__14 = new(big.Int).Div(new(big.Int).Sub(p_new_xF__11, p_xb), PRECISION_FACTOR)
				p_output_yF_next_stage__15 = getSwapXToYOutPreprocessed(xb, xa, input_xF_next_stage__14, k, k2, xa, xb, m, n)
				p_output_y__16 = new(big.Int).Div(new(big.Int).Mul(new(big.Int).Add(p_delta_yF_this_stage__13, p_output_yF_next_stage__15), f_denominator__5), f_numerator__4)
				temp20 = Min(p_output_y__16, p_output_y_max__12)
			} else {
				tp := new(big.Int).Add(p_new_xF__11, p_m)
				p_new_yF__17 = new(big.Int).Sub(new(big.Int).Div(p_k, tp), p_m)
				p_delta_yF__18 = new(big.Int).Sub(p_current_yF__9, p_new_yF__17)
				p_output_y__19 = new(big.Int).Div(new(big.Int).Mul(p_delta_yF__18, f_denominator__5), f_numerator__4)
				temp20 = new(big.Int).Mul(p_output_y__19, big.NewInt(1))
			}
			temp31 = temp20
		} else {
			f_numerator__21, f_denominator__22, _, _ = solveFBottomRight(currentX, currentY, n, k2)
			p_current_xF__23 = new(big.Int).Div(new(big.Int).Mul(new(big.Int).Mul(currentX, f_numerator__21), PRECISION_FACTOR), f_denominator__22)
			p_current_yF__24 = new(big.Int).Div(new(big.Int).Mul(new(big.Int).Mul(currentY, f_numerator__21), PRECISION_FACTOR), f_denominator__22)
			p_input_xF__25 = new(big.Int).Div(new(big.Int).Mul(new(big.Int).Div(new(big.Int).Mul(new(big.Int).Mul(inputX, f_numerator__21), PRECISION_FACTOR), f_denominator__22), preprocessingNumerator), preprocessingDenominator)
			p_new_xF__26 = new(big.Int).Add(p_current_xF__23, p_input_xF__25)
			tp := new(big.Int).Sub(p_new_xF__26, p_n)
			p_new_yF__27 = new(big.Int).Div(p_k2, tp)
			if p_current_yF__24.Cmp(p_new_yF__27) > 0 {
				temp28 = new(big.Int).Sub(p_current_yF__24, p_new_yF__27)
			} else {
				temp28 = big.NewInt(0)
			}
			p_delta_yF__29 = temp28
			p_output_y__30 = new(big.Int).Div(new(big.Int).Mul(p_delta_yF__29, f_denominator__22), f_numerator__21)
			temp31 = new(big.Int).Mul(p_output_y__30, big.NewInt(1))
		}
		temp32 = temp31
	}

	return temp32
}

func compareFraction(firstNumerator, firstDenominator, secondNumerator, secondDenominator *big.Int) bool {
	return new(big.Int).Mul(firstNumerator, secondDenominator).Cmp(new(big.Int).Mul(firstDenominator, secondNumerator)) < 0
}

func solveFBottomRight(x, y, n, k2 *big.Int) (a, b, c, d *big.Int) {
	return solveFUpperLeft(y, x, n, k2)
}

func solveFUpperLeft(x, y, n, k2 *big.Int) (a, b, c, d *big.Int) {
	xn := new(big.Int).Mul(x, n)
	xy := new(big.Int).Mul(x, y)
	// (xn * xn + ( 4 * xy * k2 )
	temp2 := new(big.Int).Mul(new(big.Int).Mul(big.NewInt(4), xy), k2)
	temp := new(big.Int).Add(new(big.Int).Mul(xn, xn), temp2)
	temp1 := new(big.Int).Sqrt(temp)
	numerator := new(big.Int).Add(xn, temp1)
	denominator := new(big.Int).Mul(big.NewInt(2), xy)
	xF := mulW(x, numerator, denominator)
	return numerator, denominator, k2, new(big.Int).Mul(xF, xF)
}

func solveFMiddle(x, y, m, k *big.Int) (a_, b_, c_, d_ *big.Int) {
	xy := new(big.Int).Mul(x, y)
	x_plus_y := new(big.Int).Add(x, y)
	b := new(big.Int).Mul(x_plus_y, m)
	temp4 := new(big.Int).Mul(m, m)
	temp5 := new(big.Int).Sub(k, temp4)
	temp6 := new(big.Int).Mul(new(big.Int).Mul(big.NewInt(4), xy), temp5)
	temp := new(big.Int).Add(new(big.Int).Mul(b, b), temp6)
	temp3 := new(big.Int).Sqrt(temp)
	numerator := new(big.Int).Sub(temp3, b)
	denominator := new(big.Int).Mul(big.NewInt(2), xy)
	xF := mulW(x, numerator, denominator)
	xf_plus_m := new(big.Int).Add(xF, m)
	return numerator, denominator, k, new(big.Int).Mul(xf_plus_m, xf_plus_m)
}

func mulW(multiplier, numerator, denominator *big.Int) *big.Int {
	return new(big.Int).Div(new(big.Int).Mul(multiplier, numerator), denominator)
}

func Max(currentX, currentY *big.Int) *big.Int {
	if currentX.Cmp(currentY) > 0 {
		return currentX
	}
	return currentY
}

func Min(x, y *big.Int) *big.Int {
	if x.Cmp(y) < 0 {
		return x
	}
	return y
}
