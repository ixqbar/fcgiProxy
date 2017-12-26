package proxy

import (
	"log"
	"github.com/jonnywang/go-kits/redis"
)

func init()  {
	redis.Logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

var Logger = redis.Logger