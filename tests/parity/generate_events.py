"""Generate event/search/tag parity cases by executing the pinned Python SDK."""
import ast
import base64
import copy
import inspect
import json
import textwrap
from dataclasses import asdict,is_dataclass
from datetime import datetime,date,timezone
from decimal import Decimal
from types import SimpleNamespace
import httpx
from pydantic import BaseModel
from polymarket._internal.actions import gamma
from polymarket._internal.dispatch import sync_paginate_keyset,sync_paginate_page_based
from polymarket.models.gamma import event
from generate_market_types import ROOT,manifest


def normalize(value):
    if isinstance(value,Decimal):return str(value)
    if isinstance(value,datetime):
        value=value.replace(tzinfo=timezone.utc) if value.tzinfo is None else value.astimezone(timezone.utc)
        return value.strftime('%Y-%m-%dT%H:%M:%S')+('.'+f'{value.microsecond:06d}'.rstrip('0') if value.microsecond else '')+'Z'
    if isinstance(value,date):return value.isoformat()
    if isinstance(value,BaseModel):return {key:normalize(getattr(value,key)) for key,field in type(value).model_fields.items() if not field.deprecated}
    if is_dataclass(value):return normalize(asdict(value))
    if isinstance(value,dict):return {k:normalize(v) for k,v in value.items()}
    if isinstance(value,(tuple,list)):return [normalize(v) for v in value]
    return value

cases=[]
def add(name,operation,args,response,cursor=None):
    case=dict(name=name,operation=operation,input=copy.deepcopy(args),response=copy.deepcopy(response),request=None)
    if cursor is not None:case['cursor']=cursor
    class Transport:
        def get_json(self,path,params=None):
            query=httpx.QueryParams(params or {})
            case['request']=dict(path=path,query={key:query.get_list(key) for key in query})
            return copy.deepcopy(response)
    transport=Transport()
    ctx=SimpleNamespace(gamma=transport)
    try:
        kwargs=copy.deepcopy(args)
        if operation=='get_event':
            spec=gamma.get_event_spec(**(dict(id=None,slug=None,url=None,include_best_lines=None,include_chat=None,include_template=None,locale=None)|kwargs))
        elif operation in ('get_event_tags','get_market_tags'):
            spec=getattr(gamma,operation+'_spec')(**kwargs)
        elif operation=='list_events':
            size=kwargs.pop('page_size',20)
            spec=gamma.list_events_spec(**kwargs)
        elif operation=='search':
            size=kwargs.pop('page_size',10)
            spec=gamma.search_spec(**kwargs)
        else:raise AssertionError(operation)
        if operation in ('list_events','search'):
            factory=sync_paginate_keyset if operation=='list_events' else sync_paginate_page_based
            pager=factory(ctx,spec,page_size=size)
            if cursor is not None:pager=pager.from_cursor(cursor)
            value=pager.first_page()
        else:value=spec.parse(transport.get_json(spec.path,params={k:v for k,v in (spec.params or {}).items() if v is not None}))
        case['expected']={'value':normalize(value)}
    except Exception as error:case['expected']={'error':type(error).__name__}
    cases.append(case)
    return case

market=dict(id='42',outcomes=['Yes','No'],outcomePrices=['0.55','0.45'],clobTokenIds=['123','456'])
minimal=dict(id=123)
full=dict(id=123,parentEventId=50,version='v2',ticker='TICK',slug='sample',title='Event',subtitle='Subtitle',description='Details',category='sports',subcategory='soccer',image='image',icon='icon',featuredImage='feature',createdAt='2026-10-01 12:00:00+00',updatedAt='2026-10-01T12:00:00.123456789+08:00',published_at='2026-10-01',eventMetadata={'nested':{'note':'中文','count':9007199254740993}})
# Cover every scalar flat field in the actual upstream normalizer, rather than
# hand-picking a subset of the Event model.
normalizer=ast.parse(textwrap.dedent(inspect.getsource(event.Event._normalize_event)))
wire=next(node.value for node in ast.walk(normalizer) if isinstance(node,ast.Return) and isinstance(node.value,ast.Dict))
for key,group in zip(wire.keys,wire.values):
    if not isinstance(group,ast.Dict):continue
    model=event.Event.model_fields[key.value].annotation
    for field,expr in zip(group.keys,group.values):
        if not isinstance(expr,ast.Call):continue
        field_type=str(model.model_fields[field.value].annotation)
        wire_key=expr.args[0].value
        if 'datetime.datetime' in field_type:value='2026-10-01T23:45:01+08:00'
        elif 'datetime.date' in field_type:value='2026-10-01'
        elif 'decimal.Decimal' in field_type:value='123.000000000000000001'
        elif 'bool' in field_type:value=True
        elif 'float' in field_type:value='1.25'
        elif 'int' in field_type:value='3'
        elif 'str' in field_type:value='value-'+field.value
        else:continue
        full[wire_key]=value
