"""Generate market, pagination and history fixtures using verified Python source."""
import copy
import hashlib
import importlib.metadata
import json
from dataclasses import asdict, is_dataclass
from datetime import datetime, date, timezone
from decimal import Decimal
from pathlib import Path
from types import SimpleNamespace
import httpx
import polymarket
from pydantic import BaseModel
from polymarket._internal.actions import gamma, data
from polymarket._internal.dispatch import sync_paginate_keyset
from polymarket.pagination import Page, Paginator

ROOT=Path(__file__).resolve().parents[2]
manifest=json.loads((ROOT/'tests/parity/upstream.json').read_text())
assert importlib.metadata.version('polymarket-client')==manifest['version']
source=Path(polymarket.__file__).parent
for entry in manifest['source_hashes']:
 assert hashlib.sha256((source/Path(entry['path']).relative_to('src/polymarket')).read_bytes()).hexdigest()==entry['sha256']

def normalize(value):
 if isinstance(value,Decimal): return str(value)
 if isinstance(value,datetime):
  value=value.replace(tzinfo=timezone.utc) if value.tzinfo is None else value.astimezone(timezone.utc)
  return value.strftime('%Y-%m-%dT%H:%M:%S')+('.'+f'{value.microsecond:06d}'.rstrip('0') if value.microsecond else '')+'Z'
 if isinstance(value,date): return value.isoformat()
 if isinstance(value,BaseModel):
  # Only the deprecated market-level position_ids duplicates outcome fields.
  return {key:normalize(getattr(value,key)) for key,field in type(value).model_fields.items() if not field.deprecated}
 if is_dataclass(value): return {key:normalize(item) for key,item in asdict(value).items()}
 if isinstance(value,dict): return {key:normalize(item) for key,item in value.items()}
 if isinstance(value,(list,tuple)): return [normalize(item) for item in value]
 return value

cases=[]
def add(name,op,args,response,cursor=None):
 case=dict(name=name,operation=op,input=copy.deepcopy(args),response=copy.deepcopy(response),request=None)
 if cursor is not None: case['cursor']=cursor
 class Transport:
  def get_json(self,path,params=None):
   case['request']=dict(path=path,query={key:httpx.QueryParams(params).get_list(key) for key in httpx.QueryParams(params)})
   return copy.deepcopy(response)
 transport=Transport()
 ctx=SimpleNamespace(gamma=transport,data=transport)
 try:
  if op=='get_market':
   spec=gamma.get_market_spec(**(dict(id=None,slug=None,url=None,include_tag=None,locale=None)|args))
   value=spec.parse(transport.get_json(spec.path,params={k:v for k,v in spec.params.items() if v is not None}))
  else:
   kwargs=copy.deepcopy(args)
   if op=='list_markets':
    size=kwargs.pop('page_size',20)
    spec=gamma.list_markets_spec(**kwargs)
   else:
    size=kwargs.get('page_size',10000)
    spec=data.build_list_price_history_spec(**kwargs)
   pager=sync_paginate_keyset(ctx,spec,page_size=size)
   if cursor is not None: pager=pager.from_cursor(cursor)
   value=pager.first_page()
  case['expected']={'value':normalize(value)}
 except Exception as error: case['expected']={'error':type(error).__name__}
 cases.append(case)
 return case

market=dict(id='42',slug='example',question='Will it happen?',outcomes='["Yes","No"]',outcomePrices='["0.55","0.45"]',clobTokenIds='["123","456"]',conditionId='0x'+'ab'*32,active=True,closed=False)
full=market|dict(version='v2',groupItemTitle='Test',description='Description',category='Science',image='image',icon='icon',
 comboStatus='new_future_status',acceptingOrders=True,enableOrderBook='true',negRisk=False,startDate='2026-01-01 12:30:00+00',endDate='2026-10-28T20:00:00+08:00',closedTime='',
 volume='1000000000000000000000000000.123456789',volumeNum=123.5,volume24hr='2',volume1wk='3',volume1mo='4',volume1yr='5',volumeAmm='6',volumeClob='7',liquidity='8',liquidityNum=9,liquidityClob='10',
 bestBid='0.54',bestAsk='0.56',lastTradePrice=0.55,spread='0.02',oneHourPriceChange='0.01',oneDayPriceChange='0.02',oneWeekPriceChange='0.03',oneMonthPriceChange='0.04',oneYearPriceChange='0.05',
 orderMinSize='5',orderPriceMinTickSize='0.01',secondsDelay='2',feesEnabled=True,feeType='test',feeSchedule=dict(exponent=2,rate=0.01,takerOnly='true',rebateRate='0.001'),
 onchainEventId='0x'+'ab'*29+'000000',requestId='0x'+'cd'*31+'00',resolutionStatus='active',questionID='question',negRiskRequestID='negrequest',umaResolutionStatus='proposed',resolutionSource='source',resolvedBy='0xowner',
 clobRewards=[dict(id="7",conditionId='0xab',assetAddress='asset',rewardsAmount='10',rewardsDailyRate=0.1,startDate='2026-01-01',endDate=None)],rewardsMinSize='5',rewardsMaxSpread='0.2',holdingRewardsEnabled=True,
 sportsMarketType='winner',line=1.5,gameId='game',gameStartTime='2026-10-28T12:00:00Z',events=[dict(id=7,slug='event',title='Event')],tags='[{"id":"8","slug":"tag","label":"Tag"}]',positionIds='["987","654"]')
