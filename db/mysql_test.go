package db

import (
	"fmt"
	"fsfc/models"
	"testing"
	"time"
)

func Test_Init(t *testing.T) {
	db = GetDB()
	fmt.Println(db)
}

func Test_Insert(t *testing.T) {
	db = GetDB()

	sqlStr := "insert into rsyncFile(rsyncFile_Filename, rsyncFile_RsyncTime) values(?,?)"

	result, err := db.Exec(sqlStr, "afe.png", time.Now())
	if err != nil {
		fmt.Printf("insert failed, err:%v\n", err)
		return
	}
	theID, err := result.LastInsertId() // 新插入数据的id
	if err != nil {
		fmt.Printf("get lastinsert ID failed, err:%v\n", err)
		return
	}
	fmt.Printf("insert success, the id is %v\n", theID)
}

func Test_Query(t *testing.T) {
	sqlStr := "select rsyncFile_Id, rsyncFile_Filename, rsyncFile_RsyncTime from rsyncFile where rsyncFile_Id=?"
	var r models.RsyncFile
	err := db.QueryRow(sqlStr, 1).Scan(&r.RsyncFileId, &r.RsyncFileFilename, &r.RsyncFileRsyncTime)
	if err != nil {
		fmt.Printf("scan failed, err:%v\n", err)
		return
	}
	fmt.Printf("RsyncFileId:%v RsyncFileFilename:%v RsyncFileRsyncTime:%v\n", r.RsyncFileId, r.RsyncFileFilename, r.RsyncFileRsyncTime)
}

func Test_Set(t *testing.T) {
	sqlStr := "update user set age=? where id = ?"
	ret, err := db.Exec(sqlStr, 39, 3)
	if err != nil {
		fmt.Printf("update failed, err:%v\n", err)
		return
	}
	n, err := ret.RowsAffected() // 操作影响的行数
	if err != nil {
		fmt.Printf("get RowsAffected failed, err:%v\n", err)
		return
	}
	fmt.Printf("update success, affected rows:%d\n", n)
}

func Test_Delete(t *testing.T) {
	sqlStr := "delete from user where id = ?"
	ret, err := db.Exec(sqlStr, 3)
	if err != nil {
		fmt.Printf("delete failed, err:%v\n", err)
		return
	}
	n, err := ret.RowsAffected() // 操作影响的行数
	if err != nil {
		fmt.Printf("get RowsAffected failed, err:%v\n", err)
		return
	}
	fmt.Printf("delete success, affected rows:%d\n", n)
}
