package proxy

import (
	"sync"
	"database/sql"
	"fmt"
)

type mysqlDao struct {
	sync.Mutex
	name string
	db *sql.DB
}

func NewMysqlDao(name string, mysqlConfig MysqlConfig) *mysqlDao {
	if len(mysqlConfig.Ip) == 0 || len(mysqlConfig.Database) == 0 {
		return nil
	}

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

	return &mysqlDao{
		name:name,
		db:db,
	}
}

func (obj *mysqlDao) Close() {
	if obj.db != nil {
		obj.db.Close()
	}

	Logger.Printf("mysql db[%s] closed", obj.name)
}
