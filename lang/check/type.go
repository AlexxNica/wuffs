// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package check

import (
	"fmt"
	"math/big"

	a "github.com/google/puffs/lang/ast"
	t "github.com/google/puffs/lang/token"
)

type typeChecker struct {
	c       *Checker
	idMap   *t.IDMap
	typeMap TypeMap

	errFilename string
	errLine     uint32
}

func (c *typeChecker) checkVars(n *a.Node) error {
	c.errFilename, c.errLine = n.Raw().FilenameLine()

	if n.Kind() == a.KVar {
		v := n.Var()
		name := v.Name()
		if _, ok := c.typeMap[name]; ok {
			return fmt.Errorf("check: duplicate var %q", c.idMap.ByID(name))
		}
		if err := c.checkTypeExpr(v.XType()); err != nil {
			return err
		}
		if value := v.Value(); value != nil {
			// TODO: check that value doesn't mention the variable itself.
		}
		c.typeMap[name] = v.XType()
		return nil
	}
	for _, l := range n.Raw().SubLists() {
		for _, m := range l {
			if err := c.checkVars(m); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *typeChecker) checkStatement(n *a.Node) error {
	c.errFilename, c.errLine = n.Raw().FilenameLine()

	switch n.Kind() {
	case a.KAssert:
		o := n.Assert()
		cond := o.Condition()
		if err := c.checkExpr(cond); err != nil {
			return err
		}
		if !cond.MType().Eq(TypeExprBoolean) {
			return fmt.Errorf("check: assert condition %q, of type %q, does not have a boolean type",
				cond.String(c.idMap), cond.MType().String(c.idMap))
		}
		// TODO: check the actual assertion.
		return nil

	case a.KAssign:
		return c.checkAssign(n.Assign())

	case a.KVar:
		o := n.Var()
		if value := o.Value(); value != nil {
			if err := c.checkExpr(value); err != nil {
				return err
			}
			// TODO: check that value.Type is assignable to o.TypeExpr().
		}
		return nil

	case a.KWhile:
		o := n.While()
		if cond := o.Condition(); cond != nil {
			if err := c.checkExpr(cond); err != nil {
				return err
			}
			if !cond.MType().Eq(TypeExprBoolean) {
				return fmt.Errorf("check: for-loop condition %q, of type %q, does not have a boolean type",
					cond.String(c.idMap), cond.MType().String(c.idMap))
			}
		}
		for _, m := range o.Body() {
			if err := c.checkStatement(m); err != nil {
				return err
			}
		}
		return nil

	case a.KIf:
		for o := n.If(); o != nil; o = o.ElseIf() {
			cond := o.Condition()
			if err := c.checkExpr(cond); err != nil {
				return err
			}
			if !cond.MType().Eq(TypeExprBoolean) {
				return fmt.Errorf("check: if condition %q, of type %q, does not have a boolean type",
					cond.String(c.idMap), cond.MType().String(c.idMap))
			}
			for _, m := range o.BodyIfTrue() {
				if err := c.checkStatement(m); err != nil {
					return err
				}
			}
			for _, m := range o.BodyIfFalse() {
				if err := c.checkStatement(m); err != nil {
					return err
				}
			}
		}
		return nil

	case a.KReturn:
		o := n.Return()
		if value := o.Value(); value != nil {
			if err := c.checkExpr(value); err != nil {
				return err
			}
			// TODO: type-check that value is assignable to the return value.
			// This needs the context of what func we're in.
		}
		return nil

	case a.KBreak:
		// TODO: check that we're in a for loop.
		return nil

	case a.KContinue:
		// TODO: check that we're in a for loop.
		return nil
	}

	return fmt.Errorf("check: unrecognized ast.Kind (%d) for checkStatement", n.Kind())
}

func (c *typeChecker) checkAssign(n *a.Assign) error {
	lhs := n.LHS()
	rhs := n.RHS()
	if err := c.checkExpr(lhs); err != nil {
		return err
	}
	if err := c.checkExpr(rhs); err != nil {
		return err
	}
	lTyp := lhs.MType()
	rTyp := rhs.MType()

	if n.Operator() == t.IDEq {
		if (rTyp == TypeExprIdealNumber && lTyp.IsNumType()) || lTyp.EqIgnoringRefinements(rTyp) {
			return nil
		}
		return fmt.Errorf("check: cannot assign %q of type %q to %q of type %q",
			rhs.String(c.idMap), rTyp.String(c.idMap), lhs.String(c.idMap), lTyp.String(c.idMap))
	}

	if !lTyp.IsNumType() {
		return fmt.Errorf("check: assignment %q: assignee %q, of type %q, does not have numeric type",
			c.idMap.ByID(n.Operator()), lhs.String(c.idMap), lTyp.String(c.idMap))
	}

	switch n.Operator() {
	case t.IDShiftLEq, t.IDShiftREq:
		if rTyp.IsNumType() {
			return nil
		}
		return fmt.Errorf("check: assignment %q: shift %q, of type %q, does not have numeric type",
			c.idMap.ByID(n.Operator()), rhs.String(c.idMap), rTyp.String(c.idMap))
	}

	if rTyp == TypeExprIdealNumber || lTyp.EqIgnoringRefinements(rTyp) {
		return nil
	}
	return fmt.Errorf("check: assignment %q: %q and %q, of types %q and %q, do not have compatible types",
		c.idMap.ByID(n.Operator()),
		lhs.String(c.idMap), rhs.String(c.idMap),
		lTyp.String(c.idMap), rTyp.String(c.idMap),
	)
}

func (c *typeChecker) checkExpr(n *a.Expr) error {
	switch n.ID0().Flags() & (t.FlagsUnaryOp | t.FlagsBinaryOp | t.FlagsAssociativeOp) {
	case 0:
		return c.checkExprOther(n)
	case t.FlagsUnaryOp:
		return c.checkExprUnaryOp(n)
	case t.FlagsBinaryOp:
		return c.checkExprBinaryOp(n)
	case t.FlagsAssociativeOp:
		return c.checkExprAssociativeOp(n)
	}
	return fmt.Errorf("check: unrecognized token.Key (0x%X) for checkExpr", n.ID0().Key())
}

func (c *typeChecker) checkExprOther(n *a.Expr) error {
	switch n.ID0() {
	case 0:
		switch id1 := n.ID1(); {
		case id1.IsNumLiteral():
			z := big.NewInt(0)
			s := c.idMap.ByID(id1)
			if _, ok := z.SetString(s, 0); !ok {
				return fmt.Errorf("check: invalid numeric literal %q", s)
			}
			n.SetConstValue(z)
			n.SetMType(TypeExprIdealNumber)
			return nil

		case id1.IsIdent():
			if c.typeMap != nil {
				if typ, ok := c.typeMap[id1]; ok {
					n.SetMType(typ)
					return nil
				}
			}
			// TODO: look for (global) names (constants, funcs, structs).
			return fmt.Errorf("check: unrecognized identifier %q", c.idMap.ByID(id1))

		case id1 == t.IDFalse:
			n.SetConstValue(zero)
			n.SetMType(TypeExprBoolean)
			return nil

		case id1 == t.IDTrue:
			n.SetConstValue(one)
			n.SetMType(TypeExprBoolean)
			return nil

		case id1 == t.IDUnderscore:
			// TODO.

		case id1 == t.IDThis:
			// TODO.
		}

	case t.IDOpenParen:
		// n is a function call.
		// TODO.

	case t.IDOpenBracket:
		// n is an index.
		// TODO.

	case t.IDColon:
		// n is a slice.
		// TODO.

	case t.IDDot:
		// n is a selector.
		// TODO.
	}
	return fmt.Errorf("check: unrecognized token.Key (0x%X) for checkExprOther", n.ID0().Key())
}

func (c *typeChecker) checkExprUnaryOp(n *a.Expr) error {
	rhs := n.RHS().Expr()
	if err := c.checkExpr(rhs); err != nil {
		return err
	}
	rTyp := rhs.MType()

	switch n.ID0() {
	case t.IDXUnaryPlus, t.IDXUnaryMinus:
		if !numeric(rTyp) {
			return fmt.Errorf("check: unary %q: %q, of type %q, does not have a numeric type",
				c.idMap.ByID(n.ID0().AmbiguousForm()), rhs.String(c.idMap), rTyp.String(c.idMap))
		}
		if cv := rhs.ConstValue(); cv != nil {
			if n.ID0() == t.IDXUnaryMinus {
				cv = big.NewInt(0).Neg(cv)
			}
			n.SetConstValue(cv)
		}
		n.SetMType(rTyp)
		return nil

	case t.IDXUnaryNot:
		if !rTyp.Eq(TypeExprBoolean) {
			return fmt.Errorf("check: unary %q: %q, of type %q, does not have a boolean type",
				c.idMap.ByID(n.ID0().AmbiguousForm()), rhs.String(c.idMap), rTyp.String(c.idMap))
		}
		if cv := rhs.ConstValue(); cv != nil {
			n.SetConstValue(btoi(cv.Cmp(zero) == 0))
		}
		n.SetMType(TypeExprBoolean)
		return nil
	}
	return fmt.Errorf("check: unrecognized token.Key (0x%X) for checkExprUnaryOp", n.ID0().Key())
}

func (c *typeChecker) checkExprBinaryOp(n *a.Expr) error {
	lhs := n.LHS().Expr()
	if err := c.checkExpr(lhs); err != nil {
		return err
	}
	lTyp := lhs.MType()
	op := n.ID0()
	if op == t.IDXBinaryAs {
		rhs := n.RHS().TypeExpr()
		if err := c.checkTypeExpr(rhs); err != nil {
			return err
		}
		if numeric(lTyp) && rhs.IsNumType() {
			n.SetMType(rhs)
			return nil
		}
		return fmt.Errorf("check: cannot convert expression %q, of type %q, as type %q",
			lhs.String(c.idMap), lTyp.String(c.idMap), rhs.String(c.idMap))
	}
	rhs := n.RHS().Expr()
	if err := c.checkExpr(rhs); err != nil {
		return err
	}
	rTyp := rhs.MType()

	switch op {
	default:
		if !numeric(lTyp) {
			return fmt.Errorf("check: binary %q: %q, of type %q, does not have a numeric type",
				c.idMap.ByID(op.AmbiguousForm()), lhs.String(c.idMap), lTyp.String(c.idMap))
		}
		if !numeric(rTyp) {
			return fmt.Errorf("check: binary %q: %q, of type %q, does not have a numeric type",
				c.idMap.ByID(op.AmbiguousForm()), rhs.String(c.idMap), rTyp.String(c.idMap))
		}
	case t.IDXBinaryNotEq, t.IDXBinaryEqEq:
		// No-op.
	case t.IDXBinaryAnd, t.IDXBinaryOr:
		if lTyp != TypeExprBoolean {
			return fmt.Errorf("check: binary %q: %q, of type %q, does not have a boolean type",
				c.idMap.ByID(op.AmbiguousForm()), lhs.String(c.idMap), lTyp.String(c.idMap))
		}
		if rTyp != TypeExprBoolean {
			return fmt.Errorf("check: binary %q: %q, of type %q, does not have a boolean type",
				c.idMap.ByID(op.AmbiguousForm()), rhs.String(c.idMap), rTyp.String(c.idMap))
		}
	}

	switch op {
	default:
		if !lTyp.EqIgnoringRefinements(rTyp) && lTyp != TypeExprIdealNumber && rTyp != TypeExprIdealNumber {
			return fmt.Errorf("check: binary %q: %q and %q, of types %q and %q, do not have compatible types",
				c.idMap.ByID(op.AmbiguousForm()),
				lhs.String(c.idMap), rhs.String(c.idMap),
				lTyp.String(c.idMap), rTyp.String(c.idMap),
			)
		}
	case t.IDXBinaryShiftL, t.IDXBinaryShiftR:
		if (lTyp == TypeExprIdealNumber) && (rTyp != TypeExprIdealNumber) {
			return fmt.Errorf("check: binary %q: %q and %q, of types %q and %q; "+
				"cannot shift an ideal number by a non-ideal number",
				c.idMap.ByID(op.AmbiguousForm()),
				lhs.String(c.idMap), rhs.String(c.idMap),
				lTyp.String(c.idMap), rTyp.String(c.idMap),
			)
		}
	}

	if l, r := lhs.ConstValue(), rhs.ConstValue(); l != nil && r != nil {
		if err := c.setConstValueBinaryOp(n, l, r); err != nil {
			return err
		}
	}

	if comparisonOps[0xFF&op.Key()] {
		n.SetMType(TypeExprBoolean)
	} else if lTyp != TypeExprIdealNumber {
		n.SetMType(lTyp)
	} else {
		n.SetMType(rTyp)
	}

	return nil
}

func (c *typeChecker) setConstValueBinaryOp(n *a.Expr, l *big.Int, r *big.Int) error {
	switch n.ID0() {
	case t.IDXBinaryPlus:
		n.SetConstValue(big.NewInt(0).Add(l, r))
	case t.IDXBinaryMinus:
		n.SetConstValue(big.NewInt(0).Sub(l, r))
	case t.IDXBinaryStar:
		n.SetConstValue(big.NewInt(0).Mul(l, r))
	case t.IDXBinarySlash:
		if r.Cmp(zero) == 0 {
			return fmt.Errorf("check: division by zero in const expression %q", n.String(c.idMap))
		}
		// TODO: decide on Euclidean division vs other definitions. See "go doc
		// math/big int.divmod" for details.
		n.SetConstValue(big.NewInt(0).Div(l, r))
	case t.IDXBinaryShiftL:
		if r.Cmp(zero) < 0 || r.Cmp(ffff) > 0 {
			return fmt.Errorf("check: shift %q out of range in const expression %q",
				n.RHS().Expr().String(c.idMap), n.String(c.idMap))
		}
		n.SetConstValue(big.NewInt(0).Lsh(l, uint(r.Uint64())))
	case t.IDXBinaryShiftR:
		if r.Cmp(zero) < 0 || r.Cmp(ffff) > 0 {
			return fmt.Errorf("check: shift %q out of range in const expression %q",
				n.RHS().Expr().String(c.idMap), n.String(c.idMap))
		}
		n.SetConstValue(big.NewInt(0).Rsh(l, uint(r.Uint64())))
	case t.IDXBinaryAmp:
		n.SetConstValue(big.NewInt(0).And(l, r))
	case t.IDXBinaryAmpHat:
		n.SetConstValue(big.NewInt(0).AndNot(l, r))
	case t.IDXBinaryPipe:
		n.SetConstValue(big.NewInt(0).Or(l, r))
	case t.IDXBinaryHat:
		n.SetConstValue(big.NewInt(0).Xor(l, r))
	case t.IDXBinaryNotEq:
		n.SetConstValue(btoi(l.Cmp(r) != 0))
	case t.IDXBinaryLessThan:
		n.SetConstValue(btoi(l.Cmp(r) < 0))
	case t.IDXBinaryLessEq:
		n.SetConstValue(btoi(l.Cmp(r) <= 0))
	case t.IDXBinaryEqEq:
		n.SetConstValue(btoi(l.Cmp(r) == 0))
	case t.IDXBinaryGreaterEq:
		n.SetConstValue(btoi(l.Cmp(r) >= 0))
	case t.IDXBinaryGreaterThan:
		n.SetConstValue(btoi(l.Cmp(r) > 0))
	case t.IDXBinaryAnd:
		n.SetConstValue(btoi((l.Cmp(zero) != 0) && (r.Cmp(zero) != 0)))
	case t.IDXBinaryOr:
		n.SetConstValue(btoi((l.Cmp(zero) != 0) || (r.Cmp(zero) != 0)))
	}
	return nil
}

func (c *typeChecker) checkExprAssociativeOp(n *a.Expr) error {
	// TODO.
	return fmt.Errorf("check: unrecognized token.Key (0x%X) for checkExprAssociativeOp", n.ID0().Key())
}

func (c *typeChecker) checkTypeExpr(n *a.TypeExpr) error {
	for ; n != nil; n = n.Inner() {
		switch n.PackageOrDecorator() {
		case 0:
			name := n.Name()
			if name.IsNumType() || name == t.IDBool {
				for _, bound := range n.Bounds() {
					if bound == nil {
						continue
					}
					if err := c.checkExpr(bound); err != nil {
						return err
					}
					if bound.ConstValue() == nil {
						return fmt.Errorf("check: %q is not constant", bound.String(c.idMap))
					}
				}
				return nil
			}
			if n.InclMin() != nil || n.ExclMax() != nil {
				// TODO: reject.
			}
			// TODO: see if name refers to a struct type.
			return fmt.Errorf("check: %q is not a type", c.idMap.ByID(name))

		case t.IDPtr:
			// No-op.

		case t.IDOpenBracket:
			aLen := n.ArrayLength()
			if err := c.checkExpr(aLen); err != nil {
				return err
			}
			if aLen.ConstValue() == nil {
				return fmt.Errorf("check: %q is not constant", aLen.String(c.idMap))
			}

		default:
			return fmt.Errorf("check: unrecognized node for checkTypeExpr")
		}
	}
	return nil
}

var comparisonOps = [256]bool{
	t.IDXBinaryNotEq >> t.KeyShift:       true,
	t.IDXBinaryLessThan >> t.KeyShift:    true,
	t.IDXBinaryLessEq >> t.KeyShift:      true,
	t.IDXBinaryEqEq >> t.KeyShift:        true,
	t.IDXBinaryGreaterEq >> t.KeyShift:   true,
	t.IDXBinaryGreaterThan >> t.KeyShift: true,
}
