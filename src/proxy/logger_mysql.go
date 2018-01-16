package proxy

import (
	"sync"
	_ "github.com/go-sql-driver/mysql"
)

type messageDao struct {
	mysqlDao *mysqlDao
}

var messageMysqlDao *messageDao
var messageOnce sync.Once

func LoggerMessageDao() *messageDao {
	messageOnce.Do(func() {
		messageMysqlDao = &messageDao{
			mysqlDao:NewMysqlDao("logger", Config.LoggerMysqlConfig),
		}
	})

	return messageMysqlDao
}

func (obj *messageDao) RecordMessage(pubSubMessage *PubSubMessage) (bool) {
	if obj.mysqlDao == nil {
		Logger.Print("mysql logger not open")
		return false
	}

	logMessage, ok := pubSubMessage.Message.(LogMessage)
	if !ok {
		return false
	}

	stmtIns, err := obj.mysqlDao.db.Prepare("INSERT INTO access_logs(user_id,user_ip,user_agent,resource,type,content,time) VALUES (?,?,?,?,?,?,?)")
	if err != nil {
		Logger.Print(err)
		return false
	}
	defer stmtIns.Close()

	if _, err = stmtIns.Exec(
		logMessage.UserID,
		pubSubMessage.IP,
		pubSubMessage.UserAgent,
		logMessage.Resource,
		logMessage.Type,
		logMessage.Content,
		pubSubMessage.Time); err != nil {
		Logger.Print(err)
		return false
	}

	return true
}

func (obj *messageDao) Close() {
	if obj.mysqlDao == nil {
		return
	}

	Logger.Printf("mysql dao[%s] will close", obj.mysqlDao.name)

	obj.mysqlDao.Close()
}