package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/config"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/database"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/routes"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/golang-jwt/jwt/v5"
)

// @title RootAccess CTF API
// @version 1.0.0
// @description This is the backend API for the RootAccess CTF Platform.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/Uttam-Mahata/RootAccess/issues
// @contact.email contact@rootaccess.live

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host rootaccess.live
// @BasePath /
// @query.collection.format multi

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

var (
	ginLambda *ginadapter.GinLambda
	appConfig *config.Config
)

func main() {
	// Local development or containerized deployment
	appConfig = config.LoadConfig()
	cfg := appConfig

	// Connect to Database
	database.ConnectDB(cfg.MongoURI, cfg.DBName)

	// Connect to 6 Upstash Redis instances
	database.ConnectRedisRegistry(map[string]string{
		"auth":       cfg.RedisURLAuth,
		"scoreboard": cfg.RedisURLScoreboard,
		"challenge":  cfg.RedisURLChallenge,
		"websocket":  cfg.RedisURLWebSocket,
		"analytics":  cfg.RedisURLAnalytics,
		"general":    cfg.RedisURLGeneral,
	})

	r := routes.SetupRouter(cfg)

	// Check if running in AWS Lambda
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		log.Println("Running in AWS Lambda mode")
		ginLambda = ginadapter.New(r)
		lambda.Start(UnifiedHandler)
		return
	}

	log.Printf("Server running on port %s", cfg.Port)
	r.Run(":" + cfg.Port)
}

// UnifiedHandler handles both REST and WebSocket API events
func UnifiedHandler(ctx context.Context, event json.RawMessage) (interface{}, error) {
	cfg := appConfig

	// 1. Try to unmarshal as standard APIGatewayProxyRequest (REST)
	var restReq events.APIGatewayProxyRequest
	if err := json.Unmarshal(event, &restReq); err == nil && restReq.HTTPMethod != "" {
		return ginLambda.ProxyWithContext(ctx, restReq)
	}

	// 2. Try to unmarshal as APIGatewayWebsocketProxyRequest (WebSocket)
	var wsReq events.APIGatewayWebsocketProxyRequest
	if err := json.Unmarshal(event, &wsReq); err == nil && wsReq.RequestContext.ConnectionID != "" {
		path := "/ws/default"
		switch wsReq.RequestContext.RouteKey {
		case "$connect":
			path = "/ws/connect"
		case "$disconnect":
			path = "/ws/disconnect"
		}

		headers := map[string]string{
			"X-Forwarded-For": wsReq.RequestContext.Identity.SourceIP,
			"X-Connection-Id": wsReq.RequestContext.ConnectionID,
		}

		// Try to extract user_id from token in query params for $connect
		if wsReq.RequestContext.RouteKey == "$connect" {
			if tokenStr, ok := wsReq.QueryStringParameters["token"]; ok {
				// Parse token to get user info
				token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
					return []byte(cfg.JWTSecret), nil
				})
				if err == nil && token.Valid {
					if claims, ok := token.Claims.(jwt.MapClaims); ok {
						if uid, ok := claims["user_id"].(string); ok {
							headers["X-User-Id"] = uid
						}
					}
				}
			}
		}

		// Convert WebSocket event to a synthetic REST request for Gin
		syntheticReq := events.APIGatewayProxyRequest{
			Path:           path,
			HTTPMethod:     "POST",
			Headers:        headers,
			RequestContext: events.APIGatewayProxyRequestContext{
				// Minimum context needed for gin-adapter
			},
			Body: wsReq.Body,
		}

		return ginLambda.ProxyWithContext(ctx, syntheticReq)
	}

	return nil, fmt.Errorf("unsupported event type")
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}