full.update(markets=[market,dict(id='skip',outcomes=['A','B','C'])],series=[dict(id=4,slug='series',title='Series',subtitle='sub',description='d',image='i',icon='c',active=True,closed=False,archived=False,volume='9',liquidity=1.5,startDate='2026-01-01')],tags=[dict(id='7',slug='tag',label='Tag')],eventCreators=[dict(id='8',creatorName='Name',creatorHandle='Handle',creatorURL='url',creatorImage='image',createdAt='2026-01-01',updatedAt='2026-01-02')],externalPartners=[dict(id='9',externalId='external',partner=dict(id=10,slug='partner',name='Partner'),createdAt='2026-01-01',updatedAt='2026-01-02')],bestLines=[dict(id='11',lineType='spread',line=1.5)],teams=[dict(id=12,name='Team',league='L',record='R',logo='logo',abbreviation='T',alias='A',providerId='13',color='red',ordering='home',createdAt='2026-01-01',updatedAt='2026-01-02')],sport=dict(id=14,sport='sport',name='Sport',image='i',resolution='r',ordering='o',tags='t',series='s',createdAt='2026-01-01',updatedAt='2026-01-02'))
image=dict(id='15',imageUrlSource='source',imageUrlOptimized='optimized',imageSizeKbSource=1.5,imageSizeKbOptimized='0.5',imageOptimizedComplete='true',imageOptimizedLastUpdated=1720000000,relID='16',field='image',relname='events')
for key in ('imageOptimized','iconOptimized','featuredImageOptimized'):full[key]=image
add('event/full','get_event',{'id':'123'},full)
add('event/minimal','get_event',{'id':'123'},minimal)
for args in [{},{'id':'1','slug':'s'},{'id':'','slug':'s'},{'slug':'a/b +中文😀'},{'url':'https://polymarket.com/event/a%2Fb?x=1#f'},{'url':'https://www.polymarket.com/event/sample/'},{'url':'https://polymarket.com/event/a/b'},{'url':'https://polymarket.com/market/a'},{'url':'http://polymarket.com/event/a'},{'url':'https://polymarket.com:443/event/a'},{'url':'https://evil.test/event/a'},{'id':'123','include_best_lines':True,'include_chat':False,'include_template':True,'locale':'中文'},{'id':'123','include_best_lines':False,'include_chat':True,'include_template':False,'locale':''}]:
    add('event/lookup/'+repr(args),'get_event',args,minimal)
for key,values in {
'id':[None,True,1.5,'123'], 'parentEventId':[None,False,1.5,42,''], 'version':['bad','',None],
'active':['yes',2], 'volume':[True,'NaN','',0.000001], 'eventDate':['','bad','2026-10-01T23:45:00-08:00',1720000000],
'finishedTimestamp':['bad',1720000000,'2026-10-01'], 'publishedAt':['2026-01-01'],
'markets':[None,'[]','null',{},[None],[dict(outcomes=[])],[dict(outcomes=[1,2])],[dict(outcomes='bad')]],
'series':['[]',[dict(id=4)],[dict(id=True)]], 'tags':['[]',[dict(id=1)]],
'eventCreators':[[dict(id='1',creatorUrl='',creatorURL='fallback')],[dict(id='1',creatorUrl=None)],[dict(id='1',creatorUrl=False)],[dict(id='1',creatorUrl=0)],'bad'],
'externalPartners':[[dict(id=1,externalId='x',partner='ignored')],[dict(id=1,externalId='x',partner={})]],
'bestLines':[None,[],False,'',[dict(id=1)]], 'teams':[None,[dict(id=1,ordering='away')],[dict(id=1,ordering='bad')]],
'eventMetadata':[None,[],{'a':[1,True,None]}],
'imageOptimized':[{},{'id':'1','imageOptimizedLastUpdated':''},{'id':'1','imageOptimizedLastUpdated':'1720000000000'},{'id':'1','imageOptimizedLastUpdated':True}],
}.items():
    for value in values:add('event/'+key+'/'+repr(value),'get_event',{'id':'123'},minimal|{key:value})
