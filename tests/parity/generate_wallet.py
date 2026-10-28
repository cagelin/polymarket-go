"""Exercise chain call builders, combo derivation and session inputs against Python."""
import json
from dataclasses import asdict
from support import ROOT,manifest,normalize
from polymarket._internal.actions.relayer import calls,positions
from polymarket._internal.actions import session_keys
from polymarket._internal.environment import PRODUCTION_CONFIG as c
items=[]
for v2 in [False,True]:
 for action in ['split','merge','redeem']:
  for amount in [0,1,2**255,-1,2**256]:
   condition='0x03'+'ab'*30 if v2 else '0x'+'ab'*32
   row={'operation':'call','v2':v2,'action':action,'amount':str(amount),'condition':condition}
   try:
    if v2:
     fn=getattr(calls,'router_'+action+'_call');kw={'router':c.protocol_v2_router,'condition_id':condition,'amount':amount}
     if action=='redeem':kw['outcome_index']=1
    else:
     fn=getattr(calls,{'split':'ctf_split_position_call','merge':'ctf_merge_positions_call','redeem':'ctf_redeem_positions_call'}[action]);kw={'collateral':c.collateral_token,'condition_id':condition}
     if action=='redeem':kw['ctf']=c.collateral_adapter
     else:kw.update(target=c.collateral_adapter,amount=amount)
    call=fn(**kw);row['data']=call.data;row['to']=call.to
   except Exception as err:row['error']=type(err).__name__
   items.append(row)
ids=[str(int('01'+'ab'*30+'00',16)),str(int('02'+'bc'*30+'01',16))]
for legs in [[],ids[:1],ids,ids[::-1],ids+ids[:1],[ids[0],str(int(ids[0])+1)],['-1'],['1'],[str(2**256)]]:
 row={'operation':'combo','legs':legs}
 try:
  canonical=positions.canonicalize_combo_legs(legs);result=positions.derive_combo_position_context(canonical)
  row.update(condition=result.condition_id,ids=list(result.position_ids),data=calls.combinatorial_prepare_condition_call(combinatorial_module=c.combinatorial_module,legs=list(canonical)).data)
 except Exception as err:row['error']=type(err).__name__
 items.append(row)
for scopes in [[],['ALL'],['CLOB','COMBOSRFQ'],['all'],[''],['ALL','CLOB'],['future'],['CLOB','CLOB']]:
 row={'operation':'scopes','scopes':scopes}
 try:row['expected']=list(session_keys._parse_scopes(scopes))
 except Exception as err:row['error']=type(err).__name__
 items.append(row)
(ROOT/'testdata/wallet-parity.json').write_text(json.dumps({'upstream_version':manifest['version'],'upstream_commit':manifest['commit'],'cases':items},indent=2)+'\n')
print('Generated',len(items),'wallet cases')
