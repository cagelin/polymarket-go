"""Generate authenticated response models from the pinned Python definitions."""
import inspect,json,re,subprocess,types
from typing import get_args,get_origin,Literal
from enum import IntEnum
from datetime import datetime
from decimal import Decimal
from generate_data_types import ROOT,gotype,fieldname,typename,rule,stripped,ids
from polymarket.models.clob import account,api_key,builder,notifications,rewards,cancel,order_heartbeats,order_response
from polymarket.models import collateral_return
classes=[collateral_return.CollateralReturnPlanResponse,account.OpenOrder,account.MakerOrder,account.ClobTrade,api_key.BuilderApiKeyInfo,builder.BuilderFeeRates,cancel.CancelOrdersResponse,order_heartbeats.OrderHeartbeat,order_heartbeats.OrderHeartbeatMismatchResponse,order_response.RawOrderResponse,order_response.AcceptedOrder,order_response.RejectedOrder,rewards.UserEarning,rewards.TotalUserEarning,rewards.UserRewardsEarning]
existing=set()
for path in (ROOT/'models').glob('*.go'):
 if path.name!='trading_models.go':existing.update(re.findall(r'^type (\w+) ',path.read_text(),re.M))
for name in notifications.__all__:
 value=getattr(notifications,name)
 if hasattr(value,'model_fields'):classes.append(value)
def collect(a):
 if hasattr(a,'model_fields') and typename(a.__name__) not in existing and a not in classes:
  classes.append(a)
 for arg in get_args(a):collect(arg)
for cls in classes:
 for f in cls.model_fields.values():collect(f.annotation)
classes=list(dict.fromkeys(classes))
def modeltype(a):
 base=stripped(a)
 if get_origin(base) is types.UnionType and str in get_args(base):return "string"
 if get_origin(base) is dict:
  result='map[string]'+gotype(get_args(base)[1]);return result
 if get_origin(base) is Literal and isinstance(get_args(base)[0],IntEnum):return 'NotificationType'
 return gotype(a).replace('CtfConditionID','ConditionID').replace('ApiKey','APIKey')
parts=[]
for cls in classes:
 name=typename(cls.__name__).replace('ApiKey','APIKey')
 fields=[(k,f) for k,f in cls.model_fields.items() if not f.deprecated and not (f.description or '').startswith('Deprecated')]
 parts += [f'// {name} is an authenticated exchange response.',f'type {name} struct {{']
 for key,f in fields:parts.append(fieldname(key)+' '+modeltype(f.annotation)+' `json:"'+key+'"`')
 parts+=['}','// UnmarshalJSON validates canonical fields and exchange wire aliases.',f'func (m *{name}) UnmarshalJSON(raw []byte) error {{']
 if name=='BuilderFeeRates':parts+=['normalized,err:=normalizeBuilderFees(raw);if err!=nil{return err};raw=normalized']
 parts += [f'type plain {name}','var result plain',('if err=decodeData(raw,&result,[]dataField{' if name=='BuilderFeeRates' else 'if err:=decodeData(raw,&result,[]dataField{')]
 for key,f in fields:
  r=rule(cls,key,f);a=stripped(f.annotation)
  sources='\n'.join(inspect.getsource(v.func) for v in cls.__pydantic_decorators__.field_validators.values() if key in v.info.fields)
  extra=[]
  if get_origin(f.default_factory) is dict:r=r.replace('},', ',fallback:json.RawMessage(`{}`)},') if 'fallback:' not in r else r
  if 'parse_e6_decimal_string' in sources:r=r.replace('kind:"data_decimal"','kind:"e6_decimal"')
  if 'validate_optional_combo_condition_id' in sources:extra.append('validate_combo_condition_id')
  if key=='position_summary':r=r[:-2]+',fallback:json.RawMessage(`{}`)},'
  if 'validate_condition_id(' in sources and 'condition_feed' not in r:extra.append('condition_feed')
  if a is Decimal and 'parse_decimal_string' in sources:r=r.replace('kind:"data_decimal"','kind:"string_decimal"')
  if a is datetime:r=re.sub('kind:"[^"]+"','kind:"clob_epoch_iso"',r)
  if '_empty_string_to_none' in sources:extra.append('empty_none')
  if '_normalize_trade_status' in sources:extra.append('trade_status')
  if cls is order_response.RawOrderResponse and key in ['making_amount','taking_amount']:extra.append('empty_zero')
  if '_parse_expiration_timestamp' in sources:extra.append('zero_time')
  if (cls is api_key.BuilderApiKeyInfo or '_parse_epoch_ms' in sources) and a is datetime:r=r.replace('clob_epoch_iso','reward_epoch_ms')
  if key=='type' and get_origin(a) is Literal and isinstance(get_args(a)[0],IntEnum):
   r=re.sub(r'allowed:\[\]string\{[^}]+\},?','',r)
   r=r[:-2]+',kind:"notification_kind",minimum:'+str(int(get_args(a)[0]))+'},'
  if '_parse_notification_id' in sources:r=r.replace('kind:"int"','kind:"notification_id"')
  if cls is order_heartbeats.OrderHeartbeat:r=r[:-2]+',kind:"nonempty_string"},'
  if extra:
   if 'steps:[]string{' in r:r=r.replace('steps:[]string{','steps:[]string{'+','.join(json.dumps(x) for x in extra)+',')
   else:r=r[:-2]+',steps:[]string{'+','.join(json.dumps(x) for x in extra)+'}},'
  parts.append(r.replace(",,",","))
 parts+=['});err!=nil{return err}',f'*m={name}(result)','return nil','}']
header='''// Authenticated models ported from the pinned SDK. This file is linted.

package models
import (
 "encoding/json"
 "time"
 "github.com/ericlagergren/decimal"
)
'''
for name,base in sorted(ids):
 if name not in existing and name not in ['CtfConditionID']:header+=f'// {name.replace("ApiKey","APIKey")} identifies an account resource.\ntype {name.replace('ApiKey','APIKey')} {base}\n'
(ROOT/'models/trading_models.go').write_text(header+'\n'.join(parts))
subprocess.run(['gofmt','-w',str(ROOT/'models/trading_models.go')],check=True)
