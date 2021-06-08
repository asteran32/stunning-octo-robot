package service

import (
	"app/db"
	"app/model"
	"app/redis"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/twinj/uuid"
)

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

type AccessDetails struct {
	AccessUuid string
	UserEmail  string
}

var SECRET_KEY = "MySecretKey1$1$234"

func SignIn(c *gin.Context) {
	var user model.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err = db.UserSignIn(user.Email, user.Password)
	if err != nil {
		if err == db.ErrINVALIDEMAIL {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()}) //Err:403
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) //Err:500
		return
	}

	// Create JWT token
	ts, err := GenerateJWT(user.Email)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	// Redis
	saveErr := CreateAuth(user.Email, ts)
	if saveErr != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": saveErr.Error()})
	}

	tokens := map[string]string{
		"accessToken":  ts.AccessToken,
		"refreshToken": ts.RefreshToken,
	}

	c.JSON(http.StatusOK, tokens)

}

// SignUp is resgister function
func SignUp(c *gin.Context) {
	var user model.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = db.UserSignUp(user)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func VerifyToken(c *gin.Context) {
	// 1. Get access Token from header and parse
	clientToken := c.GetHeader("Authorization")
	if clientToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "message": "Authorization Token is required"})
		// c.Abort()
		return
	}

	extracted := strings.Split(clientToken, "Bearer ")
	if len(extracted) == 2 {
		clientToken = strings.TrimSpace(extracted[1])
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Incorrect Format of Authorization Token "})
		// c.Abort()
		return
	}

	claims := jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(clientToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			log.Println("Invalid Token Signature")
			c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "message": "Invalid Token Signature"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Bad Request"})
		return
	}
	if !parsedToken.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "message": "Invalid Token"})
		return
	}

	// 2. Check token vaild
	// claims, ok := token.Claims.(jwt.MapClaims)
	// if !ok && !token.Valid {
	// 	log.Println("invaild token claims")
	// 	c.JSON(http.StatusUnprocessableEntity, "invalid uuid")
	// 	return
	// }

	// 3. Redis - Extract Meta data for find Reids Redis 저장소에서 조회할 토큰 메타데이터를 추출
	uuid, ok := claims["access_uuid"].(string)
	if !ok {
		log.Println("invaild uuid token claims")
		c.JSON(http.StatusUnprocessableEntity, "invalid uuid")
		return
	}

	email, ok := claims["user_id"].(string)
	if !ok {
		log.Println("invaild email token claims")
		c.JSON(http.StatusUnprocessableEntity, "invalid uuid")
		return
	}

	details := AccessDetails{
		AccessUuid: uuid,
		UserEmail:  email,
	}

	// 4. Redis - Fining and Fetch
	client := redis.GetClient()
	userEmail, err := client.Get(details.AccessUuid).Result()
	if err != nil {
		log.Println("Redis fetch error")
		c.JSON(http.StatusUnprocessableEntity, "invalid uuid in redis")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": userEmail})
}

// JWT
// https://codeburst.io/using-jwt-for-authentication-in-a-golang-application-e0357d579ce2
// https://covenant.tistory.com/203
func GenerateJWT(email string) (*TokenDetails, error) {
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * 15).Unix()
	td.AccessUuid = uuid.NewV4().String()

	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUuid = uuid.NewV4().String()

	var err error
	//Creating Access Token
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user_id"] = email
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return nil, err
	}

	//Creating Refresh Token
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = email
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return nil, err
	}
	return td, nil
}

func CreateAuth(email string, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) //Converting Unix to UTC
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	client := redis.GetClient()
	//redis set
	errAccess := client.Set(td.AccessUuid, email, at.Sub(now)).Err()
	if errAccess != nil {
		return errAccess
	}

	errRefresh := client.Set(td.RefreshUuid, email, rt.Sub(now)).Err()
	if errRefresh != nil {
		return errRefresh
	}

	return nil
}
