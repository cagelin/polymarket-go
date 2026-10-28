"""Emit public stream payload schemas from the verified Python source."""
import inspect,json,subprocess
from datetime import datetime
from typing import get_args,get_origin,Literal
from generate_data_types import ROOT,gotype,fieldname,typename,rule,stripped
from polymarket.models.clob import market_events,user_events
from polymarket.models import sports_events,rtds_events,price_events
classes=[]
for module in [market_events,sports_events,rtds_events,user_events,price_events]:
 for name in module.__all__:
  cls=getattr(module,name)
  if hasattr(cls,'model_fields') and cls.__module__==module.__name__ and cls not in classes:classes.append(cls)
parts=[]
for cls in classes:
 name=typename(cls.__name__);fields=[(k,f) for k,f in cls.model_fields.items() if not f.deprecated and not (f.description or '').startswith('Deprecated')]
 parts += [f'// {name} represents a public stream message or payload.',f'type {name} struct {{']
 for key,f in fields:
  a=stripped(f.annotation)
  typ='string' if key in ['source','status'] and cls.__module__ in [user_events.__name__,price_events.__name__] else 'json.RawMessage' if a is object else 'int64' if get_origin(a) is Literal and all(type(v) is int for v in get_args(a)) else gotype(f.annotation)
  if key=='status' and cls.__module__==user_events.__name__:typ='*string'
  parts.append(fieldname(key)+' '+typ+' `json:"'+key+'"`')
 parts += ['}','// UnmarshalJSON validates stream fields and preserves exact amounts.',f'func(m *{name}) UnmarshalJSON(raw []byte)error{{']
 if name in ['SportsGameResult','CryptoPricesChainlinkTwapPayload','EquityPriceUpdatePayload']:
  parts += [f'normalized,err:=normalizeStreamPayload(raw,"{name}");if err!=nil{{return err}};raw=normalized']
 parts += [f'type plain {name}','var result plain',('if err=decodeData(raw,&result,[]dataField{' if name in ['SportsGameResult','CryptoPricesChainlinkTwapPayload','EquityPriceUpdatePayload'] else 'if err:=decodeData(raw,&result,[]dataField{')]
 for key,f in fields:
  a=stripped(f.annotation);r=rule(cls,key,f)
  texts='\n'.join(inspect.getsource(v.func) for v in cls.__pydantic_decorators__.field_validators.values() if key in v.info.fields)
  steps=[]
  if key in ['source','status'] and cls.__module__ in [user_events.__name__,price_events.__name__]:
   r=r.replace('kind:"json",','')
  if cls.__module__==user_events.__name__ and key=='status':typ='*string'
  if '_coerce_optional_decimalish' in texts:steps.append('optional_text')
  if '_uppercase_order_side' in texts or '_uppercase_string' in texts:steps.append('uppercase')
  if '_normalize_trade_status' in texts:steps.append('trade_status')
  if '_parse_expiration_timestamp' in texts:steps.append('zero_time')
  if 'validate_optional_condition_id' in texts:steps.append('condition_feed')
  if a is datetime:
   kind='stream_epoch_ms' if key=='timestamp' and cls.__module__==market_events.__name__ else 'reward_epoch_ms_unsigned' if cls.__module__!=market_events.__name__ else 'game_start_time'
   if 'parse_optional_datetime' in texts:kind='datetime'
   if cls.__module__==user_events.__name__:kind='clob_epoch_iso'
   if cls.__module__==price_events.__name__:kind='epoch_datetime'
   r=r.replace('kind:"epoch_datetime"','kind:'+json.dumps(kind))
  if cls.__module__==price_events.__name__ and key=='window_seconds':r=r.replace('allowed:[]string{"60"}','kind:"int",steps:[]string{"price_window"}')
  elif key=='window_seconds':r=r.replace('allowed:[]string{"30","60"}','kind:"int",steps:[]string{"twap_window"}')
  if steps:r=r.replace('name:'+json.dumps(key),'name:'+json.dumps(key)+',steps:[]string{'+','.join(json.dumps(v) for v in steps)+'}')
  parts.append(r)
 parts += ['});err!=nil{return err}',f'*m={name}(result)','return nil','}']
header='''// Stream schemas ported from the pinned Python SDK. This file is linted.

package models

import (
 "encoding/json"
 "time"

 "github.com/ericlagergren/decimal"
)
'''
(ROOT/'models/stream_models.go').write_text(header+'\n'.join(parts))
subprocess.run(['gofmt','-w',str(ROOT/'models/stream_models.go')],check=True)
