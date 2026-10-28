"""Pin approval pairs to the exact upstream environment, including its required operator set."""
from support import ROOT
from polymarket.environments import PRODUCTION
from polymarket._internal.actions.relayer.approvals import _required_trading_approvals
import json,subprocess
first,second=_required_trading_approvals(PRODUCTION._config)
s='''package data

import "github.com/cagelam/polymarket-go/models"

// RequiredApprovals returns fresh values so callers cannot mutate the pinned catalog.
//nolint:gosec // This catalog contains public contract addresses, not credentials.
func RequiredApprovals() models.MissingTradingApprovals {
return models.MissingTradingApprovals{
ERC20:[]models.ERC20TradingApproval{
'''
for row in first:s+='{TokenAddress:'+json.dumps(row.token_address)+',Spender:'+json.dumps(row.spender)+',Amount:maxUint256()},\n'
s+='},\nERC1155:[]models.ERC1155TradingApproval{\n'
for row in second:s+='{TokenAddress:'+json.dumps(row.token_address)+',Operator:'+json.dumps(row.operator)+'},\n'
s+='},\n}\n}\n'
(ROOT/'internal/data/approval_catalog.go').write_text(s)
subprocess.run(['gofmt','-w',str(ROOT/'internal/data/approval_catalog.go')],check=True)
