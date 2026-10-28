"""Keep the root SDK's canonical type names in sync with the domain packages."""
from pathlib import Path
import re
import subprocess

ROOT=Path(__file__).resolve().parents[1]
exports=[]
for directory in ['models','errors','internal/clob','internal/gamma','internal/data','internal/rfq','internal/trading','streams']:
    for path in sorted((ROOT/directory).glob('*.go')):
        if path.name.endswith('_test.go'):continue
        source=path.read_text()
        for name in re.findall(r'^type ([A-Z]\w*)\b',source,re.M):
            if name.endswith('Client') or (directory in ['streams','internal/trading'] and name=='Config'):continue
            exports.append(('type',name,directory))
        for block in re.findall(r'^const \((.*?)^\)',source,re.M|re.S):
            for name in re.findall(r'^\s*([A-Z]\w*)\s',block,re.M):exports.append(('const',name,directory))
exports.sort(key=lambda item:item[1])
assert len({name for _,name,_ in exports})==len(exports),'duplicate public name'
text='// Canonical public names; implementations live in domain packages.\n// Regenerate with scripts/generate_exports.py. This file is linted.\n\npackage polymarket\n\nimport (\n'
for directory in sorted({directory for _,_,directory in exports}):text+='"github.com/cagelam/polymarket-go/'+directory+'"\n'
text+='"github.com/cagelam/polymarket-go/pagination"\n)\n'
for kind,name,directory in exports:
    text+=f'\n// {name} is the canonical SDK {"type" if kind=="type" else "value"}.\n{kind} {name} = {Path(directory).name}.{name}\n'
text+='''
// Page holds a batch of results and continuation metadata.
type Page[T any] = pagination.Page[T]
// Paginator lazily retrieves results.
type Paginator[T any] = pagination.Paginator[T]
// Asset selects one CLOB asset.
func Asset(id ClobAssetID) AssetQuery { return clob.Asset(id) }
'''
(ROOT/'types.go').write_text(text)
subprocess.run(['gofmt','-w',str(ROOT/'types.go')],check=True)
