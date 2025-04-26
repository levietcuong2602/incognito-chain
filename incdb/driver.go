package incdb

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/pkg/errors"
)

// Driver defines a structure for backend drivers to use when they registered
// themselves as a backend which implements the Database interface.
type Driver struct {
	DbType string
	Open   func(args ...interface{}) (Database, error)
}

var drivers = make(map[string]*Driver)

// RegisterDriver registers the driver d.
func RegisterDriver(d Driver) error {
	if _, exists := drivers[d.DbType]; exists {
		return errors.Wrapf(errors.New("Driver is already registered"), d.DbType)
	}
	drivers[d.DbType] = &d
	return nil
}

// Open opens the db connection.
func Open(typ string, args ...interface{}) (Database, error) {
	d, exists := drivers[typ]
	if !exists {
		return nil, errors.Wrapf(errors.New("Driver is not registered"), typ)
	}
	return d.Open(args...)
}

// Open opens the db connection.
func OpenMultipleDB(typ string) (map[int]Database, error) {
	m := make(map[int]Database)
	d, exists := drivers[typ]
	if !exists {
		return nil, errors.Wrapf(errors.New("Driver is not registered"), typ)
	}
	for i := -1; i < common.MaxShardNumber; i++ {
		newPath := config.Config().GetShardDataDir(i)
		db, err := d.Open(newPath)
		if err != nil {
			return nil, errors.WithStack(fmt.Errorf("Open database error %+v", err))
		}
		m[i] = db
	}
	return m, nil
}

func OpenDBWithPath(typ string, path string) (Database, error) {
	d, exists := drivers[typ]
	if !exists {
		return nil, errors.Wrapf(errors.New("Driver is not registered"), typ)
	}
	db, err := d.Open(path)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("Open database error %+v", err))
	}
	return db, nil
}
