package oprecord

import (
	"fmt"
	"testing"
)

func TestAPICall(t *testing.T) {
	data, err := CallAPILayer()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(data))
}
