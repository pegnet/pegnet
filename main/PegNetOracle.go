package main

import (
	"fmt"

	"github.com/pegnet/OracleRecord"
)

func main() {
	var opr oprecord.OraclePriceRecord
	opr.GetOPRecord()
	fmt.Println(opr)
}

/*   Not used right now.  structures are there if you want to use it

 */
