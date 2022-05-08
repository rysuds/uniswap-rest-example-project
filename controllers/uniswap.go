package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"suds/uniswap_rest/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hasura/go-graphql-client"
)

type UniswapClient struct {
	*graphql.Client
}

/*
Get a single Asset based on block number

Graphql:
{
  token(id: "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"){
    symbol
    volumeUSD
  }
}

*/
func (u *UniswapClient) GetAsset(c *gin.Context) {
	assetId := c.Param("id")
	var query struct {
		Token struct {
			Id        graphql.String
			Symbol    graphql.String
			VolumeUSD graphql.String `graphql:"volumeUSD"`
		} `graphql:"token(id: $id)"`
	}
	variables := map[string]interface{}{
		"id": graphql.String(assetId),
	}

	err := u.Client.Query(context.Background(), &query, variables)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	parsedVolume, err := strconv.ParseFloat(string(query.Token.VolumeUSD), 64)
	if err != nil {
		c.Error(err)
	}
	asset := models.Asset{
		ID:        string(query.Token.Id),
		Symbol:    string(query.Token.Symbol),
		VolumeUSD: parsedVolume,
	}
	c.JSON(http.StatusOK, gin.H{"asset": asset})
}

/*
Get at Pools an Asset sits in

Graphql:
{
  token(id:"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"){
 	id
    whitelistPools {
      id
      token0 {
        id
      }
      token1 {
        id
      }
    }
 }
}
*/
func (u UniswapClient) GetAssetPools(c *gin.Context) {
	// Given an AssetID, return all Pools
	assetId := c.Param("id")
	var query struct {
		Token struct {
			Id             graphql.String
			WhitelistPools []struct {
				Id     graphql.String
				Token0 struct {
					Symbol graphql.String
				}
				Token1 struct {
					Symbol graphql.String
				}
			}
		} `graphql:"token(id: $id)"`
	}
	variables := map[string]interface{}{
		"id": graphql.String(assetId),
	}

	err := u.Client.Query(context.Background(), &query, variables)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if err != nil {
		c.Error(err)
	}
	fmt.Println(query)
	var pools []models.Pool
	for _, pool := range query.Token.WhitelistPools {
		pools = append(pools, models.Pool{
			ID:           string(pool.Id),
			Asset0Symbol: string(pool.Token0.Symbol),
			Asset1Symbol: string(pool.Token1.Symbol),
		})
	}
	c.JSON(http.StatusOK, gin.H{"pools": pools})
}

/*
Get total volume of Asset given a time period. If no period specified, default to a week from current time

Graphql:
{
  tokenDayDatas(where: {date_gte: 1620172800, date_lte: 1630172800, token: "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"}) {
		volumeUSD
    date
  }
}
*/
func (u UniswapClient) GetAssetVolume(c *gin.Context) {
	assetId := c.Param("id")

	var query struct {
		TokenDayDatas []struct {
			VolumeUSD graphql.String `graphql:"volumeUSD"`
			Date      graphql.String
		} `graphql:"tokenDayDatas(where: {date_gte: $date_gte, date_lte: $date_lte, token: $id})"`
	}
	endTimeParam := c.DefaultQuery("endTime", fmt.Sprintf("%d", time.Now().Unix()))
	endTime, err := strconv.ParseInt(endTimeParam, 10, 64)
	fmt.Println(endTime)
	if err != nil {
		c.Error(err)
	}

	startTimeParam := c.DefaultQuery("startTime", fmt.Sprintf("%d", time.Now().Unix()-models.WEEK_IN_SECONDS))
	startTime, err := strconv.ParseInt(startTimeParam, 10, 64)
	fmt.Println(startTime)
	if err != nil {
		c.Error(err)
	}

	variables := map[string]interface{}{
		"id":       graphql.String(assetId),
		"date_gte": graphql.Int(startTime),
		"date_lte": graphql.Int(endTime),
	}

	qerr := u.Client.Query(context.Background(), &query, variables)
	if qerr != nil {
		fmt.Println(qerr)
	}

	// TODO: There is a bug here, it is only grabbing one day when it should be grabbing multiple
	var sumVolume float64 = 0
	for _, dayData := range query.TokenDayDatas {
		parsedVolume, err := strconv.ParseFloat(string(dayData.VolumeUSD), 64)
		if err != nil {
			fmt.Println(qerr)
		}
		fmt.Println(parsedVolume)
		sumVolume += parsedVolume
	}

	result := models.VolumePerTimeWindow{
		TokenId:        assetId,
		StartTime:      startTime,
		EndTime:        endTime,
		TotalVolumeUSD: sumVolume,
	}
	c.JSON(http.StatusOK, gin.H{"volume": result})
}

