package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"main/Controller"
	"main/MiddleWare"
	"net/http"
	"runtime"
)

const f = `
          ⣠⠤⠖⠚⢉⣩⣭⡭⠛⠓⠲⠦⣄⡀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⢀⡴⠋⠁⠀⠀⠊⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠳⢦⡀⠀
⠀⠀⠀⠀⢀⡴⠃⢀⡴⢳⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⣆
⠀⠀⠀⠀⡾⠁⣠⠋⠀⠈⢧⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⢧
⠀   ⣸⠁⢰⠃⠀⠀⠀⠈⢣⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⣇
⠀⠀⠀⡇⠀⡾⡀⠀⠀⠀⠀⣀⣹⣆⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢹
⠀⠀⢸⠃⢀⣇⡈⠀⠀⠀⠀⠀⠀⢀⡑⢄⡀⢀⡀⠀⠀⠀⠀⠀⠀⢸⡇
⠀⠀⢸⠀⢻⡟⡻⢶⡆⠀⠀⠀⠀⡼⠟⡳⢿⣦⡑⢄⠀⠀⠀⠀⠀⢸⡇
⠀⠀⣸⠀⢸⠃⡇⢀⠇⠀⠀⠀⠀⡼⠀ ⠀⠈⣿⡗⠂⠀⠀⠀⠀⢸⠁
⠀⠀⡏⠀⣼⠀⢳⠊⠀⠀⠀⠀⠀⠀⠱⣀⣀⠔⣸⠁⠀⠀⠀⠀⢠⡟⠀
⠀⠀⡇⢀⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠠⠀⡇⠀⠀⠀⠀⠀⢸⠃⠀
⠀⢸⠃⠘⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⠁⠀⠀⢀⠀⠀⣾⠀⠀
⠀⣸⠀⠀⠹⡄⠀⠀⠈⠁⠀⠀⠀⠀⠀⠀⠀⡞⠀⠀⠀⠸⠀⠀⡇⠀⠀
⠀⡏⠀⠀⠀⠙⣆⠀⠀⠀⠀⠀⠀⠀⢀⣠⢶⡇⠀⠀⢰⡀⠀⠀⡇⠀⠀
⢰⠇⡄⠀⠀⠀⡿⢣⣀⣀⣀⡤⠴⡞⠉⠀⢸⠀⠀⠀⣿⡇⠀⠀⣧⠀⠀
⣸⠀⡇⠀⠀⠀⠀⠀⠀⠉⠀⠀⠀⢹⠀⠀⢸⠀⠀⢀⣿⠀⠁⠀⢸⠀⠀
⣿⠀⡇⠀⠀⠀⠀⠀⢀⡤⠤⠶⠶⠾⠤⠄⢸⠀⡀⠸⣿⣀⠀⠀⠈⣇⠀
⡇⠀⡇⠀⠀⡀⠀⡴⠋⠀⠀⠀⠀⠀⠀⠀⠸⡌⣵⡀⢳⡇⠀⠀⠀⢹⡀
⡇⠀⠇⠀⠀⡇⡸⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⠮⢧⣀⣻⢂⠀⠀⠀⢧
⣇⠀⢠⠀⠀⢳⠇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀ ⠀ ⠈⡎⣆⠀⠀⠘
⢻⠀⠈⠰⠀⢸⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠰⠘⢮⣧⡀⠀
⠸⡆⠀⠀⠇⣾⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⠆⠀⠀⠀⠀⠀⠀⠀⠙⠳⣄
`

func main() {
	fmt.Println(f)
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:8080", "http://localhost:8081"}
	config.AllowCredentials = true
	router.Use(cors.New(config))
	router.GET("/getNum", func(c *gin.Context) {
		log.Println(runtime.NumGoroutine())
		c.JSON(http.StatusOK, nil)
	})
	homePage := router.Group("/homePage")
	{
		homePage.GET("/index", Controller.GetIndex)
	}
	account := router.Group("/account")
	{
		account.POST("/AuthCode", Controller.GetAuth)
		account.POST("/Login", Controller.Login)
		account.POST("/Register", Controller.Register)
		account.POST("/LogOut", MiddleWare.Auth(), Controller.LogOut)
		account.GET("/Info", MiddleWare.Auth(), Controller.Info)
		account.POST("/EditInfo", MiddleWare.Auth(), Controller.EditInfo)
		account.POST("/Forget", Controller.ForgetPassword)
		account.POST("/ChangePd", MiddleWare.Auth(), Controller.ChangePassword)
	}
	staff := router.Group("/staff", MiddleWare.Auth())
	{
		staff.GET("/getStaff", Controller.GetStaff)
		staff.POST("/addStaff", Controller.AddStaff)
		staff.GET("/fireStaff", Controller.DeleteStaff)
		staff.GET("/getStaffInfo", Controller.GetStaffInfo)
		staff.POST("/EditInfo", Controller.ChangeStaffInfo)
	}
	company := router.Group("/company", MiddleWare.Auth())
	{
		company.GET("/getJointVenture", Controller.GetJointVenture)
		company.POST("/makeFriends", Controller.MakeFriend)
		company.POST("/deleteFriends", Controller.DeleteFriend)
		company.GET("/getFriendsInfo", Controller.GetFriendsInfo)
		company.POST("/sendMessage", Controller.SendMessageToFriends)
	}
	order := router.Group("/order", MiddleWare.Auth())
	{
		order.GET("/getAllOrder", Controller.GetAllOrder)
		order.POST("/submitOrder", Controller.BindOrder)
		order.POST("/askForPrice", Controller.AskForPrice)
		order.GET("/getAllBargain", Controller.GetAllBargain)
		order.POST("/replyBargain", Controller.ReplyBargain)
		order.GET("/getOrderInfo", Controller.GetOrderInfo)
		order.POST("/chooseAgent", Controller.SubmitCompanyChoose)
	}
	message := router.Group("/message", MiddleWare.Auth())
	{
		message.GET("/getMessage", Controller.GetMessage)
		message.GET("/getMessageInfo", Controller.GetMessageInfo)
		message.GET("/deleteMessage", Controller.DeleteMessage)
		message.POST("/company/reply", Controller.ReplyFriend)
	}
	router.GET("/ws", MiddleWare.Auth(), Controller.BuildSocket)
	//router.Run(":8080")
	router.RunTLS(":8081", "./key/cunyuqing.online_bundle.pem", "./key/cunyuqing.online.key")
}
