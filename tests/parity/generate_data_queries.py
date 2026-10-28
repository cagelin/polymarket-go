"""Execute all public Data methods and retain request/response/cursor expectations."""
import copy,inspect,json,base64
from types import SimpleNamespace
import httpx
from support import ROOT,manifest,normalize
from polymarket.clients.public import PublicClient
ops=json.loads((ROOT/'tests/parity/data-operations.json').read_text())
models=json.loads((ROOT/'testdata/data-models-parity.json').read_text())['cases']
baselines={case['model']:case['response'] for case in models if case['name'].endswith('/full')}
condition='0x'+'ab'*32
combo='0x03'+'ab'*30
cases=[]
class Transport:
 def get_json(self,path,params=None):
  case['request']={'path':path,'query':dict(httpx.QueryParams(params or {}).multi_items())}
  case['request']['query']={key:httpx.QueryParams(params or {}).get_list(key) for key in (params or {})}
  return copy.deepcopy(case['response'])
 def get_bytes(self,path,params=None):
  self.get_json(path,params);return b'accounting data'
def add(op,label,inp,response,cursor=None):
 global case
 case={'name':op['op']+'/'+label,'operation':op['op'],'input':inp,'response':response}
 if cursor is not None:case['cursor']=cursor
 client=object.__new__(PublicClient);client._ctx=SimpleNamespace(data=Transport())
 try:
  value=getattr(client,op['op'])(**copy.deepcopy(inp))
  if op['paginated']:
   if cursor is not None:value=value.from_cursor(cursor)
   value=value.first_page()
  if isinstance(value,bytes):value=base64.b64encode(value).decode()
  case['expected']={'value':normalize(value)}
 except Exception as err:case['expected']={'error':type(err).__name__}
 cases.append(case)
 return case
for op in ops:
 name=op['op'];params=inspect.signature(getattr(PublicClient,name)).parameters
 inp={}
 if 'user' in params and params['user'].default is inspect.Parameter.empty:inp['user']='wallet'
 if name=='list_positions':inp['user']='wallet'
 if name in ['list_market_holders','get_token_references','get_resolutions']:inp['condition_ids']=[condition]
 if name=='get_event_live_volume':inp['event_ids']=[1]
 typ=op['output'].removeprefix('[]');typ={'Activity':'TradeActivity','ComboActivity':'ComboSplitActivity','BiggestWinner':'MarketBiggestWinner'}.get(typ,typ)
 value=baselines.get(typ,{})
 data=[value] if op['paginated'] or op['output'].startswith('[]') else value
 response={'data':data}
 if op['paginated']:response['pagination']={'has_more':True,'next_cursor':'next/ token'}
 result=add(op,'default',inp,response)
 if 'error' in result['expected']:raise AssertionError(result)
 if op['paginated']:
  cursor=result['expected']['value']['next_cursor']
  add(op,'resume',inp|{'page_size':2},response,cursor)
  add(op,'bad-cursor',inp,response,'invalid')
  for meta in [{},{'has_more':False},{'has_more':True,'next_cursor':None},{'has_more':False,'next_cursor':'extra'},{'has_more':1,'next_cursor':'x'},{'has_more':True,'next_cursor':''}]:add(op,'pagination/'+repr(meta),inp,{'data':data,'pagination':meta})
 for raw in [None,[],{}, {'data':None},{'data':{}},{'data':[]}, {'data':[None]}]:add(op,'shape/'+repr(raw),inp,raw)
 for field in params:
  values=[]
  if field in ['self']:continue
  if field=='user':values=['','wallet']
  elif field.startswith('condition_id'):values=[[],[condition],[condition,condition.upper().replace('0X','0x')],[combo],[combo+'01'],['bad']]
  elif field.startswith('event_id'):values=[[],[0],[1,1,2],[2147483648]]
  elif field=='asset_ids':values=[[],['00012','12','  13  ',''],['x'],['1'*79]]
  elif field=='question_id':values=['','question']
  elif field in ['start','end','updated_after','updated_before']:values=[0,1,253402300800]
  elif field=='full_history':values=[True]
  elif field in ['page_size','bucket_limit']:values=[0,1,1001]
  elif field in ['include_pnl','taker_only','include_archived']:values=[False,True]
  elif field in ['filter_amount','min_balance']:values=[-1.0,0.0,1.25]
  elif field=='activity_types':values=[[],['TRADE','DEPOSIT'],['bad']]
  elif field=='status':values=[['OPEN','OPEN'],[],['REDEEMABLE','OPEN'],['bad']] if name=='list_combo_positions' else ['OPEN','MERGEABLE','CLOSED','bad']
  elif field=='side':values=['BUY','SELL','bad']
  elif field=='title':values=['  ','a'*201,' a ']
  elif field=='category':values=['','sports']
  elif field=='sort_direction':values=['ASC','DESC','bad']
  elif field=='sort_by':values=['CURRENT_VALUE','PNL','bad']
  elif field=='filter_type':values=['TOKENS','CASH','bad']
  elif field in ['interval','window','fidelity']:values=['day','1d','bad']
  for val in values:add(op,field+'/'+repr(val),inp|{field:val},response)
 if 'full_history' in params:add(op,'full-history-conflict',inp|{'full_history':True,'start':1},response)
 if name=='list_combo_positions':add(op,'watermarks-reversed',inp|{'updated_after':2,'updated_before':1},response)
 if name=='list_positions':
  for val in [{'condition_id':[condition]},{'condition_id':[condition],'status':'MERGEABLE'},{'user':'wallet','status':'CLOSED','include_archived':True}]:add(op,'constraints/'+repr(val),val,response)
(ROOT/'testdata/data-queries-parity.json').write_text(json.dumps({'upstream_version':manifest['version'],'upstream_commit':manifest['commit'],'cases':cases},indent=2)+'\n')
print('Generated',len(cases),'Data request cases')
# Generate statically typed Go invocation, leaving common assertions in the SDK suite.
text='''package polymarket
import ("context"; "encoding/json"; "fmt")
func callDataParity(ctx context.Context,c *PublicClient,operation string,input json.RawMessage,cursor *string)(any,error){
switch operation {
'''
for op in ops:
 text+=f'case "{op["op"]}":\nvar p {op["params"]};if err:=json.Unmarshal(input,&p);err!=nil{{return nil,err}}\n'
 if op['paginated']:text+=f'pager,err:=c.{op["method"]}(p);if err!=nil{{return nil,err}};if cursor!=nil{{pager=pager.FromCursor(cursor)}};return pager.FirstPage(ctx)\n'
 else:text+=f'return c.{op["method"]}(ctx,p)\n'
text+='default:return nil,fmt.Errorf("unknown operation %s",operation)\n}}\n'
(ROOT/'data_dispatch_test.go').write_text(text)
