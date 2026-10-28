"""Pin maker quote arithmetic, signatures and event parsing to Python."""
import json
from decimal import Decimal
from dataclasses import asdict
from eth_account import Account
from support import ROOT,manifest,normalize
from polymarket._internal.rfq import RfqQuoterSession,_parse_quote_request,_parse_confirmation_request,_parse_trade
from polymarket._internal import rfq as source
from polymarket._internal.environment import PRODUCTION_CONFIG as c
from polymarket._internal.wallet import derive_proxy_wallet_address,derive_safe_wallet_address,derive_beacon_deposit_wallet_address
from polymarket.models import ApiKeyCreds
from polymarket.rfq import RfqQuoteSource
signer=Account.from_key('0x'+f'{1:064x}')
source.secrets.randbits=lambda _:19
source.time.time=lambda:1720000000
wallets={'EOA':signer.address,'POLY_PROXY':derive_proxy_wallet_address(signer.address,c.wallet_derivation),'GNOSIS_SAFE':derive_safe_wallet_address(signer.address,c.wallet_derivation),'DEPOSIT_WALLET':derive_beacon_deposit_wallet_address(signer.address,c.wallet_derivation)}
cases=[]
request={'type':'RFQ_REQUEST','rfq_id':'request-1','requestor_public_id':'requestor','leg_position_ids':['123','456'],'condition_id':'0x03'+'ab'*30,'yes_position_id':'123','no_position_id':'124','direction':'BUY','side':'YES','requested_size':{'unit':'notional','value_e6':'7000000'},'submission_deadline':1720000030}
for kind,wallet in wallets.items():
 session=RfqQuoterSession(chain_id=137,credentials=ApiKeyCreds(key='key',secret='c2VjcmV0',passphrase='pass'),exchange=c.exchange_v3,headers=None,logger=None,signer=signer,url='wss://example.invalid',wallet=wallet,wallet_type=kind)
 for direction in ['BUY','SELL']:
  for origin in ['collateral','inventory']:
   for price in [1,333333,999999]:
    raw=request|{'direction':direction};event=_parse_quote_request(raw,session)
    size=source._default_quote_size(event.requested_size,price)
    expected=session._sign_quote_order(request=event,price_e6=price,size_e6=size,source=RfqQuoteSource(origin))
    cases.append({'wallet':wallet,'kind':kind,'request':raw,'price':str(Decimal(price)/1000000),'source':origin,'expected':{'type':'RFQ_QUOTE','rfq_id':event.rfq_id,'price_e6':str(price),'size_e6':str(size),'signed_order':expected}})
(ROOT/'testdata/rfq-session-parity.json').write_text(json.dumps({'upstream_version':manifest['version'],'upstream_commit':manifest['commit'],'cases':cases},indent=2)+'\n')
print('Generated',len(cases),'RFQ maker quote cases')
