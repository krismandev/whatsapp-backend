package lib

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	//load mysql driver
	"github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	//_ "github.com/lib/pq"
)

// DatabaseConnection : interface for database action
type DatabaseConnection interface {
	InitDB(fn string) error
	Open() (*sql.DB, error)
	Close()
	GetRows(rows *sql.Rows) (map[int]map[string]string, error)

	GetFirstRow() (string, error)
	Query(strSQL string, args ...interface{}) (*sql.Rows, error)
	Exec(strSQL string, args ...interface{}) (int64, error)
	Queryf(strSQL string, args ...interface{}) (*sql.Rows, error)
	Execf(strSQL string, args ...interface{}) (int64, error)

	// Added by Budianto
	GetRowsbyIndex(rows *sql.Rows) (map[int]map[int]string, error)
	GetFirstData(strSQL string, args ...interface{}) (string, error)
	GetFirstRowByQuery(strSQL string, key string, args ...interface{}) (string, error)
}

// InitDB is use for Initialize DB Connection
func InitDB(dbType string,dbURL string) DBConnection {
	var dbConn *DBConnection	
	dbConn, err := NewConnection()
	dbConn.DBType=dbType
	dbConn.DBURL=dbURL
	if err != nil {
		log.Errorf("Unable to initialize database %+v", err)
		os.Exit(1)
	}
	dbConn.DB, err = dbConn.Open(dbType,dbURL)
	if err != nil {
		log.Errorf("Unable to open database %+v", err)
		os.Exit(1)
	}
	return *dbConn
}

// DBConnection is use for database connection
type DBConnection struct {
	DB *sql.DB
	DBType 	string
	DBURL	string
}

//var c.DB *sql.DB

// NewConnection is use to create db connection
func NewConnection() (*DBConnection, error) {
	var c DBConnection
	return &c, nil
}

// Open function prepares DBConnection for future connection to database
func (c DBConnection) Open(dbType string,dbURL string) (*sql.DB, error) {
	c.Close()

	var err error
	// Open database connection
	dbConn, err := sql.Open(dbType,dbURL)
	if err != nil {
		return nil, err
	}

	dbConn.SetMaxOpenConns(100)
	dbConn.SetConnMaxLifetime(time.Minute * 1)
	dbConn.SetMaxIdleConns(100)

	err = dbConn.Ping()
	if err != nil {
		return nil, err
	}
	log.Infof("Database connection %s", "Initiated")
	return dbConn, nil
}

// Close function closes existing DBConnection
//
func (c DBConnection) Close() {
	if c.DB != nil {
		log.Debug("Closing previous database connection.")
		c.DB.Close()
		c.DB = nil
	}
}

//GetRows parses recordset into map
func (c DBConnection) GetRows(rows *sql.Rows) (map[int]map[string]string,int, error) {
	var results map[int]map[string]string
	results = make(map[int]map[string]string)
	if (rows !=nil){
		defer rows.Close()
	}	
	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil,0, err
	}

	// Make a slice for the values
	values := make([]sql.RawBytes, len(columns))

	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Fetch rows
	counter := 1
	for rows.Next() {
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil,0, err
		}

		// initialize the second layer
		results[counter] = make(map[string]string)

		// Now do something with the data.
		// Here we just print each column as a string.
		var value string
		for i, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			results[counter][columns[i]] = value
		}
		counter++
	}
	if err = rows.Err(); err != nil {
		return nil,0, err
	}
	
	return results,counter, nil
}

//GetFirstRow parse and gets column value in first record
func (c DBConnection) GetFirstRow(rows *sql.Rows, key string) (string, error) {
	if (rows !=nil){
		defer rows.Close()
	}	
	results,_, err := c.GetRows(rows)
	if err != nil {
		return "", err
	}
	
	return results[1][key], nil
}

