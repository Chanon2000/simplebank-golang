package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4" // เติม v4 ด้วย เนื่องจาก auto import ของ vscode มันไม่ใส่ให้ (แต่ v5 จะเป็น version ล่าสุดนะแต่ไม่ใช่เพราะว่า code มันจะต่างจากของผู้สอนเกินไป)
)

const minSecretKeySize = 32

// JWTMaker is a JSON Web Token maker
type JWTMaker struct {
	secretKey string
}

// NewJWTMaker creates a new JWTMaker
func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize { // เพื่อให้มั้นใจว่า secretKey จะไม่สั้นเกินไปเพื่อ security ที่ดีขึ้น
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}
	return &JWTMaker{secretKey}, nil // return type เป็น Maker interface นั้นก็คือจะต้อง return struct ที่มี CreateToken และ VerifyToken method ก็ได้ โดยในที่นี้เรา return เป็น JWTMaker struct นั้นเอง
}

// CreateToken creates a new token for a specific username and duration
func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload) // Payload จะต้องมี Valid() method เนื่องจาก jwt มันต้องการเพื่อใช้ภายใน
	return jwtToken.SignedString([]byte(maker.secretKey))
}

// VerifyToken checks if the token is valid or not
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC) // คือทำการ verify header เพื่อให้มั้นใจว่า signing algorithm (เพื่อป้องกัน trivial attack machanism เช่น set header เป็น none เพื่อ bypass การ verify เป็นต้น)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc) // &Payload{} คือ empty payload object // ParseWithClaims คือทำการ parse token
	// keyFunc คือ function ที่ได้รับ token ที่ parsed แล้วแต่ยังไม่ได้ verified ซึ่ง jwt-go จะใช้มันเพื่อ verify token นั้นเอง
	if err != nil {
		verr, ok := err.(*jwt.ValidationError) // เนื่องจาก ParseWithClaims จะ return error ออกมาเป็น jwt.ValidationError ซึ่งครั้งนี้ก็ทำการ convert err เป็น type jwt.ValidationError
		if ok && errors.Is(verr.Inner, ErrExpiredToken) { // เนื่องจาก ParseWithClaims จะเก็บ err ไว้ใน Inner field (ลอง inpect code ที่ ParseWithClaims function ดูได้)
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}

	return payload, nil
}
