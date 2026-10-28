"""Generate committed fixtures by executing the pinned official Python SDK.

Run with the parity virtual environment; the Go SDK never imports Python.
"""
from __future__ import annotations

import copy
import hashlib
import importlib.metadata
import json
from datetime import datetime
from decimal import Decimal
from pathlib import Path

from pydantic import BaseModel
from polymarket._internal.actions import clob
from polymarket._internal.actions.orders import estimate
from polymarket.models.clob import PriceRequest
import polymarket
import httpx
from polymarket.clients import _transport

ROOT = Path(__file__).resolve().parents[2]
VERSION = "0.14.0"
COMMIT = "7032ed06175b0bbea77b2d9e898bd9f252af2041"
assert importlib.metadata.version("polymarket-client") == VERSION
inventory = json.loads((ROOT / "tests/parity/upstream.json").read_text())
assert inventory["version"] == VERSION and inventory["commit"] == COMMIT
package_root = Path(polymarket.__file__).resolve().parent
for entry in inventory["source_hashes"]:
    relative = Path(entry["path"]).relative_to("src/polymarket")
    actual = hashlib.sha256((package_root / relative).read_bytes()).hexdigest()
    assert actual == entry["sha256"], f"Upstream source changed: {relative}"


def normalize(value):
    if isinstance(value, Decimal):
        return str(value)
    if isinstance(value, datetime):
        return value.strftime("%Y-%m-%dT%H:%M:%S") + (
            ("." + f"{value.microsecond:06d}".rstrip("0")) if value.microsecond else ""
        ) + "Z"
    if isinstance(value, BaseModel):
        # Go exposes canonical fields only; Python compatibility aliases are omitted.
        return normalize(value.model_dump(exclude={"market", "token_id"}))
    if isinstance(value, dict):
        return {str(k): normalize(v) for k, v in value.items()}
    if isinstance(value, (tuple, list)):
        return [normalize(v) for v in value]
    return value


cases = []


def add(name, operation, args, response):
    case = dict(name=name, operation=operation, input=copy.deepcopy(args), response=copy.deepcopy(response))
    try:
        if operation == "estimate_market_price":
            asset, notional = estimate._validate_estimate_inputs(
                asset_id=args.get("asset_id"), token_id=args.get("token_id"),
                side=args.get("side", ""), amount=args.get("amount"), shares=args.get("shares"),
                order_type=args.get("order_type", "FOK"),
            )
            path, payload = clob.build_order_book_request(asset_id=asset)
            case["request"] = dict(method="GET", path=path, query=payload)
            result = estimate._resolve_market_price_context(
                clob.parse_order_book(response), token_id=asset, side=args["side"],
                notional=notional, order_type=args.get("order_type", "FOK"),
            ).price
        else:
            kwargs = copy.deepcopy(args)
            if operation == "prices":
                kwargs["requests"] = [PriceRequest(**entry) for entry in kwargs["requests"]]
            path, payload = getattr(clob, "build_" + operation + "_request")(**kwargs)
            batch = operation in {"midpoints", "prices", "order_books", "spreads", "last_trade_prices"}
            case["request"] = dict(method="POST" if batch else "GET", path=path)
            case["request"]["body" if batch else "query"] = payload
            result = getattr(clob, "parse_" + operation)(response)
        case["expected"] = {"value": normalize(result)}
    except Exception as error:
        case["expected"] = {"error": type(error).__name__}
    cases.append(case)


book = dict(market="condition", asset_id="123", timestamp="1720000000123",
            bids=[dict(price="0.3", size="8"), dict(price="0.4", size="2")],
            asks=[dict(price="0.7", size="8"), dict(price="0.6", size="2")],
            min_order_size="5", tick_size="0.01", neg_risk=False,
            last_trade_price="", hash="book-hash")
responses = dict(midpoint={"mid": "0.5000"}, midpoints={"123": "0.5000"},
                 price={"price": "0.4"}, prices={"123": {"BUY": "0.4", "SELL": "0.6"}},
                 order_book=book, order_books=[book], spread={"spread": "0.2"},
                 spreads={"123": "0.2"}, last_trade_price={"price": "0.4", "side": "BUY"},
                 last_trade_prices=[dict(token_id="123", price="0.4", side="SELL")])