// Query sends SELECT command to database
func (c DBConnection) Query(strSQL string, args ...interface{}) (*sql.Rows, error) {
	// if no DBConnection, return
	//
	if c.DB == nil {
		return nil, fmt.Errorf("database needs to be initiated first")
	}
	check, errCheck := c.CheckDB(true)
	if !check {
		return nil, errCheck
	}

	//if strSQL, found = sqlCommandMap[strSQL]; !found {
	rows, err := c.DB.Query(strSQL, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

//Exec executes UPDATE/INSERT/DELETE statements and returns rows affected
//
func (c DBConnection) Exec(strSQL string, args ...interface{}) (int64,int, error) {
	// if no DBConnection, return
	var errNumber int =0
	if c.DB == nil {
		return 0,errNumber, fmt.Errorf("Please OpenConnection prior Query")
	}
	check, errCheck := c.CheckDB(true)
	if !check {
		return 0,errNumber, errCheck
	}

	res, err := c.DB.Exec(strSQL, args...)
	if err != nil {
		if mysqlError,Ok :=err.(*mysql.MySQLError); Ok{
			errNumber=int(mysqlError.Number)
		}
		return 0,errNumber, err		
	}

	rows, err := res.RowsAffected()

	if err != nil {		
		if mysqlError,Ok :=err.(*mysql.MySQLError); Ok{
			errNumber=int(mysqlError.Number)
		}
		return 0,errNumber, err
	}
	return rows, errNumber, nil
}

// InsertGetLastID is use for ...
func (c DBConnection) InsertGetLastID(strSQL string, args ...interface{}) (int64, error) {
	// if no DBConnection, return
	//
	if c.DB == nil {
		return 0, fmt.Errorf("Please OpenConnection prior Query")
	}
	check, errCheck := c.CheckDB(true)
	if !check {
		return 0, errCheck
	}

	// Execute the query
	res, err := c.DB.Exec(strSQL, args...)
	if err != nil {
		return 0,err
		//panic(err.Error()) // proper error handling instead of panic in your app
	}

	rows, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return rows, nil
}

//GetRowsbyIndex parses Get row using index of row and column
func (c DBConnection) GetRowsbyIndex(rows *sql.Rows) (map[int]map[int]string, int, error) {
	var results map[int]map[int]string
	results = make(map[int]map[int]string)
	if (rows !=nil){
		defer rows.Close()
	}	
	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, 0, err
	}

	//Get Column name
	values := make([]sql.RawBytes, len(columns))

	//Define dynamic variables base on column name
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Fetch rows
	counter := 1 //row count
	for rows.Next() {
		// get RawBytes from data
		// Assign value to dynamic variables
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, 0, err
		}

		// initialize the second layer
		results[counter] = make(map[int]string)

		// Now do something with the data.
		// Here we just print each column as a string.
		var value string
		for i, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			results[counter][i] = value
		}
		counter++
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	
	//[row][column]
	return results, counter, nil
}

//GetCustomRowColumn parses Get any number of  row or column where 0 = get all
//row start from 1, column index start from 0
//maxRow,maxColumn = row/column count, start from
func (c DBConnection) GetCustomRowColumn(rows *sql.Rows, maxRow int, maxColumn int) (map[int]map[int]string, error) {
	var results map[int]map[int]string
	results = make(map[int]map[int]string)
	if (rows !=nil){
		defer rows.Close()
	}	
	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	//Get Column name
	values := make([]sql.RawBytes, len(columns))

	//Define dynamic variables base on column name
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Fetch rows
	rowCounter := 1
	for rows.Next() {
		// get RawBytes from data
		// Assign value to dynamic variables
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		// initialize the second layer
		results[rowCounter] = make(map[int]string)

		// Now do something with the data.
		// Here we just print each column as a string.
		var value string
		for colCounter, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			results[rowCounter][colCounter] = value
			if (maxColumn > 0) && (colCounter+1 >= maxColumn) {
				break
			}
		}
		if (maxRow > 0) && (rowCounter >= maxRow) {
			break
		}
		rowCounter++
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return results, nil
}

