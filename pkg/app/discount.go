package app

import (
	"fmt"

	"github.com/bojanz/currency"
	"github.com/dnawand/go-membershipapi/pkg/domain"
)

type DiscountService struct{}

func NewDiscountService() *DiscountService {
	return &DiscountService{}
}

func (ds *DiscountService) ApplyDiscountOnPrice(price domain.Money, voucher domain.Voucher) (domain.Money, error) {
	switch voucher.Type {
	case domain.VoucherFixedAmount:
		return applyFixedAmount(price, voucher)
	case domain.VoucherPercentage:
		return applyPercentage(price, voucher)
	default:
		return domain.Money{}, &domain.ErrInvalidArgument{Msg: "invalid voucher type"}
	}
}

func (ds *DiscountService) ApplyDiscountOnTax(price domain.Money, tax domain.Money, voucher domain.Voucher) (domain.Money, error) {
	switch voucher.Type {
	case domain.VoucherFixedAmount:
		voucherAsPercentage, _ := fixedAmountToPercentage(price, tax, voucher)
		return applyFixedAmount(price, voucherAsPercentage)
	case domain.VoucherPercentage:
		return applyPercentage(price, voucher)
	default:
		return domain.Money{}, &domain.ErrInvalidArgument{Msg: "invalid voucher type"}
	}
}

func applyFixedAmount(price domain.Money, voucher domain.Voucher) (domain.Money, error) {
	priceAmount, err := currency.NewAmount(price.Number, string(price.Code))
	if err != nil {
		return domain.Money{}, &domain.ErrInvalidArgument{Msg: fmt.Sprintf("invalid Money values: %w", err)}
	}

	discountAmount, _ := currency.NewAmount(voucher.Discount, string(price.Code))
	priceAmountWithDiscount, _ := priceAmount.Sub(discountAmount)

	return domain.Money{Code: price.Code, Number: priceAmountWithDiscount.Number()}, nil
}

func applyPercentage(price domain.Money, voucher domain.Voucher) (domain.Money, error) {
	priceAmount, err := currency.NewAmount(price.Number, string(price.Code))
	if err != nil {
		return domain.Money{}, &domain.ErrInvalidArgument{Msg: fmt.Sprintf("invalid Money values: %w", err)}
	}

	pivot, _ := currency.NewAmount(voucher.Discount, string(price.Code))
	pivot, _ = pivot.Div("100")
	discountAmount, _ := priceAmount.Mul(pivot.Round().Number())
	priceAmountWithDiscount, _ := priceAmount.Sub(discountAmount)

	return domain.Money{Code: price.Code, Number: priceAmountWithDiscount.Number()}, nil
}

func fixedAmountToPercentage(price domain.Money, tax domain.Money, voucher domain.Voucher) (domain.Voucher, error) {
	priceAmount, _ := currency.NewAmount(price.Number, string(price.Code))
	discountAmount, _ := currency.NewAmount(string(voucher.Discount), string(price.Code))
	pivot, _ := discountAmount.Mul("100")
	fixedAmountAsPercentage, _ := pivot.Div(priceAmount.Number())

	return domain.Voucher{
		ID:       voucher.ID,
		Type:     domain.VoucherPercentage,
		Discount: fixedAmountAsPercentage.Round().Number(),
		IsActive: voucher.IsActive,
	}, nil
}