for timestamp in [-0.001,-1720000000.123456,1720000000.1234567,1720000000123.456,20000000000,20000000001,'2026-10-01T12:34:56.123456789Z','bad',False]:
    case=add('event/image_timestamp/'+repr(timestamp),'get_event',{'id':'123'},minimal|{'imageOptimized':{'id':'1','imageOptimizedLastUpdated':timestamp}})
    if isinstance(timestamp,float) and timestamp<0 and timestamp!=int(timestamp):
        # UP-001: pinned pydantic-core mirrors the fractional part of negative
        # float timestamps. Python's datetime and Go agree on the correct instant.
        case['compatibility_difference']='UP-001'
        case['python_value']=copy.deepcopy(case['expected']['value'])
        corrected=normalize(datetime.fromtimestamp(timestamp,timezone.utc))
        case['expected']['value']['display']['image_optimized']['image_optimized_last_updated']=corrected
normalized=normalize(event.Event.parse_response(full))
add('event/normalized','get_event',{'id':'123'},normalized)
for key in ['state','schedule','metrics','display','trading','resolution','estimation','sports','partners','markets','series','tags','creators']:
    add('event/normalized_null/'+key,'get_event',{'id':'123'},normalized|{key:None})
add('event/normalized_empty_sports','get_event',{'id':'123'},normalized|{'sports':{}})
for key in ['best_lines','teams']:
    sports=copy.deepcopy(normalized);sports['sports'][key]=None
    add('event/sports_null/'+key,'get_event',{'id':'123'},sports)

page=dict(events=[minimal,full],next_cursor='next/中文😀')
filters=dict(ascending=False,closed=True,cyom=False,end_date_max='max',end_date_min='',ended=True,event_date='2026-10-01',event_week=0,exclude_tag_ids=[1,2],featured=False,featured_order=True,game_ids=[3],ids=[4,5],include_best_lines=True,include_chat=False,include_children=True,include_template=False,liquidity_max=1e16,liquidity_min=0.000001,live=True,locale='中文',order=' volume,liquidity ',parent_event_id=0,partner_slug='',recurrence='daily',related_tags=True,series_ids=[6],slug=['x','😀'],start_date_max='max',start_date_min='min',start_time_max='time',start_time_min='',tag_ids=[7],tag_match='all',tag_slug='tag',title_search='test',version=['v1','v2'],volume_max=10.5,volume_min=0.0,page_size=2)
first=add('events/all_filters','list_events',filters,page)
continuation=first['expected']['value']['next_cursor']
add('events/resume','list_events',filters|{'page_size':3},dict(events=[],next_cursor=None),continuation)
add('events/wrong_query','list_events',{},page,continuation)
for args in [{},{'closed':True},{'page_size':0},{'page_size':-1},{'page_size':5000},{'recurrence':'weekly'},{'recurrence':'monthly'},{'recurrence':'bad'},{'recurrence':''},{'tag_match':'any'},{'tag_match':'bad'},{'version':[]},{'version':['v1','v2','v1']},{'version':['bad']},{'version':['v2']},{'slug':[],'ids':[]}]:
    add('events/params/'+repr(args),'list_events',args,page)
for response in [None,[],{},dict(events=None),dict(events=[None]),dict(events=[]),dict(events=[],next_cursor=''),dict(events=[],next_cursor=2),dict(events=[minimal],next_cursor=None)]:
    add('events/response/'+repr(response),'list_events',{},response)

