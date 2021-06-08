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
		api.POST("/auth/signin", service.SignIn)   //login
		api.POST("/auth/signup", service.SignUp)   //register
		api.GET("/test/user", service.VerifyToken) //
	}

	opcua := r.Group("/opcua")
	{
		opcua.Any("/client", service.ReadOPC) //opcua websocket
		opcua.GET("/csv", service.ReadCSV)    //opcua plc nuc database connect
	}

	log.Fatal(r.Run(add))
}
