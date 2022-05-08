package main

import (
	"suds/uniswap_rest/controllers"
	"suds/uniswap_rest/models"

	"github.com/gin-gonic/gin"
	"github.com/hasura/go-graphql-client"
)

// func ErrorHandler(logger *zap.Logger) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		c.Next()

// 		for _, ginErr := range c.Errors {
// 			logger.Error(ginErr.Error())
// 		}
// 	}
// }

func main() {
	router := gin.Default()
	// logger, _ := zap.NewDevelopment()
	// router.Use(ErrorHandler(logger))

	// Initialize graphql client
	uniswapClient := &controllers.UniswapClient{graphql.NewClient(models.UNNISWAP_GRAPH_ENDPOINT, nil)}

	router.GET("/api/assets/:id", uniswapClient.GetAsset)
	router.GET("/api/assets/:id/pools", uniswapClient.GetAssetPools)
	router.GET("/api/assets/:id/volume", uniswapClient.GetAssetVolume)
	router.GET("/api/blocks/:blocknumber/swaps", uniswapClient.GetSwapsPerBlock)
	router.GET("/api/blocks/:blocknumber/assets", uniswapClient.GetAssetsSwappedPerBlock)

	router.Run()
}
