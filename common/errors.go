package common

import "errors"

var (
	ERR_LOCK_ALREADY_REQUIRED = errors.New("锁已被占用")
	ERR_NO_LOCAL_IP_FOUND     = errors.New("获取网卡IP失败")
)
