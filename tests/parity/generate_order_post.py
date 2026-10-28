"""Execute pinned order payload construction and accepted/rejected normalization."""
import json,copy
from support import ROOT,manifest,normalize
from polymarket.models.clob.orders import SignedOrder
from polymarket._internal.actions.orders.post import build_post_order_request,parse_order_response
from polymarket._internal.hmac import build_hmac_signature
source=json.loads((ROOT/'testdata/signing-parity.json').read_text())['cases']
cases=[]
for item in source:
 if item['operation']!='order':continue
 raw=copy.deepcopy(item['order']);raw['token_id']=raw.pop('asset_id');raw['signature']=item['expected']
 raw={key:value for key,value in raw.items() if key in SignedOrder.__dataclass_fields__}
 for post_only in [False,True]:
  raw['post_only']=post_only;order=SignedOrder(**raw);_,body=build_post_order_request(order,owner_api_key='key')
  canonical=json.dumps(body,sort_keys=True,separators=(',',':'),ensure_ascii=False)
  go=normalize(order);go['asset_id']=go.pop('token_id')
  cases.append({'order':go,'body':body,'signature':build_hmac_signature(secret='c2VjcmV0',timestamp=1720000000,method='POST',path='/order',body=canonical)})
responses=[]
base={'errorMsg':'','makingAmount':'1.234','takingAmount':'2.345','orderID':'order','success':True,'status':'live'}
for change in [{},{'status':'matched'},{'status':'delayed'},{'status':'unknown'},{'success':False},{'orderID':''},{'makingAmount':''},{'takingAmount':''},{'tradeIDs':['trade'],'transactionsHashes':['hash']},{'makingAmount':1},{'success':None}]:
 raw=base|change;case={'response':raw}
 try:case['expected']={'value':normalize(parse_order_response(raw))}
 except Exception as err:case['expected']={'error':type(err).__name__}
 responses.append(case)
from polymarket.models.clob.order_response import _ERROR_MSG_TO_CODE
for message in list(_ERROR_MSG_TO_CODE)+['not enough balance / allowance for this order','unknown failure','']:
 for status in ['','unmatched']:
  raw=base|{'errorMsg':message,'success':False,'status':status};responses.append({'response':raw,'expected':{'value':normalize(parse_order_response(raw))}})
# Run the complete settlement poller with a deterministic in-memory transport.
from types import SimpleNamespace
from polymarket._internal.actions.orders.settlement import wait_for_order_fill_settlement_sync
baseline=next(c['response'] for c in json.loads((ROOT/'testdata/data-models-parity.json').read_text())['cases'] if c['name']=='ClobTrade/full')
settlements=[]
for states in [[],['CONFIRMED'],['FAILED'],['MINED'],['CONFIRMED','FAILED'],['CONFIRMED','CONFIRMED'],['missing-hash']]:
 trades=[]
 for i,state in enumerate(states):
  trade=baseline|{'id':str(i),'status':'CONFIRMED' if state=='missing-hash' else state,'transaction_hash':None if state=='missing-hash' else 'hash'}
  trades.append(trade)
 ids=[t['id'] for t in trades];ids=ids+ids
 order=SimpleNamespace(order_id='order',trade_ids=ids,transactions_hashes=['existing'])
 class SettlementTransport:
  def get_json(self,path,params):
   return {'data':[t for t in trades if t['id']==params['id']],'next_cursor':'LTE='}
 case={'states':states,'trades':trades,'trade_ids':ids,'existing':['existing']}
 try:case['expected']={'value':normalize(wait_for_order_fill_settlement_sync(SettlementTransport(),order,timeout_s=0))}
 except Exception as err:case['expected']={'error':type(err).__name__}
 if states==['missing-hash']:
  assert case['expected']['value']==[None]
  case['python_value']=case['expected']['value'];case['difference']='UP-002'
  case['expected']={'error':'UnexpectedResponseError'}
 settlements.append(case)
(ROOT/'testdata/order-post-parity.json').write_text(json.dumps({'upstream_version':manifest['version'],'upstream_commit':manifest['commit'],'cases':cases,'responses':responses,'settlements':settlements},indent=2)+'\n')
print('Generated',len(cases)+len(responses)+len(settlements),'order post cases')
