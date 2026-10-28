package polymarket

import (
	"context"

	"github.com/cagelam/polymarket-go/models"
	"github.com/cagelam/polymarket-go/pagination"
)

// ListTrades defaults an omitted user to the authenticated wallet.
func (c *SecureClient) ListTrades(p ListTradesParams) (*pagination.Paginator[models.Trade], error) {
	if p.User == nil {
		wallet := string(c.Wallet())
		p.User = &wallet
	}
	return c.PublicClient.ListTrades(p)
}

// ListActivity defaults an omitted user to the authenticated wallet.
func (c *SecureClient) ListActivity(p ListActivityParams) (*pagination.Paginator[models.Activity], error) {
	if p.User == "" {
		p.User = string(c.Wallet())
	}
	return c.PublicClient.ListActivity(p)
}

// ListComboActivity defaults an omitted user to the authenticated wallet.
func (c *SecureClient) ListComboActivity(p ListComboActivityParams) (*pagination.Paginator[models.ComboActivity], error) {
	if p.User == "" {
		p.User = string(c.Wallet())
	}
	return c.PublicClient.ListComboActivity(p)
}

// ListPositions defaults an omitted user to the authenticated wallet.
func (c *SecureClient) ListPositions(p ListPositionsParams) (*pagination.Paginator[models.Position], error) {
	if p.User == nil {
		wallet := string(c.Wallet())
		p.User = &wallet
	}
	return c.PublicClient.ListPositions(p)
}

// ListComboPositions defaults an omitted user to the authenticated wallet.
func (c *SecureClient) ListComboPositions(p ListComboPositionsParams) (*pagination.Paginator[models.ComboPosition], error) {
	if p.User == "" {
		p.User = string(c.Wallet())
	}
	return c.PublicClient.ListComboPositions(p)
}

// GetPortfolioValue defaults an omitted user to the authenticated wallet.
func (c *SecureClient) GetPortfolioValue(ctx context.Context, p GetPortfolioValueParams) (*models.PortfolioValue, error) {
	if p.User == "" {
		p.User = string(c.Wallet())
	}
	return c.PublicClient.GetPortfolioValue(ctx, p)
}

// GetUserStats defaults an omitted user to the authenticated wallet.
func (c *SecureClient) GetUserStats(ctx context.Context, p GetUserStatsParams) (*models.UserStats, error) {
	if p.User == "" {
		p.User = string(c.Wallet())
	}
	return c.PublicClient.GetUserStats(ctx, p)
}

// GetUserPnL defaults an omitted user to the authenticated wallet.
func (c *SecureClient) GetUserPnL(ctx context.Context, p GetUserPnLParams) (*models.UserPnLSeries, error) {
	if p.User == "" {
		p.User = string(c.Wallet())
	}
	return c.PublicClient.GetUserPnL(ctx, p)
}

// GetUserVolume defaults an omitted user to the authenticated wallet.
func (c *SecureClient) GetUserVolume(ctx context.Context, p GetUserVolumeParams) (*models.UserVolume, error) {
	if p.User == "" {
		p.User = string(c.Wallet())
	}
	return c.PublicClient.GetUserVolume(ctx, p)
}

// GetTraderLeaderboardStanding defaults an omitted user to the authenticated wallet.
func (c *SecureClient) GetTraderLeaderboardStanding(ctx context.Context, p GetTraderLeaderboardStandingParams) (*models.TraderLeaderboardStanding, error) {
	if p.User == "" {
		p.User = string(c.Wallet())
	}
	return c.PublicClient.GetTraderLeaderboardStanding(ctx, p)
}

// DownloadAccountingSnapshot defaults an omitted user to the authenticated wallet.
func (c *SecureClient) DownloadAccountingSnapshot(ctx context.Context, p DownloadAccountingSnapshotParams) ([]byte, error) {
	if p.User == "" {
		p.User = string(c.Wallet())
	}
	return c.PublicClient.DownloadAccountingSnapshot(ctx, p)
}
