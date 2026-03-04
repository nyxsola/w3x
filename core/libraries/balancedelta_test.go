package libraries

import (
	"math/big"
	"testing"
)

func TestNewBalanceDelta_Success(t *testing.T) {
	a0 := big.NewInt(100)
	a1 := big.NewInt(-200)

	delta := NewBalanceDelta(a0, a1)

	if delta.Amount0.Cmp(a0) != 0 || delta.Amount1.Cmp(a1) != 0 {
		t.Fatalf("values not set correctly")
	}
}

func TestBalanceDelta_Add_Success(t *testing.T) {
	a := NewBalanceDelta(big.NewInt(10), big.NewInt(20))
	b := NewBalanceDelta(big.NewInt(5), big.NewInt(-10))

	res := a.Add(b)

	expected0 := big.NewInt(15)
	expected1 := big.NewInt(10)

	if res.Amount0.Cmp(expected0) != 0 ||
		res.Amount1.Cmp(expected1) != 0 {
		t.Fatalf("add result incorrect")
	}
}

func TestBalanceDelta_Sub_Success(t *testing.T) {
	a := NewBalanceDelta(big.NewInt(20), big.NewInt(10))
	b := NewBalanceDelta(big.NewInt(5), big.NewInt(3))

	res := a.Sub(b)

	expected0 := big.NewInt(15)
	expected1 := big.NewInt(7)

	if res.Amount0.Cmp(expected0) != 0 ||
		res.Amount1.Cmp(expected1) != 0 {
		t.Fatalf("sub result incorrect")
	}
}

func TestBalanceDelta_Equal(t *testing.T) {
	a := NewBalanceDelta(big.NewInt(100), big.NewInt(200))
	b := NewBalanceDelta(big.NewInt(100), big.NewInt(200))
	c := NewBalanceDelta(big.NewInt(1), big.NewInt(2))

	if !a.Equal(b) {
		t.Fatalf("expected equal")
	}

	if a.Equal(c) {
		t.Fatalf("expected not equal")
	}
}
