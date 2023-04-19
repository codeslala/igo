package main

import (
	"errors"
	"fmt"

	"github.com/codeslala/igo/zlog"
)

func main() {
	//将数据写入日志文件
	zlog.Datalog("this is a test log", true)
	//记录日志，sync 为是否实时写入文件，true- 实时写入文件，false-写入内存
	zlog.Info("this is a test info log", false)
	zlog.Error(fmt.Errorf("this is a test error log, err=%v", errors.New("error log")), true)
	zlog.Warning("this is a test warning log", false)
	zlog.Trace("this is a test trace log", false)
}
