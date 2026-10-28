"""Regenerate every model and parity fixture from the pinned Python environment."""
import os,subprocess,sys
from pathlib import Path
root=Path(__file__).resolve().parents[1]
names=['market_types','event_types','catalog_types','data_types','data_query_types','data_variants','public_clob_types','stream_types','stream_events','approval_catalog',None,'discovery','events','catalog_queries','data_models','data_queries','public_clob_queries','approvals','stream_parity','decimal','union_parity','surface']
for name in names:
 script='generate'+('_'+name if name else '')+'.py'
 subprocess.run([sys.executable,str(root/'tests/parity'/script)],cwd=root,check=True)
subprocess.run([sys.executable,str(root/'scripts/generate_exports.py')],cwd=root,check=True)
subprocess.run([os.environ.get('GOLANGCI_LINT','golangci-lint'),'fmt'],cwd=root,check=True)
