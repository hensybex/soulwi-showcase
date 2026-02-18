// internal/service/apple_service.go
package service

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hensybex/soulwi_go_back/internal/model"
)

// Apple URLs
const (
	AppleProductionURL = "https://buy.itunes.apple.com/verifyReceipt"
	AppleSandboxURL    = "https://sandbox.itunes.apple.com/verifyReceipt"
	AppleJWKSURL       = "https://appleid.apple.com/auth/keys"
)

// In-memory cache constants
const (
	appleKeysCacheTTL = 1 * time.Hour
)

// In-memory cache for Apple's public keys
type inMemoryAppleKeyCache struct {
	sync.RWMutex
	keys   map[string]*ecdsa.PublicKey
	expiry time.Time
}

func (c *inMemoryAppleKeyCache) Get() (map[string]*ecdsa.PublicKey, bool) {
	c.RLock()
	defer c.RUnlock()
	if c.keys == nil || time.Now().After(c.expiry) {
		return nil, false
	}
	return c.keys, true
}

func (c *inMemoryAppleKeyCache) Set(keys map[string]*ecdsa.PublicKey, ttl time.Duration) {
	c.Lock()
	defer c.Unlock()
	c.keys = keys
	c.expiry = time.Now().Add(ttl)
}

// Receipt validation response structures
type AppleReceiptResponse struct {
	Status             int                       `json:"status"`
	Environment        string                    `json:"environment"`
	Receipt            AppleReceipt              `json:"receipt"`
	LatestReceiptInfo  []AppleReceiptInfo        `json:"latest_receipt_info"`
	PendingRenewalInfo []ApplePendingRenewalInfo `json:"pending_renewal_info"`
	IsRetryable        bool                      `json:"is-retryable"`
}

type AppleReceipt struct {
	ReceiptType                string             `json:"receipt_type"`
	AdamId                     int64              `json:"adam_id"`
	AppItemId                  int64              `json:"app_item_id"`
	BundleId                   string             `json:"bundle_id"`
	ApplicationVersion         string             `json:"application_version"`
	DownloadId                 int64              `json:"download_id"`
	VersionExternalIdentifier  int64              `json:"version_external_identifier"`
	ReceiptCreationDate        string             `json:"receipt_creation_date"`
	ReceiptCreationDateMs      string             `json:"receipt_creation_date_ms"`
	ReceiptCreationDatePst     string             `json:"receipt_creation_date_pst"`
	RequestDate                string             `json:"request_date"`
	RequestDateMs              string             `json:"request_date_ms"`
	RequestDatePst             string             `json:"request_date_pst"`
	OriginalPurchaseDate       string             `json:"original_purchase_date"`
	OriginalPurchaseDateMs     string             `json:"original_purchase_date_ms"`
	OriginalPurchaseDatePst    string             `json:"original_purchase_date_pst"`
	OriginalApplicationVersion string             `json:"original_application_version"`
	InApp                      []AppleReceiptInfo `json:"in_app"`
}

type AppleReceiptInfo struct {
	Quantity                    string `json:"quantity"`
	ProductId                   string `json:"product_id"`
	TransactionId               string `json:"transaction_id"`
	OriginalTransactionId       string `json:"original_transaction_id"`
	PurchaseDate                string `json:"purchase_date"`
	PurchaseDateMs              string `json:"purchase_date_ms"`
	PurchaseDatePst             string `json:"purchase_date_pst"`
	OriginalPurchaseDate        string `json:"original_purchase_date"`
	OriginalPurchaseDateMs      string `json:"original_purchase_date_ms"`
	OriginalPurchaseDatePst     string `json:"original_purchase_date_pst"`
	ExpiresDate                 string `json:"expires_date"`
	ExpiresDateMs               string `json:"expires_date_ms"`
	ExpiresDatePst              string `json:"expires_date_pst"`
	WebOrderLineItemId          string `json:"web_order_line_item_id"`
	IsTrialPeriod               string `json:"is_trial_period"`
	IsInIntroOfferPeriod        string `json:"is_in_intro_offer_period"`
	IsUpgraded                  string `json:"is_upgraded"`
	InAppOwnershipType          string `json:"in_app_ownership_type"`
	SubscriptionGroupIdentifier string `json:"subscription_group_identifier"`
}

