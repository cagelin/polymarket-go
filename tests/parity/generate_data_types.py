"""Generate typed Data API models and explicit coercion tables from verified source."""
from pathlib import Path
from typing import get_args,get_origin,Literal
from enum import Enum
from datetime import datetime,date
from decimal import Decimal
import inspect,re,json,subprocess
from polymarket.models.data import common,portfolio,activity,analytics,leaderboard,tokens,resolutions
from generate_market_types import ROOT,named,go_name,ids

modules=[portfolio,activity,analytics,leaderboard,tokens,resolutions]
classes=[]
for module in modules:
 for name in module.__all__:
  cls=getattr(module,name)
  if hasattr(cls,'model_fields') and cls.__name__!='PriceHistoryPoint':classes.append(cls)
existing=set()
for path in (ROOT/'models').glob('*.go'):
 if path.name not in ['data_models.go','data_enums.go']:existing.update(re.findall(r'^type (\w+) ',path.read_text(),re.M))
enums={}
def typename(name):return named(name).replace('Pnl','PnL')
def fieldname(name):return go_name(name).replace('Pnl','PnL').replace('Usdc','USDC')
def stripped(a):
 args=get_args(a)
 if type(None) in args:
  rest=[v for v in args if v is not type(None)]
  if len(rest)==1:return rest[0]
  return str
 return a
def gotype(a):
 args=get_args(a)
 if type(None) in args:
  rest=[v for v in args if v is not type(None)]
  base="ClobAssetID" if len(rest)>1 and all(getattr(v,"__supertype__",None) is str for v in rest) else gotype(stripped(a))
  return base if base.startswith(('[]','map[','*')) else '*'+base
 if a is str:return 'string'
 if a is int:return 'int64'
 if a is bool:return 'bool'
 if a is float:return 'float64'
 if a is Decimal:return '*decimal.Big'
 if a is datetime:return 'time.Time'
 if a is date:return 'Date'
 if get_origin(a) is tuple:return '[]'+gotype(args[0])
 if get_origin(a) is dict:return 'map[string]json.RawMessage'
 if get_origin(a) is Literal:
  if all(isinstance(v,bool) for v in args):return 'bool'
  if all(isinstance(v,Enum) for v in args):return gotype(type(args[0]))
  if args==('BUY','SELL'):return 'OrderSide'
  return 'string'
 if args and all(getattr(v,'__supertype__',None) is str for v in args):return 'ClobAssetID'
 if isinstance(a,type) and issubclass(a,Enum):enums[typename(a.__name__)]=a;return typename(a.__name__)
 if hasattr(a,'__supertype__'):
  name=typename(a.__name__);ids.add((name,'int64' if a.__supertype__ is int else 'string'));return name
 return typename(a.__name__)
