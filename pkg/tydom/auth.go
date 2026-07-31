package tydom

import (
	"crypto/tls"
	"net/http"

	"github.com/icholy/digest"
)

func (tc *Client) Authenticate(password string) (string, error) {
	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, // #nosec G402
		Timeout:   httpRequestTimeout,
	}
	newuri := tc.TargetHost
	newuri.Scheme = "https"

	res, err := client.Get(newuri.String())
	if err != nil {
		return "", &Error{Msg: "error connecting to tydom", Err: err}
	}

	defer res.Body.Close()
	header := res.Header.Get("WWW-Authenticate")
	chal, err := digest.ParseChallenge(header)
	if err != nil {
		return "", &Error{Msg: "failed to parse digest challenge", Err: err}
	}

	// use it to create credentials for the next request
	cred, err := digest.Digest(chal, digest.Options{
		Username: tc.GatewayID,
		Password: password,
		Method:   "GET",
		URI:      res.Request.URL.RequestURI(),
		Count:    1,
	})
	if err != nil {
		return "", &Error{Msg: "error generate auth digest", Err: err}
	}

	tc.Logger.Debug("creds", "data", cred.String())
	return cred.String(), nil
}
