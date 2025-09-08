package routes

import (
	controllers "simple-forum/app/controller"
	"simple-forum/app/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) *gin.Engine {

	// 公共路由
	public := r.Group("/api")
	{
		public.POST("/user/reg", controllers.Register)
		public.POST("/user/login", controllers.Login)
	}

	// 受保护的路由
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("/student/post", controllers.CreatePost)             //发布新帖子
		protected.PUT("/student/post", controllers.UpdatePost)              //修改自己的帖子
		protected.DELETE("/student/post", controllers.DeletePost)           //删除自己的帖子
		protected.POST("/student/post/restore", controllers.RestorePost)    //恢复自己的帖子
		protected.GET("/student/post", controllers.GetPosts)                //应该是获取所有帖子
		protected.POST("/student/report-post", controllers.GetPosts)        //举报帖子
		protected.GET("/student/report-post", controllers.GetReportResults) //学生查看审核结果
		protected.GET("admin/report", controllers.ViewReportApproval)       //获取所有未审批的举报帖子列表
		protected.POST("admin/report", controllers.ApproveReport)           //审核被举报的帖子
		protected.GET("student/likes", controllers.GetPostLikes)            //获取点赞数(进阶需求)
		protected.POST("/student/likes", controllers.Like)                  //点赞(进阶需求)
	}

	return r
}
