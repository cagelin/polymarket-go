"""Generate the remaining Gamma models from the pinned, verified SDK."""
from pathlib import Path
import re,subprocess
from polymarket.models.gamma import common,tag,series,comment,profile
from generate_market_types import ROOT,emit_classes,ids
classes=[common.CategoryReference,common.CollectionReference,common.SeriesReference,common.TemplateReference,common.Chat,common.SportsMarketTypes]
for module in [tag,series,comment,profile]:classes += [getattr(module,name) for name in module.__all__]
parts=emit_classes(classes)
existing=set()
for path in (ROOT/'models').glob('*.go'):
 if path.name!='catalog_models.go':existing.update(re.findall(r'^type (\w+) ',path.read_text(),re.M))
header='''// Catalog models ported from the pinned Python SDK. This file is linted.

package models

import (
 "time"
 "github.com/ericlagergren/decimal"
)
'''
for name,base in sorted(ids):
 if name not in existing:header+=f'\n// {name} identifies a catalog resource.\ntype {name} {base}\n'
(ROOT/'models/catalog_models.go').write_text(header+'\n'+parts)
subprocess.run(['gofmt','-w',str(ROOT/'models/catalog_models.go')],check=True)
