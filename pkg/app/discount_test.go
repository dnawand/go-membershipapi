package app

import (
	"testing"

	"github.com/dnawand/go-membershipapi/pkg/domain"
	"github.com/stretchr/testify/assert"
)

func TestApplyDiscountOnPrice(t *testing.T) {
	discountService := NewDiscountService()

	t.Run("test voucher with fixed amount", func(t *testing.T) {
		price := domain.Money{
			Code:   domain.CurrencyEUR,
			Number: "100.00",
		}
		voucher := domain.Voucher{
			Type:     domain.VoucherFixedAmount,
			Discount: "5",
		}

		priceWithDiscount, err := discountService.ApplyDiscountOnPrice(price, voucher)
		assert.NoError(t, err)
		assert.Equal(t, "95.00", priceWithDiscount.Number)
	})

	t.Run("test voucher with percentage", func(t *testing.T) {
		price := domain.Money{
			Code:   domain.CurrencyEUR,
			Number: "115.00",
		}
		voucher := domain.Voucher{
			Type:     domain.VoucherPercentage,
			Discount: "8",
		}

		priceWithDiscount, err := discountService.ApplyDiscountOnPrice(price, voucher)
		assert.NoError(t, err)
		assert.Equal(t, "105.80", priceWithDiscount.Number)
	})

	t.Run("test invalid price Number value", func(t *testing.T) {
		price := domain.Money{
			Code:   domain.CurrencyEUR,
			Number: "115.00.",
		}
		voucher := domain.Voucher{
			Type:     domain.VoucherPercentage,
			Discount: "8",
		}

		_, err := discountService.ApplyDiscountOnPrice(price, voucher)
		assert.Error(t, err)

		var errInvalidArgument *domain.ErrInvalidArgument
		assert.ErrorAs(t, err, &errInvalidArgument)
	})
}

func TestApplyDiscountOnTax(t *testing.T) {
	discountService := NewDiscountService()

	t.Run("test voucher with fixed amount", func(t *testing.T) {
		price := domain.Money{
			Code:   domain.CurrencyEUR,
			Number: "100.00",
		}
		tax := domain.Money{
			Code:   domain.CurrencyEUR,
			Number: "40.00",
		}
		voucher := domain.Voucher{
			Type:     domain.VoucherFixedAmount,
			Discount: "7",
		}

		taxWithDiscount, err := discountService.ApplyDiscountOnTax(price, tax, voucher)
		assert.NoError(t, err)
		assert.Equal(t, "37.20", taxWithDiscount.Number)
	})

	t.Run("test voucher with percentage", func(t *testing.T) {
		price := domain.Money{
			Code:   domain.CurrencyEUR,
			Number: "100.00",
		}
		tax := domain.Money{
			Code:   domain.CurrencyEUR,
			Number: "40.00",
		}
		voucher := domain.Voucher{
			Type:     domain.VoucherFixedAmount,
			Discount: "19",
		}

		taxWithDiscount, err := discountService.ApplyDiscountOnTax(price, tax, voucher)
		assert.NoError(t, err)
		assert.Equal(t, "32.40", taxWithDiscount.Number)
	})

	t.Run("test invalid tax Number value", func(t *testing.T) {
		price := domain.Money{
			Code:   domain.CurrencyEUR,
			Number: "100.00",
		}
		tax := domain.Money{
			Code:   domain.CurrencyEUR,
			Number: "40.00.23",
		}
		voucher := domain.Voucher{
			Type:     domain.VoucherFixedAmount,
			Discount: "19",
		}

		_, err := discountService.ApplyDiscountOnTax(price, tax, voucher)
		assert.Error(t, err)

		var errInvalidArgument *domain.ErrInvalidArgument
		assert.ErrorAs(t, err, &errInvalidArgument)
	})
}
