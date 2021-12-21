package main

import (
	"fmt"
	"io/ioutil"

	"github.com/sunreaver/sqlsplit"
)

func main() {
	data, _ := ioutil.ReadFile("example.sql")
	outs := sqlsplit.Split(string(data))
	fmt.Printf("共解析出%v条sql\n", len(outs))
	for idx, v := range outs {
		fmt.Printf("第%v条--------------------------------------\n", idx+1)
		fmt.Println("类型:", v.Type)
		fmt.Println(v.SQL)
	}
}
