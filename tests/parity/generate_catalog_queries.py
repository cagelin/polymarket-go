"""Record the pinned client's catalog and comment behavior, including offset limits."""
import copy,inspect,json,base64
from types import SimpleNamespace
import httpx
from support import ROOT,manifest,normalize
from polymarket.clients.public import PublicClient
from polymarket.clients._transport import _clean_params
ops=['get_series','get_tag','get_related_tags','get_related_tag_resources','get_sports','get_sports_market_types','get_public_profile','get_comment_thread','list_series','list_tags','list_teams','list_comments','list_comments_by_user_address']
outputs=['Series','Tag','RelatedTag','Tag','SportsMetadata','SportsMarketTypes','PublicProfile','Comment','Series','Tag','Team','Comment','Comment']
models=json.loads((ROOT/'testdata/data-models-parity.json').read_text())['cases']
baselines={case['model']:case['response'] for case in models if case['name'].endswith('/full')}
baselines['Team']={'id':1}
baselines['SportsMetadata']={'id':1,'sport':'nba','image':'image','resolution':'source','ordering':'1','tags':'1','series':'1'}
cases=[]
class Transport:
 def get_json(self,path,params=None):
  params=_clean_params(params)
  case['request']={'path':path,'query':{key:httpx.QueryParams(params or {}).get_list(key) for key in (params or {})}}
  return copy.deepcopy(case['response'])
def add(op,label,inp,response,cursor=None):
 global case
 case={'name':op+'/'+label,'operation':op,'input':copy.deepcopy(inp),'response':response}
 if cursor is not None:case['cursor']=cursor
 client=object.__new__(PublicClient);client._ctx=SimpleNamespace(gamma=Transport())
 try:
  value=getattr(client,op)(**copy.deepcopy(inp))
  if op.startswith('list_'):
   if cursor is not None:value=value.from_cursor(cursor)
   value=value.first_page()
  case['expected']={'value':normalize(value)}
 except Exception as err:case['expected']={'error':type(err).__name__}
 cases.append(case);return case
for op,typ in zip(ops,outputs):
 params=inspect.signature(getattr(PublicClient,op)).parameters
 inp={}
 if op in ['get_series','get_tag','get_related_tags','get_related_tag_resources','get_comment_thread']:inp['id']='123'
 if 'address' in params:inp['address']='wallet'
 if op=='list_comments':inp.update(parent_entity_id='123',parent_entity_type='Event')
 item=copy.deepcopy(baselines.get(typ,{}))
 if typ=='Comment':item['parentCommentID']=None
 response=[item] if op.startswith('list_') or op in ['get_related_tags','get_related_tag_resources','get_sports','get_comment_thread'] else item
 if op=='list_comments':response={'comments':[item],'next_cursor':'next'}
 result=add(op,'default',inp,response)
 if 'error' in result['expected']:raise AssertionError(result)
 if op.startswith('list_'):
  filled=[item]*2;page={'comments':filled,'next_cursor':'next'} if op=='list_comments' else filled
  initial=add(op,'full-page',inp|{'page_size':2},page)
  cursor=initial['expected']['value']['next_cursor'];assert cursor
  add(op,'resume',inp|{'page_size':1},page,cursor)
  add(op,'bad-cursor',inp,page,'invalid')
 for raw in [None,[],{},[None]]:add(op,'shape/'+repr(raw),inp,raw)
 for field in params:
  values=[]
  if field in ['id','address','parent_entity_id']:values=['','a/b']
  elif field=='parent_entity_type':values=['Series','bad']
  elif field=='slug':values=[[],['one','two']] if op=='list_series' else ['','one']
  elif field=='page_size':values=[0,1,101]
  elif field in ['ascending','closed','exclude_events','include_template','omit_empty','is_carousel','get_positions','holders_only']:values=[False,True]
  elif field=='recurrence':values=['daily','bad']
  elif field in ['locale','status','order']:values=['','createdAt','id','likes'] if field=='order' else ['','en']
  elif field in ['abbreviation','league','name']:values=[[],['one','two']]
  elif field=='provider_ids':values=[[],[1,2]]
  for val in values:
   updated=inp|{field:val}
   # Comment protocol depends on filters, as in the real client.
   r=response
   if op=='list_comments' and (updated.get('get_positions') or updated.get('holders_only') or (updated.get('order') is not None and updated['order'] not in ('id','createdAt'))):r=[item]
   add(op,field+'/'+repr(val),updated,r)
 if 'slug' in params and op!='list_series':
  add(op,'slug-lookup',{'slug':'tag'},response)
  add(op,'empty-selector',{},response)
 if op in ['list_comments','list_comments_by_user_address']:
  selected=inp|{'page_size':100}
  if op=='list_comments':selected['get_positions']=True
  first=add(op,'cap-first',selected,[item]*100)
  cursor=first['expected']['value']['next_cursor']
  second=add(op,'cap-second',selected,[item]*100,cursor)
  third=add(op,'cap-last',selected,[item]*100,second['expected']['value']['next_cursor'])
  add(op,'cap-exceeded',selected,[item],third['expected']['value']['next_cursor'])
(ROOT/'testdata/catalog-queries-parity.json').write_text(json.dumps({'upstream_version':manifest['version'],'upstream_commit':manifest['commit'],'cases':cases},indent=2)+'\n')
print('Generated',len(cases),'catalog request cases')
text='''package polymarket
import ("context";"encoding/json";"fmt")
func callCatalogParity(ctx context.Context,c *PublicClient,operation string,input json.RawMessage,cursor *string)(any,error){
switch operation {
'''
for op in ops:
 method=''.join(v.title() for v in op.split('_'))
 custom={
 'get_series':('struct { ID SeriesID; Locale *string }','c.GetSeries(ctx,p.ID,p.Locale)'),
 'get_public_profile':('struct { Address EVMAddress }','c.GetPublicProfile(ctx,p.Address)'),
 'get_comment_thread':('struct { ID CommentID; GetPositions *bool `json:"get_positions"` }','c.GetCommentThread(ctx,p.ID,p.GetPositions)'),
 'get_tag':('TagQuery','c.GetTag(ctx,p)'),
 'get_related_tags':('RelatedTagsQuery','c.GetRelatedTags(ctx,p)'),
 'get_related_tag_resources':('RelatedTagResourcesQuery','c.GetRelatedTagResources(ctx,p)'),
 }
 text+=f'case "{op}":\n'
 if op in ['get_sports','get_sports_market_types']:text+=f'return c.{method}(ctx)\n';continue
 typ,call=custom.get(op,(method+'Params',None))
 text+=f'var p {typ};if err:=json.Unmarshal(input,&p);err!=nil{{return nil,err}}\n'
 if op.startswith('list_'):text+=f'pager,err:=c.{method}(p);if err!=nil{{return nil,err}};if cursor!=nil{{pager=pager.FromCursor(cursor)}};return pager.FirstPage(ctx)\n'
 else:text+=f'return {call}\n'
text+='default:return nil,fmt.Errorf("unknown operation %s",operation)\n}}\n'
(ROOT/'catalog_dispatch_test.go').write_text(text)
