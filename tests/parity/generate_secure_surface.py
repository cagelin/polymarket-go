"""Track all nonexperimental secure members and the explicit wallet-milestone remainder."""
import inspect,json
from support import ROOT,manifest
from polymarket.clients import SecureClient,AsyncSecureClient
from generate_market_types import go_name
wallet=set('open_rfq_session authorize_session_key fetch_session_keys revoke_session_key approve_erc20 approve_erc1155_for_all transfer_erc20 get_trading_approvals_state setup_trading_approvals setup_gasless_wallet is_gasless_ready execute_transaction split_position merge_positions merge_multiple_positions redeem_positions plan_collateral_return execute_collateral_return_plan'.split())
items=[]
for name in sorted({name for cls in [SecureClient,AsyncSecureClient] for name,value in inspect.getmembers(cls) if not name.startswith('_') and (callable(value) or isinstance(value,property)) and 'Experimental:' not in (inspect.getdoc(value) or '')}):
 mapped=go_name(name).replace('Pnl','PnL').replace('Rfq','RFQ').replace('Api','API').replace('Erc','ERC')
 items.append({'python':name,'go':mapped,'stage':3 if name in wallet else 2,'excluded':name in {'setup_gasless_wallet','is_gasless_ready'}})
(ROOT/'testdata/secure-surface.json').write_text(json.dumps({'upstream_version':manifest['version'],'upstream_commit':manifest['commit'],'members':items},indent=2)+'\n')
print('Generated',len(items),'secure client members')
