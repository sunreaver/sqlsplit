package sqlsplit

import (
	"fmt"
	"os"
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
	if len(ps) != len(expectedExampleCases) {
		t.Fatalf("split error, expect %d, got %d", len(expectedExampleCases), len(ps))
	}

	for i, tc := range expectedExampleCases {
		caseName := fmt.Sprintf("Case_%02d", i+1)
		t.Run(caseName, func(t *testing.T) {
			got := ps[i]
			if got.Type != tc.WantType {
				t.Errorf("[%s] Type mismatch: got %v, want %v", caseName, got.Type, tc.WantType)
			}
			if got.SQL != tc.WantSQL {
				t.Errorf("[%s] SQL mismatch:\ngot:\n%s\nwant:\n%s", caseName, got.SQL, tc.WantSQL)
			}
		})
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
			name:    "MySQL DELIMITER 切换切分",
			input:   "DELIMITER //\nCREATE PROCEDURE p() BEGIN SELECT 1; END //\nDELIMITER ;\nSELECT 2;",
			wantLen: 2,
			wantSQL: []string{
				"CREATE PROCEDURE p() BEGIN SELECT 1; END",
				"SELECT 2",
			},
			wantType: []SQLTYPE{DDL, DQL},
		},
		{
			name:     "Oracle 独立行斜杠结束符",
			input:    "DECLARE\n  x INT;\nBEGIN\n  NULL;\nEND;\n/\nSELECT 100;",
			wantLen:  2,
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
