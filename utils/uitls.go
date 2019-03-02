package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pegnet/OracleRecord/common"
)

func PullValue(line string, howMany int) string {
	i := 0
	//fmt.Println(line)
	var pos int
	for i < howMany {
		//find the end of the howmany-th tag
		pos = strings.Index(line, ">")
		line = line[pos+1:]
		//fmt.Println(line)
		i = i + 1
	}
	//fmt.Println("line:", line)
	pos = strings.Index(line, "<")
	//fmt.Println("POS:", pos)
	line = line[0:pos]
	//fmt.Println(line)
	return line
}
func FloatStringToInt(floatString string) int64 {
	//fmt.Println(floatString)
	if floatString == "-" {
		return 0
	}
	if strings.TrimSpace(floatString) == "" {
		return 0
	}
	floatValue, err := strconv.ParseFloat(floatString, 64)
	if err != nil {
		fmt.Println("ParseError:", floatString)
		return 0
	} else {
		return int64(floatValue * common.PointMultiple)
	}

}
