package google

import (
	"context"
	"encoding/json"
	rfc7807 "maryan_api/pkg/problem"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type UserInfoOAUTH struct {
	Name    string `json:"name"`
	SurName string `json:"surName"`
	Email   string `json:"email"`
}

var config = oauth2.Config{
	ClientID:     "953257087625-dm82cn9b20a19526g33cmu1di1q34nju.apps.googleusercontent.com",
	ClientSecret: "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",
	RedirectURL:  "http://localhost:3000",
	Endpoint:     google.Endpoint,
	Scopes:       []string{"profile", "email"},
}

func GetCredentialsByCode(code string, ctx context.Context, client *http.Client) (UserInfoOAUTH, error) {
	token, err := config.Exchange(ctx, code)

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

	res, err := clientOAUTH.Get("https://www.googleapis.com/oauth2/v3/userinfo")

	if err != nil {
		return UserInfoOAUTH{}, badGateway(err)
	}

	defer res.Body.Close()

	var credentials UserInfoOAUTH

	err = json.NewDecoder(res.Body).Decode(&credentials)

	if err != nil {
		return UserInfoOAUTH{}, rfc7807.Internal("Parsing Google Response Error", err.Error())
	}

	return credentials, nil
}