batch_ops = {"midpoints", "order_books", "spreads", "last_trade_prices"}
for operation, response in responses.items():
    args = {"asset_ids": ["123", "456"]} if operation in batch_ops else {"asset_id": "123"}
    if operation == "price":
        args["side"] = "BUY"
    if operation == "prices":
        args = {"requests": [dict(asset_id="123", side="BUY"), dict(asset_id="456", side="SELL")]}
    add(operation + "/valid", operation, args, response)
    add(operation + "/null", operation, args, None)
    add(operation + "/wrong_shape", operation, args, [False] if isinstance(response, dict) else {})

for value in [None, True, 0.5, 1, "", "NaN", "Infinity", "-Infinity", "garbage", " 0.1230 ",
              "1_0", "١.٢", "1e-20", "1E+30", "1e-100001", "1e100001", "1e-999999999999999999", "1e999999999999999999", "1e1000000000000000000", "1e-1000000000000000000", "1e-1999999999999999997", "1e-1999999999999999998", "99e999999999999999999", "-0.00", "0.123456789012345678901234567890123"]:
    add("decimal/" + repr(value), "midpoint", {"asset_id": "123"}, {"mid": value})
for response in [{}, {"mid": "0.5", "future_field": True}]:
    add("midpoint/fields/" + str(len(response)), "midpoint", {"asset_id": "123"}, response)
for args in [{}, {"asset_id": ""}, {"asset_id": " "}, {"asset_id": "x&side=SELL"}]:
    add("input/" + repr(args), "midpoint", args, {"mid": "0.5"})
for args in [{}, {"asset_ids": []}, {"asset_ids": [""]}, {"asset_ids": ["123", "123"]}]:
    add("batch_input/" + repr(args), "midpoints", args, {})
for side in ["", "buy", "SELL", "UNKNOWN"]:
    add("price/side/" + side, "price", {"asset_id": "123", "side": side}, {"price": "0.5"})
for requests in [[], [dict(asset_id="", side="BUY")], [dict(asset_id="123", side="bad")]]:
    add("prices/input/" + repr(requests), "prices", {"requests": requests}, {})
add("prices/bad_response_side", "prices", {"requests": [dict(asset_id="123", side="BUY")]}, {"123": {"bad": "0.2"}})
for field in book:
    payload = copy.deepcopy(book)
    del payload[field]
    add("book/missing/" + field, "order_book", {"asset_id": "123"}, payload)
for field, values in {
    "timestamp": [None, "", "0", "١٢٣", "253402300799999", "-1", "+1", " 1 ", 1720000000123, "invalid", "999999999999999999999"],
    "neg_risk": ["yes", "0", "false", 1, 2, None],
    "bids": [None, [], [{}], [None]], "asks": [None],
    "last_trade_price": [None, "0", 0.4, "NaN"],
    "tick_size": [None, 0.01, "Infinity"],
}.items():
    for value in values:
        add("book/" + field + "/" + repr(value), "order_book", {"asset_id": "123"}, book | {field: value})
alias_book = copy.deepcopy(book)
alias_book["token_id"] = alias_book.pop("asset_id")
add("book/token_alias", "order_book", {"asset_id": "123"}, alias_book)
add("book/canonical_precedes_alias", "order_book", {"asset_id": "123"}, book | {"token_id": "999", "condition_id": "ignored"})
for response in [{"price": "0", "side": ""}, {"price": "bad", "side": ""}, {"price": "0.4", "side": "bad"}, {"price": "0.4"}]:
    add("last_trade/" + repr(response), "last_trade_price", {"asset_id": "123"}, response)
add("last_trades/empty", "last_trade_prices", {"asset_ids": ["123"]}, [])
add("books/null_item", "order_books", {"asset_ids": ["123"]}, [None])
for side, key in [("BUY", "amount"), ("SELL", "shares")]:
    for order_type in ["FOK", "FAK"]:
        for value in ["0.1", "1.2", "1.2000000000000000000000000001", "2", "3", "100", "0", "-1", "NaN", None]:
            add(f"estimate/{side}/{order_type}/{value}", "estimate_market_price",
                dict(asset_id="123", side=side, order_type=order_type, **{key: value}), book)
