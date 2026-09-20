package sqlsplit

import (
	"os"
	"strings"
	"testing"
)

var insqls = []byte{}

func TestMain(m *testing.M) {
	var err error
	insqls, err = os.ReadFile("example/example.sql")
	if err != nil {
		panic(err)
	}
	m.Run()
}

func TestSplitExample(t *testing.T) {
	ps := Split(string(insqls))
	if len(ps) != 44 {
		t.Fatalf("split error, expect 44, got %d", len(ps))
	}

	// 1. 验证常规查询
	if ps[0].Type != DQL || !strings.Contains(ps[0].SQL, "SELECT id, username, email, status") {
		t.Errorf("ps[0] unexpected: %s, %s", ps[0].Type, ps[0].SQL)
	}

	// 2. 验证复核 SQL (SELECT FOR UPDATE) 为 DQL
	if ps[4].Type != DQL || !strings.Contains(ps[4].SQL, "FOR UPDATE") {
		t.Errorf("ps[4] unexpected: %s, %s", ps[4].Type, ps[4].SQL)
	}

	// 3. 验证复核 SQL (INSERT SELECT) 为 DML
	if ps[7].Type != DML || !strings.Contains(ps[7].SQL, "INSERT INTO vip_customer_archive") {
		t.Errorf("ps[7] unexpected: %s, %s", ps[7].Type, ps[7].SQL)
	}

	// 4. 验证视图创建为 DDL
	if ps[8].Type != DDL || !strings.Contains(ps[8].SQL, "CREATE VIEW v_active_customer_overview") {
		t.Errorf("ps[8] unexpected: %s, %s", ps[8].Type, ps[8].SQL)
	}

	// 5. 验证 CTE 查询与 CTE 插入
	if ps[11].Type != DQL || !strings.Contains(ps[11].SQL, "WITH RECURSIVE dept_tree") {
		t.Errorf("ps[11] unexpected: %s, %s", ps[11].Type, ps[11].SQL)
	}
	if ps[12].Type != DML || !strings.Contains(ps[12].SQL, "INSERT INTO user_rank_snapshot") {
		t.Errorf("ps[12] unexpected: %s, %s", ps[12].Type, ps[12].SQL)
	}

	// 6. 验证复杂 Oracle 存储过程完整性（包含内部所有循环、分支、异常）
	if ps[14].Type != DDL || !strings.Contains(ps[14].SQL, "SP_FTP_ACCT_DATA") || !strings.Contains(ps[14].SQL, "end loop;") {
		t.Errorf("ps[14] unexpected Oracle procedure: %s", ps[14].SQL)
	}
	if ps[15].Type != DDL || !strings.Contains(ps[15].SQL, "SP_FTP_DEL_DATA") || !strings.Contains(ps[15].SQL, "SQLERRM") {
		t.Errorf("ps[15] unexpected Oracle procedure SP_FTP_DEL_DATA")
	}

	// 7. 验证 MySQL DELIMITER 存储过程（已消费 DELIMITER 指令，无多余 //）
	if ps[19].Type != DDL || !strings.Contains(ps[19].SQL, "sp_mysql_calc_salary") || strings.HasSuffix(ps[19].SQL, "//") {
		t.Errorf("ps[19] unexpected MySQL procedure: %s", ps[19].SQL)
	}

	// 8. 验证 PostgreSQL $$ 块函数为 DDL
	if ps[23].Type != DDL || !strings.Contains(ps[23].SQL, "get_user_full_name") {
		t.Errorf("ps[23] unexpected PG function: %s", ps[23].SQL)
	}

	// 9. 验证事务控制与权限控制
	if ps[25].Type != TTL || !strings.HasSuffix(ps[25].SQL, "COMMIT") {
		t.Errorf("ps[25] unexpected: %s, %s", ps[25].Type, ps[25].SQL)
	}
	if ps[28].Type != DCL || !strings.HasPrefix(ps[28].SQL, "GRANT") {
		t.Errorf("ps[28] unexpected: %s, %s", ps[28].Type, ps[28].SQL)
	}
}

