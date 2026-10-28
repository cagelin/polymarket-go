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
    if isinstance(value,BaseModel):return {key:normalize(getattr(value,key)) for key,field in type(value).model_fields.items() if not field.deprecated and not (field.description or '').startswith('Deprecated')}
    if is_dataclass(value):return normalize(asdict(value))
    if isinstance(value,dict):return {k:normalize(v) for k,v in value.items()}
    if isinstance(value,(tuple,list)):return [normalize(v) for v in value]
    return value
