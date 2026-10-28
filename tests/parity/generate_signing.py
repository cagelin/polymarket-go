"""Execute Python authentication, wallet derivation and four-mode order signatures."""
import json
from eth_account import Account
from support import ROOT,manifest
from polymarket.environments import PRODUCTION
from polymarket._internal.l1_auth import sign_api_key_auth
from polymarket._internal.actions.auth import build_l1_auth_headers
from polymarket._internal.hmac import build_hmac_signature
from polymarket._internal.wallet import derive_proxy_wallet_address,derive_safe_wallet_address,derive_uups_deposit_wallet_address,derive_beacon_deposit_wallet_address,wrap_deposit_wallet_signature
from polymarket._internal.actions.orders.typed_data import build_order_typed_data,build_order_signature
from polymarket._internal.actions.orders.types import UnsignedOrder,BYTES32_ZERO
c=PRODUCTION._config;cases=[]
# These are deliberately public, synthetic test keys, never funded wallets.
for private in [1,2,123456789]:
 key='0x'+f'{private:064x}';signer=Account.from_key(key)
 wallets={'EOA':signer.address,'POLY_PROXY':derive_proxy_wallet_address(signer.address,c.wallet_derivation),'GNOSIS_SAFE':derive_safe_wallet_address(signer.address,c.wallet_derivation),'UUPS':derive_uups_deposit_wallet_address(signer.address,c.wallet_derivation),'BEACON':derive_beacon_deposit_wallet_address(signer.address,c.wallet_derivation)}
 for nonce in [0,1,99999999]:
  cases.append({'operation':'l1','key':key,'timestamp':1720000000,'nonce':nonce,'expected':build_l1_auth_headers(sign_api_key_auth(signer,chain_id=137,timestamp=1720000000,nonce=nonce))})
 cases.append({'operation':'wallets','key':key,'expected':wallets})
 for kind in range(4):
  for asset in ['123',str(1<<60)]:
   for side in ['BUY','SELL']:
    for session in ([False,True] if kind==3 else [False]):
     wallet=wallets[['EOA','POLY_PROXY','GNOSIS_SAFE','BEACON'][kind]]
     exchange=c.exchange_v3 if asset=='123' else c.standard_exchange
     order=UnsignedOrder(chain_id=137,builder=BYTES32_ZERO,exchange_address=exchange,expiration=0,maker=wallet,maker_amount=1234567,metadata=BYTES32_ZERO,order_type='GTC',salt=12345,side=side,signature_type=kind,signer=wallet if kind==3 else signer.address,taker_amount=7654321,timestamp=1720000000123,token_id=asset,protocol_version='3' if asset=='123' else '2')
     signed='0x'+signer.sign_typed_data(full_message=build_order_typed_data(order)).signature.hex()
     signature=build_order_signature(order,signed)
     if kind==3:signature=wrap_deposit_wallet_signature(signer=signer.address,signer_type='SESSION_KEY' if session else 'OWNER',signature=signature)
     raw={name:getattr(order,name) for name in order.__dataclass_fields__};raw['asset_id']=raw.pop('token_id');raw['signature']='';raw['post_only']=False
     cases.append({'operation':'order','key':key,'order':raw,'session':session,'exchange':exchange,'expected':signature})
for secret in ['c2VjcmV0','c2VjcmV0==','-_8=','+/8=','','a','YQ#','YQ###','YWJj=','YWJj===','not-a-secret','中文']:
 for body in [None,'','{"unicode":"中文","id":"a&b"}']:
  item={'operation':'hmac','secret':secret,'timestamp':1720000000,'method':'POST','path':'/order','body':body}
  try:item['expected']=build_hmac_signature(secret=secret,timestamp=item['timestamp'],method=item['method'],path=item['path'],body=body)
  except Exception as err:item['error']=type(err).__name__
  cases.append(item)
(ROOT/'testdata/signing-parity.json').write_text(json.dumps({'upstream_version':manifest['version'],'upstream_commit':manifest['commit'],'cases':cases},indent=2)+'\n')
print('Generated',len(cases),'signing cases')
