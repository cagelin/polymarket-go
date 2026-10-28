"""Execute authenticated Python account calls against an in-memory transport."""
import copy,inspect,json
from types import SimpleNamespace
import httpx
from support import ROOT,manifest,normalize
from polymarket.clients.secure import SecureClient
from polymarket.clients._transport import _clean_params
from polymarket._internal.wallet import WalletType
baselines={c['model']:c['response'] for c in json.loads((ROOT/'testdata/data-models-parity.json').read_text())['cases'] if c['name'].endswith('/full')}
ops={
'fetch_api_keys':({}, {'apiKeys':['key1','key2']}),
'delete_api_key':({},'OK'),
'create_builder_api_key':({}, {'key':'builder','secret':'c2VjcmV0','passphrase':'pass'}),
'fetch_builder_api_keys':({},['builder',{'key':'builder2'}]),
'get_closed_only_mode':({}, {'closed_only':True}),
'list_open_orders':({}, {'data':[baselines['OpenOrder']],'next_cursor':'next'}),
'get_order':({'order_id':'order-1'},baselines['OpenOrder']),
'list_account_trades':({}, {'data':[baselines['ClobTrade']],'next_cursor':'next'}),
'get_notifications':({},[baselines['OrderFillNotification']]),
'drop_notifications':({'ids':['1','2']},None),
'get_balance_allowance':({'asset_type':'COLLATERAL'},{'balance':'123456789012345678901234567890','allowances':{'address':'999'}}),
'send_order_heartbeat':({}, {'heartbeat_id':'next'}),
'cancel_order':({'order_id':'order-1'},{'canceled':['order-1'],'not_canceled':{}}),
'cancel_orders':({'order_ids':['a','b']},{'canceled':['a','b'],'not_canceled':{}}),
'cancel_all':({}, {'canceled':[],'not_canceled':{}}),
'cancel_market_orders':({'market':'market'},{'canceled':[],'not_canceled':{}}),
'get_order_scoring':({'order_id':'order-1'},{'scoring':True}),
'get_orders_scoring':({'order_ids':['a','b']},{'a':True,'b':False}),
'list_user_earnings_for_day':({'date':'2026-10-28'},{'data':[baselines['UserEarning']],'next_cursor':'next'}),
'get_total_earnings_for_user_for_day':({'date':'2026-10-28'},[baselines['TotalUserEarning']]),
'list_user_earnings_and_markets_config':({'date':'2026-10-28'},{'data':[baselines['UserRewardsEarning']],'next_cursor':'next','total_count':10}),
'get_reward_percentages':({}, {'condition':1.5}),
}
cases=[]
class Transport:
 def request(self,method,path,params=None,json=None,**kwargs):
  query=httpx.QueryParams(_clean_params(params) or {})
  case['request']={'method':method,'path':path,'query':{key:query.get_list(key) for key in query},'body':json}
  return copy.deepcopy(case['response'])
 def get_json(self,path,params=None):return self.request('GET',path,params)
 def post_json(self,path,json=None):return self.request('POST',path,json=json)
 def delete_json(self,path,json=None):return self.request('DELETE',path,json=json)
 def delete(self,path,params=None):return self.request('DELETE',path,params)
def add(op,label,inp,response,cursor=None):
 global case
 case={'name':op+'/'+label,'operation':op,'input':inp,'response':response}
 if cursor is not None:case['cursor']=cursor
 client=object.__new__(SecureClient);client._ended=False;client._ctx=SimpleNamespace(secure_clob=Transport(),wallet_type='EOA')
 try:
  value=getattr(client,op)(**copy.deepcopy(inp))
  if op.startswith('list_'):
   if cursor is not None:value=value.from_cursor(cursor)
   value=value.first_page()
  case['expected']={'value':normalize(value)}
 except Exception as err:case['expected']={'error':type(err).__name__}
 cases.append(case);return case
for op,(inp,response) in ops.items():
 first=add(op,'default',inp,response)
 if 'error' in first['expected']:raise AssertionError(first)
 if op.startswith('list_'):
  add(op,'resume',inp,response,'next');add(op,'empty-cursor',inp,response,'')
  for cursor in [None,'','LTE=',5]:add(op,'next/'+repr(cursor),inp,response|{'next_cursor':cursor})
 for raw in [None,{},[],{'data':None}]:add(op,'response/'+repr(raw),inp,raw)
 for field in inspect.signature(getattr(SecureClient,op)).parameters:
  if field in ['self','token_id']:continue
  values=[None,'','sample']
  if field=='date':values=['20260101','2026-W01-1','2026W01','2026-02-30','0000-01-01']
  elif field=='asset_type':values=['CONDITIONAL','CONDITIONAL-V2','bad']
  elif field in ['ids','order_ids']:values=[[],[''],[' ','2']]
  elif field in ['sponsored','no_competition']:values=[True,False]
  elif field=='page_size':values=[0,1,500,501]
  for value in values:
   # Go's nonoptional string has no None value; the empty default is covered.
   if field=='heartbeat_id' and value is None:continue
   add(op,field+'/'+repr(value),inp|{field:value},response)
(ROOT/'testdata/account-queries-parity.json').write_text(json.dumps({'upstream_version':manifest['version'],'upstream_commit':manifest['commit'],'cases':cases},indent=2)+'\n')
print('Generated',len(cases),'account request cases')
