package main

import (
	"database/sql"
	"fmt"
	"log"
	httpImport "net/http"
	"os"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/lib/pq"
	"github.com/philLITERALLY/wodland-service/internal/data"
	"github.com/philLITERALLY/wodland-service/internal/data/db"
	"github.com/philLITERALLY/wodland-service/internal/http"
)

var (
	idKey       = "id"
	usernameKey = "username"
	roleKey     = "role"
)

func helloHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	user, _ := c.Get(usernameKey)

	c.JSON(200, gin.H{
		"Claims": map[string]interface{}{
			"id":       claims[idKey],
			"username": claims[usernameKey],
			"role":     claims[roleKey],
		},
		"Context": map[string]interface{}{
			"id":       user.(*data.User).ID,
			"username": user.(*data.User).Username,
			"role":     user.(*data.User).Role,
		},
	})
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	// Set up heroku database connection
	dataSource, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}

	router := gin.New()
	router.Use(gin.Logger())

	// The jwt middleware
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "test zone",
		Key:         []byte("secret key"),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: usernameKey,
		PayloadFunc: func(dataInterface interface{}) jwt.MapClaims {
			if v, ok := dataInterface.(*data.User); ok {
				return jwt.MapClaims{
					idKey:       v.ID,
					usernameKey: v.Username,
					roleKey:     v.Role,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &data.User{
				ID:       int(claims[idKey].(float64)),
				Username: claims[usernameKey].(string),
				Role:     claims[roleKey].(string),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals data.Login
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}

			// Fetch login user
			user, err := db.GetUser(dataSource, loginVals)
			if err != nil {
				fmt.Printf("fetching user err: %v \n", err)
				return nil, jwt.ErrFailedAuthentication
			}

			return &user, nil
		},
		Authorizator: func(dataInterface interface{}, c *gin.Context) bool {
			if _, ok := dataInterface.(*data.User); ok {
				return true
			}

			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	})

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	errInit := authMiddleware.MiddlewareInit()
	if errInit != nil {
		log.Fatal("authMiddleware.MiddlewareInit() Error:" + errInit.Error())
	}

	router.POST("/login", authMiddleware.LoginHandler)
	router.POST("/logout", authMiddleware.LogoutHandler)
	router.GET("/refresh_token", authMiddleware.RefreshHandler)
	router.GET("/hello", authMiddleware.MiddlewareFunc(), helloHandler) // Test endpoint delete when done
	router.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Printf("NoRoute claims: %#v\n", claims)
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	// Endpoint to get single WOD and any attempts at it
	router.GET("/WOD/:wodID", authMiddleware.MiddlewareFunc(), http.GetWOD(dataSource))

	// Endpoint to get WODs (can be filtered)
	router.GET("/WODs", authMiddleware.MiddlewareFunc(), http.GetWODs(dataSource))

	// Endpoint to get WODs (can be filtered)
	router.GET("/Activities", authMiddleware.MiddlewareFunc(), http.GetActivities(dataSource))

	// Endpoint to create a WOD (and add an attempt if supplied)
	router.POST("/WOD", authMiddleware.MiddlewareFunc(), http.AddWOD(dataSource))

	// Endpoint to add an Activity
	router.POST("/Activity", authMiddleware.MiddlewareFunc(), http.AddActivity(dataSource))

	if err := httpImport.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
