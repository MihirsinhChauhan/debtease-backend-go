package service

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestNextUnsecuredEmiRatio(t *testing.T) {
	oldTotal := decimal.NewFromInt(10000)
	oldRatio := decimal.NewFromFloat(0.4)
	newEMI := decimal.NewFromInt(2000)

	// Unsecured new loan: (4000 + 2000) / 12000 = 0.5
	got := NextUnsecuredEmiRatio(oldTotal, oldRatio, newEMI, false)
	exp := decimal.NewFromFloat(0.5)
	if !got.Equal(exp) {
		t.Fatalf("unsecured loan: got %s want %s", got, exp)
	}

	// Secured new loan: (4000 + 0) / 12000
	got2 := NextUnsecuredEmiRatio(oldTotal, oldRatio, newEMI, true)
	exp2 := oldTotal.Mul(oldRatio).Div(oldTotal.Add(newEMI))
	if !got2.Equal(exp2) {
		t.Fatalf("secured loan: got %s want %s", got2, exp2)
	}
}
