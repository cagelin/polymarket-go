"""Compare SDK arithmetic with Python's default decimal context."""
import json,random
from decimal import Decimal,localcontext
from support import ROOT,manifest
values=['0','-0.00','1','-1','0.1','9.9999999999999999999999999999','1e-100001','1e-1000026','1e-1000027','5e-1000027','6e-1000027','1e999999','1e1000000','1e-999999999999999999','1e999999999999999999']
rng=random.Random(20261028)
for _ in range(80):
 values.append(str(rng.randrange(-10**40,10**40))+'e'+str(rng.choice([-1000067,-1000060,-1000030,-60,-20,0,30,999970])))
cases=[]
for op in ['add','sub','mul']:
 pairs=[(x,y) for x in values[:15] for y in values[:15]]+[(rng.choice(values),rng.choice(values)) for _ in range(150)]
 for x,y in pairs:
  item={'operation':op,'x':x,'y':y}
  try:
   with localcontext() as ctx:
    result=getattr(ctx,{'add':'add','sub':'subtract','mul':'multiply'}[op])(Decimal(x),Decimal(y))
   item['value']=str(result)
  except Exception as err:item['error']=type(err).__name__
  cases.append(item)
(ROOT/'testdata/decimal-parity.json').write_text(json.dumps({'upstream_version':manifest['version'],'upstream_commit':manifest['commit'],'cases':cases},indent=2)+'\n')
print('Generated',len(cases),'decimal arithmetic cases')
