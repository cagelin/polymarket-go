"""Generate event/search models from the same source-verified SDK as markets."""
import re
import subprocess
from polymarket.models.gamma import common, event, search
from generate_market_types import ROOT, emit_classes, ids

classes=[common.ImageOptimization,common.Partner,common.BestLine,common.Team,common.SportsMetadata,common.TagReference]
classes += [getattr(event,n) for n in event.__all__]
classes += [search.Profile,search.SearchTag,search.SearchResults]
parts=emit_classes(classes)
existing=set(re.findall(r'^type (\w+) ',(ROOT/'models/market_models.go').read_text(),re.M))
header='''// Event and search models ported from polymarket-client 0.14.0.
// Regenerate with tests/parity/generate_event_types.py. These files are linted.

package models

import (
 "encoding/json"
 "time"

 "github.com/ericlagergren/decimal"
)
'''
for name,base in sorted(ids):
    if name not in existing:
        header+=f'\n// {name} identifies a {name.removesuffix("ID")} resource.\ntype {name} {base}\n'
header+='''
// TeamOrdering identifies a team's side in a sports event.
type TeamOrdering string

// Supported team positions.
const (
 TeamHome TeamOrdering = "home"
 TeamAway TeamOrdering = "away"
)
'''
(ROOT/'models/event_models.go').write_text(header+'\n'+parts)
subprocess.run(['gofmt','-w',str(ROOT/'models/event_models.go')],check=True)