add('market/full','get_market',{'id':'42'},full)
add('market/minimal','get_market',{'id':'42'},dict(id='42',outcomes=['A','B']))
for args in [{'slug':'a/b +中文'},{'url':'https://polymarket.com/event/e/m?x=1#f'},{'url':'https://www.polymarket.com/market/a%2Fb'},{'url':'https://polymarket.com/event/e'},{'id':'42','include_tag':False,'locale':''},{},{'id':'','slug':'s'},{'id':'1','slug':'2'},{'url':'http://polymarket.com/market/a'},{'url':'https://evil.test/market/a'},{'url':'https://polymarket.com/events/a'},{'url':'https://polymarket.com:443/market/a'}]:
 add('lookup/'+repr(args),'get_market',args,market)
for key,values in {
 'id':[None,42], 'outcomes':[None,[],['A'],['A','B','C'],'bad',[1,2]],
 'outcomePrices':['bad',[True],[None],['NaN'],[1,0],['.2','.8','bad']],
 'clobTokenIds':[None,[123]], 'positionIds':[[123]], 'conditionId':['',None,'0xab','bad'],
 'version':['v3','',None], 'active':['yes',2], 'volume':[True,'NaN','',None,0.000001,1e20],
 'startDate':['bad',1,'2026-01-02','2026-01-02T12:30:00.123456789Z','2026-01-02T12:30+08:00'], 'resolutionStatus':['future',''], 'umaResolutionStatus':['future',''],
 'onchainEventId':['bad',''], 'requestId':['bad',''], 'events':[[None],[{'id':7}],{'id':'7'}],
 'tags':[[{'id':7}]], 'feeSchedule':[{},dict(exponent='2',rate='0.1',takerOnly=False,rebateRate='0')],
 'clobRewards':[[{}]], 'secondsDelay':['2.0',1.5],
}.items():
 for value in values: add('market/'+key+'/'+repr(value),'get_market',{'id':'42'},market|{key:value})
normalized=normalize(gamma.get_market_spec(id='42',slug=None,url=None,include_tag=None,locale=None).parse(full))
add('market/normalized','get_market',{'id':'42'},normalized)
for key in ['state','outcomes','metrics','prices','trading','resolution','rewards','sports','events','tags']:
 copy_market=copy.deepcopy(normalized);copy_market[key]=None
 add('market/normalized_null/'+key,'get_market',{'id':'42'},copy_market)

page=dict(markets=[market],next_cursor='cursor-中文😀/+')
all_filters=dict(ascending=False,closed=False,clob_token_ids=['123','456'],condition_ids=['0xab'],cyom=True,decimalized=False,end_date_max='2026-01-01',end_date_min='',game_id='g',ids=[1,2],include_tag=True,liquidity_num_max=1.0,liquidity_num_min=0.000001,locale='中文',market_maker_addresses=['a','b'],order=' volume, liquidity ,endDate ',onchain_event_ids=['event'],position_ids=['987'],question_ids=['q'],related_tags=False,request_ids=['r'],resolution_status='active',rfq_enabled=True,rewards_min_size=0.0,slug=['<>&é😀','x'],sports_market_types=['winner'],start_date_max='max',start_date_min='min',tag_id=0,tag_match='all',uma_resolution_status='requested',version=['v1','v2'],volume_num_max=1e16,volume_num_min=1e-7,page_size=2)
first=add('markets/all_filters','list_markets',all_filters,page)
next_cursor=first['expected']['value']['next_cursor']
add('markets/resume','list_markets',all_filters,dict(markets=[],next_cursor=None),next_cursor)
add('markets/wrong_query_cursor','list_markets',{'closed':True},page,next_cursor)
for args in [{},{'page_size':0},{'page_size':-1},{'page_size':5000},{'version':[]},{'version':['v1','v2','v1']},{'version':['bad']},{'version':['v2']},{'tag_match':'bad'},{'resolution_status':'bad'},{'request_ids':['x']*101},{'onchain_event_ids':['x']*101},{'slug':[]},{'order':'volume'},{'order':' liquidity , unknown '},{'page_size':2}]:
 add('markets/params/'+repr(args),'list_markets',args,page)