profile=dict(id='p',name='name',user=1,referral='ref',createdBy='2',updatedBy=3,createdAt='2026-01-01',updatedAt='2026-01-02',utmSource='source',utmMedium='medium',utmCampaign='campaign',utmContent='content',utmTerm='term',walletActivated=True,pseudonym='pseudo',displayUsernamePublic=True,profileImage='image',bio='bio',proxyWallet='proxy',profileImageOptimized=image,isCloseOnly=False,isCertReq=True,certReqDate='2026-01-01',discordUsername='d',xUsername='x',verifiedBadge=True,dubPartnerId='id',termsAcceptedAt='2026-01-01',viewOnlyAcknowledgedAt='2026-01-02',isReferralRestricted=False)
results=dict(events=[full],tags=[dict(id='7',label='Tag',slug='tag',event_count='3')],profiles=[profile],pagination=dict(hasMore=True,totalResults=123))
filters=dict(q='中文 😀 election',ascending=False,cache=True,events_status='active',events_tag=['tag1','tag2'],exclude_tag_ids=[1,2],keep_closed_markets=0,optimized=False,presets=['a','b'],recurrence='weekly',search_profiles=True,search_tags=False,sort='volume',page_size=2)
first=add('search/all_filters','search',filters,results)
search_cursor=first['expected']['value']['next_cursor']
add('search/resume_original_size','search',filters|{'page_size':3},results,search_cursor)
add('search/wrong_query','search',{'q':'other'},results,search_cursor)
add('search/event_cursor','search',{'q':'x'},results,continuation)
for args in [dict(q=''),dict(q=' '),dict(q='x'),dict(q='x',page_size=0),dict(q='x',page_size=-1),dict(q='x',page_size=5000),dict(q='x',recurrence='bad'),dict(q='x',recurrence='daily'),dict(q='x',recurrence='monthly'),dict(q='x',presets=[],events_tag=[],exclude_tag_ids=[])]+[dict(q='x',sort=sort) for sort in ['volume','volume_24hr','liquidity','competitive','closed_time','start_date','end_date','bad','']]:
    add('search/params/'+repr(args),'search',args,results)
for response in [None,[],{},dict(events=None,tags=None,profiles=None),dict(events={}),dict(events='[]'),dict(tags=[None]),dict(profiles=[None]),dict(profiles=[dict(wallet='wallet',proxyWallet='ignored')]),dict(profiles=[dict(wallet=None,proxyWallet='ignored')]),dict(profiles=[dict(user=True)]),dict(profiles=[dict(id=1)])]:
    add('search/response/'+repr(response),'search',{'q':'x'},response)
for pagination in [None,{},[],dict(hasMore=None,totalResults=None),dict(hasMore=1),dict(hasMore='true'),dict(totalResults=True),dict(totalResults='3'),dict(totalResults=3.0),dict(totalResults=-1),dict(hasMore=False,totalResults=0),dict(hasMore=True)]:
    add('search/pagination/'+repr(pagination),'search',{'q':'x'},dict(pagination=pagination))
for field,values in {'v':[None,2,True,1.0],'svc':['data',None],'p':['/events/keyset'],'f':['bad'],'pg':[0,-1,True,'2',2.0,None],'s':[0,-1,False,'2',2.0,None]}.items():
    for value in values:
        payload=json.loads(base64.b64decode(search_cursor));payload[field]=value
        mutated=base64.b64encode(json.dumps(payload,separators=(',',':'),sort_keys=True).encode()).decode()
        add('search/cursor/'+field+'/'+repr(value),'search',filters,results,mutated)
for cursor in ['', 'bad', search_cursor+'\n']:
    add('search/bad_cursor/'+repr(cursor),'search',filters,results,cursor)

tag=dict(id='7',label='Tag',slug='tag',forceShow=True,publishedAt='2026-01-01',createdBy='1',updatedBy=2,createdAt='2026-01-01',updatedAt='2026-01-02',forceHide=False,isCarousel=True,requiresTranslation=False,activeEventsCount='3')
for operation in ['get_event_tags','get_market_tags']:
    add(operation+'/full',operation,{'id':'a/b 中文'},[tag])
    add(operation+'/empty_id',operation,{'id':''},[])
    for response in [None,{},[],[None],[dict(id=1)],[dict(id='1',forceShow='yes')]]:
        add(operation+'/response/'+repr(response),operation,{'id':'123'},response)

(ROOT/'testdata/events-parity.json').write_text(json.dumps(dict(upstream_version=manifest['version'],upstream_commit=manifest['commit'],cases=cases),ensure_ascii=False,indent=2)+'\n')
print('Generated',len(cases),'event/search/tag cases')
