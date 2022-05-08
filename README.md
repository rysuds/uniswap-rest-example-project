# Uniswap REST API for the Graph

## Description

The API currently works by making GraphQL calls to the archived subgraph for UniswapV3. Input params from the REST endpoints are parsed and injected into GraphQL Queries. The responnses from the server are parsed and returned to the user as JSON unmarshalled structs.

Because I wanted to get this out quickly and get a workig solution, there is alot that could be improved (see next section). All in all, great learning experience, and at face value things seem to workas intended (I think)!

### What I could improve on

- Unit testing
  - Adding mock outputs for each of the controller functions via sample "happy path" json files
  - Validate parsinng logic works as expected
- Input validation for endpoints
  - Disallow incorrect types for both path and query parameters
    - e.g. `:id` should always be a hex string
- Default pagination
  - In the current state, we grab all records that match the condition of the query, this could be costly at scale so we should add a default paginnation limit
- Better utilization of efficient Go patterns
  - I'm pretty new with Go, so I could be better about using some of the patterns for efficiency
    - Pointers
    - Goroutines

### How to run

```
go run main.go
```

By default the server will run on PORT 8080

## Endpoints

---

### `GET /api/assets/:id`

Get an `Asset` object via it's `id`

**Parameters**

- `:id` [Path]

**Example**

```curl
curl localhost:8080/api/assets/0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48
```

Status Code: `200`

```json
{
  "asset": {
    "id": "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
    "symbol": "USDC",
    "volume_usd": 323047705878.95746
  }
}
```

### `GET /api/assets/:id/pools`

For a given `Asset`, retrieve all of the pools it is currently swapping in. The asset could be `token0` or `token1`

**Parameters**

- `:id` [Path]

**Example**

```curl
curl localhost:8080/api/assets/0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48/pools
```

Status Code: `200`

```json
{
  "pools": [
    {
      "id": "0x020c349a0541d76c16f501abc6b2e9c98adae892",
      "asset0_symbol": "USDC",
      "asset1_symbol": "SNX"
    },
    {
      "id": "0x07a6e955ba4345bae83ac2a6faa771fddd8a2011",
      "asset0_symbol": "MATIC",
      "asset1_symbol": "USDC"
    },
    ...
    {
      "id": "0xfad57d2039c21811c8f2b5d5b65308aa99d31559",
      "asset0_symbol": "LINK",
      "asset1_symbol": "USDC"
    }
  ]
}
```

### `GET /api/assets/:id/volume`

For a given `Asset`, retrieve the total volume swapped. User can also specify a `startTime` and/or `endTime` epoch integer query params to get the volume over a time range

_Note_: This endpoint is definitely buggy, it is currently only returning the volume of a single day aggregate. I tried debugging for quite some time but got stuck :/

**Parameters**

- `:id` (`string`) _Path_
- `startTime` (`int`) _Query_
- `endTime` (`int`) _Query_

**Example**

```curl
curl -s -X GET 'localhost:8080/api/assets/0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48/volume?startTime=1620172800&endTime=1620259200'
```

Status Code: `200`

```json
{
  "volume": {
    "TokenId": "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
    "start_time": "1620172800",
    "end_time": "1621359201",
    "total_volume_USD": 6711234.708437207
  }
}
```

### `GET /api/blocks/:blocknumber/swaps`

For a given `Block` return all of the swaps that took place during that time. The approach I used here was to get all of the trannsactions per block, then grab all of the swaps per transaction and bubble them up

**Parameters**

- `:blocknumber` (`int`) _Path_

**Example**

```curl
curl localhost:8080/api/blo
cks/14732439/swaps
```

Status Code: `200`

```json
{
  "count": 2,
  "swaps": [
    {
      "id": "0xdaba7b4d0e601022032337a20c737bc45b6e356e0f523832bb6dca024fe83d4e#28034",
      "Amount0": -25154.830567,
      "Amount1": 25259.687517518712,
      "asset0": {
        "id": "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
        "symbol": "USDC"
      },
      "asset1": {
        "id": "0xa47c8bf37f92abed4a126bda807a7b7498661acd",
        "symbol": "UST"
      }
    },
    {
      "id": "0xf65426771632624c09308d1b3855021e3501b478e171f861f105d3d6cd2e1fea#83429",
      "Amount0": 2.5250575024654442,
      "Amount1": -367.8205895069069,
      "asset0": {
        "id": "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
        "symbol": "WETH"
      },
      "asset1": {
        "id": "0xc18360217d8f7ab5e7c516566761ea12ce7f9d72",
        "symbol": "ENS"
      }
    }
  ]
}
```

### `GET /api/blocks/:blocknumber/assets`

For a given `Block` return all of the assets swapped durinng that time. I used the same approach as the previous endpoint, but just recorded the asset pairs as well (keeping mind of duplicates)

**Parameters**

- `:blocknumber` (`int`) _Path_

**Example**

```curl
curl localhost:8080/api/blocks/14732439/assets
```

Status Code: `200`

```json
{
  "assets": [
    { "id": "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", "symbol": "USDC" },
    { "id": "0xa47c8bf37f92abed4a126bda807a7b7498661acd", "symbol": "UST" },
    { "id": "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", "symbol": "WETH" },
    { "id": "0xc18360217d8f7ab5e7c516566761ea12ce7f9d72", "symbol": "ENS" }
  ],
  "count": 4
}
```
