package service

import (
	"context"
	"net/http"
	"slices"
	"time"

	"github.com/nazarkurii/marshrutka_api/config"
	"github.com/nazarkurii/marshrutka_api/internal/domain/tickets/repo"
	"github.com/nazarkurii/marshrutka_api/internal/entity"
	"github.com/nazarkurii/marshrutka_api/internal/infrastructure/clients/stripe"
	"github.com/nazarkurii/marshrutka_api/pkg/auth"
	"github.com/nazarkurii/marshrutka_api/pkg/dbutil"
	"github.com/nazarkurii/marshrutka_api/pkg/hypermedia"
	rfc7807 "github.com/nazarkurii/marshrutka_api/pkg/problem"

	"github.com/d3code/uuid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/skip2/go-qrcode"
)

type Ticket interface {
	Purchase(ctx context.Context, userID uuid.UUID, newTicket entity.NewTicketJSON) (string, error)
	PurchaseFailed(ctx context.Context, sessionID, token string) error
	PurchaseSucceded(ctx context.Context, sessionID, token string) error
	GetTickets(ctx context.Context, paginationStr dbutil.PaginationStr, userID uuid.UUID) ([]entity.CustomerTicket, hypermedia.Links, error)
}

type serviceImpl struct {
	repo   repo.Ticket
	client *http.Client
}

func (s *serviceImpl) GetTickets(ctx context.Context, paginationStr dbutil.PaginationStr, userID uuid.UUID) ([]entity.CustomerTicket, hypermedia.Links, error) {
	pagination, err := paginationStr.ParseWithCondition(dbutil.Condition{"user_id = ? AND (SELECT succeeded FROM ticket_payments WHERE ticket_id = `tickets`.id)", []any{userID}}, []string{}, "created_at")
	if err != nil {

		return nil, nil, err
	}

	tickets, connections, total, err, empty := s.repo.GetTickets(ctx, pagination)
	if err != nil || empty {
		return nil, nil, err
	}

	var respose = make([]entity.CustomerTicket, len(tickets))

	for i, ticket := range tickets {
		connectionIndex := slices.IndexFunc(connections, func(connection entity.Connection) bool {
			return connection.ID == ticket.ConnectionID
		})

		if connectionIndex == -1 {
			return nil, nil, rfc7807.DB("internal")
		}

		respose[i] = entity.CustomerTicket{
			Ticket:     ticket,
			Connection: connections[connectionIndex].Simplify(),
			Expired:    connections[connectionIndex].DepartureTime.Before(time.Now().UTC()),
		}
	}

	return respose, hypermedia.Pagination(paginationStr, total), nil
}

func (s *serviceImpl) PurchaseFailed(ctx context.Context, sessionID, token string) error {
	_, err := auth.VerifyAccessToken(token, config.PaymentSecretKey())
	if err != nil {
		return rfc7807.Unauthorized("unauthorized", "Unauthorized Error", "Unauthorized Error")
	}
	err = stripe.CancelPaymentIntent(sessionID)
	if err != nil {
		return rfc7807.BadGateway("payment-cancelation",
			"Payment Cancelation Error", err.Error())
	}
	err = s.repo.RemovePassengerStops(ctx, sessionID)
	if err != nil {
		return err
	}
	return nil
}

func (s *serviceImpl) PurchaseSucceded(ctx context.Context, sessionID, token string) error {
	_, err := auth.VerifyAccessToken(token, config.PaymentSecretKey())
	if err != nil {
		return rfc7807.Unauthorized("unauthorized", "Unauthorized Error", "Unauthorized Error")
	}
	return s.repo.PaymentSucceeded(ctx, sessionID)
}

func (s *serviceImpl) Purchase(ctx context.Context, userID uuid.UUID, newTicket entity.NewTicketJSON) (string, error) {
	email, phoneNumber, err := newTicket.ParseContaanctInfo()

	connection, takenSeats, err := s.repo.GetConnectionByID(ctx, newTicket.ConnectionID, len(newTicket.Passengers))
	if err != nil {
		return "", err
	}

	ticketID := uuid.New()

	seats, err := newTicket.Validate(connection, takenSeats, ticketID, connection.LuggageVolumeLeft)
	if err != nil {
		return "", err
	}

	pickUpAdress, dropOffAdress, err := newTicket.ParseAdresses(ctx, s.client, connection.DepartureCountryID, connection.DestinationCountryID)
	if err != nil {
		return "", err
	}

	passengers, err := newTicket.ParsePassengers(ticketID)
	if err != nil {
		return "", err
	}

	token, err := auth.GenerateAccessToken(config.PaymentSecretKey(), jwt.MapClaims{
		"expires": time.Now().Add(time.Minute * 15).Unix(),
	})
	if err != nil {
		return "", err
	}

	redirectURL, sessionID, err := stripe.CreateStripeCheckoutSession((int64(connection.Price)*int64(len(newTicket.SeatIDs)))+int64(newTicket.LuggagePrice()), "/connection/purchase-ticket", token)
	if err != nil {
		return "", rfc7807.BadGateway("payment", "Payment Error", err.Error())
	}

	qrCode, err := qrcode.Encode(ticketID.String(), qrcode.Highest, 256)
	if err != nil {
		return "", rfc7807.Internal("QR-Code Encoding Error", err.Error())
	}

	ticket := &entity.Ticket{
		ID:              ticketID,
		UserID:          userID,
		PhoneNumber:     phoneNumber,
		Email:           email,
		ConnectionID:    connection.ID,
		Seats:           seats,
		Passengers:      passengers,
		PickUpAdressID:  pickUpAdress.ID,
		PickUpAdress:    *pickUpAdress,
		DropOffAdressID: dropOffAdress.ID,
		DropOffAdress:   *dropOffAdress,
		Payment: entity.TicketPayment{
			TicketID:  ticketID,
			Price:     connection.Price + newTicket.LuggagePrice(),
			Method:    entity.PaymentMethodCard,
			SessionID: sessionID,
			Succeeded: false,
		},
		LuggageVolume: newTicket.LuggageVolume(),
		QRCode:        qrCode,
	}

	err = s.repo.SaveTicket(ctx, ticket)
	if err != nil {
		return "", err
	}

	err = s.repo.CreatePassengerStops(ctx, sessionID)
	if err != nil {
		return "", err
	}
	return redirectURL, nil
}

func NewTicketService(repo repo.Ticket, client *http.Client) Ticket {
	return &serviceImpl{
		repo,
		client,
	}
}
