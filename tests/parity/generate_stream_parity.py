"""Compare stream event parsing, subscription frames and filtering with upstream."""
import copy,json
from support import ROOT,manifest,normalize
from polymarket.models.clob.market_events import parse_market_event
from polymarket.models.rtds_events import parse_rtds_event
from polymarket.models.sports_events import parse_sports_event
from polymarket.streams import MarketSpec,SportsSpec,CommentsSpec,CryptoPricesSpec,CryptoPricesChainlinkTwapSpec,EquityPricesSpec
from polymarket._internal.streams.clob import market_protocol as mp
from polymarket._internal.streams.rtds import protocol as rp
baselines={c['model']:c['response'] for c in json.loads((ROOT/'testdata/data-models-parity.json').read_text())['cases'] if c['name'].endswith('/full')}
wire_topics={'prices.crypto.binance':'crypto_prices','prices.crypto.chainlink':'crypto_prices_chainlink','prices.crypto.chainlink.twap':'crypto_prices_twap_sixty','prices.equity.pyth':'equity_prices','comments':'comments'}
events=[]
def add(service,label,raw):
 case={'service':service,'name':label,'raw':raw}
 try:case['expected']={'value':normalize({'market':parse_market_event,'sports':parse_sports_event,'rtds':parse_rtds_event}[service](raw))}
 except Exception:case['expected']={'error':True}
 events.append(case);return case
for name,base in baselines.items():
 if not name.endswith('Event') or 'topic' not in base or 'payload' not in base:continue
 raw=copy.deepcopy(base);topic=raw['topic'];service='market' if topic=='market' else 'sports' if topic=='sports' else 'rtds'
 if service=='market':raw={'event_type':raw['type']}|raw['payload']
 elif service=='sports':raw=raw['payload']
 else:
  raw['topic']=wire_topics[topic]
  if topic=='prices.crypto.chainlink.twap':raw['payload'].update(window_s=60,full_accuracy_value='12345678901234567890123456789')
 initial=add(service,name,raw)
 if 'error' in initial['expected']:raise AssertionError(initial)
 for key in list(raw):
  absent=copy.deepcopy(raw);absent.pop(key);add(service,name+'/missing/'+key,absent)
  add(service,name+'/null/'+key,raw|{key:None})
 if service=='market':add(service,name+'/envelope',base)
 if 'timestamp' in raw:
  for value in ['0','1720000000001',1720000000001,'','-1',-1,True,'2024-07-03T09:46:40+08:00',1.25]:add(service,name+'/timestamp/'+repr(value),raw|{'timestamp':value})
 if name=='CryptoPricesChainlinkTwapEvent':
  for key,values in {'full_accuracy_value':['-1','0','1.2',None],'window_s':[30,60.0,True],'value':[None,'bad',1.25]}.items():
   for value in values:add(service,name+'/'+key+'/'+repr(value),raw|{'payload':raw['payload']|{key:value}})
 if name=='EquityPricesUpdateEvent':add(service,name+'/full-accuracy',raw|{'payload':raw['payload']|{'full_accuracy_value':'1.234567890123456789'}})
 if service=='sports':add(service,name+'/timestamp-fallback',raw|{'finishedTimestamp':None,'finished_timestamp':1720000000001})
for service in ['market','sports','rtds']:
 for raw in [None,[],1,{}, {'type':'unknown','topic':'unknown','payload':{}}]:add(service,'invalid/'+repr(raw),raw)
(ROOT/'testdata/streams-events-parity.json').write_text(json.dumps({'upstream_version':manifest['version'],'upstream_commit':manifest['commit'],'cases':events},indent=2)+'\n')
print('Generated',len(events),'stream event cases')
constructors={'market':MarketSpec,'sports':SportsSpec,'comments':CommentsSpec,'crypto':CryptoPricesSpec,'twap':CryptoPricesChainlinkTwapSpec,'equity':EquityPricesSpec}
selections=[
 ('market',{'asset_ids':['123']}),('market',{'asset_ids':['z','123','123'],'custom_feature_enabled':True}),('market',{'asset_ids':[]}),('market',{'asset_ids':['']}),
 ('sports',{}),('comments',{}),('comments',{'types':['comment_created','reaction_removed'],'parent_entity_id':123,'parent_entity_type':'Event'}),('comments',{'types':[]}),('comments',{'types':['bad']}),('comments',{'parent_entity_type':'Series'}),
 ('crypto',{'topic':'prices.crypto.binance'}),('crypto',{'topic':'prices.crypto.chainlink','symbols':['sample']}),('crypto',{'topic':'bad'}),('crypto',{'topic':'prices.crypto.binance','symbols':[]}),
 ('twap',{'window_seconds':60}),('twap',{'window_seconds':30,'symbols':['sample']}),('twap',{'window_seconds':1}),
 ('equity',{'symbol':'SAMPLE'}),('equity',{'symbol':'sample','types':['subscribe']}),('equity',{'symbol':'sample','types':[]}),('equity',{'symbol':''})]
protocol=[]
def actual(spec):return constructors[spec['kind']](**spec['input'])
def frame(service,specs):
 if service=='market':return mp.build_initial_frame(mp.derive_state([mp.MarketSubscription(tuple(s.asset_ids),s.custom_feature_enabled) for s in specs]))
 if service=='rtds':return rp.build_subscribe_frame(rp.derive_state(specs).values())
 return None
for index,(kind,inp) in enumerate(selections):
 case={'name':kind+'/'+str(index),'spec':{'kind':kind,'input':inp}}
 try:
  spec=actual(case['spec']);service='market' if kind=='market' else 'sports' if kind=='sports' else 'rtds';case['service']=service
  matcher=mp.match_for(mp.MarketSubscription(tuple(spec.asset_ids),spec.custom_feature_enabled)) if service=='market' else (lambda _:True) if service=='sports' else rp.matcher_for(spec)
  case['expected']={'initial':frame(service,[spec]),'matches':[ev['name'] for ev in events if ev['service']==service and 'error' not in ev['expected'] and matcher({'market':parse_market_event,'sports':parse_sports_event,'rtds':parse_rtds_event}[service](ev['raw']))]}
 except Exception:case['expected']={'error':True}
 protocol.append(case)
transitions=[]
for kind,before,after in [('market',[{'asset_ids':['a','b']}],[{'asset_ids':['b','c'],'custom_feature_enabled':True}]),('market',[{'asset_ids':['a']}],[{'asset_ids':['a'],'custom_feature_enabled':True}]),('market',[{'asset_ids':['a']}],[]),('comments',[{}],[{'types':['comment_removed','comment_created']}]),('comments',[{}],[])]:
 service='market' if kind=='market' else 'rtds';b=[constructors[kind](**p) for p in before];a=[constructors[kind](**p) for p in after]
 if service=='market':b=[mp.MarketSubscription(tuple(s.asset_ids),s.custom_feature_enabled) for s in b];a=[mp.MarketSubscription(tuple(s.asset_ids),s.custom_feature_enabled) for s in a];diff=mp.diff_state_frames(mp.derive_state(b),mp.derive_state(a))
 else:diff=rp.diff_state_frames(rp.derive_state(b),rp.derive_state(a))
 transitions.append({'service':service,'before':[{'kind':kind,'input':p} for p in before],'after':[{'kind':kind,'input':p} for p in after],'expected':diff})
(ROOT/'testdata/streams-protocol-parity.json').write_text(json.dumps({'cases':protocol,'transitions':transitions},indent=2)+'\n')
