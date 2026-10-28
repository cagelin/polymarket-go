"""Authenticated user and price event parsing compared with the pinned Python SDK."""
import copy,json
from support import ROOT,manifest,normalize
from polymarket.models.clob.user_events import parse_user_event
from polymarket._internal.streams.realtime.protocol import parse_price_event
baselines={c['model']:c['response'] for c in json.loads((ROOT/'testdata/data-models-parity.json').read_text())['cases'] if c['name'].endswith('/full')}
cases=[]
def add(service,label,raw):
 case={'service':service,'name':label,'raw':raw}
 try:
  value=(parse_user_event if service=='user' else parse_price_event)(copy.deepcopy(raw))
  case['expected']={'value':normalize(value)} if value is not None else {'error':True}
 except Exception:case['expected']={'error':True}
 cases.append(case)
for name in ['UserOrderEvent','UserTradeEvent']:
 base=baselines[name];raw=copy.deepcopy(base['payload']);raw['event_type']=base['type']
 add('user',name,raw);add('user',name+'/envelope',base)
 missing_event=raw.copy();missing_event.pop('event_type');missing_event['type']=base['type'];add('user',name+'/type-is-not-event-type',missing_event)
 for key in raw:
  missing=raw.copy();missing.pop(key);add('user',name+'/missing/'+key,missing);add('user',name+'/null/'+key,raw|{key:None})
 for key in ['side','status','type','timestamp']:
  if key in raw:
   for value in ['buy','sell','MATCHED','matched','canceled','new_status','1720000000.123456',1720000000,True]:add('user',name+'/'+key+'/'+repr(value),raw|{key:value})
for channel in ['price.crypto','price.crypto.twap','price.equity','price.equity.twap']:
 point={'timestamp':1720000000123,'value':1.25,'full_accuracy_value':'12345678901234567890.123456789'}
 payload={'symbol':'BTCUSD' if '.crypto' in channel else 'AAPL','source':'future-source'}
 if channel.endswith('.twap'):payload['window_seconds']=60
 if '.equity' in channel:payload['market_status']='closed'
 for snapshot in [False,True]:
  base={'v':1,'channel':channel,'seq':5,'dropped':2,'ts':1720000000130,'snapshot':snapshot,'payload':payload|({'data':[point]} if snapshot else point|{'received_at':1720000000124,'is_carried_forward':True})}
  label=channel+'/'+str(snapshot);add('price',label,base)
  for key in base:
   missing=base.copy();missing.pop(key);add('price',label+'/missing/'+key,missing);add('price',label+'/null/'+key,base|{key:None})
  for key,value in base['payload'].items():
   missing=copy.deepcopy(base);missing['payload'].pop(key);add('price',label+'/payload/missing/'+key,missing)
   for bad in [None,True,{},'']:
    malformed=copy.deepcopy(base);malformed['payload'][key]=bad;add('price',label+'/payload/'+key+'/'+repr(bad),malformed)
  for changes in [{'v':True},{'v':1.0},{'seq':-1},{'seq':1.0},{'ts':1.0},{'dropped':-1}]:add('price',label+'/'+repr(changes),base|changes)
(ROOT/'testdata/secure-stream-events-parity.json').write_text(json.dumps({'upstream_version':manifest['version'],'upstream_commit':manifest['commit'],'cases':cases},indent=2)+'\n')
print('Generated',len(cases),'secure stream events')
