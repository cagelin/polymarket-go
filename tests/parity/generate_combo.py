"""Execute requester combo RFQ validation, response parsing and acceptance signatures offline."""
import copy,json
from types import SimpleNamespace
from decimal import Decimal as D
from unittest.mock import patch
from eth_account import Account
from support import ROOT,manifest,normalize
from polymarket._internal.actions import combo_rfq as combo
from polymarket.environments import PRODUCTION
from polymarket._internal.wallet import derive_proxy_wallet_address,derive_safe_wallet_address,derive_beacon_deposit_wallet_address
from polymarket.auth import BuilderApiKey
from polymarket.rfq import ComboQuote
config=PRODUCTION._config;signer=Account.from_key('0x'+'00'*31+'01')
ctx=SimpleNamespace(wallet=signer.address,wallet_type='EOA',signer_type='OWNER',signer=signer,environment_config=config,api_key=BuilderApiKey(key='builder',secret='c2VjcmV0',passphrase='pass'))
base=dict(leg_position_ids=['123','456'],direction='BUY',amount='10',size=None,side='YES')
request=combo.build_combo_quote_request_body(ctx,**base)[1]
quote={'rfq_id':'rfq-1','status':'AWAITING_REQUESTER_ACCEPTANCE','expires_at':1720000000000,'builder_code':'0x'+'ab'*32,'request':request|{'rfq_id':'rfq-1','condition_id':'0x03'+'ab'*30,'yes_position_id':'123','no_position_id':'456'},'quote':{'quote_id':'quote-1','blended_price_e6':'500000','maker_amount_e6':'10000000','taker_amount_e6':'20000000','total_required_e6':'10100000'}}
cases=[]
def add(op,label,raw):
 case={'operation':op,'name':op+'/'+label,'input':raw}
 try:
  if op=='request':result=combo.build_combo_quote_request_body(ctx,**raw)[1]
  elif op=='quote':result=combo._parse_combo_quote_result(raw,request=request)
  elif op=='status':result=combo._parse_rfq_status(raw,expected_rfq_id='rfq-1')
  elif op=='model':result=ComboQuote.model_validate(raw)
  elif op=='accept':
   with patch.object(combo.secrets,'randbits',return_value=12345),patch.object(combo.time,'time',return_value=1720000000):result=combo._build_accept_request_body(ctx,ComboQuote.model_validate(raw))
  case['expected']={'value':normalize(result)}
 except Exception as err:case['expected']={'error':type(err).__name__}
 cases.append(case)
add('request','default',base)
for key,values in {'leg_position_ids':[[],['1'],['01','1'],['٤٥٦','123'],['-1','2'],['2','10'],['0','1']], 'amount':['0','-1','0.0000001','1.123456','1e30','NaN'],'direction':['SELL','buy','unknown'],'side':['NO','yes','']}.items():
 for value in values:add('request',key+'/'+repr(value),base|{key:value})
add('request','sell',base|{'direction':'SELL','amount':None,'size':'2'})
add('quote','default',quote)
for path in [[],['request'],['quote']]:
 target=quote if not path else quote[path[0]]
 for key in target:
  for value in [None,'',{},True]:
   changed=copy.deepcopy(quote);current=changed if not path else changed[path[0]];current[key]=value;add('quote','/'.join(path+[key])+repr(value),changed)
for code in ['NO_QUOTES','SIZE_TOO_LARGE','UNKNOWN']:
 add('quote','no-quote/'+code,{'rfq_id':'rfq-1','status':'FAILED','error':{'code':code,'message':'no usable quote'}})
for status in ['EXECUTING','MATCHED','CONFIRMED','FILLED','FAILED','EXPIRED','CANCELED','bad']:
 body={'rfq_id':'rfq-1','status':status,'tx_hash':'0xhash'};add('status',status,body)
for change in [{'rfq_id':'other'},{'tx_hash':None},{'tx_hash':''},{'taker_order_hash':None},{'error':{'code':'future','message':'message'}},{'error':{'code':'future'}}]:add('status',repr(change),{'rfq_id':'rfq-1','status':'EXECUTING'}|change)
accepted=normalize(combo._parse_combo_quote_result(quote,request=request).quote)
for key in accepted:
 for value in [None,{},'']:
  add('model',key+'/'+repr(value),accepted|{key:value})
 missing=accepted.copy();missing.pop(key);add('model','missing/'+key,missing)
for value in ['0','1000000','1000000000000000000000000000000000000']:
 changed=copy.deepcopy(quote);changed['quote']['maker_amount_e6']=value;add('quote','decimal/'+value,changed)

for kind,wallet in [('EOA',signer.address),('POLY_PROXY',derive_proxy_wallet_address(signer.address,config.wallet_derivation)),('GNOSIS_SAFE',derive_safe_wallet_address(signer.address,config.wallet_derivation)),('DEPOSIT_WALLET',derive_beacon_deposit_wallet_address(signer.address,config.wallet_derivation))]:
 ctx.wallet_type=kind;ctx.wallet=wallet
 for direction in ['BUY','SELL']:
  value=accepted|{'direction':direction,'net_receive':'9' if direction=='SELL' else None}
  before=len(cases);add('accept',kind+'/'+direction,value);cases[-1]['wallet']=wallet;cases[-1]['wallet_type']=kind
  if 'error' in cases[-1]['expected']:raise AssertionError(cases[-1])
(ROOT/'testdata/combo-parity.json').write_text(json.dumps({'upstream_version':manifest['version'],'upstream_commit':manifest['commit'],'request':normalize(request),'cases':cases},indent=2)+'\n')
print('Generated',len(cases),'combo cases')
