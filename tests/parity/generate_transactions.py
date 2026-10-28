"""Execute pinned Python ABI encoders and transaction signers without a network."""
import inspect,json
from eth_account import Account
from support import ROOT,manifest
from polymarket._internal.actions.relayer import calls as abi
from polymarket._internal.actions.relayer.signing.deposit_wallet import sign_deposit_wallet_batch
from polymarket._internal.actions.relayer.signing.safe import sign_safe_transaction
from polymarket._internal.actions.relayer.signing.proxy import build_proxy_transaction_hash,sign_proxy_message
from polymarket._internal.wallet import wrap_deposit_wallet_signature
from polymarket._internal.environment import PRODUCTION_CONFIG as config
cases=[]
key='0x'+f'{1:064x}';signer=Account.from_key(key)
wallet='0x1111111111111111111111111111111111111111';target='0x2222222222222222222222222222222222222222'
for count in [1,2,4]:
 for value in [0,1,2**255]:
  calls=[abi.TransactionCall(to=target,data='0x'+'12ab'*i,value=value) for i in range(count)]
  row={'key':key,'wallet':wallet,'nonce':'19','deadline':'1720000600','calls':[{'to':c.to,'data':c.data,'value':str(c.value)} for c in calls]}
  row['proxy_data']=abi.encode_proxy_call(calls);row['safe_data']=abi.encode_safe_multisend_call(calls)
  deposit=sign_deposit_wallet_batch(signer,wallet=wallet,calls=calls,nonce=row['nonce'],deadline=row['deadline'],chain_id=137)
  row['deposit_signature']=deposit
  row['session_signature']=wrap_deposit_wallet_signature(signer=signer.address,signer_type='SESSION_KEY',signature=deposit)
  row['safe_signature']=sign_safe_transaction(signer,safe_address=wallet,to=target,data=row['safe_data'],value=value,operation=1,nonce='19',chain_id=137)
  digest=build_proxy_transaction_hash(from_address=signer.address,to=target,data=row['proxy_data'],relayer_fee='0',gas_price='0',gas_limit='200000',nonce='19',relay_hub=config.relay_hub,relay=wallet)
  row['proxy_signature']=sign_proxy_message(signer,digest)
  row['legacy_transaction']='0x'+signer.sign_transaction({'nonce':19,'gasPrice':30_000_000_000,'gas':200000,'to':target,'value':value,'data':calls[0].data,'chainId':137}).raw_transaction.hex()
  cases.append(row)
(ROOT/'testdata/transactions-parity.json').write_text(json.dumps({'upstream_version':manifest['version'],'upstream_commit':manifest['commit'],'cases':cases},indent=2)+'\n')
print('Generated',len(cases),'transaction byte and signature cases')
