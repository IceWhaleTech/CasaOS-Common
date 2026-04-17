package external

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/utils/constants"
	http2 "github.com/IceWhaleTech/CasaOS-Common/utils/http"
	"github.com/IceWhaleTech/CasaOS-Common/utils/jwt"
	"github.com/orca-zhang/ecache"
)

const (
	UserServiceAddressFilename = "user-service.url"
	GatewaySockFilename        = "zimaos-gateway.sock"
)

var (
	cachedPublicKey        *ecdsa.PublicKey
	lastUpdate             time.Time
	validParseTokenCache   = ecache.NewLRUCache(2, 8, time.Minute)
	invalidParseTokenCache = ecache.NewLRUCache(2, 16, time.Minute)
	readUserServiceAddress = getAddress
	userServiceAddressFile = filepath.Join(constants.DefaultRuntimePath, UserServiceAddressFilename)
	gatewaySockFile        = filepath.Join(constants.DefaultRuntimePath, GatewaySockFilename)
)

var (
	errTokenInvalid = errors.New("token is invalid")
	errTokenExpired = errors.New("token is expired")
)

type tokenCacheSentinel uint8

const (
	tokenCacheSentinelInvalid tokenCacheSentinel = iota + 1
	tokenCacheSentinelExpired
)

type ParsedToken struct {
	Valid     bool   `json:"valid"`
	ExpiresAt int64  `json:"expires_at"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	UserID    int    `json:"user_id"`
}

func GetPublicKey(runtimePath string) (*ecdsa.PublicKey, error) {
	if cachedPublicKey != nil && time.Since(lastUpdate) < 10*time.Second {
		return cachedPublicKey, nil
	}

	address, err := getAddress(filepath.Join(runtimePath, UserServiceAddressFilename))
	if err != nil {
		return nil, err
	}

	jwksURL, err := url.JoinPath(address, jwt.JWKSPath)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch JWKS: received status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read JWKS response: %w", err)
	}

	var jwks jwt.JWKS
	err = json.Unmarshal(body, &jwks)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWKS: %w", err)
	}

	// Use the first key in JWKS to validate the JWT
	if len(jwks.Keys) == 0 {
		return nil, fmt.Errorf("no keys found in JWKS")
	}

	key := jwks.Keys[0]
	xBytes, err := base64.RawURLEncoding.DecodeString(key.X)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWK x value: %w", err)
	}

	yBytes, err := base64.RawURLEncoding.DecodeString(key.Y)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWK y value: %w", err)
	}

	cachedPublicKey = &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}

	lastUpdate = time.Now()

	return cachedPublicKey, nil
}

func ParseToken(token string) (*ParsedToken, error) {
	normalizedToken := strings.TrimSpace(token)
	if normalizedToken == "" {
		return nil, errors.New("token is empty")
	}

	cacheKey := normalizedToken

	if cachedEntry, found := invalidParseTokenCache.Get(cacheKey); found {
		switch cachedEntry.(tokenCacheSentinel) {
		case tokenCacheSentinelInvalid:
			return nil, errTokenInvalid
		case tokenCacheSentinelExpired:
			return nil, errTokenExpired
		}
	}

	if cachedEntry, found := validParseTokenCache.Get(cacheKey); found {
		token := cachedEntry.(*ParsedToken)
		if token.isExpired() {
			validParseTokenCache.Del(cacheKey)
			invalidParseTokenCache.Put(cacheKey, tokenCacheSentinelExpired)
			return nil, errTokenExpired
		}

		return token, nil
	}

	requestBody, err := json.Marshal(struct {
		Token string `json:"token"`
	}{Token: normalizedToken})
	if err != nil {
		return nil, err
	}

	var resp *http.Response
	address, err := readUserServiceAddress(userServiceAddressFile)
	if err == nil {
		parseTokenURL, err := url.JoinPath(address, "/v1/users/parse-token")
		if err != nil {
			return nil, err
		}

		resp, err = http2.Post(parseTokenURL, requestBody, 30*time.Second)
		if err != nil {
			return nil, fmt.Errorf("failed to parse token: %w", err)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://unix/v1/users/parse-token", bytes.NewBuffer(requestBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err = (&http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					var dialer net.Dialer
					return dialer.DialContext(ctx, "unix", gatewaySockFile)
				},
			},
		}).Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to parse token: %w", err)
		}
	} else {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to parse token: received status code %d", resp.StatusCode)
	}

	var parsedResp struct {
		Success int         `json:"success"`
		Message string      `json:"message"`
		Data    ParsedToken `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsedResp); err != nil {
		return nil, fmt.Errorf("failed to decode parse token response: %w", err)
	}

	if !parsedResp.Data.Valid {
		invalidParseTokenCache.Put(cacheKey, tokenCacheSentinelInvalid)
		return nil, errTokenInvalid
	}

	if parsedResp.Data.isExpired() {
		invalidParseTokenCache.Put(cacheKey, tokenCacheSentinelExpired)
		return nil, errTokenExpired
	}

	validParseTokenCache.Put(cacheKey, &parsedResp.Data)

	return &parsedResp.Data, nil
}

func (p *ParsedToken) isExpired() bool {
	return p.ExpiresAt <= time.Now().Unix()
}
