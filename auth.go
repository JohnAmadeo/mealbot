package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

const (
	InvalidAccessToken = "Invalid access token"
	Issuer             = "https://mealbot.auth0.com/"
	Audience           = "https://mealbot-2.herokuapp.com/"
	JSONWebKeySet      = "https://mealbot.auth0.com/.well-known/jwks.json"
)

type CustomJWTMiddleware struct {
	ValidationKeyGetter jwt.Keyfunc
	SigningMethod       jwt.SigningMethod
}

func (mw *CustomJWTMiddleware) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := mw.CheckJWT(w, r)
		if err != nil {
			fmt.Println(err)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (mw *CustomJWTMiddleware) CheckJWT(
	w http.ResponseWriter,
	r *http.Request,
) error {
	// preflight request
	if r.Method == "OPTIONS" {
		return nil
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return errors.New("No authorization header")
	}

	authStr := strings.Split(authHeader, " ")
	if len(authStr) != 2 || strings.ToLower(authStr[0]) != "bearer" {
		return errors.New("Authorization header format must be Bearer {token}")
	}

	token := authStr[1]
	parsedToken, err := jwt.Parse(token, mw.ValidationKeyGetter)
	if err != nil {
		return err
	}

	if !parsedToken.Valid {
		return errors.New("Token is invalid")
	}

	if mw.SigningMethod.Alg() != parsedToken.Header["alg"] {
		return errors.New("Token must use 'alg' signing method")
	}

	return nil
}

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

func GetFakeAuthHandler(handler http.Handler) http.Handler {
	return handler
}

func GetAuthHandler(handler http.Handler) http.Handler {
	mw := &CustomJWTMiddleware{
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
				return token, errors.New("PEM Certificate failed")
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
	}

	return mw.Handler(handler)
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
