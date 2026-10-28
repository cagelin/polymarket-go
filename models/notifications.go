package models

import (
	"encoding/json"
	"fmt"

	"github.com/cagelam/polymarket-go/internal/codec"
)

// NotificationType identifies the notification payload.
type NotificationType int

// Known account notification types.
const (
	NotificationOrderCancellation   NotificationType = 1
	NotificationOrderFill           NotificationType = 2
	NotificationMarketRegistered    NotificationType = 3
	NotificationMarketResolved      NotificationType = 4
	NotificationRewardPayout        NotificationType = 5
	NotificationChildCommentCreated NotificationType = 6
	NotificationYieldPayout         NotificationType = 7
	NotificationOrderFillFailed     NotificationType = 8
	NotificationAutoRedeemed        NotificationType = 9
	NotificationComboAutoRedeemed   NotificationType = 10
	NotificationTakerTierUpgraded   NotificationType = 11
)

// Notification is one of the known typed account notification variants.
type Notification interface{ notificationVariant() }

// ParseNotifications decodes the account's discriminated notification list.
func ParseNotifications(raw []byte) ([]Notification, error) {
	rows, err := codec.ParseList[json.RawMessage](raw)
	if err != nil {
		return nil, err
	}
	result := make([]Notification, 0, len(rows))
	for _, row := range rows {
		item, parseErr := ParseNotification(row)
		if parseErr != nil {
			return nil, parseErr
		}
		result = append(result, item)
	}
	return result, nil
}

// ParseNotification selects a payload by its numeric wire type.
func ParseNotification(raw []byte) (Notification, error) {
	fields, err := codec.Object(raw)
	if err != nil {
		return nil, err
	}
	var kind float64
	if string(fields["type"]) == "true" {
		kind = 1
	} else if err = json.Unmarshal(fields["type"], &kind); err != nil {
		return nil, err
	}
	var result Notification
	switch kind {
	case float64(NotificationAutoRedeemed):
		result = &AutoRedeemedNotification{}
	case float64(NotificationChildCommentCreated):
		result = &ChildCommentCreatedNotification{}
	case float64(NotificationComboAutoRedeemed):
		result = &ComboAutoRedeemedNotification{}
	case float64(NotificationMarketRegistered):
		result = &MarketRegisteredNotification{}
	case float64(NotificationMarketResolved):
		result = &MarketResolvedNotification{}
	case float64(NotificationOrderCancellation):
		result = &OrderCancellationNotification{}
	case float64(NotificationOrderFillFailed):
		result = &OrderFillFailedNotification{}
	case float64(NotificationOrderFill):
		result = &OrderFillNotification{}
	case float64(NotificationRewardPayout):
		result = &RewardPayoutNotification{}
	case float64(NotificationTakerTierUpgraded):
		result = &TakerTierUpgradedNotification{}
	case float64(NotificationYieldPayout):
		result = &YieldPayoutNotification{}
	default:
		return nil, fmt.Errorf("unknown notification type")
	}
	if err = json.Unmarshal(raw, result); err != nil {
		return nil, err
	}
	return result, nil
}
func (*AutoRedeemedNotification) notificationVariant()        {}
func (*ChildCommentCreatedNotification) notificationVariant() {}
func (*ComboAutoRedeemedNotification) notificationVariant()   {}
func (*MarketRegisteredNotification) notificationVariant()    {}
func (*MarketResolvedNotification) notificationVariant()      {}
func (*OrderCancellationNotification) notificationVariant()   {}
func (*OrderFillFailedNotification) notificationVariant()     {}
func (*OrderFillNotification) notificationVariant()           {}
func (*RewardPayoutNotification) notificationVariant()        {}
func (*TakerTierUpgradedNotification) notificationVariant()   {}
func (*YieldPayoutNotification) notificationVariant()         {}
