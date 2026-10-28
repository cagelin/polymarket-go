"""Run public rewards, builder and combo catalog calls against the pinned SDK."""
import copy,inspect,json
from types import SimpleNamespace
import httpx
from support import ROOT,manifest,normalize
from polymarket.clients.public import PublicClient
from polymarket.clients._transport import _clean_params
ops={'list_builder_trades':'BuilderTrade','list_reward_markets':'RewardMarket','list_current_rewards':'CurrentReward','list_market_rewards':'MarketReward','get_current_rebates':'RebatedFee','list_combo_markets':'ComboMarket'}
baselines={c['model']:c['response'] for c in json.loads((ROOT/'testdata/data-models-parity.json').read_text())['cases'] if c['name'].endswith('/full')}
cases=[]
class Transport:
 def get_json(self,path,params=None):
  params=_clean_params(params)
  query=httpx.QueryParams(str(httpx.QueryParams(params or {})))
  case['request']={'path':path,'query':{key:query.get_list(key) for key in query}}
  return copy.deepcopy(case['response'])
def add(op,label,inp,response,cursor=None):
 global case
 case={'name':op+'/'+label,'operation':op,'input':inp,'response':response}
 if cursor is not None:case['cursor']=cursor
 client=object.__new__(PublicClient);client._ctx=SimpleNamespace(clob=Transport(),rfq=Transport())
 try:
  value=getattr(client,op)(**copy.deepcopy(inp))
  if op.startswith('list_'):
   if cursor is not None:value=value.from_cursor(cursor)
   value=value.first_page()
  case['expected']={'value':normalize(value)}
 except Exception as err:case['expected']={'error':type(err).__name__}
 cases.append(case);return case
for op,typ in ops.items():
 inp={};params=inspect.signature(getattr(PublicClient,op)).parameters
 if op=='list_builder_trades':inp['builder_code']='0x'+'ab'*32
 if op=='list_market_rewards':inp['condition_id']='0x'+'ab'*32
 if op=='get_current_rebates':inp={'date':'2026-10-28','maker_address':'0x'+'ab'*20}
 item=baselines[typ]
 response={'data':[item],'next_cursor':'next'}
 if op=='list_combo_markets':response={'markets':[item],'next_cursor':'next'}
 if op=='get_current_rebates':response=[item]
 first=add(op,'default',inp,response)
 if 'error' in first['expected']:raise AssertionError(first)
 if op.startswith('list_'):
  add(op,'resume',inp,response,first['expected']['value']['next_cursor'])
  add(op,'empty-cursor',inp,response,'')
  for value in [None,'','LTE=',7]:add(op,'cursor/'+repr(value),inp,response|{'next_cursor':value})
 for raw in [None,{},[],{'data':None},{'data':[None]}]:add(op,'shape/'+repr(raw),inp,raw)
 for field in params:
  values=[]
  if field in ['token_id','self']:continue
  if field in ['builder_code','condition_id','maker_address']:values=['','0xab','0x'+'00'*32]
  elif field=='date':values=['2026-02-30','20260101','2026-01-01']
  elif field=='page_size':values=[0,1,501]
  elif field=='market_id':values=[-1,1]
  elif field=='event_ids':values=[[],[1,2],[0]]
  elif field in ['tag_slugs','exclude']:values=[[],['a','b'],['']]
  elif field=='sponsored':values=[True,False]
  elif field=='sort_direction':values=['ASC','DESC','bad']
  elif field=='order_by':values=['price','bad']
  elif field.startswith(('min_','max_')):values=[-1.0,0.0,1.25]
  else:values=['','sample']
  for value in values:add(op,field+'/'+repr(value),inp|{field:value},response)
(ROOT/'testdata/public-clob-queries-parity.json').write_text(json.dumps({'upstream_version':manifest['version'],'upstream_commit':manifest['commit'],'cases':cases},indent=2)+'\n')
print('Generated',len(cases),'public CLOB request cases')
text='''package polymarket
import ("context";"encoding/json";"fmt")
func callPublicClobParity(ctx context.Context,c *PublicClient,operation string,input json.RawMessage,cursor *string)(any,error){
switch operation {
'''
for op in ops:
 method=''.join(v.title() for v in op.split('_'))
 text+=f'case "{op}":\nvar p {method}Params;if err:=json.Unmarshal(input,&p);err!=nil{{return nil,err}}\n'
 if op.startswith('list_'):text+=f'pager,err:=c.{method}(p);if err!=nil{{return nil,err}};if cursor!=nil{{pager=pager.FromCursor(cursor)}};return pager.FirstPage(ctx)\n'
 else:text+=f'return c.{method}(ctx,p)\n'
text+='default:return nil,fmt.Errorf("unknown operation %s",operation)\n}}\n'
(ROOT/'public_clob_dispatch_test.go').write_text(text)
