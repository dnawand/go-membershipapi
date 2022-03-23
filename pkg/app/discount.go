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
	if voucher.Discount == "" {
		return price, nil
	}

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
	if voucher.Discount == "" {
		return tax, nil
	}

	if voucher.Type != domain.VoucherFixedAmount && voucher.Type != domain.VoucherPercentage {
		return domain.Money{}, &domain.ErrInvalidArgument{Msg: "invalid voucher type"}
	}

	if voucher.Type == domain.VoucherFixedAmount {
		v, err := fixedAmountToPercentage(price, tax, voucher)
		if err != nil {
			return domain.Money{}, fmt.Errorf("error when converting tax fixedAmount to percentage: %w", err)
		}
		voucher = v
	}

	return applyPercentage(tax, voucher)
}

func applyFixedAmount(money domain.Money, voucher domain.Voucher) (domain.Money, error) {
	priceAmount, err := currency.NewAmount(money.Number, string(money.Code))
	if err != nil {
		return domain.Money{}, &domain.ErrInvalidArgument{Msg: fmt.Sprintf("invalid Money values: %v", err)}
	}

	discountAmount, _ := currency.NewAmount(voucher.Discount, string(money.Code))
	if err != nil {
		return domain.Money{}, fmt.Errorf("error when calculating discount amount: %w", err)
	}
	priceAmountWithDiscount, _ := priceAmount.Sub(discountAmount)
	if err != nil {
		return domain.Money{}, fmt.Errorf("error when subtracting discount from price: %w", err)
	}

	return domain.Money{Code: money.Code, Number: formatCurrency(priceAmountWithDiscount).Number()}, nil
}

func applyPercentage(money domain.Money, voucher domain.Voucher) (domain.Money, error) {
	priceAmount, err := currency.NewAmount(money.Number, string(money.Code))
	if err != nil {
		return domain.Money{}, &domain.ErrInvalidArgument{Msg: fmt.Sprintf("invalid Money values: %v", err)}
	}

	discountPercentageAmount, err := currency.NewAmount(voucher.Discount, string(money.Code))
	if err != nil {
		return domain.Money{}, fmt.Errorf("error when getting percentage discount: %w", err)
	}
	fmt.Println(discountPercentageAmount.Number())

	pivot, err := discountPercentageAmount.Div("100")
	if err != nil {
		return domain.Money{}, fmt.Errorf("error when getting pivot value: %w", err)
	}

	discountAmount, err := priceAmount.Mul(pivot.Number())
	if err != nil {
		return domain.Money{}, fmt.Errorf("error calculating discount amount: %w", err)
	}

	priceAmountWithDiscount, err := priceAmount.Sub(discountAmount)
	if err != nil {
		return domain.Money{}, fmt.Errorf("error when subtracting discount from price: %w", err)
	}

	return domain.Money{Code: money.Code, Number: formatCurrency(priceAmountWithDiscount).Number()}, nil
}

func fixedAmountToPercentage(price domain.Money, tax domain.Money, voucher domain.Voucher) (domain.Voucher, error) {
	priceAmount, err := currency.NewAmount(price.Number, string(price.Code))
	if err != nil {
		return domain.Voucher{}, fmt.Errorf("invalid tax values: %w", err)
	}

	discountAmount, err := currency.NewAmount(voucher.Discount, string(price.Code))
	if err != nil {
		return domain.Voucher{}, fmt.Errorf("invalid tax values: %w", err)
	}

	pivot, err := discountAmount.Mul("100")
	if err != nil {
		return domain.Voucher{}, fmt.Errorf("error getting pivot value from discount amount: %w", err)
	}

	fixedAmountAsPercentage, err := pivot.Div(priceAmount.Number())
	if err != nil {
		return domain.Voucher{}, fmt.Errorf("error during pivot division: %w", err)
	}

	return domain.Voucher{
		ID:       voucher.ID,
		Type:     domain.VoucherPercentage,
		Discount: formatCurrency(fixedAmountAsPercentage).Number(),
		IsActive: voucher.IsActive,
	}, nil
}

func formatCurrency(ammout currency.Amount) currency.Amount {
	return ammout.RoundTo(2, currency.RoundHalfUp)
}
