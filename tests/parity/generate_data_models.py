"""Exercise every Data model against the actual pinned Python parser."""
from typing import get_args,get_origin,Literal
from datetime import datetime,date
from decimal import Decimal
from enum import Enum
from pathlib import Path
import copy,json,subprocess
from generate_data_types import classes,typename,ROOT
from support import normalize,manifest
from generate_public_clob_types import classes as clob_classes
from generate_catalog_types import classes as catalog_classes
from generate_stream_types import classes as stream_classes
from generate_trading_types import classes as trading_classes
classes = classes + clob_classes + catalog_classes + stream_classes + trading_classes

def value_for(a,name,full):
 if name=="onchain_event_id":return "0x"+"ab"*29+"000000"
 if name=="request_id":return "0x"+"ab"*31+"00"
 args=get_args(a)
 if type(None) in args:
  rest=[v for v in args if v is not type(None)]
  if not full:return None
  return value_for(rest[0],name,full)
 if get_origin(a) is Literal:return args[0].value if isinstance(args[0],Enum) else args[0]
 if isinstance(a,type) and issubclass(a,Enum):return next(iter(a)).value
 if a is object:return {"rate":1}
 if a is str:
  if 'condition' in name:return '0x'+'ab'*32
  return 'sample'
 if a is bool:return True
 if a is int:return 7
 if a is float:return 1.25
 if a is Decimal:return '1.234567890123456789'
 if a is datetime:return 1720000000
 if a is date:return '2026-10-01'
 if get_origin(a) is tuple:
  if len(args)==2 and args[1] is not Ellipsis:return [value_for(v,name,full) for v in args]
  return [value_for(args[0],name,full)] if full else []
 if get_origin(a) is dict:return {'key':'value'}
 if hasattr(a,'__supertype__'):
  if 'ComboCondition' in a.__name__:return '0x03'+'ab'*30
  if 'Condition' in a.__name__:return '0x'+'ab'*32
  if a.__supertype__ is int:return 7
  return '123'
 if hasattr(a,'model_fields'):return payload(a,full)
 if args:return value_for(args[0],name,full)
 raise TypeError((a,name))
def payload(cls,full):
 result={}
 for name,field in cls.model_fields.items():
  if field.deprecated or (field.description or '').startswith('Deprecated') or (not full and not field.is_required()):continue
  alias=field.validation_alias
  wire=alias.choices[0] if hasattr(alias,'choices') else alias if isinstance(alias,str) else name
  result[wire]=value_for(field.annotation,name,full)
  if cls.__name__ in ["CollateralReturnOperation","CollateralReturnPositionAmount"] and name=="amount":result[wire]="1234567"
  if (cls.__module__.startswith('polymarket.models.gamma') or cls.__name__=='CommentRemovedPayload') and cls.__name__!='ImageOptimization' and result[wire] is not None and (field.annotation is datetime or datetime in get_args(field.annotation)):
   result[wire]='2024-07-03T09:46:40Z'
 if cls.__module__=='polymarket.models.clob.market_events' and 'timestamp' in result and result['timestamp'] is not None:result['timestamp']='1720000000000'
 if cls.__name__=='CryptoPricesChainlinkTwapPayload':
  result.pop('window_s',None);result['window_seconds']=30
 if cls.__name__ == "ComboMarket": result["condition_id"]='0x'+'ab'*32
 if cls.__name__ == "RedeemActivity": result["is_combo"] = False
 return result
cases=[]
def add(name,cls,raw):
 case=dict(name=name,model=typename(cls.__name__).replace("ApiKey","APIKey"),response=copy.deepcopy(raw))
 try:case['expected']={'value':normalize(cls.parse_response(copy.deepcopy(raw)))}
 except Exception as err:case['expected']={'error':type(err).__name__}
 cases.append(case)
