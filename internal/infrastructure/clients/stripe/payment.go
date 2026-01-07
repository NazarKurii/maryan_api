package stripe

import (
	"github.com/nazarkurii/marshrutka_api/config"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

func InitStripe() {
	stripe.Key = config.StripSekretKey()
}

func CreateStripeCheckoutSession(amount int64, base, token string) (string, string, error) {
	params := &stripe.CheckoutSessionParams{
		Mode:       stripe.String("payment"),
		SuccessURL: stripe.String(config.APIURL() + base + "/succeded/{CHECKOUT_SESSION_ID}/" + token),
		CancelURL:  stripe.String(config.APIURL() + base + "/failed/{CHECKOUT_SESSION_ID}/" + token),

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

func CancelPaymentIntent(sessionID string) error {
	_, err := session.Expire(sessionID, nil)
	return err
}
