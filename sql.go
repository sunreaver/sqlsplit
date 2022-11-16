package sqlsplit

import "strings"

type SQLTYPE string

/*
1. DDL（数据定义语言）
    DDL 主要是指如下的四种SQL 语句，以 CREATE、DROP、ALRET开头和 TRUNCATE TABLE 语句。这里主要说一下 TRUNCATE TABLE ，截断表的数据，也就是删除表中的数据，删除这些数据的时候，系统不做日志，因此无法恢复，删除的速度比较快；而DELETE 语句也是删除表中的记录，但它要写日志，删除的数据可以恢复，数据量大的时候删除比较慢。

2. DML（数据操纵语言）
    它们是UPDATE、INSERT、DELETE，就象它的名字一样，这4条命令是用来对数据库里的数据进行操作的语言。

3. DQL（数据查询语言）
    例如：SELECT语句

4. TCL（事务处理语言）
    事物处理语言是指提交、回滚和保留点3句SQL，既是commit、rollback和savepoint。事务是指一系列的连续的不可分割的数据库操作，这些操作要么同时成功，要么同时失败。oracle 的默认事务模型是显式事务模型，即执行完DML后必须手动提交或回滚。

5. DCL（数据控制语言）
    是指授予权限和回收权限语句，既是grant、revoke、deny 等语句。
*/

const (
	DCL SQLTYPE = "DCL" // （数据控制语言）
	DDL SQLTYPE = "DDL" // （数据定义语言）
	DML SQLTYPE = "DML" // （数据操纵语言）
	DQL SQLTYPE = "DQL" // （数据查询语言）
	TTL SQLTYPE = "TTL" // （事务处理语言）
)

func SQLType(raw string) SQLTYPE {
	raw = strings.ToUpper(raw)
	if strings.HasPrefix(raw, "AUDIT") ||
		strings.HasPrefix(raw, "COMMENT") ||
		strings.HasPrefix(raw, "CONNECT") ||
		strings.HasPrefix(raw, "DISCONNECT") ||
		strings.HasPrefix(raw, "EXIT") ||
		strings.HasPrefix(raw, "GRANT") ||
		strings.HasPrefix(raw, "NOAUDIT") ||
		strings.HasPrefix(raw, "QUIT") ||
		strings.HasPrefix(raw, "REVOKE") ||
		// 删除外键表的时候，SET FOREIGN_KEY_CHECKS=0 的问题
		// 同时 DCL 类型 使得 SET不参与explain检测
		strings.HasPrefix(raw, "SET") ||
		strings.HasPrefix(raw, "RENAME") {
		return DCL
	} else if strings.HasPrefix(raw, "DROP") ||
		strings.HasPrefix(raw, "ALTER") ||
		strings.HasPrefix(raw, "TRUNCATE") ||
		strings.HasPrefix(raw, "CREATE") {
		return DDL
	} else if strings.HasPrefix(raw, "INSERT") ||
		strings.HasPrefix(raw, "UPDATE") ||
		strings.HasPrefix(raw, "CALL") ||
		strings.HasPrefix(raw, "DELETE") ||
		strings.HasPrefix(raw, "MERGE") {
		return DML
	} else if strings.HasPrefix(raw, "COMMIT") ||
		strings.HasPrefix(raw, "ROLLBACK") ||
		strings.HasPrefix(raw, "SAVEPOINT") {
		return TTL
	}
	return DQL
}
