"""Compare all tagged account response variants with the official parsers."""
import json,copy
from support import ROOT,manifest,normalize
from polymarket.models.data.activity import parse_activity,parse_combo_activity,_KNOWN_ACTIVITY_TYPES,_COMBO_ACTIVITY_TYPES
from polymarket.models.data.leaderboard import parse_biggest_winners
models=json.loads((ROOT/'testdata/data-models-parity.json').read_text())['cases']
full={c['model']:c['response'] for c in models if c['name'].endswith('/full')}
cases=[]
for union,classes,parse in [('Activity',[c.__name__ for c in _KNOWN_ACTIVITY_TYPES.values()]+['ComboTradeActivity','UnknownActivity'],parse_activity),('ComboActivity',[c.__name__ for c in _COMBO_ACTIVITY_TYPES.values()],parse_combo_activity),('BiggestWinner',['MarketBiggestWinner','ComboBiggestWinner'],lambda p:parse_biggest_winners([p])[0])]:
 raws=[(name,full[name]) for name in classes]
 raws += [(str(v),v) for v in [None,[],False,{}, {'type':'FUTURE','kind':'future','extra':7}]]
 if union=='Activity':
  raws += [('unknown-type/'+str(v),full['UnknownActivity']|{'type':v}) for v in ['FUTURE',None,False,7,{}]]
  raws += [('combo-flag/'+str(v),full['ComboTradeActivity']|{'is_combo':v}) for v in [True,False,'true',1]]
 for name,raw in raws:
  item={'name':union+'/'+name,'union':union,'response':raw}
  try:
   result=parse(copy.deepcopy(raw));item['expected']={'value':normalize(result),'variant':type(result).__name__}
  except Exception as err:item['expected']={'error':type(err).__name__}
  cases.append(item)
(ROOT/'testdata/unions-parity.json').write_text(json.dumps({'upstream_version':manifest['version'],'upstream_commit':manifest['commit'],'cases':cases},indent=2)+'\n')
print('Generated',len(cases),'union cases')
