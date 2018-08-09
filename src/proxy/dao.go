package proxy

import (
	"database/sql"
	"fmt"
	"sync"
	"time"
)

type mysqlDao struct {
	sync.Mutex
	name        string
	mysqlConfig *TMysqlConfig
	db          *sql.DB
}

func NewMysqlDao(name string, mysqlConfig TMysqlConfig) *mysqlDao {
	if len(mysqlConfig.Ip) == 0 || len(mysqlConfig.Database) == 0 {
		return nil
	}

	db := ConnectMysql(&mysqlConfig)
	if db == nil {
		return nil
	}

	return &mysqlDao{
		name:        name,
		mysqlConfig: &mysqlConfig,
		db:          db,
	}
}

func ConnectMysql(mysqlConfig *TMysqlConfig) *sql.DB {
	source := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		mysqlConfig.Username,
		mysqlConfig.Password,
		mysqlConfig.Ip,
		mysqlConfig.Port,
		mysqlConfig.Database,
	)

	Logger.Printf("start connect to mysql server %s", source)

	db, err := sql.Open("mysql", source)
	if err != nil {
		Logger.Print("connect mysql server failed")
		return nil
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(time.Duration(10) * time.Second)

	return db
}

func (obj *mysqlDao) Close() {
	if obj.db != nil {
		obj.db.Close()
	}

	Logger.Printf("mysql db[%s] closed", obj.name)
}

func (obj *mysqlDao) Reconnect() {
	obj.Lock()
	defer obj.Unlock()

	Logger.Printf("mysql db[%s] try reconnect", obj.name)

	obj.db = ConnectMysql(obj.mysqlConfig)
}
