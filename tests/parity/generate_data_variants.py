import sys,json
import generate_market_types
from polymarket.models.data.activity import _KNOWN_ACTIVITY_TYPES,_COMBO_ACTIVITY_TYPES
from pathlib import Path
text='''package models

import (
 "encoding/json"
 "fmt"
 "github.com/cagelam/polymarket-go/internal/codec"
)
'''
for union,mapping,extra in [('Activity',_KNOWN_ACTIVITY_TYPES,['ComboTradeActivity','UnknownActivity']),('ComboActivity',_COMBO_ACTIVITY_TYPES,[]),('BiggestWinner',{'market':type('MarketBiggestWinner',(),{}),'combo':type('ComboBiggestWinner',(),{})},[])]:
 marker=union[0].lower()+union[1:]+'Variant'
 text+=f'// {union} is one of the concrete public response variants. Use a type switch to inspect it.\ntype {union} interface {{ {marker}() }}\n'
 for name in [v.__name__ for v in mapping.values()]+extra:text+=f'func (*{name}) {marker}() {{}}\n'
 plural={'Activity':'Activities','ComboActivity':'ComboActivities','BiggestWinner':'BiggestWinners'}[union]
 text+=f'''// Parse{plural} validates and decodes the tagged response variants.
func Parse{plural}(raw []byte) ([]{union},error) {{
 items,err:=codec.ParseList[json.RawMessage](raw);if err!=nil {{return nil,err}}
 result:=make([]{union},0,len(items))
 for _,item:=range items {{value,parseErr:=Parse{union}(item);if parseErr!=nil{{return nil,parseErr}};result=append(result,value)}}
 return result,nil
}}
// Parse{union} preserves the concrete response variant.
func Parse{union}(raw []byte) ({union},error) {{
 fields,err:=codec.Object(raw);if err!=nil{{return nil,err}}
 var kind string
 _ = json.Unmarshal(fields["{'kind' if union=='BiggestWinner' else 'type'}"],&kind)
 var result {union}
 switch kind {{
'''
 for value,cls in mapping.items():text+=f'case {json.dumps(str(value))}:result=&{cls.__name__}{{}}\n'
 if union=='Activity':
  text+='''default:
  original,encodeErr:=json.Marshal(fields);if encodeErr!=nil{return nil,encodeErr}
  fields["raw"]=original
  fields["type"],err=json.Marshal(kind);if err!=nil{return nil,err}
  raw,err=json.Marshal(fields);if err!=nil{return nil,err}
  result=&UnknownActivity{}
 }
 if kind=="TRADE" && string(fields["is_combo"])=="true" {result=&ComboTradeActivity{}}
'''
 else:text+='default:return nil,fmt.Errorf("unknown '+union+' variant %q",kind)\n}\n'
 text+='if err=json.Unmarshal(raw,result);err!=nil{return nil,err};return result,nil\n}\n'
(generate_market_types.ROOT/'models/data_variants.go').write_text(text)
