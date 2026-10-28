"""Generate public CLOB and RFQ response models from verified upstream definitions."""
import inspect,re,json,subprocess
from datetime import datetime
from decimal import Decimal
from generate_data_types import ROOT,gotype,fieldname,typename,rule,ids,stripped
from polymarket.models.clob import rewards,builder
from polymarket.models import rfq
classes=[rewards.CurrentRewardConfig,rewards.CurrentReward,rewards.MarketRewardConfig,rewards.MarketRewardToken,rewards.MarketReward,rewards.RewardMarket,rewards.RebatedFee,builder.BuilderTrade,rfq.ComboMarketOutcome,rfq.ComboMarketOutcomes,rfq.ComboMarket]
parts=[]
for cls in classes:
 name=typename(cls.__name__)
 fields=[(key,f) for key,f in cls.model_fields.items() if not f.deprecated and not (f.description or '').startswith('Deprecated')]
 parts += [f'// {name} represents a public CLOB or RFQ response.',f'type {name} struct {{']
 for key,f in fields:parts.append(fieldname(key)+' '+gotype(f.annotation).replace('CtfConditionID','ConditionID')+' `json:"'+key+'"`')
 parts+=['}', '// UnmarshalJSON validates response values and wire aliases.',f'func (m *{name}) UnmarshalJSON(raw []byte) error {{']
 if name=='ComboMarket':parts+=['normalized,err:=normalizeComboMarket(raw);if err!=nil {return err};raw=normalized']
 parts += [f'type plain {name}','var result plain',('if err=decodeData(raw,&result,[]dataField{' if name=='ComboMarket' else 'if err:=decodeData(raw,&result,[]dataField{')]
 for key,f in fields:
  r=rule(cls,key,f);a=stripped(f.annotation)
  if key=='condition_id' and cls is not builder.BuilderTrade:r=r.replace('name:"condition_id"','name:"condition_id",steps:[]string{"condition_feed"}')
  if cls is rfq.ComboMarket and key=='id':r=r.replace('name:"id"','name:"id",kind:"id"')
  if cls is builder.BuilderTrade:
   if a is Decimal:r=r.replace('kind:"data_decimal"','kind:"string_decimal"')
   if a is datetime:r=r.replace('kind:"epoch_datetime"','kind:"clob_epoch_iso"')
  elif a is datetime:r=r.replace('kind:"epoch_datetime"','kind:"reward_epoch_ms"')
  parts.append(r)
 parts+=['});err!=nil{return err}',f'*m={name}(result)','return nil','}']
header='''// Public CLOB models ported from the pinned Python SDK. This file is linted.

package models

import (
 "encoding/json"
 "time"

 "github.com/ericlagergren/decimal"
)
'''
existing=set()
for path in (ROOT/'models').glob('*.go'):
 if path.name!='public_clob_models.go':existing.update(re.findall(r'^type (\w+) ',path.read_text(),re.M))
for name,base in sorted(ids):
 if name not in existing and name!='CtfConditionID':header+=f'// {name} identifies a public CLOB resource.\ntype {name} {base}\n'
(ROOT/'models/public_clob_models.go').write_text(header+'\n'.join(parts))
subprocess.run(['gofmt','-w',str(ROOT/'models/public_clob_models.go')],check=True)
