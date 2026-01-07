package service

import (
	"context"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/nazarkurii/marshrutka_api/config"
	"github.com/nazarkurii/marshrutka_api/internal/domain/parcel/repo"
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

type Parcel interface {
	FindConnections(ctx context.Context, request entity.FindParcelConnectionsRequest) ([]entity.ConnectionParcel, error)
	Purchase(ctx context.Context, userID uuid.UUID, connectionID string, newParcel entity.PurchaseParcelRequest) (string, error)
	PurchaseSucceded(ctx context.Context, sessionID, token string) error
	PurchaseFailed(ctx context.Context, sessionID, token string) error
	GetParcels(ctx context.Context, paginationStr dbutil.PaginationStr, userID uuid.UUID) ([]entity.CustomerParcel, hypermedia.Links, error)
	GetConnectionByID(ctx context.Context, idStr, widthStr, heightStr, lengthStr string) (entity.CustomerConnection, error)
}

type serviceImpl struct {
	repo   repo.Parcel
	client *http.Client
}

func (s *serviceImpl) fetchConnections(ctx context.Context, req entity.FindParcelConnectionsRequestParsed) ([]entity.Connection, error) {
	connections, err := s.repo.GetConnectionsByMonth(ctx, req.From, req.To, int(req.Month), req.Year)
	if err != nil {
		return nil, err
	}
	if len(connections) < 1 {
		return nil, rfc7807.BadRequest(
			"calendar-unavailable-yet",
			"Calendar Unavailable Yet Error",
			"There has been no planning performed for this future.",
		)
	}
	return connections, nil
}

func pickBestConnectionsPerDay(connections []entity.Connection) map[int]entity.Connection {
	bestPerDay := make(map[int]entity.Connection)
	for _, c := range connections {
		day := config.MustParseToLocalByUUID(c.DepartureTime, c.DepartureCountryID).Day()
		if existing, ok := bestPerDay[day]; !ok || existing.LuggageVolumeLeft < c.LuggageVolumeLeft {
			bestPerDay[day] = c
		}
	}
	return bestPerDay
}

func buildCurrentMonth(req entity.FindParcelConnectionsRequestParsed, bestPerDay map[int]entity.Connection) []entity.ConnectionParcel {
	daysCount := daysIn(req.Month, req.Year)
	connections := make([]entity.ConnectionParcel, daysCount)

	for i := 0; i < daysCount; i++ {
		dayNum := i + 1
		date := time.Date(req.Year, req.Month, dayNum, 0, 0, 0, 0, config.MustGetLocationFromCountryID(req.From))
		weekday := normalizeWeekday(date.Weekday())

		if value, ok := bestPerDay[dayNum]; ok {
			connections[i] = value.ToParcelConnection(true, weekday, dayNum, true, int(config.CalulateParcelPrice(uint(req.Width), uint(req.Height), uint(req.Length))))

		} else {
			connections[i] = entity.ConnectionParcel{
				Usable:         false,
				DayNumber:      weekday,
				DayMonth:       dayNum,
				IsCurrentMonth: true,
			}
		}
	}
	return connections
}

func fillPreviousMonth(connections []entity.ConnectionParcel, req entity.FindParcelConnectionsRequestParsed) []entity.ConnectionParcel {
	firstDayNumber := connections[0].DayNumber
	if firstDayNumber <= 1 {
		return connections
	}

	firstDate := time.Date(req.Year, req.Month, 1, 0, 0, 0, 0, config.MustGetLocationFromCountryID(req.From))
	daysToAdd := firstDayNumber - 1
	monthBefore := make([]entity.ConnectionParcel, daysToAdd)

	for i := daysToAdd; i >= 1; i-- {
		dateAdded := firstDate.Add(-time.Duration(24*i) * time.Hour)
		weekday := normalizeWeekday(dateAdded.Weekday())

		monthBefore[daysToAdd-i] = entity.ConnectionParcel{
			Usable:         false,
			DayNumber:      weekday,
			DayMonth:       dateAdded.Day(),
			IsCurrentMonth: false,
		}
	}

	return append(monthBefore, connections...)
}

func fillNextMonth(connections []entity.ConnectionParcel, req entity.FindParcelConnectionsRequestParsed) []entity.ConnectionParcel {
	lastDayNumber := connections[len(connections)-1].DayNumber
	if lastDayNumber >= 7 {
		return connections
	}

	lastDate := time.Date(req.Year, req.Month, daysIn(req.Month, req.Year), 0, 0, 0, 0, config.MustGetLocationFromCountryID(req.From))
	daysToAdd := 7 - lastDayNumber
	monthAfter := make([]entity.ConnectionParcel, daysToAdd)

	for i := 1; i <= daysToAdd; i++ {
		dateAdded := lastDate.Add(time.Duration(24*i) * time.Hour)
		weekday := normalizeWeekday(dateAdded.Weekday())

		monthAfter[i-1] = entity.ConnectionParcel{
			Usable:         false,
			DayNumber:      weekday,
			DayMonth:       dateAdded.Day(),
			IsCurrentMonth: false, // optional: add IsNextMonth if your struct supports
		}
	}

	return append(connections, monthAfter...)
}

func (s *serviceImpl) FindConnections(ctx context.Context, requestUnparsed entity.FindParcelConnectionsRequest) ([]entity.ConnectionParcel, error) {
	req, err := requestUnparsed.Parse()
	if err != nil {
		return nil, err
	}

	connections, err := s.fetchConnections(ctx, req)
	if err != nil {
		return nil, err
	}

	bestPerDay := pickBestConnectionsPerDay(connections)
	connectionsMonth := buildCurrentMonth(req, bestPerDay)
	connectionsMonth = fillPreviousMonth(connectionsMonth, req)
	connectionsMonth = fillNextMonth(connectionsMonth, req)

	return connectionsMonth, nil
}

func normalizeWeekday(weekday time.Weekday) int {
	if weekday == 0 {
		return 7
	}
	return int(weekday)
}

func daysIn(month time.Month, year int) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func (s *serviceImpl) GetParcels(ctx context.Context, paginationStr dbutil.PaginationStr, userID uuid.UUID) ([]entity.CustomerParcel, hypermedia.Links, error) {
	pagination, err := paginationStr.ParseWithCondition(dbutil.Condition{"user_id = ? AND (SELECT succeeded FROM parcel_payments WHERE parcel_id = `parcels`.id)", []any{userID}}, []string{}, "created_at")
	if err != nil {

		return nil, nil, err
	}

	parcels, connections, total, err, empty := s.repo.GetParcels(ctx, pagination)
	if err != nil || empty {
		return nil, nil, err
	}

	var respose = make([]entity.CustomerParcel, len(parcels))

	for i, parcel := range parcels {
		connectionIndex := slices.IndexFunc(connections, func(connection entity.Connection) bool {
			return connection.ID == parcel.ConnectionID
		})

		if connectionIndex == -1 {
			return nil, nil, rfc7807.DB("internal")
		}

		respose[i] = entity.CustomerParcel{
			Parcel:     parcel,
			Connection: connections[connectionIndex].ToCustomer(nil, 0, 0, 0),
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
	err = s.repo.RemoveParcelStops(ctx, sessionID)
	if err != nil {

		return err
	}

	err = s.repo.DeleteParcels(ctx, sessionID)
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

func (s *serviceImpl) Purchase(ctx context.Context, userID uuid.UUID, connectionID string, newParcel entity.PurchaseParcelRequest) (string, error) {
	req, params := newParcel.Parse(connectionID)
	if params != nil {
		return "", rfc7807.BadRequest("invalid-request", "Invalid Request Error", "The request is not valid.", params...)
	}

	connection, err := s.repo.GetConnectionByID(ctx, req.ConnectionID)
	if err != nil {
		return "", err
	}

	if connection.LuggageVolumeLeft < uint(req.Height*req.Length*req.Width) {
		return "", rfc7807.New(http.StatusConflict, "too-big-lugage-volume", "Too big Luggage Volume Error", "Provided luggage params makes volume that exceeds the remainig.")
	}

	pickUpAdress := req.PickUpAdress.ToAddress(connection.DepartureCountryID)
	params = pickUpAdress.Validate()
	dropOffAdress := req.DropOffAdress.ToAddress(connection.DepartureCountryID)
	params = dropOffAdress.Validate()
	if params != nil {
		return "", rfc7807.BadRequest("invalid-request", "Invalid Request Error", "The request is not valid.", params...)
	}

	if err := pickUpAdress.Prepare(ctx, s.client); err != nil {
		return "", err
	}

	if err := dropOffAdress.Prepare(ctx, s.client); err != nil {
		return "", err
	}

	token, err := auth.GenerateAccessToken(config.PaymentSecretKey(), jwt.MapClaims{
		"expires": time.Now().Add(time.Minute * 15).Unix(),
	})
	if err != nil {
		return "", err
	}

	redirectURL, sessionID, err := stripe.CreateStripeCheckoutSession(int64(config.CalulateParcelPrice(uint(req.Height), uint(req.Length), uint(req.Width))), "/connection/purchase-parcel", token)
	if err != nil {
		return "", rfc7807.BadGateway("payment", "Payment Error", err.Error())
	}

	parcelID := uuid.New()

	qrCode, err := qrcode.Encode(parcelID.String(), qrcode.Highest, 256)
	if err != nil {
		return "", rfc7807.Internal("QR-Code Encoding Error", err.Error())
	}

	parcel := entity.Parcel{
		ID:                  parcelID,
		UserID:              userID,
		ConnectionID:        connection.ID,
		SenderPhoneNumber:   req.Sender.PhoneNumber,
		SenderEmail:         req.Sender.Email,
		RecieverPhoneNumber: req.Recievier.PhoneNumber,
		RecieverEmail:       req.Recievier.Email,

		SenderName:        req.Sender.FirstName,
		SenderLastName:    req.Sender.LastName,
		RecieverFirstName: req.Recievier.FirstName,
		RecieverLastName:  req.Recievier.LastName,
		PickUpAdressID:    pickUpAdress.ID,
		PickUpAdress:      pickUpAdress,
		DropOffAdressID:   dropOffAdress.ID,
		DropOffAdress:     dropOffAdress,
		Payment: entity.ParcelPayment{
			ParcelID:  parcelID,
			Price:     int(config.CalulateParcelPrice(uint(req.Height), uint(req.Length), uint(req.Width))),
			Method:    entity.PaymentMethodCard,
			SessionID: sessionID,
		},
		LuggageVolume: uint(req.Height * req.Length * req.Width),
		Width:         req.Width,
		Height:        req.Height,
		Length:        req.Length,
		QRCode:        qrCode,
		Weight:        req.Weight,
		Type:          req.Type,
	}

	err = s.repo.Create(ctx, &parcel)
	if err != nil {
		return "", err
	}

	err = s.repo.CreateParcelStops(ctx, sessionID)
	if err != nil {
		return "", err
	}
	return redirectURL, nil
}

func (c *serviceImpl) GetConnectionByID(ctx context.Context, idStr, widthStr, heightStr, lengthStr string) (entity.CustomerConnection, error) {

	id, err := uuid.Parse(idStr)
	if err != nil {
		return entity.CustomerConnection{}, rfc7807.BadRequest("parsing", "Parsing Error", err.Error())
	}
	width, err := strconv.Atoi(widthStr)
	if err != nil {
		return entity.CustomerConnection{}, rfc7807.BadRequest("parsing", "Parsing Error", err.Error())
	}
	height, err := strconv.Atoi(heightStr)
	if err != nil {
		return entity.CustomerConnection{}, rfc7807.BadRequest("parsing", "Parsing Error", err.Error())
	}
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return entity.CustomerConnection{}, rfc7807.BadRequest("parsing", "Parsing Error", err.Error())
	}

	connetion, err := c.repo.GetConnectionByID(ctx, id)
	if err != nil {
		return entity.CustomerConnection{}, err
	}

	customeConnection := connetion.ToCustomer(nil, 0, 0, 0)
	customeConnection.Price = int(config.CalulateParcelPrice(uint(width), uint(height), uint(length)))
	return customeConnection, nil
}

func NewParcelService(repo repo.Parcel, client *http.Client) Parcel {
	return &serviceImpl{
		repo,
		client,
	}
}
