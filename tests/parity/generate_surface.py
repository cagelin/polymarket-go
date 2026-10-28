"""Record the actual public client surface, including the asynchronous subscriber."""
import inspect,json
from support import ROOT,manifest
from polymarket.clients import PublicClient,AsyncPublicClient
from generate_market_types import go_name
members=sorted({name for cls in [PublicClient,AsyncPublicClient] for name,value in inspect.getmembers(cls) if not name.startswith('_') and (callable(value) or isinstance(value,property)) and 'Experimental:' not in (inspect.getdoc(value) or '')})
items=[{'python':name,'go':go_name(name).replace('Pnl','PnL')} for name in members]
(ROOT/'testdata/public-surface.json').write_text(json.dumps({'upstream_version':manifest['version'],'upstream_commit':manifest['commit'],'members':items},indent=2)+'\n')
print('Generated',len(items),'public client members')