//GetFirstRowByQuery parse and gets column value in first record by query
func (c DBConnection) GetFirstRowByQuery(strSQL string, args ...interface{}) (map[int]map[int]string, error) {
	var rowret *sql.Rows
	if (rowret !=nil){
		defer rowret.Close()
	}	
	var err error
	if c.DB == nil {
		return nil, fmt.Errorf("Please OpenConnection prior Query")
	}
	check, errCheck := c.CheckDB(true)
	if !check {
		return nil, errCheck
	}

	rowret, err = c.Query(strSQL, args...)
	if err != nil {
		return nil, err
	}

	results, err := c.GetCustomRowColumn(rowret, 1, 0)
	if err != nil {
		return nil, err
	}
	
	return results, nil
}

//GetFirstData get result from sql command that return only 1 row and 1 column ony
func (c DBConnection) GetFirstData(strSQL string, args ...interface{}) (string, error) {
	
	if c.DB == nil {
		return "", fmt.Errorf("Please OpenConnection prior Query")
	}
	check, errCheck := c.CheckDB(true)
	if !check {
		return "", errCheck
	}

	rowret, err := c.Query(strSQL, args...)
	if (rowret !=nil){
		defer rowret.Close()
	}	

	if err != nil {
		return "", err
	}
	results, err := c.GetCustomRowColumn(rowret, 1, 1)
	if err != nil {
		return "", err
	}
	
	return results[1][0], nil
}

//GetFirstColumn get result from sql command that return only 1 column ony
func (c DBConnection) GetFirstColumn(strSQL string,args ...interface{}) ([]string, error) {
	
	if c.DB == nil {
		return nil, fmt.Errorf("Please OpenConnection prior Query")
	}
	check, errCheck := c.CheckDB(true)
	if !check {
		return nil, errCheck
	}

	rowret, err := c.Query(strSQL, args...)
	if (rowret !=nil){
		defer rowret.Close()
	}	

	if err != nil {
		return nil, err
	}
	results, err := c.GetCustomRowColumn(rowret, 0, 1)
	if err != nil {
		return nil, err
	}
	var data []string
	for _,col:=range results {
		data=append(data, col[0])
	}

	return data, nil
}

//SelectQuery parse and gets column value in first record by query
func (c DBConnection) SelectQuery(strSQL string, args ...interface{}) (map[int]map[int]string, int, error) {
	var rowret *sql.Rows
	var err error
	if (rowret !=nil){
		defer rowret.Close()
	}	
	if c.DB == nil {
		return nil, 0, fmt.Errorf("Please OpenConnection prior Query")
	}
	check, errCheck := c.CheckDB(true)
	if !check {
		return nil, 0, errCheck
	}

	rowret, err = c.Query(strSQL, args...)
	if err != nil {
		return nil, 0, err
	}

	results, rowCount, err := c.GetRowsbyIndex(rowret)
	if err != nil {
		return nil, 0, err
	}	
	return results, rowCount - 1, nil
}

//SelectQueryByFieldName parse and gets column value by field name
func (c DBConnection) SelectQueryByFieldName(strSQL string, args ...interface{}) (map[int]map[string]string,int, error) {
	var rowret *sql.Rows
	var err error
	if (rowret !=nil){
		defer rowret.Close()
	}	
	if c.DB == nil {
		return nil,0, fmt.Errorf("Please OpenConnection prior Query")
	}
	check, errCheck := c.CheckDB(true)
	if !check {
		return nil,0,errCheck
	}

	rowret, err = c.Query(strSQL, args...)
	if err != nil {
		return nil,0,  err
	}

	results,rowCount,  err := c.GetRows(rowret)
	if err != nil {
		return nil,0, err
	}
	

	return results,rowCount -1, nil
}

// CheckDB  function prepares DBConnection for future connection to database
func (c DBConnection) CheckDB(Reconnect bool) (bool, error) {
	var err error
	err = c.DB.Ping()
	if err != nil {
		if Reconnect {
			c = InitDB(c.DBType,c.DBURL)
		} else {
			log.Errorf("Database connection %s", "Failed")
			return false, fmt.Errorf("Database connection failed")
		}

	}

	return true, err
}