type ApplePendingRenewalInfo struct {
	ExpirationIntent          string `json:"expiration_intent"`
	AutoRenewProductId        string `json:"auto_renew_product_id"`
	OriginalTransactionId     string `json:"original_transaction_id"`
	IsInBillingRetryPeriod    string `json:"is_in_billing_retry_period"`
	ProductId                 string `json:"product_id"`
	AutoRenewStatus           string `json:"auto_renew_status"`
	GracePeriodExpiresDate    string `json:"grace_period_expires_date"`
	GracePeriodExpiresDateMs  string `json:"grace_period_expires_date_ms"`
	GracePeriodExpiresDatePst string `json:"grace_period_expires_date_pst"`
	OfferCodeRefName          string `json:"offer_code_ref_name"`
	PriceIncreaseStatus       string `json:"price_increase_status"`
}

// Server notification structures (App Store Server Notifications v2)
type AppleServerNotification struct {
	SignedPayload string `json:"signedPayload"`
}

type AppleNotificationPayload struct {
	NotificationType string                `json:"notificationType"`
	Subtype          string                `json:"subtype,omitempty"`
	NotificationUUID string                `json:"notificationUUID"`
	Data             AppleNotificationData `json:"data"`
	Version          string                `json:"version"`
	SignedDate       int64                 `json:"signedDate"`
}

type AppleNotificationData struct {
	AppAppleId            int64  `json:"appAppleId"`
	BundleId              string `json:"bundleId"`
	BundleVersion         string `json:"bundleVersion"`
	Environment           string `json:"environment"`
	SignedRenewalInfo     string `json:"signedRenewalInfo"`
	SignedTransactionInfo string `json:"signedTransactionInfo"`
}

// JWKS structures for Apple public keys
type AppleJWKS struct {
	Keys []AppleJWK `json:"keys"`
}

type AppleJWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	X   string `json:"x"`
	Y   string `json:"y"`
	Crv string `json:"crv"`
}

type AppStoreTransaction struct {
	TransactionID         string `json:"transactionId"`
	OriginalTransactionID string `json:"originalTransactionId"`
	ProductID             string `json:"productId"`
	Type                  string `json:"type"`
	PurchaseDate          int64  `json:"purchaseDate"`
	ExpiresDate           int64  `json:"expiresDate"`
	IsTrialPeriod         bool   `json:"isTrialPeriod"`
}

type AppleService interface {
	ValidateReceipt(ctx context.Context, receiptData string) (*AppleReceiptResponse, error)
	ValidateTestJWS(ctx context.Context, jwsToken string) (*AppStoreTransaction, error)
	ParseServerNotification(ctx context.Context, signedPayload string) (*AppleNotificationPayload, error)
	ConvertReceiptInfoToSubscription(receiptInfo AppleReceiptInfo, userID uint, environment model.Environment) *model.Subscription
}

type appleService struct {
	appSharedSecret string
	bundleID        string
	httpClient      *http.Client
	keyCache        *inMemoryAppleKeyCache
}

func NewAppleService(appSharedSecret, bundleID string) AppleService {
	return &appleService{
		appSharedSecret: appSharedSecret,
		bundleID:        bundleID,
		keyCache:        &inMemoryAppleKeyCache{},
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *appleService) ValidateReceipt(ctx context.Context, receiptData string) (*AppleReceiptResponse, error) {
	payload := map[string]interface{}{
		"receipt-data": receiptData,
		"password":     s.appSharedSecret,
	}

	response, err := s.makeReceiptRequest(ctx, AppleProductionURL, payload)
	if err != nil {
		return nil, err
	}

	if response.Status == 21007 {
		response, err = s.makeReceiptRequest(ctx, AppleSandboxURL, payload)
		if err != nil {
			return nil, err
		}
	}

	if response.Status != 0 {
		return nil, fmt.Errorf("apple receipt validation failed with status: %d", response.Status)
	}

	return response, nil
}

func (s *appleService) ValidateTestJWS(ctx context.Context, jwsToken string) (*AppStoreTransaction, error) {
	certBytes, err := os.ReadFile("/app/StoreKitTestCertificate.cer")
	if err != nil {
		return nil, fmt.Errorf("failed to read cert file: %w", err)
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		block, _ := pem.Decode(certBytes)
		if block == nil {
			return nil, fmt.Errorf("failed to decode pem block")
		}
		cert, err = x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse cert from pem: %w", err)
		}
	}

	publicKey, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an ECDSA public key")
	}

	token, err := jwt.Parse(jwsToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWS: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	jsonBytes, err := json.Marshal(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal claims: %w", err)
	}

	var transaction AppStoreTransaction
	if err := json.Unmarshal(jsonBytes, &transaction); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	return &transaction, nil
}

