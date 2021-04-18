package server

func (s *HttpService) setRouter() {
	// 后台管理系统接口
	// 增加新用户，传入用户邮箱地址、用户名，返回userid、用户密码
	s.router.POST("/v1/api/user", s.AddUser)
	// 查询用户列表，支持offset、limit、邮箱、用户名、userId查询用户
	s.router.GET("/v1/api/users", s.QueryUserList)

	// 为用户添加APIKEY、SECRETKEY、交易所
	s.router.POST("/v1/api/apikey", s.AddApikey)
	// 查询用户所有APIKEY
	s.router.GET("/v1/api/apikeys", s.QueryUserApikeys)

	// 添加策略信息，策略名、策略描述、策略参数类型
	s.router.POST("/v1/api/strategy", s.AddStrategy)
	// 查询所有策略信息
	s.router.GET("/v1/api/strategies", s.QueryStrategies)
	// 修改策略描述、策略参数类型
	s.router.PUT("/v1/api/strategy", s.UpdateStrategy)
	// 删除策略
	// s.router.DELETE("/v1/api/strategy", s.DelStrategy)

	// 添加模板
	s.router.POST("/v1/api/strategy/model", s.AddStrategyModel)
	// 查询策略所有模板
	s.router.GET("/v1/api/strategy/models", s.QueryStrategyModels)
	// 修改模板
	s.router.PUT("/v1/api/strategy/model", s.UpdateStrategyModel)
	// 删除模板
	s.router.DELETE("/v1/api/strategy/model", s.DelStrategyModel)

	// 增加策略任务,apikey&strategy&instrumentid唯一标识一个策略任务
	s.router.POST("/v1/api/strategy/task", s.AddStrategyTask)
	// 查询策略任务,查询apikey的所有策略,策略任务是否在运行
	s.router.GET("/v1/api/strategy/tasks", s.QueryStrategyTasks)
	// 更新策略参数
	s.router.PUT("/v1/api/strategy/task", s.UpdateStrategyTask)
	// 删除策略
	s.router.DELETE("/v1/api/strategy/task", s.DelStrategyTask)

	// 开始策略任务,调度时选择负载最小的节点
	s.router.POST("/v1/api/strategy/task/start", s.StartStrategyTask)
	// 停止策略
	s.router.POST("/v1/api/strategy/task/stop", s.StopStrategyTask)

	// 小程序接口
	// 登录，用户通过邮箱地址和密码登录
	s.router.POST("/v1/api/login", s.UserLogin)
	// 修改密码，用户首次登录小程序需要修改密码
	s.router.POST("/v1/api/password/change", s.ChangePassword)
	// 查询用户资产信息，总资产、初始余额、当前余额、真实余额、当前收益
	s.router.GET("/v1/api/user/property", s.QueryProperty)
	// 查询用户收益数据，按天计算，返回最近一个月的数据，今日收益、今日最大回撤
	s.router.GET("/v1/api/user/profit", s.QueryProfit)
	// 查询用户策略信息，策略类型、交易品种、当前收益、当前持仓、运行简要、当前挂单、运行状态
	s.router.GET("/v1/api/user/strategy", s.QueryUserStrategyData)
	// 查询用户交易记录
	s.router.GET("/v1/api/user/trade", s.QueryUserTrade)
}

/*
1.增加策略任务,apikey&strategy&instrumentid唯一标识一个策略任务,要加分布式锁
2.查询策略任务,查询策略是否在运行,查询apikey的所有策略
3.开始策略任务,调度时选择负载最小的节点,要加分布式锁
4.更新策略参数,要加分布式锁
5.停止策略,要加分布式锁
6.删除策略,要加分布式锁
7.节点挂掉重新调度此节点上的策略,要加分布式锁

策略manager启动时向etcd注册,/service/strategymanager/10.22.33.55:32154 --> {"uptime":111111111, "available":true}

调度服务增加一个策略任务,/strategy/task/$apikey/$strategy/$instrumentid --> {uptime":111111111, "available":true, "param"："xxxx"}

调度服务watch /service/strategymanager,有节点下线就重新调度节点上的任务,再删除/scheduler/node/10.22.33.55:32154 --prefix
                                                                          /scheduler/strategy/$apikey/$strategy/$instrumentid

调度服务调度时记录节点上运行的任务,/scheduler/node/10.22.33.55:32154/$apikey/$strategy/$instrumentid --> {"uptime":111111111}
                                /scheduler/strategy/$apikey/$strategy/$instrumentid --> {"addr":"10.22.33.55:32154", "uptime":111111111, "available":true}

定时任务遍历/strategy/task,检查addr是否在线,不在线则重新调度
*/