def rule(cls,name,field):
 a=stripped(field.annotation);kind='';steps=[];allowed=[]
 if a is Decimal:kind='data_decimal'
 elif a is datetime:kind='epoch_datetime'
 elif a is date:kind='date'
 elif a is bool:kind='bool'
 elif a is int:kind='int'
 elif a is float:kind='float'
 elif get_origin(a) is tuple:kind='sequence'
 if get_origin(a) is Literal:
  values=get_args(a)
  if all(isinstance(v,bool) for v in values):kind='literal_bool_true' if values[0] else 'literal_bool_false'
  else:allowed=[str(v) for v in values]
 elif isinstance(a,type) and issubclass(a,Enum):allowed=[str(v.value) for v in a]
 validators=[]
 for v in cls.__pydantic_decorators__.field_validators.values():
  if name in v.info.fields:
   text=inspect.getsource(v.func);validators.append(v.func.__name__)
   for parser in ['optional_text','optional_outcome_index','date_from_calendar_string','optional_event_id','required_event_id','optional_datetime_from_epoch_seconds','datetime_from_epoch_seconds','optional_datetime_from_epoch_or_iso','datetime_from_epoch_or_iso','validate_condition_id_response','validate_optional_condition_id_response','validate_combo_condition_id']:
    if parser in text or v.func.__name__==parser:steps.append(parser)
 if cls.__name__=='TraderLeaderboardStanding' and name in ['pnl_rank','volume_rank']:steps.append('zero_rank')
 if cls.__name__=='Resolution':
  if name in ['proposed_price','reproposed_price','price']:steps.append('unset_price')
  if name=='payouts':kind='payout_pair'
 if cls.__name__=='OpenInterest' and name=='condition_id':steps.insert(0,'global_condition')
 if cls.__name__=='TokenReference':
  if name in ['asset_id','opposite_asset_id']:kind='asset_reference'
  if name in ['condition_id','structural_condition_id','neg_risk_market_id']:kind='condition_reference'
  if name=='event_id':kind='event_reference'
 strict=any(getattr(meta,'strict',False) for meta in field.metadata)
 if strict and kind in ['bool','int']:kind='strict_'+kind
 for variant in ['optional_datetime_from_epoch_seconds','datetime_from_epoch_seconds','optional_datetime_from_epoch_or_iso','datetime_from_epoch_or_iso']:
  if variant in steps:
   kind={'optional_datetime_from_epoch_seconds':'optional_epoch','datetime_from_epoch_seconds':'epoch','optional_datetime_from_epoch_or_iso':'optional_epoch_or_iso','datetime_from_epoch_or_iso':'epoch_or_iso'}[variant]
   # Only the longest explicit validator applies.
   if variant.startswith('optional_'):break
 steps=[v for v in steps if 'datetime_from_' not in v]
 # Response optional validators invoke their required counterpart internally.
 if 'validate_optional_condition_id_response' in steps:steps=[v for v in steps if v!='validate_condition_id_response']
 if 'optional_event_id' in steps and 'required_event_id' in steps:steps.remove('required_event_id')
 values=['name:'+json.dumps(name)]
 alias=field.validation_alias
 aliases=alias.choices if hasattr(alias,'choices') else [alias] if isinstance(alias,str) else []
 if aliases:values.append('aliases:[]string{'+','.join(json.dumps(v) for v in aliases)+'}')
 if kind:values.append('kind:'+json.dumps(kind))
 if field.is_required():values.append('required:true')
 if type(None) in get_args(field.annotation):values.append('nullable:true')
 if steps:values.append('steps:[]string{'+','.join(json.dumps(v) for v in dict.fromkeys(steps))+'}')
 if allowed:values.append('allowed:[]string{'+','.join(json.dumps(v) for v in allowed)+'}')
 if not field.is_required() and field.default is not None and isinstance(field.default,(str,bool,int,list,tuple,dict)):
  values.append('fallback:json.RawMessage('+json.dumps(json.dumps(field.default))+')')
 if field.default_factory is dict:values.append('fallback:json.RawMessage(`{}`)')
 for meta in field.metadata:
  if hasattr(meta,'ge'):values+=['minimum:'+str(meta.ge),'bounded:true']
  if hasattr(meta,'le'):values+=['maximum:'+str(meta.le)]
 return '{'+','.join(values)+'},'
parts=[]
for cls in classes:
 name=typename(cls.__name__)
 fields=[(k,v) for k,v in cls.model_fields.items() if not v.deprecated]
 parts += ['// '+name+' represents '+(cls.__doc__ or 'a Data API response').splitlines()[0].replace('`', '').rstrip('.')+'.','type '+name+' struct {']
 for key,field in fields:parts.append(fieldname(key)+' '+gotype(field.annotation)+' `json:"'+key+'"`')
 parts+=['}','// UnmarshalJSON validates Data API field types and wire aliases.','func (m *'+name+') UnmarshalJSON(raw []byte) error {','type plain '+name,'var result plain','if err:=decodeData(raw,&result,[]dataField{']
 for key,field in fields:parts.append(rule(cls,key,field))
 parts+=['});err!=nil { return err }']
 if name=='RedeemActivity':parts+=['if result.IsCombo { if _,err:=codec.Condition(string(result.ConditionID),"combo");err!=nil { return err } }']
 parts+=['*m='+name+'(result)','return nil','}']
header='''// Data models ported from polymarket-client 0.14.0. This file is linted.

package models

import (
 "encoding/json"
 "time"
 "github.com/ericlagergren/decimal"
 "github.com/cagelam/polymarket-go/internal/codec"
)
'''
for name,base in sorted(ids):
 if name not in existing:header+=f'// {name} identifies a Data API resource.\ntype {name} {base}\n'
(ROOT/'models/data_models.go').write_text(header+'\n'+'\n'.join(parts))
# Include filter-only enumerations as well as model fields.
for name in common.__all__:
 value=getattr(common,name)
 if isinstance(value,type) and issubclass(value,Enum):enums[typename(value.__name__)]=value
text='// Data API enumerations from the pinned SDK. This file is linted.\n\npackage models\n'
for name,cls in sorted(enums.items()):
 text+=f'\n// {name} is a supported Data API value.\ntype {name} string\n\n// Supported {name} values.\nconst (\n'
 for v in cls:
  suffix=fieldname(v.name.lower())
  text+=f'{name}{suffix} {name} = '+json.dumps(v.value)+'\n'
 text+=')\n'
(ROOT/'models/data_enums.go').write_text(text)
subprocess.run(['gofmt','-w',str(ROOT/'models/data_models.go'),str(ROOT/'models/data_enums.go')],check=True)
