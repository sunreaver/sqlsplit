package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sunreaver/sqlsplit"
)

func main() {
	f := flag.String("f", "example.sql", "解析的文件")
	if len(*f) == 0 {
		os.Exit(1)
	}
	data, e := ioutil.ReadFile(*f)
	if e != nil {
		fmt.Printf("打开文件 %v 失败: %v", *f, e.Error())
		os.Exit(1)
	}
	outs := sqlsplit.Split(string(data))
	fmt.Printf("共解析出%v条sql\n", len(outs))
	for idx, v := range outs {
		fmt.Printf("第%v条--------------------------------------\n", idx+1)
		fmt.Println("类型:", v.Type)
		fmt.Println(v.SQL)
	}
}
