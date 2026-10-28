package streams

import "slices"

type serverSubscription struct {
	Topic string `json:"topic"`
	Type  string `json:"type"`
}
type serverState struct {
	assets []string
	custom bool
	all    bool
	topics []serverSubscription
}

func deriveState(selections []selection) serverState {
	var state serverState
	state.assets = []string{}
	for _, s := range selections {
		state.all = state.all || s.all
		for _, asset := range s.assets {
			if !slices.Contains(state.assets, asset) {
				state.assets = append(state.assets, asset)
			}
		}
		state.custom = state.custom || s.custom
		if s.service != "rtds" {
			continue
		}
		topic := map[string]string{"comments": "comments", "prices.crypto.binance": "crypto_prices", "prices.crypto.chainlink": "crypto_prices_chainlink", "prices.equity.pyth": "equity_prices"}[s.topic]
		if s.window == 30 {
			topic = "crypto_prices_twap_thirty"
		}
		if s.window == 60 {
			topic = "crypto_prices_twap_sixty"
		}
		types := []string{"update"}
		if s.topic == "comments" {
			types = s.types
		}
		for _, kind := range types {
			entry := serverSubscription{Topic: topic, Type: kind}
			if !slices.Contains(state.topics, entry) {
				state.topics = append(state.topics, entry)
			}
		}
	}
	if state.all {
		state.assets = []string{}
	}
	slices.Sort(state.assets)
	return state
}
func initialFrames(service string, state serverState) []any {
	switch service {
	case "market":
		return []any{map[string]any{"type": "market", "assets_ids": state.assets, "custom_feature_enabled": state.custom}}
	case "rtds":
		return []any{map[string]any{"action": "subscribe", "subscriptions": state.topics}}
	default:
		return nil
	}
}
func diffFrames(service string, before, after serverState) []any {
	if service == "user" {
		return userDiff(before, after)
	}
	var frames []any
	switch service {
	case "market":
		var added, removed []string
		for _, id := range after.assets {
			if !slices.Contains(before.assets, id) {
				added = append(added, id)
			}
		}
		for _, id := range before.assets {
			if !slices.Contains(after.assets, id) {
				removed = append(removed, id)
			}
		}
		if len(added) == 0 && before.custom != after.custom && len(after.assets) > 0 {
			added = []string{after.assets[0]}
		}
		if len(added) > 0 {
			frames = append(frames, map[string]any{"operation": "subscribe", "assets_ids": added, "custom_feature_enabled": after.custom})
		}
		if len(removed) > 0 {
			frames = append(frames, map[string]any{"operation": "unsubscribe", "assets_ids": removed})
		}
	case "rtds":
		var added, removed []serverSubscription
		for _, s := range after.topics {
			if !slices.Contains(before.topics, s) {
				added = append(added, s)
			}
		}
		for _, s := range before.topics {
			if !slices.Contains(after.topics, s) {
				removed = append(removed, s)
			}
		}
		if len(added) > 0 {
			frames = append(frames, map[string]any{"action": "subscribe", "subscriptions": added})
		}
		if len(removed) > 0 {
			frames = append(frames, map[string]any{"action": "unsubscribe", "subscriptions": removed})
		}
	}
	return frames
}

func (c *Client) initialFrames(service string, state serverState) []any {
	if service != "user" {
		return initialFrames(service, state)
	}
	credentials := c.config.Credentials
	frame := map[string]any{"type": "user", "auth": map[string]string{"apiKey": credentials.Key, "secret": credentials.Secret, "passphrase": credentials.Passphrase}}
	if !state.all {
		frame["markets"] = state.assets
	}
	return []any{frame}
}
func userDiff(before, after serverState) []any {
	if after.all {
		if !before.all && len(before.assets) > 0 {
			return []any{map[string]any{"operation": "unsubscribe", "markets": before.assets}}
		}
		return nil
	}
	if before.all {
		if len(after.assets) > 0 {
			return []any{map[string]any{"operation": "subscribe", "markets": after.assets}}
		}
		return nil
	}
	frames := diffFrames("market", before, after)
	for _, value := range frames {
		if frame, ok := value.(map[string]any); ok {
			frame["markets"] = frame["assets_ids"]
			delete(frame, "assets_ids")
			delete(frame, "custom_feature_enabled")
		}
	}
	return frames
}
