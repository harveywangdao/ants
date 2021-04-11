package server

func (s *HttpService) setRouter() {
	// 后台管理系统接口
	// 1.增加新用户，传入用户邮箱地址、用户名，返回userid、用户密码
	s.router.POST("/v1/api/user", s.AddUser)
	// 2.查询用户列表，支持offset、limit、邮箱、用户名、userId查询用户
	s.router.GET("/v1/api/users", s.QueryUserList)

	// 3.为用户添加APIKEY、SECRETKEY、交易费率、交易所、策略(一个用户可以有多个apikey) --> 增加之后就启动策略
	s.router.POST("/v1/api/apikey", s.AddApikey)
	// 4.查询用户所有APIKEY
	s.router.GET("/v1/api/apikeys", s.QueryUserApikeys)
	// 5.修改用户的APIKEY使用的策略
	s.router.PUT("/v1/api/apikey", s.ChangeStrategy)

	// 6.添加策略信息，策略类型、币种、策略参数
	s.router.POST("/v1/api/strategy", s.AddStrategy)
	// 7.查询所有策略信息
	s.router.GET("/v1/api/strategies", s.QueryStrategies)
	// 8.修改策略参数
	s.router.PUT("/v1/api/strategy", s.UpdateStrategy)
	// 9.删除策略
	s.router.DELETE("/v1/api/strategy", s.DelStrategy)

	// 10.启动单个用户的单个APIKEY的策略
	s.router.POST("/v1/api/apikey/strategy/start", s.StartApikeyStrategy)
	// 11.停止单个用户的单个APIKEY的策略
	s.router.POST("/v1/api/apikey/strategy/stop", s.StopApikeyStrategy)

	// 12.暂停某一策略的所有APIKEY
	s.router.POST("/v1/api/strategy/pause", s.PauseStrategy)
	// 13.恢复运行某一策略的所有APIKEY
	s.router.POST("/v1/api/strategy/resume", s.ResumeStrategy)
	// 14.将A策略的所有APIKEY迁移到B策略
	s.router.POST("/v1/api/strategy/migrate", s.MigrateStrategy)

	// 小程序接口
	// 1.登录，用户通过邮箱地址和密码登录
	s.router.POST("/v1/api/login", s.UserLogin)
	// 2.修改密码，用户首次登录小程序需要修改密码
	s.router.POST("/v1/api/password/change", s.ChangePassword)
	// 3.查询用户资产信息，总资产、初始余额、当前余额、真实余额、当前收益
	s.router.GET("/v1/api/user/property", s.QueryProperty)
	// 4.查询用户收益数据，按天计算，返回最近一个月的数据，今日收益、今日最大回撤
	s.router.GET("/v1/api/user/profit", s.QueryProfit)
	// 5.查询用户策略信息，策略类型、交易品种、当前收益、当前持仓、运行简要、当前挂单、运行状态
	s.router.GET("/v1/api/user/strategy", s.QueryUserStrategyData)
	// 6.查询用户交易记录
	s.router.GET("/v1/api/user/trade", s.QueryUserTrade)
}

/*
1.停止单个用户的单个APIKEY的策略
  1.1更新数据库该APIKEY为stop,等待同步

2.启动单个用户的单个APIKEY的策略
  2.1更新数据库该APIKEY为run,等待同步

3.暂停某一策略的所有APIKEY
  3.1更新数据库该策略从run到pause,等待同步

4.恢复运行某一策略的所有APIKEY
  4.1更新数据库该策略从pause到run,等待同步

5.将A策略的所有APIKEY迁移到B策略
  5.1更新数据库为B策略,等待同步

6.节点挂掉,怎么重新启动挂掉节点上的策略,不区分节点是正常退出还是异常退出

调度服务收到请求,APIKEY需要绑定一个策略服务,调度服务向etcd存储APIKEY信息,/scheduler/apikey/$apikey --> {"strategyName":"grid","param":0.1}

调度服务watch /scheduler/apikey,有新的apikey加入就去分配策略服务的节点/scheduler/strategy/strategyName/10.22.33.55:32154 --> {"apikeys":["apikey1","apikey2"]}
                               记录映射关系到 /scheduler/apikey_strategy/$apikey --> {"node":"10.22.33.55:32154"}

                               删除apikey,先删除/scheduler/apikey_strategy/$apikey
                               再更新/scheduler/strategy/strategyName/10.22.33.55:32154 --> {"apikeys":["apikey1"]}

                               更新策略,新删除再增加

调度服务watch /service/strategy,有新的节点加入就增加/scheduler/strategy/strategyName/10.22.33.55:32154 --> {}
                               有节点退出就删除/scheduler/strategy/strategyName/10.22.33.55:32154,并重新调度apikey1,apikey2

策略服务启动时向etcd注册地址,/service/strategy/strategyName/10.22.33.55:32154 --> {"uptime":111111111, "available":true, "strategyName":"grid"}

动态检测数据库数据与etcd的一致性,允许存在延迟
*/