/*
For a given block, return aggregated result of Swaps and all Assets in every swap.
Aggregation is done so via grabbing all Transactions per block and iterating through

Graphql:
{
  transactions(where: {blockNumber: 14732439}){
    id
    blockNumber
    swaps {
			id
      token0 {
				id
        symbol
      }
      amount0
      token1 {
				id
        symbol
      }
      amount1
    }
  }
}

*/
func (u UniswapClient) GetSwapResult(c *gin.Context) models.SwapResult {
	blockNumber := c.Param("blocknumber")
	var query struct {
		Transactions []struct {
			Id          graphql.String
			BlockNumber graphql.String
			Swaps       []struct {
				Id      graphql.String
				Amount0 graphql.String
				Amount1 graphql.String
				Token0  struct {
					Id     graphql.String
					Symbol graphql.String
				}
				Token1 struct {
					Id     graphql.String
					Symbol graphql.String
				}
			}
		} `graphql:"transactions(where: {blockNumber: $blocknumber})"`
	}

	variables := map[string]interface{}{
		"blocknumber": graphql.String(blockNumber),
	}
	err := u.Client.Query(context.Background(), &query, variables)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return models.SwapResult{}
	}

	// Iterate through transactions and record Awaps and Assets
	var swaps []models.Swap
	var assets []models.Asset
	assetSet := make(map[string]struct{})
	for _, txn := range query.Transactions {
		for _, swap := range txn.Swaps {
			amount0, err := strconv.ParseFloat(string(swap.Amount0), 64)
			if err != nil {
				c.Error(err)
			}
			amount1, err := strconv.ParseFloat(string(swap.Amount1), 64)
			if err != nil {
				c.Error(err)
			}
			asset0 := models.Asset{
				ID:     string(swap.Token0.Id),
				Symbol: string(swap.Token0.Symbol),
			}
			asset1 := models.Asset{
				ID:     string(swap.Token1.Id),
				Symbol: string(swap.Token1.Symbol),
			}
			if _, ok := assetSet[asset0.Symbol]; !ok {
				assetSet[asset0.Symbol] = struct{}{}
				assets = append(assets, asset0)
			}
			if _, ok := assetSet[asset1.Symbol]; !ok {
				assetSet[asset1.Symbol] = struct{}{}
				assets = append(assets, asset1)
			}
			swaps = append(swaps, models.Swap{
				ID:      string(swap.Id),
				Amount0: float64(amount0),
				Amount1: float64(amount1),
				Asset0: models.Asset{
					ID:     string(swap.Token0.Id),
					Symbol: string(swap.Token0.Symbol),
				},
				Asset1: models.Asset{
					ID:     string(swap.Token1.Id),
					Symbol: string(swap.Token1.Symbol),
				},
			})
		}
	}
	blockNumberInt, err := strconv.ParseInt(blockNumber, 10, 64)
	if err != nil {
		c.Error(err)
	}
	swapResult := models.SwapResult{
		BlockNumber: blockNumberInt,
		Swaps:       swaps,
		Assets:      assets,
	}
	return swapResult
}

// Return array of Swaps given a block number
func (u UniswapClient) GetSwapsPerBlock(c *gin.Context) {
	swapResult := u.GetSwapResult(c)
	swaps := swapResult.Swaps
	c.JSON(http.StatusOK, gin.H{"swaps": swaps, "count": len(swaps)})
}

// Return array of assets swapped given a block number
func (u UniswapClient) GetAssetsSwappedPerBlock(c *gin.Context) {
	swapResult := u.GetSwapResult(c)
	assets := swapResult.Assets
	c.JSON(http.StatusOK, gin.H{"assets": assets, "count": len(assets)})
}
