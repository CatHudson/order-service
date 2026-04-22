package util

import (
	sdecimal "github.com/shopspring/decimal"
	gdecimal "google.golang.org/genproto/googleapis/type/decimal"
	"google.golang.org/genproto/googleapis/type/money"
)

const nanosPerUnit = 1_000_000_000

// MoneyToDecimal converts google.type.Money to shopspring/decimal.
// The currency_code is returned separately — the caller decides what to do with it.
func MoneyToDecimal(m *money.Money) sdecimal.Decimal {
	units := sdecimal.NewFromInt(m.GetUnits())
	nanos := sdecimal.NewFromInt(int64(m.GetNanos())).
		Div(sdecimal.NewFromInt(nanosPerUnit))

	return units.Add(nanos)
}

// DecimalToMoney converts a shopspring/decimal back to google.type.Money.
func DecimalToMoney(d *sdecimal.Decimal) *money.Money {
	if d == nil {
		return nil
	}
	units := d.IntPart()
	nanos := d.Sub(sdecimal.NewFromInt(units)).
		Mul(sdecimal.NewFromInt(nanosPerUnit)).
		IntPart()

	return &money.Money{
		CurrencyCode: "USD",
		Units:        units,
		Nanos:        int32(nanos), //nolint: gosec // always fits into int32
	}
}

// DecimalFromProto converts google.type.Decimal to shopspring/decimal.
func DecimalFromProto(d *gdecimal.Decimal) sdecimal.Decimal {
	dec, err := sdecimal.NewFromString(d.GetValue())
	if err != nil {
		return sdecimal.Zero
	}
	return dec
}

// DecimalToProto converts shopspring/decimal to google.type.Decimal.
func DecimalToProto(d *sdecimal.Decimal) *gdecimal.Decimal {
	if d == nil {
		return nil
	}
	return &gdecimal.Decimal{
		Value: d.String(),
	}
}
