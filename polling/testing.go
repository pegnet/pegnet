package polling

import (
	"fmt"

	"github.com/zpatrick/go-config"
)

// NewTestingDataSource is for unit test.
// Having a testing data source is for unit test mocking
var NewTestingDataSource = func(config *config.Config, source string) (IDataSource, error) {
	return nil, fmt.Errorf("this is a testing datasource for unit tests only")
}
