package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ung/api/internal/models"
)

const (
	SubscriptionContextKey contextKey = "subscription"
)

// Plan types
const (
	PlanTypeFree     = "free"
	PlanTypePro      = "pro"
	PlanTypeBusiness = "business"
)

// Feature entitlements
const (
	EntitlementPro      = "pro"
	EntitlementBusiness = "business"
)

// RevenueCatConfig holds configuration for RevenueCat integration
type RevenueCatConfig struct {
	APIKey  string
	BaseURL string
	Enabled bool
}

// SubscriptionMiddleware checks user subscription status via RevenueCat
func SubscriptionMiddleware(config *RevenueCatConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUser(r)
			if user == nil {
				respondError(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Build subscription info
			subInfo := &models.SubscriptionInfo{
				UserID:       user.ID,
				PlanType:     user.PlanType,
				IsActive:     true,
				Entitlements: make(map[string]models.RevenueCatEntitlement),
			}

			// If RevenueCat is enabled and user has a subscription ID, check status
			if config != nil && config.Enabled && user.SubscriptionID != nil && *user.SubscriptionID != "" {
				rcSub, err := fetchRevenueCatSubscription(config, *user.SubscriptionID)
				if err == nil && rcSub != nil {
					subInfo = rcSub
					subInfo.UserID = user.ID
				}
			}

			// Add subscription info to context
			ctx := context.WithValue(r.Context(), SubscriptionContextKey, subInfo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePlan middleware checks if user has required plan level
func RequirePlan(requiredPlans ...string) func(http.Handler) http.Handler {
	planMap := make(map[string]bool)
	for _, p := range requiredPlans {
		planMap[p] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sub := GetSubscription(r)
			if sub == nil {
				respondError(w, "Subscription info not available", http.StatusInternalServerError)
				return
			}

			// Check if user's plan is in the required plans list
			if !planMap[sub.PlanType] {
				respondError(w, "This feature requires a "+requiredPlans[0]+" plan or higher", http.StatusForbidden)
				return
			}

			if !sub.IsActive {
				respondError(w, "Your subscription is not active", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireEntitlement middleware checks if user has specific entitlement
func RequireEntitlement(entitlement string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sub := GetSubscription(r)
			if sub == nil {
				respondError(w, "Subscription info not available", http.StatusInternalServerError)
				return
			}

			// Check if user has the entitlement
			ent, ok := sub.Entitlements[entitlement]
			if !ok || !ent.IsActive {
				respondError(w, "This feature requires the "+entitlement+" entitlement", http.StatusForbidden)
				return
			}

			// Check if entitlement is expired
			if ent.ExpiresDate != nil && ent.ExpiresDate.Before(time.Now()) {
				respondError(w, "Your "+entitlement+" entitlement has expired", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetSubscription retrieves subscription info from request context
func GetSubscription(r *http.Request) *models.SubscriptionInfo {
	if sub, ok := r.Context().Value(SubscriptionContextKey).(*models.SubscriptionInfo); ok {
		return sub
	}
	return nil
}

// RevenueCatSubscriberResponse represents the RevenueCat API response
type RevenueCatSubscriberResponse struct {
	Subscriber struct {
		Entitlements map[string]struct {
			ExpiresDate       string `json:"expires_date"`
			ProductIdentifier string `json:"product_identifier"`
			PurchaseDate      string `json:"purchase_date"`
		} `json:"entitlements"`
		Subscriptions map[string]struct {
			ExpiresDate         string `json:"expires_date"`
			PurchaseDate        string `json:"purchase_date"`
			WillRenew           bool   `json:"will_renew"`
			ProductIdentifier   string `json:"product_identifier"`
			IsSandbox           bool   `json:"is_sandbox"`
			UnsubscribeDetectedAt string `json:"unsubscribe_detected_at,omitempty"`
		} `json:"subscriptions"`
	} `json:"subscriber"`
}

// fetchRevenueCatSubscription fetches subscription data from RevenueCat API
func fetchRevenueCatSubscription(config *RevenueCatConfig, appUserID string) (*models.SubscriptionInfo, error) {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.revenuecat.com/v1"
	}

	url := fmt.Sprintf("%s/subscribers/%s", config.BaseURL, appUserID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RevenueCat API error: %s - %s", resp.Status, string(body))
	}

	var rcResp RevenueCatSubscriberResponse
	if err := json.NewDecoder(resp.Body).Decode(&rcResp); err != nil {
		return nil, err
	}

	// Convert to our subscription info model
	subInfo := &models.SubscriptionInfo{
		PlanType:     PlanTypeFree,
		IsActive:     false,
		Entitlements: make(map[string]models.RevenueCatEntitlement),
	}

	// Process entitlements
	for name, ent := range rcResp.Subscriber.Entitlements {
		expiresDate, _ := time.Parse(time.RFC3339, ent.ExpiresDate)
		purchaseDate, _ := time.Parse(time.RFC3339, ent.PurchaseDate)

		isActive := expiresDate.After(time.Now())

		subInfo.Entitlements[name] = models.RevenueCatEntitlement{
			IsActive:          isActive,
			ProductIdentifier: ent.ProductIdentifier,
			ExpiresDate:       &expiresDate,
			PurchaseDate:      purchaseDate,
		}

		if isActive {
			subInfo.IsActive = true
			// Set plan type based on entitlement
			switch name {
			case EntitlementBusiness:
				subInfo.PlanType = PlanTypeBusiness
			case EntitlementPro:
				if subInfo.PlanType != PlanTypeBusiness {
					subInfo.PlanType = PlanTypePro
				}
			}
		}
	}

	// Find the latest expiry date
	var latestExpiry time.Time
	for _, sub := range rcResp.Subscriber.Subscriptions {
		expiresDate, err := time.Parse(time.RFC3339, sub.ExpiresDate)
		if err == nil && expiresDate.After(latestExpiry) {
			latestExpiry = expiresDate
		}
	}
	if !latestExpiry.IsZero() {
		subInfo.ExpiresAt = &latestExpiry
	}

	return subInfo, nil
}

// SubscriptionController handles subscription-related endpoints
type SubscriptionController struct {
	config *RevenueCatConfig
}

// NewSubscriptionController creates a new subscription controller
func NewSubscriptionController(config *RevenueCatConfig) *SubscriptionController {
	return &SubscriptionController{config: config}
}

// GetStatus handles GET /api/v1/subscription
func (c *SubscriptionController) GetStatus(w http.ResponseWriter, r *http.Request) {
	sub := GetSubscription(r)
	if sub == nil {
		// Return default free subscription info
		user := GetUser(r)
		sub = &models.SubscriptionInfo{
			UserID:       user.ID,
			PlanType:     user.PlanType,
			IsActive:     true,
			Entitlements: make(map[string]models.RevenueCatEntitlement),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    sub,
	})
}

// VerifyPurchase handles POST /api/v1/subscription/verify
// This endpoint allows the frontend to verify a purchase and update user subscription
func (c *SubscriptionController) VerifyPurchase(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AppUserID string `json:"app_user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if c.config == nil || !c.config.Enabled {
		respondError(w, "RevenueCat integration is not configured", http.StatusServiceUnavailable)
		return
	}

	// Fetch subscription from RevenueCat
	subInfo, err := fetchRevenueCatSubscription(c.config, req.AppUserID)
	if err != nil {
		respondError(w, "Failed to verify subscription: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    subInfo,
	})
}
