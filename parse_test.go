package sqlsplit

import (
	"io/ioutil"
	"testing"
)

var insqls = []byte{}

func TestMain(m *testing.M) {
	var err error
	insqls, err = ioutil.ReadFile("example/example.sql")
	if err != nil {
		panic(err)
	}
	m.Run()
}

func TestSplit(t *testing.T) {
	ps := Split(string(insqls)) // 测试分词
	if len(ps) != 50 {
		t.Errorf("split error, expect 49, got %d", len(ps))
	}
}
