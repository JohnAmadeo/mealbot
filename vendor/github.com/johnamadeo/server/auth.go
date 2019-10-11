package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
)

const (
	InvalidAccessToken = "Invalid access token"
	Issuer             = "https://intouch-android.auth0.com/"
	Audience           = "https://intouch-android-backend.herokuapp.com/"
	JSONWebKeySet      = "https://intouch-android.auth0.com/.well-known/jwks.json"
)

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

type Message struct {
	Message string
}

func GetAuthHandler(handler http.HandlerFunc) http.Handler {
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			checkAud := verifyAudience(token.Claims, Audience)

			if checkAud != nil {
				return token, errors.New("Invalid audience")
			}

			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(Issuer, true)
			if !checkIss {
				return token, errors.New("Invalid issuer")
			}

			cert, err := getPEMCertificate(token)
			if err != nil {
				panic(err.Error())
			}

			result, err := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
			if err != nil {
				return token, errors.New("Failed to parse RSA public key from PEM certificate")
			}

			return result, nil
		},

		// When set, the middleware verifies that tokens are signed with the specific signing algorithm
		// If the signing method is not constant the ValidationKeyGetter callback can be used to implement additional checks
		// Important to avoid security issues described here: https://auth0.com/blog/2015/03/31/critical-vulnerabilities-in-json-web-token-libraries/
		SigningMethod: jwt.SigningMethodRS256,
	})

	return jwtMiddleware.Handler(handler)
}

func GetFakeAuthHandler(handler http.HandlerFunc) http.Handler {
	return handler
}

// https://support.quovadisglobal.com/kb/a37/what-is-pem-format.aspx
func getPEMCertificate(token *jwt.Token) (string, error) {
	cert := ""

	resp, err := http.Get(JSONWebKeySet)
	if err != nil {
		return cert, err
	}
	defer resp.Body.Close()

	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)
	if err != nil {
		return cert, err
	}

	for k, _ := range jwks.Keys {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" +
				jwks.Keys[k].X5c[0] +
				"\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("Unable to find appropriate key")
		return cert, err
	}

	return cert, nil
}

// https://github.com/dgrijalva/jwt-go/issues/290
func verifyAudience(tokenClaims jwt.Claims, audience string) error {
	var claims map[string]interface{}
	claims, _ = tokenClaims.(jwt.MapClaims)

	if _, ok := claims["aud"]; !ok {
		return errors.New("No audience claim")
	}

	claimsMap, _ := claims["aud"].([]interface{})
	for _, item := range claimsMap {
		if item == audience {
			return nil
		}
	}

	return errors.New("Invalid audience")
}
