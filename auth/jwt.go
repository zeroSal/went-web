package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kataras/iris/v12"
)

type JWT struct {
	Base
	Secret []byte
	Expiry time.Duration
}

func NewJWT(secret []byte, expiry time.Duration) *JWT {
	return &JWT{
		Secret: secret,
		Expiry: expiry,
	}
}

func (j *JWT) Authenticate(ctx iris.Context) (interface{}, bool) {
	tokenString := j.ExtractToken(ctx)
	if tokenString == "" {
		tokenString = ctx.GetCookie("jwt")
	}
	if tokenString == "" {
		return nil, false
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return j.Secret, nil
	})

	if err != nil || !token.Valid {
		return nil, false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, false
	}

	return claims, true
}

func (j *JWT) Middleware() iris.Handler {
	return func(ctx iris.Context) {
		if _, ok := j.Authenticate(ctx); !ok {
			ctx.StatusCode(iris.StatusUnauthorized)
			ctx.StopExecution()
			return
		}
		ctx.Next()
	}
}

func (j *JWT) GenerateToken(claims map[string]interface{}) (string, error) {
	claims["exp"] = time.Now().Add(j.Expiry).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	return token.SignedString(j.Secret)
}
