package lib

import (
	"time"
	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	//"github.com/dgrijalva/jwt-go/request"
)

//ValidateToken is						success,userlogin,error
func ValidateToken(accessToken string, remoteIP string, intTransID string, JWTSecretKey string) (*jwt.StandardClaims, bool) {
	var token *jwt.Token
	token, err := jwt.ParseWithClaims(accessToken, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTSecretKey), nil
	})

	if err != nil {
		//logger.Logf("["+intTransID+"]"+"-"+"["+remoteIP+"] ValidateToken : %s", "Valid")
		return nil, false
	}
	if jwtclaim, ok := token.Claims.(*jwt.StandardClaims); ok && token.Valid {
		//logger.Logf("["+intTransID+"]"+"-"+"["+remoteIP+"] ValidateToken : %s", "Valid")
		return jwtclaim,true
	} 
	//logger.Logf("["+intTransID+"]"+"-"+"["+remoteIP+"] ValidateToken : %s", "NotValid")
	return nil,false

}

//GenerateToken is
func GenerateToken(AppID string, remoteIP string, JWTSecretKey string,MAXJWTTokenLife int) (string, error) {
	var claims jwt.StandardClaims
	var sessionID string =GetTransactionid(true)
	claims.Id = sessionID
	claims.IssuedAt = time.Now().Unix()
	claims.ExpiresAt = time.Now().Add(time.Duration(MAXJWTTokenLife) * time.Hour).Unix()
	claims.Subject = AppID
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWTSecretKey))

	if err != nil {
		return "", err
	}
	log.Debugf("["+AppID+"]"+"-"+"["+remoteIP+"] GenerateToken : %s", tokenString)
	
	return tokenString, err
}
