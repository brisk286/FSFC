package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

// 定义一个全局对象db
var db *sql.DB

// 定义一个初始化数据库的函数
func init() {
	// DSN:Data Source Name
	dsn := "root:root@tcp(127.0.0.1:3306)/fsfc?charset=utf8mb4&parseTime=True"
	// 不会校验账号密码是否正确
	// 注意！！！这里不要使用:=，我们是给全局变量赋值，然后在main函数中使用全局变量db
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("cannot open db:\n%d\n", err)
	}
	// 尝试与数据库建立连接（校验dsn是否正确）
	err = db.Ping()
	if err != nil {
		fmt.Printf("cannot ping db, dsn maybe wrong: \n%d\n", err)
	}
}

func GetDB() *sql.DB {
	return db
}

//func main() {
//	err := initDB() // 调用输出化数据库的函数
//	if err != nil {
//		fmt.Printf("init db failed,err:%v\n", err)
//		return
//	}
//}
