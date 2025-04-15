package stripe

import (
	"maryan_api/config"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

func InitStripe() {

	stripe.Key = config.StripSekretKey()
}

func CreateStripeCheckoutSession(amount int64) (string, string, error) {
	params := &stripe.CheckoutSessionParams{
		Mode:       stripe.String("payment"),
		SuccessURL: stripe.String("http://localhost:8080/connection/purchase-ticket/succeded/{CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String("http://localhost:8080/connection/purchase-ticket/failed/{CHECKOUT_SESSION_ID}"),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("eur"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Ticket"),
					},
					UnitAmount: stripe.Int64(amount),
				},
				Quantity: stripe.Int64(1),
			},
		},
	}

	s, err := session.New(params)
	if err != nil {
		return "", "", err
	}

	return s.URL, s.ID, nil
}