func (s *appleService) ParseServerNotification(ctx context.Context, signedPayload string) (*AppleNotificationPayload, error) {
	publicKeys, err := s.getApplePublicKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Apple public keys: %w", err)
	}

	token, err := jwt.Parse(signedPayload, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid kid in token header")
		}

		publicKey, exists := publicKeys[kid]
		if !exists {
			return nil, fmt.Errorf("public key not found for kid: %s", kid)
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse and validate JWT: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid JWT claims")
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal claims: %w", err)
	}

	var payload AppleNotificationPayload
	if err := json.Unmarshal(claimsJSON, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	return &payload, nil
}

func (s *appleService) getApplePublicKeys(ctx context.Context) (map[string]*ecdsa.PublicKey, error) {
	if keys, ok := s.keyCache.Get(); ok {
		return keys, nil
	}

	publicKeys, err := s.fetchApplePublicKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Apple public keys: %w", err)
	}

	s.keyCache.Set(publicKeys, appleKeysCacheTTL)
	return publicKeys, nil
}

func (s *appleService) fetchApplePublicKeys(ctx context.Context) (map[string]*ecdsa.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", AppleJWKSURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read JWKS response: %w", err)
	}

	var jwks AppleJWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWKS: %w", err)
	}

	publicKeys := make(map[string]*ecdsa.PublicKey)
	for _, jwk := range jwks.Keys {
		if jwk.Kty != "EC" || jwk.Crv != "P-256" {
			continue
		}

		publicKey, err := s.jwkToECDSAPublicKey(jwk)
		if err != nil {
			fmt.Printf("Failed to convert JWK to ECDSA for kid %s: %v\n", jwk.Kid, err)
			continue
		}

		publicKeys[jwk.Kid] = publicKey
	}

	if len(publicKeys) == 0 {
		return nil, fmt.Errorf("no valid ECDSA public keys found in JWKS")
	}

	return publicKeys, nil
}

func (s *appleService) jwkToECDSAPublicKey(jwk AppleJWK) (*ecdsa.PublicKey, error) {
	xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return nil, fmt.Errorf("failed to decode X coordinate: %w", err)
	}

	yBytes, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Y coordinate: %w", err)
	}

	x := new(big.Int).SetBytes(xBytes)
	y := new(big.Int).SetBytes(yBytes)

	publicKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	return publicKey, nil
}

func (s *appleService) makeReceiptRequest(ctx context.Context, url string, payload map[string]interface{}) (*AppleReceiptResponse, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var appleResponse AppleReceiptResponse
	if err := json.Unmarshal(body, &appleResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &appleResponse, nil
}

func (s *appleService) ConvertReceiptInfoToSubscription(receiptInfo AppleReceiptInfo, userID uint, environment model.Environment) *model.Subscription {
	subscription := &model.Subscription{
		UserID:                userID,
		ProductID:             receiptInfo.ProductId,
		OriginalTransactionID: receiptInfo.OriginalTransactionId,
		Environment:           environment,
		AutoRenewEnabled:      true,
	}

	if purchaseTime, err := s.parseAppleTime(receiptInfo.PurchaseDateMs); err == nil {
		subscription.PurchaseAt = &purchaseTime
	}

	if expiryTime, err := s.parseAppleTime(receiptInfo.ExpiresDateMs); err == nil {
		subscription.ExpiresAt = &expiryTime
	}

	if receiptInfo.IsTrialPeriod == "true" {
		subscription.Status = model.StatusTrial
		subscription.IsTrialActive = true
	} else {
		subscription.Status = model.StatusActive
		subscription.IsTrialActive = false
	}

	return subscription
}

func (s *appleService) parseAppleTime(timeMs string) (time.Time, error) {
	if timeMs == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}

	var timestamp int64
	if err := json.Unmarshal([]byte(timeMs), &timestamp); err != nil {
		return time.Time{}, err
	}

	return time.Unix(timestamp/1000, (timestamp%1000)*1000000), nil
}
