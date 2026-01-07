package google

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/nazarkurii/marshrutka_api/config"
	rfc7807 "github.com/nazarkurii/marshrutka_api/pkg/problem"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type UserInfoOAUTH struct {
	FirstName   string
	LastName    string
	Email       string
	DateOfBirth time.Time
}

func GetCredentialsByCode(code string, ctx context.Context, client *http.Client) (UserInfoOAUTH, error) {
	var cfg = oauth2.Config{
		ClientID:     config.GoogleClientID(),
		ClientSecret: config.GoogleSecretKey(),
		RedirectURL:  "postmessage",
		Endpoint:     google.Endpoint,
		Scopes:       []string{"email", "profile"},
	}

	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return UserInfoOAUTH{}, invalidCode(err)
	}

	clientOAUTH := &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(token),
			Base:   client.Transport,
		},
		Timeout: client.Timeout,
	}

	res, err := clientOAUTH.Get("https://people.googleapis.com/v1/people/me?personFields=birthdays,names,emailAddresses")
	if err != nil {
		return UserInfoOAUTH{}, badGateway(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return UserInfoOAUTH{}, rfc7807.Internal("Failed reading Google response", err.Error())
	}

	var parsed struct {
		Names []struct {
			FamilyName string `json:"familyName"`
			GivenName  string `json:"givenName"`
		} `json:"names"`
		EmailAddresses []struct {
			Value string `json:"value"`
		} `json:"emailAddresses"`
		Birthdays []struct {
			Date struct {
				Year  int `json:"year"`
				Month int `json:"month"`
				Day   int `json:"day"`
			} `json:"date"`
		} `json:"birthdays"`
	}

	err = json.Unmarshal(body, &parsed)
	if err != nil {
		return UserInfoOAUTH{}, rfc7807.Internal("Parsing Google Response Error", err.Error())
	}

	// Default safe values
	firstName, lastName, email := "", "", ""
	dob := time.Now().AddDate(-18, 0, 0) // fallback = assume 18 y/o

	if len(parsed.Names) > 0 {
		firstName = parsed.Names[0].GivenName
		lastName = parsed.Names[0].FamilyName
	}
	if len(parsed.EmailAddresses) > 0 {
		email = parsed.EmailAddresses[0].Value
	}
	if len(parsed.Birthdays) > 0 {
		d := parsed.Birthdays[0].Date
		dob = time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
	}

	return UserInfoOAUTH{
		FirstName:   firstName,
		LastName:    lastName,
		Email:       email,
		DateOfBirth: dob,
	}, nil
}
