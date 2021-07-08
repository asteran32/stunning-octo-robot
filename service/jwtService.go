package service

import (
	"app/redis"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/twinj/uuid"
)

//this should be in an env file
var SECRET_KEY = "MySecretKey1$1$234" // 나중에 .env, .yml 변경

func GenerateJWT(id string) (map[string]string, error) {
	issueAt := time.Now()
	atUUID := uuid.NewV4().String()
	atExpireAt := time.Now().Add(time.Minute * 15)

	rtUUID := uuid.NewV4().String()
	rtExpireAt := time.Now().Add(time.Hour * 24 * 7)

	// Generating Access Token
	atClaims := jwt.MapClaims{}
	atClaims["Id"] = id
	atClaims["IssuerAt"] = issueAt.Unix()
	atClaims["ExpiresAt"] = atExpireAt.Unix()
	atClaims["Access_uuid"] = atUUID

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	accessToken, err := at.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return nil, err
	}

	// Generating Refresh Token
	rtClaims := jwt.MapClaims{}
	rtClaims["Id"] = id
	rtClaims["IssuerAt"] = issueAt.Unix()
	rtClaims["ExpiresAt"] = rtExpireAt.Unix()
	rtClaims["Refresh_uuid"] = rtUUID

	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	refreshToken, err := rt.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return nil, err
	}

	tokens := map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	// Redis Set Key
	errAccess := redis.SetKey(atUUID, id, atExpireAt.Sub(issueAt))
	if errAccess != nil {
		return nil, errAccess
	}

	errRefresh := redis.SetKey(rtUUID, id, rtExpireAt.Sub(issueAt))
	if errRefresh != nil {
		return nil, errRefresh
	}

	return tokens, nil
}

func extractedMetadata(tokenStr string, ty string) (map[string]string, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil { // Faild parsing token
		return nil, err
	}
	// Check token is vaild or not
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return nil, err // invaild
	}

	var details map[string]string
	switch ty {
	case "access_token":
		uuid, ok := claims["Access_uuid"].(string)
		if !ok {
			return nil, fmt.Errorf("Invaild uuid")
		}
		id, ok := claims["Id"].(string)
		if !ok {
			return nil, fmt.Errorf("Invaild user ID")
		}
		details = map[string]string{
			"Access_uuid": uuid,
			"Id":          id,
		}

	case "refresh_token":
		uuid, ok := claims["Refresh_uuid"].(string)
		if !ok {
			return nil, fmt.Errorf("Invaild uuid")
		}
		id, ok := claims["Id"].(string)
		if !ok {
			return nil, fmt.Errorf("Invaild user ID")
		}
		details = map[string]string{
			"Refresh_uuid": uuid,
			"Id":           id,
		}
	}

	return details, nil
}

func VerifyAccessToken(tokenStr string) error {
	details, err := extractedMetadata(tokenStr, "access_token")
	if err != nil {
		return err
	}
	_, err = redis.GetKey(details["Access_uuid"])
	if err != nil {
		return err
	}

	return nil
}

func DeleteAccessToken(tokenStr string) error {
	details, err := extractedMetadata(tokenStr, "access_token")
	if err != nil {
		return err
	}
	_, err = redis.Deletekey(details["Access_uuid"])
	if err != nil {
		return err
	}

	return nil
}

func VerifyRefreshToken(tokenStr string) (map[string]string, error) {
	details, err := extractedMetadata(tokenStr, "refresh_token")
	if err != nil {
		return nil, err
	}
	// Delete the previous Refresh Token
	_, err = redis.Deletekey(details["Refresh_uuid"])
	if err != nil {
		return nil, err
	}
	// Create new pairs of refresh and access tokens and Set the tokens metadata to redis
	tokens, err := GenerateJWT(details["Id"])
	if err != nil {
		return nil, err
	}
	return tokens, nil
}
