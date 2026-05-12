package auth

import (
	"fmt"
	"time"

	"github.com/gbrlsnchs/jwt/v3"
)

type NodeClaims struct {
	jwt.Payload
	NodeID uint `json:"node_id,omitempty"`
}

type JWTManager struct {
	signingKey []byte
	algorithm  jwt.Algorithm
}

func NewJWTManager(signingKey []byte) *JWTManager {
	return &JWTManager{
		signingKey: signingKey,
		algorithm:  jwt.NewHS256(signingKey),
	}
}

func (m *JWTManager) NewNodeToken(nodeID uint) (string, error) {
	now := time.Now()
	claims := &NodeClaims{
		Payload: jwt.Payload{
			Issuer:         "Pelican Panel",
			Subject:        fmt.Sprintf("node:%d", nodeID),
			IssuedAt:       jwt.NumericDate(now),
			ExpirationTime: jwt.NumericDate(now.Add(24 * time.Hour * 365)),
		},
		NodeID: nodeID,
	}

	token, err := jwt.Sign(claims, m.algorithm)
	if err != nil {
		return "", fmt.Errorf("sign node token: %w", err)
	}
	return string(token), nil
}

func (m *JWTManager) ParseNodeToken(tokenStr string) (*NodeClaims, error) {
	var claims NodeClaims
	verifier := jwt.NewHS256(m.signingKey)
	_, err := jwt.Verify([]byte(tokenStr), verifier, &claims)
	if err != nil {
		return nil, fmt.Errorf("verify node token: %w", err)
	}
	return &claims, nil
}

func (m *JWTManager) NewSignedURL(nodeToken []byte, subject string, expiration time.Duration) (string, error) {
	now := time.Now()
	algo := jwt.NewHS256(nodeToken)
	claims := &jwt.Payload{
		Issuer:         "Pelican Panel",
		Subject:        subject,
		IssuedAt:       jwt.NumericDate(now),
		ExpirationTime: jwt.NumericDate(now.Add(expiration)),
	}
	token, err := jwt.Sign(claims, algo)
	if err != nil {
		return "", fmt.Errorf("sign url token: %w", err)
	}
	return string(token), nil
}
