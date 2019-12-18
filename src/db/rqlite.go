package db

import "github.com/rqlite/gorqlite"

// RqliteDb is the rqlite driver
type RqliteDb struct {
	conn gorqlite.Connection
}

//Init is a Db interface method.
func (s RqliteDb) Init() error {
	var err error
	s.conn, err = gorqlite.Open("http://localhost:4001")
	if err != nil {
		return err
	}

	s.conn.SetConsistencyLevel("weak")
	//s.conn.SetTimeout(10) // measured in seconds

	return nil
}

//Sync is a Db interface method.
func (s RqliteDb) Sync() {

}

//Health is a Db interface method
func (s RqliteDb) Health() bool {
	if len(s.conn.ID) > 0 {
		return true
	}
	return false
}

// Close closes the Db driver
func (s RqliteDb) Close() {
	s.conn.Close()
}

//NewRqliteDb initialize a Rqlite Db
func NewRqliteDb() {

}

// Create creates a new document
func (b RqliteDb) Create() {

}
