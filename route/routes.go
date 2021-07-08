package route

import (
	"app/service"
	"log"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func RunAPI(add string) {
	r := gin.Default()
	r.Use(static.Serve("/", static.LocalFile("./client/build", true)))

	api := r.Group("/api")
	{
		api.POST("/auth/signin", service.SignIn) //login
		api.POST("/auth/signup", service.SignUp) //register
		api.POST("/auth/signout", service.SignOut)
		api.POST("/auth/user", service.User) //
		api.POST("/token/refresh", service.Refresh)
	}

	opcua := r.Group("/opcua")
	{
		opcua.Any("/client", service.ReadOPC) //opcua websocket
	}

	log.Fatal(r.Run(add))
}