for changes in [dict(asks=[]), dict(bids=[]), dict(asset_id="999"), dict(tick_size="0.45"),
                dict(asks=[dict(price="0.6", size="2.0000000000000000000000000001")])]:
    for side in ["BUY", "SELL"]:
        key = "amount" if side == "BUY" else "shares"
        add("estimate/book/" + side + repr(changes), "estimate_market_price", dict(asset_id="123", side=side, **{key: "1.2"}), book | changes)
for args in [dict(side="bad", amount="1"), dict(side="BUY", amount="1", shares="1"),
             dict(side="SELL", amount="1", shares="1"), dict(side="BUY", amount="1", order_type="GTC")]:
    add("estimate/input/" + repr(args), "estimate_market_price", {"asset_id": "123"} | args, book)

# These amounts distinguish rounded Decimal arithmetic from exact unlimited
# arithmetic and from binary floating point, at a fill/no-fill boundary.
for side, key, quantity, target in [
    ("BUY", "amount", "2.0000000000000000000000000001", "1.20000000000000000000000000001"),
    ("SELL", "shares", "2.0000000000000000000000000001", "2.00000000000000000000000000001"),
]:
    field = "asks" if side == "BUY" else "bids"
    add("estimate/precision28/" + side, "estimate_market_price",
        dict(asset_id="123", side=side, **{key: target}),
        book | {field: [dict(price="0.6", size=quantity)]})

output = dict(upstream_version=VERSION, upstream_commit=COMMIT, cases=cases)
target = ROOT / "testdata" / "clob-parity.json"
target.parent.mkdir(exist_ok=True)
target.write_text(json.dumps(output, ensure_ascii=False, indent=2) + "\n")
print(f"Generated {len(cases)} Python cases: {target}")

transport_cases = []
for name, status, headers, body in [
    ("rate_header", 429, {"Retry-After": "1.25", "Poly-RateLimit-Remaining": "-1.5", "Poly-RateLimit-Reset": "42", "Poly-RateLimit-Tier": " standard ", "Poly-RateLimit-Warning": "TRUE"}, {}),
    ("rate_body", 429, {}, {"retry_after_seconds": 2.5}),
    ("rate_invalid_header_fallback", 429, {"Retry-After": "NaN"}, {"retry_after_seconds": 2}),
    ("rate_negative", 429, {"Retry-After": "-1"}, {"retry_after_seconds": -2}),
    ("rate_boolean", 429, {}, {"retry_after_seconds": True}),
    ("rate_string", 429, {}, {"retry_after_seconds": "2"}),
    ("rate_past_date", 429, {"Retry-After": "Sun, 06 Nov 1994 08:49:37 GMT"}, {}),
    ("rate_bad_headers", 429, {"Poly-RateLimit-Remaining": "NaN", "Poly-RateLimit-Reset": "Infinity", "Poly-RateLimit-Warning": "false"}, {}),
    ("restarting", 425, {}, {"error": "restart", "code": "restarting"}),
    ("post_only", 503, {}, {"error": "restricted", "code": "post_only_mode", "retry_after_seconds": 0}),
    ("cancel_only", 503, {}, {"error": "exchange is cancel-only"}),
    ("unavailable", 503, {}, {"error": "offline"}),
    ("not_found", 404, {}, {"error": "asset not found", "code": "missing"}),
    ("redirect", 302, {"Location": "https://example.invalid"}, {}),
    ("success_headers", 200, {"Poly-RateLimit-Remaining": "99"}, {"mid": "0.5"}),
]:
    headers = {"Content-Type": "application/json"} | headers
    response = httpx.Response(status, headers=headers, json=body, request=httpx.Request("GET", "https://example.test/midpoint"))
    expected = {"error": "", "code": "", "restriction": "", "retry_after": None}
    try:
        _transport._raise_for_response_status(response, path="/midpoint")
    except Exception as error:
        expected.update(error=type(error).__name__, code=getattr(error, "code", None) or "",
                        restriction=getattr(error, "restriction", None) or "", retry_after=getattr(error, "retry_after", None))
    update = _transport._parse_rate_limit_headers(response.headers)
    expected["rate_limit"] = None if update is None else dict(remaining=update.remaining, reset=update.reset, tier=update.tier or "", warning=update.warning)
    transport_cases.append(dict(name=name, status=status, headers=headers, body=body, expected=expected))
(ROOT / "testdata/http-parity.json").write_text(json.dumps(transport_cases, indent=2) + "\n")
print(f"Generated {len(transport_cases)} Python HTTP cases")
