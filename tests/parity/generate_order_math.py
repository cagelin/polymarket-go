"""Execute pinned Python order rounding and fee arithmetic, without network access."""
import json,itertools
from decimal import Decimal as D
from support import ROOT,manifest
from polymarket._internal.actions.orders.limit import _compute_limit_order_amounts
from polymarket._internal.actions.orders.market import _compute_market_order_amounts,adjust_buy_amount_for_fees
from polymarket._internal.actions.orders.market_data import PlatformFeeInfo
from polymarket._internal.actions.orders.context import validate_price_on_tick_grid
cases=[]
for tick,price,amount,side in itertools.product(['0.1','0.01','0.005','0.0025','0.001','0.0001'],['0.0001','0.0025','0.1','0.3333','0.5','0.995','0.9999'],['0.001','0.019','1','1.23456789','12.99999999999','999999999999.999999','1e-40','1e30'],['BUY','SELL']):
 for kind,fn in [('limit',_compute_limit_order_amounts),('market',_compute_market_order_amounts)]:
  case=dict(operation=kind,tick=tick,price=price,amount=amount,side=side)
  try:case['expected']=list(map(str,fn(price=D(price),tick_size=D(tick),side=side,**{('size' if kind=='limit' else 'amount'):D(amount)})))
  except Exception as err:case['error']=type(err).__name__
  cases.append(case)
for price,rate,exponent,spend in itertools.product(['0.0025','0.1','0.3333','0.5','0.9999'],['0','0.025','0.2'],['0','1','2','0.5','1.5'],['0.01','10','1000']):
 case=dict(operation='fees',price=price,rate=rate,exponent=exponent,spend=spend,amount='10.123456789',builder='0.001')
 try:case['expected']=[str(adjust_buy_amount_for_fees(amount=D(case['amount']),price=D(price),max_spend=D(spend),fee=PlatformFeeInfo(rate=D(rate),exponent=D(exponent)),builder_taker_fee_rate=D(case['builder'])))]
 except Exception as err:case['error']=type(err).__name__
 cases.append(case)
for tick,price in itertools.product(['0.1','0.01','0.005','0.0025','0.001','0.0001'],['0','0.0001','0.0025','0.003','0.1','0.3333','0.9999','1']):
 case=dict(operation='grid',tick=tick,price=price)
 try:case['expected']=[str(validate_price_on_tick_grid(D(price),D(tick),'price'))]
 except Exception as err:case['error']=type(err).__name__
 cases.append(case)
(ROOT/'testdata/order-math-parity.json').write_text(json.dumps({'upstream_version':manifest['version'],'upstream_commit':manifest['commit'],'cases':cases},indent=2)+'\n')
print('Generated',len(cases),'order math cases')