for response in [None,[],{},dict(markets=[]),dict(markets=[],next_cursor=''),dict(markets=[],next_cursor=2),dict(markets=None),dict(markets=[None]),dict(markets=[{'outcomes':['one','two','three']}],next_cursor='next'),dict(markets=[market,{'outcomes':[]},market]),dict(markets=[{'outcomes':'bad'}])]:
 add('markets/response/'+repr(response)[:110],'list_markets',{},response)
for cursor in ['bad','',next_cursor+'\n']:
 add('markets/bad_cursor/'+repr(cursor),'list_markets',{},page,cursor)

history=dict(data=[dict(timestamp=1720000000.123,price='0.1234567890123456789',resolution_seconds=60)],pagination=dict(has_more=True,next_cursor='history/next'))
for args in [dict(asset_id='123',interval='1h'),dict(asset_id='123',start=1720000000,end=1720000000+15*86400,bucket_seconds=60,page_size=1),dict(asset_id='123',as_of=1720000000),dict(asset_id='123'),dict(asset_id='',interval='1h'),dict(asset_id='123',interval='bad'),dict(asset_id='123',interval='1h',start=1),dict(asset_id='123',interval='1h',end=2),dict(asset_id='123',start=0,end=2),dict(asset_id='123',start=1,end=2+15*86400),dict(asset_id='123',start=3,end=2),dict(asset_id='123',as_of=1,page_size=1),dict(asset_id='123',as_of=1,bucket_seconds=60),dict(asset_id='123',as_of=253402300800),dict(asset_id='123',interval='1h',page_size=0),dict(asset_id='123',interval='1h',page_size=10001)]:
 add('history/params/'+repr(args),'history',args,history)
for interval,minimum in [('max',600),('all',600),('1m',600),('1w',300),('1d',60),('6h',60),('1h',60)]:
 for bucket in [minimum-1,minimum,86400,86401]: add(f'history/bucket/{interval}/{bucket}','history',dict(asset_id='123',interval=interval,bucket_seconds=bucket),history)
for key,values in {'timestamp':[None,True,'bad','0','1720000000.999999',253402300800],'price':[None,True,'NaN',0.5],'resolution_seconds':[None,'60',1.5,True]}.items():
 for value in values:
  response=copy.deepcopy(history);response['data'][0][key]=value
  add('history/point/'+key+'/'+repr(value),'history',dict(asset_id='123',interval='1h'),response)
for pagination in [None,{},dict(has_more=True),dict(has_more=True,next_cursor=None),dict(has_more=False,next_cursor='next'),dict(has_more=1,next_cursor='next'),dict(has_more=True,next_cursor=''),dict(has_more=False,next_cursor=None)]:
 add('history/pagination/'+repr(pagination),'history',dict(asset_id='123',interval='1h'),history|{'pagination':pagination})
first=add('history/first','history',dict(asset_id='123',interval='1h',page_size=2),history)
add('history/resume','history',dict(asset_id='123',interval='1h',page_size=3),history,first['expected']['value']['next_cursor'])
add('history/gamma_cursor','history',dict(asset_id='123',interval='1h'),history,next_cursor)
(ROOT/'testdata/discovery-parity.json').write_text(json.dumps(dict(upstream_version=manifest['version'],upstream_commit=manifest['commit'],cases=cases),ensure_ascii=False,indent=2)+'\n')
print('Generated',len(cases),'discovery/history cases')

walks=[]
for name,pages,initial in [
 ('normal',[Page(items=('a',),has_more=True,next_cursor='a'),Page(items=('b',),has_more=False)],None),
 ('missing',[Page(items=('a',),has_more=True)],None),
 ('repeat',[Page(items=('a',),has_more=True,next_cursor='a'),Page(items=('b',),has_more=True,next_cursor='a')],None),
 ('initial_repeat',[Page(items=('a',),has_more=True,next_cursor='a')],'a'),
 ('limit',[Page(items=('a',),has_more=True,next_cursor='a',limit_reached=True)],None),
 ('empty_continuing',[Page(items=(),has_more=True,next_cursor='a'),Page(items=('b',),has_more=False)],None),
]:
 calls=[]
 def fetch(cursor):
  calls.append(cursor)
  return pages[len(calls)-1]
 pager=Paginator(fetch,initial_cursor=initial)
 emitted=[];failure=''
 try:
  for page in pager: emitted.append(normalize(page))
 except Exception as error: failure=type(error).__name__
 walks.append(dict(name=name,pages=normalize(pages),initial=initial,expected=emitted,error=failure,calls=calls))
(ROOT/'testdata/pagination-parity.json').write_text(json.dumps(walks,indent=2)+'\n')
print('Generated',len(walks),'pagination walks')
