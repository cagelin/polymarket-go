import sys,inspect,typing,json,subprocess
from pathlib import Path
import generate_market_types as base
from polymarket.clients.public import PublicClient
from polymarket._internal.actions import data
from polymarket.models.data import common
from enum import Enum
root=base.ROOT
names=['list_trades','list_activity','list_combo_activity','list_positions','list_combo_positions','get_portfolio_value','get_user_stats','get_user_pnl','get_user_volume','list_market_holders','get_open_interests','get_event_live_volume','get_token_references','get_resolutions','list_trader_leaderboard','get_trader_leaderboard_standing','list_biggest_winners','list_builder_leaderboard','get_builder_volumes','download_accounting_snapshot']
outputs=['Trade','Activity','ComboActivity','Position','ComboPosition','PortfolioValue','UserStats','UserPnLSeries','UserVolume','MetaHolder','[]OpenInterest','LiveVolume','[]TokenReference','[]Resolution','TraderLeaderboardEntry','TraderLeaderboardStanding','BiggestWinner','BuilderStanding','[]BuilderVolumePoint','bytes']
paths=['trades','activity','activity/combos','positions','positions/combos','value','user-stats','user-pnl','user-volume','holders','oi','live-volume','tokens','resolutions','leaderboard','leaderboard','biggest-winners','builders/leaderboard','builders/volume','accounting/snapshot']
def name(s):return base.go_name(s).replace('Pnl','PnL').replace('Ids','IDs')
text='''package data

import (
 "context"
 "github.com/cagelam/polymarket-go/models"
 "github.com/cagelam/polymarket-go/pagination"
)
'''
meta=[]
for op,out,path in zip(names,outputs,paths):
 sig=inspect.signature(getattr(PublicClient,op)); method=name(op);param=method+'Params';isp=op.startswith('list_');qs=[]
 text+='\n// '+param+' selects filters for '+method+'. Nil optional fields are omitted.\ntype '+param+' struct {\n'
 for key,p in sig.parameters.items():
  if key=='self':continue
  typ='string';ann=str(p.annotation)
  if key in ('condition_id','condition_ids','asset_ids','activity_types','event_id','event_ids') or (key=='status' and op=='list_combo_positions'):
   typ='[]int64' if key.startswith('event_') else '[]string'
  elif key in ('page_size','bucket_limit'):typ='*int'
  elif key in ('start','end','updated_after','updated_before'):typ='*int64'
  elif key in ('filter_amount','min_balance'):typ='*float64'
  elif key in ('taker_only','include_archived','include_pnl'):typ='*bool'
  elif key=='full_history':typ='bool'
  elif p.default is None:typ='*string'
  text+=f'{name(key)} {typ} `json:"{key}"`\n'
  if key=='page_size':continue
  if typ.startswith('*'):qs.append(f'put(query,"{key}",p.{name(key)})')
  elif typ.startswith('[]'):qs.append(f'if p.{name(key)} != nil {{ query["{key}"] = append(make({typ},0,len(p.{name(key)})),p.{name(key)}...) }}')
  else:qs.append(f'query["{key}"] = p.{name(key)}')
 text+='}\n'
 result=('*pagination.Paginator[models.'+out+']') if isp else ('[]byte' if out=='bytes' else '[]models.'+out[2:] if out.startswith('[]') else '*models.'+out)
 args='p '+param if isp else 'ctx context.Context,p '+param
 text+=f'// {method} reads public Data API information.\nfunc (c *Client) {method}({args}) ({result},error) {{\n query:=map[string]any{{}}\n'+ '\n'.join(qs)+'\n'
 text+=f'if err:=normalizeQuery("{op}",query);err!=nil {{ return nil,err }}\n'
 if isp:
  cap='1000'
  if op=='list_market_holders':text+='maximum:=1000\nif p.IncludePnL!=nil && *p.IncludePnL { maximum=100 }\n';cap='maximum'
  parser={'Activity':'models.ParseActivities','ComboActivity':'models.ParseComboActivities','BiggestWinner':'models.ParseBiggestWinners'}.get(out,'parseList[models.'+out+']')
  text+=f'return listData(c,"/v2/{path}",query,p.PageSize,{cap},{parser})\n'
 elif out=='bytes':text+='return c.transport.Request(ctx,c.transport.DataURL,"GET","/v1/accounting/snapshot",queryValues(query),nil)\n'
 elif out.startswith('[]'):text+=f'return getData(ctx,c,"/v2/{path}",query,parseList[models.{out[2:]}])\n'
 else:text+=f'return getData(ctx,c,"/v2/{path}",query,parseObject[models.{out}]({str(op in ["get_user_stats","get_trader_leaderboard_standing"]).lower()}))\n'
 text+='}\n'
 meta.append(dict(op=op,method=method,params=param,output=out,paginated=isp,path=('/v1/' if out=='bytes' else '/v2/')+path))
(root/'internal/data/queries.go').write_text(text)
(root/'tests/parity/data-operations.json').write_text(json.dumps(meta,indent=2)+'\n')
# Persist generation script after imports are rooted correctly.