for cls in classes:
 for full in [False,True]:
  raw=payload(cls,full)
  add(cls.__name__+('/full' if full else '/minimal'),cls,raw)
  if 'error' in cases[-1]['expected']:
   try:cls.parse_response(raw)
   except Exception as exc:raise AssertionError((cls.__name__,str(exc.__cause__)))
 raw=payload(cls,True)
 for field,value in list(raw.items()):
  add(cls.__name__+'/null/'+field,cls,raw|{field:None})
  missing=raw.copy();missing.pop(field)
  add(cls.__name__+'/missing/'+field,cls,missing)
  if not isinstance(value,(dict,list)):add(cls.__name__+'/invalid/'+field,cls,raw|{field:{}})
  if isinstance(value,str):add(cls.__name__+'/empty/'+field,cls,raw|{field:''})
 # Common sentinels, coercions and protocol checks get explicit vectors.
 for field in ['outcome_index','condition_id','timestamp','event_id','last_event_at']:
  if field not in raw:continue
  for value in [0,999,'999','0x03'+'AB'*30,'0xab',True,'1720000000.123456']:
   add(cls.__name__+'/boundary/'+field+'/'+repr(value),cls,raw|{field:value})
# Test decimal transport and payout arithmetic beyond the previous library limit.
for cls in classes:
 raw=payload(cls,True)
 for name,field in cls.model_fields.items():
  if field.annotation is Decimal or Decimal in get_args(field.annotation):
   alias=field.validation_alias
   wire=alias.choices[0] if hasattr(alias,'choices') else alias if isinstance(alias,str) else name
   if wire not in raw:continue
   for value in ['1e-100001','1e100001','1e-999999999999999999','1e999999999999999999']:
    add(cls.__name__+'/decimal/'+name+'/'+value,cls,raw|{wire:value})
 if cls.__name__=='BuilderFeeRates':
  for value in ['0','10000','1000000','1.2300','1e30']:
   add('BuilderFeeRates/bps/'+value,cls,{'builder_maker_fee_rate_bps':value,'builder_taker_fee_rate_bps':value})
 if cls.__name__=='Resolution':
  for value in ['1000000000000000000000000000000000000','1000000.00','1e-100001','1e-1000020','1e-1000021','1e-1000027','1e1000006','-0.0000000','12345678901234567890123456789012345']:
   add('Resolution/payout/'+value,cls,raw|{'payouts':[value,value]})
(ROOT/'testdata/data-models-parity.json').write_text(json.dumps(dict(upstream_version=manifest['version'],upstream_commit=manifest['commit'],cases=cases),ensure_ascii=False,indent=2)+'\n')
print('Generated',len(cases),'Data model cases')
# Dispatch only; all expectations remain actual Python results.
text='''package models

import (
 "github.com/cagelam/polymarket-go/internal/testutil"
 "encoding/json"
 "os"
 "testing"
)

func TestPythonDataModels(t *testing.T) {
 raw,err:=os.ReadFile("../testdata/data-models-parity.json")
 if err!=nil { t.Fatal(err) }
 var fixture struct { Cases []struct { Name,Model string; Response json.RawMessage; Expected struct { Value json.RawMessage; Error string } } }
 if err=json.Unmarshal(raw,&fixture);err!=nil { t.Fatal(err) }
 for _,test:=range fixture.Cases { t.Run(test.Name,func(subtest *testing.T){
  target:=newDataParityTarget(test.Model)
  if target==nil { subtest.Fatalf("unknown model %s",test.Model) }
  decodeErr:=json.Unmarshal(test.Response,target)
  if (decodeErr!=nil)!=(test.Expected.Error!="") { subtest.Fatalf("%s decode error = %v, Python = %s",test.Model,decodeErr,test.Expected.Error) }
  if decodeErr!=nil { return }
  testutil.EqualJSON(subtest,target,test.Expected.Value)
 }) }
}

func newDataParityTarget(name string) any {
 switch name {
'''
for cls in classes:
 name=typename(cls.__name__).replace('ApiKey','APIKey');text+=f'case "{name}":return &{name}{{}}\n'
text+='default:return nil\n}\n}\n'
(ROOT/'models/data_parity_test.go').write_text(text)
subprocess.run(['gofmt','-w',str(ROOT/'models/data_parity_test.go')],check=True)
