"""Compare indexed approval identity, completeness, uint256 amounts and address casing."""
import copy,json
from support import ROOT,manifest,normalize
from polymarket.environments import PRODUCTION
from polymarket._internal.actions.relayer.approvals import build_get_trading_approvals_state_spec,_required_trading_approvals
from polymarket.clients._transport import _clean_params
import httpx
wallet='0x'+'ab'*20
first,second=_required_trading_approvals(PRODUCTION._config)
rows=[{'token':a.token_address,'spender':a.spender,'standard':'ERC20','approved':True,'amount':'max'} for a in first]+[{'token':a.token_address,'spender':a.operator,'standard':'ERC1155','approved':True} for a in second]
base={'address':wallet,'chain_id':137,'contracts':rows};cases=[]
def add(name,owner=wallet,payload=base):
 case={'name':'approvals/'+name,'operation':'get_trading_approvals_state','input':{'wallet':owner},'response':{'data':payload}}
 try:
  spec=build_get_trading_approvals_state_spec(wallet=owner,config=PRODUCTION._config)
  query=httpx.QueryParams(_clean_params(spec.params))
  case['request']={'path':spec.path,'query':{key:query.get_list(key) for key in query}}
  case['expected']={'value':normalize(spec.parse(case['response']))}
 except Exception as err:case['expected']={'error':type(err).__name__}
 cases.append(case)
add('approved')
for value in ['',wallet[2:],wallet.upper(),'0x'+'00'*20,'0xab','x'*40]:add('wallet/'+value,owner=value)
for key,values in {'address':[None,'0x'+'cd'*20], 'chain_id':[None,1,137.0,True], 'contracts':[None,[],[None]]}.items():
 for value in values:add(key+'/'+repr(value),payload=base|{key:value})
for index,row in enumerate(rows):
 changed=copy.deepcopy(rows);changed[index]['approved']=False;add('missing/'+str(index),payload=base|{'contracts':changed})
 changed=copy.deepcopy(rows);changed.pop(index);add('absent/'+str(index),payload=base|{'contracts':changed})
for value in ['0','1','00','-1','1.0','1e2',str(2**256-1),str(2**256),True,1,None]:
 changed=copy.deepcopy(rows);changed[0]['amount']=value;add('amount/'+repr(value),payload=base|{'contracts':changed})
add('duplicate',payload=base|{'contracts':rows+[rows[0]]})
add('unrelated',payload=base|{'contracts':rows+[{'token':'unknown','spender':'x','approved':'bad'},{'token':None}]})
(ROOT/'testdata/approvals-parity.json').write_text(json.dumps({'upstream_version':manifest['version'],'upstream_commit':manifest['commit'],'cases':cases},indent=2)+'\n')
print('Generated',len(cases),'approval cases')