func TestSplitGranular(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		wantLen  int
		wantSQL  []string
		wantType []SQLTYPE
	}{
		{
			name:     "连续分号与空分号过滤",
			input:    ";;;SELECT 1;;; SELECT 2;;;",
			wantLen:  2,
			wantSQL:  []string{"SELECT 1", "SELECT 2"},
			wantType: []SQLTYPE{DQL, DQL},
		},
		{
			name:     "纯注释与空语句不产生条目",
			input:    ";; -- only comment\n/* block */ ;",
			wantLen:  0,
			wantSQL:  nil,
			wantType: nil,
		},
		{
			name:     "注释附着下一条语句且末尾注释丢弃",
			input:    "-- user query\nSELECT * FROM users; -- end comment",
			wantLen:  1,
			wantSQL:  []string{"-- user query\nSELECT * FROM users"},
			wantType: []SQLTYPE{DQL},
		},
		{
			name:     "字符串与反引号内分号保护",
			input:    "SELECT 'a;b' AS s, `col;name` FROM `db;1`.`t;2` WHERE c = \"x;y\";",
			wantLen:  1,
			wantSQL:  []string{"SELECT 'a;b' AS s, `col;name` FROM `db;1`.`t;2` WHERE c = \"x;y\""},
			wantType: []SQLTYPE{DQL},
		},
		{
			name:     "SELECT FOR UPDATE 复核语句",
			input:    "SELECT * FROM account WHERE id = 1 FOR UPDATE;",
			wantLen:  1,
			wantSQL:  []string{"SELECT * FROM account WHERE id = 1 FOR UPDATE"},
			wantType: []SQLTYPE{DQL},
		},
		{
			name:     "INSERT SELECT 复合复核语句",
			input:    "INSERT INTO t1 (a, b) SELECT x, y FROM t2;",
			wantLen:  1,
			wantSQL:  []string{"INSERT INTO t1 (a, b) SELECT x, y FROM t2"},
			wantType: []SQLTYPE{DML},
		},
		{
			name:     "CREATE VIEW 语句正确切分且无末尾分号",
			input:    "CREATE VIEW v1 AS SELECT 1 AS num; CREATE OR REPLACE VIEW v2 AS SELECT 2;",
			wantLen:  2,
			wantSQL:  []string{"CREATE VIEW v1 AS SELECT 1 AS num", "CREATE OR REPLACE VIEW v2 AS SELECT 2"},
			wantType: []SQLTYPE{DDL, DDL},
		},
		{
			name:     "PostgreSQL $$ 块函数切分",
			input:    "CREATE FUNCTION f() RETURNS int AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;",
			wantLen:  1,
			wantSQL:  []string{"CREATE FUNCTION f() RETURNS int AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;"},
			wantType: []SQLTYPE{DDL},
		},
		{
			name: "MySQL DELIMITER 切换切分",
			input: "DELIMITER //\nCREATE PROCEDURE p() BEGIN SELECT 1; END //\nDELIMITER ;\nSELECT 2;",
			wantLen: 2,
			wantSQL: []string{
				"CREATE PROCEDURE p() BEGIN SELECT 1; END",
				"SELECT 2",
			},
			wantType: []SQLTYPE{DDL, DQL},
		},
		{
			name: "Oracle 独立行斜杠结束符",
			input: "DECLARE\n  x INT;\nBEGIN\n  NULL;\nEND;\n/\nSELECT 100;",
			wantLen: 2,
			wantType: []SQLTYPE{DML, DQL},
		},
		{
			name:    "空输入",
			input:   "",
			wantLen: 0,
		},
		{
			name:    "纯空白输入",
			input:   "   \n\t  ",
			wantLen: 0,
		},
		{
			name:     "CREATE TRIGGER 复合块",
			input:    "CREATE TRIGGER trg_audit AFTER INSERT ON orders FOR EACH ROW BEGIN INSERT INTO audit_log(action) VALUES('insert'); END;",
			wantLen:  1,
			wantType: []SQLTYPE{DDL},
		},
		{
			name:    "嵌套 IF/LOOP 结构",
			input:   "CREATE PROCEDURE sp_nested() BEGIN IF 1=1 THEN LOOP SELECT 1; END LOOP; END IF; END;",
			wantLen: 1,
			wantSQL: []string{"CREATE PROCEDURE sp_nested() BEGIN IF 1=1 THEN LOOP SELECT 1; END LOOP; END IF; END;"},
		},
		{
			name:     "转义单引号（反斜杠与双单引号）",
			input:    "SELECT 'it\\'s' AS a; SELECT 'it''s' AS b;",
			wantLen:  2,
			wantSQL:  []string{"SELECT 'it\\'s' AS a", "SELECT 'it''s' AS b"},
			wantType: []SQLTYPE{DQL, DQL},
		},
		{
			name:     "转义双引号",
			input:    "SELECT \"col\\\"name\" FROM t; SELECT \"col\"\"name\" FROM t;",
			wantLen:  2,
			wantType: []SQLTYPE{DQL, DQL},
		},
		{
			name:     "CTE 含括号字符串不干扰类型判断",
			input:    "WITH cte AS (SELECT '(fake)' AS val) INSERT INTO t SELECT * FROM cte;",
			wantLen:  1,
			wantType: []SQLTYPE{DML},
		},
		{
			name:     "CTE 含块注释中的括号不干扰类型判断",
			input:    "WITH cte AS (SELECT /* ) */ 1 AS val) INSERT INTO t SELECT * FROM cte;",
			wantLen:  1,
			wantType: []SQLTYPE{DML},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Split(tc.input)
			if len(got) != tc.wantLen {
				t.Fatalf("Split() len = %d, want %d. Got: %+v", len(got), tc.wantLen, got)
			}
			for i := 0; i < tc.wantLen; i++ {
				if tc.wantSQL != nil && got[i].SQL != tc.wantSQL[i] {
					t.Errorf("[%d] SQL = %q, want %q", i, got[i].SQL, tc.wantSQL[i])
				}
				if tc.wantType != nil && got[i].Type != tc.wantType[i] {
					t.Errorf("[%d] Type = %v, want %v", i, got[i].Type, tc.wantType[i])
				}
			}
		})
	}
}
