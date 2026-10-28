"""Emit typed market models from the pinned SDK; field coercions are handwritten in Go."""
from pathlib import Path
import hashlib
import importlib.metadata
import json
import subprocess
import polymarket
from typing import get_args, get_origin
from decimal import Decimal
from datetime import datetime, date
from polymarket.models.gamma import market

ROOT = Path(__file__).resolve().parents[2]
manifest=json.loads((ROOT/'tests/parity/upstream.json').read_text())
assert importlib.metadata.version('polymarket-client')==manifest['version']
source=Path(polymarket.__file__).parent
for entry in manifest['source_hashes']:
    assert hashlib.sha256((source/Path(entry['path']).relative_to('src/polymarket')).read_bytes()).hexdigest()==entry['sha256']
def go_name(name):
    special={'id':'ID','ids':'IDs','url':'URL','clob':'Clob','amm':'AMM','uma':'UMA','onchain':'Onchain'}
    return ''.join(special.get(part, part[:1].upper()+part[1:]) for part in name.split('_'))
def named(name):
    return {'UmaResolutionStatus':'UMAResolutionStatus','EvmAddress':'EVMAddress'}.get(name,name.replace('Id','ID'))
ids=set()
def go_type(annotation):
    args=get_args(annotation)
    if type(None) in args:
        rest=[a for a in args if a is not type(None)]
        inner=rest[0] if len(rest)==1 else str
        base=go_type(inner)
        return base if base.startswith(('[]','*')) else '*'+base
    if annotation is str: return 'string'
    if annotation is bool: return 'bool'
    if annotation is int: return 'int64'
    if annotation is float: return 'float64'
    if annotation is Decimal: return '*apd.Decimal'
    if annotation is datetime: return 'time.Time'
    if annotation is date: return 'Date'
    if get_origin(annotation) is tuple: return '[]'+go_type(args[0])
    if set(args)=={int,float}: return 'json.Number'
    if hasattr(annotation,'__supertype__'):
        name=named(annotation.__name__)
        ids.add((name,'int64' if annotation.__supertype__ is int else 'string'))
        return name
    return named(annotation.__name__)
def kind(cls,name,field):
    a=field.annotation; args=get_args(a)
    if type(None) in args:
        rest=[v for v in args if v is not type(None)]
        a=rest[0] if len(rest)==1 else str
    if cls=='Market' and name=='condition_id': return 'condition'
    if cls=='ClobReward' and name=='condition_id': return 'condition'
    if cls=='MarketEvent' and name=='id': return 'id'
    if cls=='MarketResolution':
        if name=='onchain_event_id': return 'event_id'
        if name=='request_id': return 'request_id'
        if name=='resolution_status': return 'resolution'
        if name=='uma_resolution_status': return 'uma'
        if name in {'question_id','neg_risk_request_id','resolved_by'}: return 'empty_text'
    if a is Decimal: return 'decimal'
    if a is datetime: return 'datetime'
    if a is date: return 'date'
    if a is bool: return 'bool'
    if a is int or getattr(a,'__supertype__',None) is int: return 'int'
    if a is float: return 'float'
    if set(get_args(a))=={int,float}: return 'number'
    if a is market.ProtocolVersion: return 'version'
    return ''
classes=[getattr(market,n) for n in market.__all__ if hasattr(getattr(market,n),'model_fields')]
parts=[]
for cls in classes:
    name=named(cls.__name__)
    fields=[(k,f) for k,f in cls.model_fields.items() if not f.deprecated]
    description=cls.__doc__.splitlines()[0].rstrip('.')
    parts.append('// '+name+' represents '+description[:1].lower()+description[1:]+'.\ntype '+name+' struct {')
    for key,field in fields: parts.append(f'\t{go_name(key)} {go_type(field.annotation)} `json:"{key}"`')
    parts.append('}\n')
    parts.append('// UnmarshalJSON validates market response fields and normalizes wire names.\nfunc (m *'+name+') UnmarshalJSON(data []byte) error {')
    if name=='Market': parts.append('\tvar err error\n\tdata, err = normalizeMarket(data)\n\tif err != nil { return err }')
    parts.append('\ttype plain '+name+'\n\tvar result plain\n\tif err := decodeGamma(data, &result, []gammaField{')
    if name=='Market': parts[-1]=parts[-1].replace('if err :=','if err =')
    for key,field in fields:
        alias=field.validation_alias
        aliases=alias.choices if hasattr(alias,'choices') else [alias] if isinstance(alias,str) else []
        values=['name: "'+key+'"']
        if aliases: values.append('aliases: []string{'+','.join('"'+v+'"' for v in aliases)+'}')
        k=kind(cls.__name__,key,field)
        if k: values.append('kind: "'+k+'"')
        if field.is_required(): values.append('required: true')
        parts.append('\t\t{'+', '.join(values)+'},')
    parts.append('\t}); err != nil { return err }\n\t*m = '+name+'(result)\n\treturn nil\n}\n')
header='''// Market models ported from polymarket-client 0.14.0.
// Regenerate with tests/parity/generate_market_types.py. These files are linted.

package polymarket

import (
 "encoding/json"
 "time"

 "github.com/cockroachdb/apd/v3"
)

// Date is an ISO 8601 calendar date (YYYY-MM-DD).
type Date string
'''
for name,base in sorted(ids):
    if name not in {'ConditionID'}:
        header+=f'\n// {name} identifies a {name.removesuffix("ID")} resource.\ntype {name} {base}\n'
for cls in [market.ProtocolVersion,market.MarketResolutionStatus,market.UmaResolutionStatus]:
    name=named(cls.__name__)
    header+=f'\n// {name} describes known market states.\ntype {name} string\n\n// Supported {name} values.\nconst (\n'
    for v in cls:
        prefix={'ProtocolVersion':'Protocol','MarketResolutionStatus':'Resolution','UmaResolutionStatus':'UMA'}[cls.__name__]
        suffix=v.name if cls is market.ProtocolVersion else v.name.title()
        header+=f'\t{prefix}{suffix} {name} = "{v.value}"\n'
    header+=')\n'
(ROOT/'market_models.go').write_text(header+'\n'+'\n'.join(parts))

subprocess.run(['gofmt','-w',str(ROOT/'market_models.go')],check=True)
