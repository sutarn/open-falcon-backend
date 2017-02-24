package db

import (
	"database/sql"
	"fmt"
	"log"
)

type TxFinale byte

const (
	TxCommit TxFinale = 1
	TxRollback TxFinale = 2
)

// Configuration of database
type DbConfig struct {
	Dsn string
	MaxIdle int
}

func (config *DbConfig) String() string {
	return fmt.Sprintf("DSN: [%s]. Max Idle: [%d]", config.Dsn, config.MaxIdle)
}

// The main functions of this file is to gives IoC(Inverse of Control) of database(RDB) objects.
//
// For exception handling, all callback method should use panic() or log.Panicf() to release the error object.
//
// You may use PanicIfError to ease your process of Error object.

// Main controller of database
type DbController struct {
	dbObject *sql.DB
	panicHandlers []PanicHandler
}

// The interface of DB callback for sql package
type DbCallback interface {
	OnDb(db *sql.DB)
}

// The function object delegates the DbCallback interface
type DbCallbackFunc func(*sql.DB)
func (f DbCallbackFunc) OnDb(db *sql.DB) {
	f(db)
}

// The interface of rows callback for sql package
type RowsCallback interface {
	NextRow(row *sql.Rows) IterateControl
}

// The function object delegates the RowsCallback interface
type RowsCallbackFunc func(*sql.Rows) IterateControl
func (callbackFunc RowsCallbackFunc) NextRow(rows *sql.Rows) IterateControl {
	return callbackFunc(rows)
}

// The interface of row callback for sql package
type RowCallback interface {
	ResultRow(row *sql.Row)
}

// The function object delegates the RowCallback interface
type RowCallbackFunc func(*sql.Row)
func (callbackFunc RowCallbackFunc) ResultRow(row *sql.Row) {
	callbackFunc(row)
}

// The interface of transaction callback for sql package
type TxCallback interface {
	InTx(tx *sql.Tx) TxFinale
}

// The function object delegates the TxCallback interface
type TxCallbackFunc func(*sql.Tx) TxFinale
func (callbackFunc TxCallbackFunc) InTx(tx *sql.Tx) TxFinale {
	return callbackFunc(tx)
}

// BuildTxForSqls builds function for exeuction of multiple SQLs
func BuildTxForSqls(queries... string) TxCallback {
	return TxCallbackFunc(func(tx *sql.Tx) TxFinale {
		for _, v := range queries {
			if _, err := tx.Exec(v); err != nil {
				PanicIfError(err)
			}
		}

		return TxCommit
	})
}

// Executes callbacks in transaction if the boot callback has true value
type ExecuteIfByTx interface {
	// First calling of database for boolean result
	BootCallback(tx *sql.Tx) bool
	// If the boot callback has true result, this callback would get called
	IfTrue(tx *sql.Tx)
}

// Extension for sql.Rows
type RowsExt sql.Rows

// converts the sql.Rows to RowsExt
func ToRowsExt(rows *sql.Rows) *RowsExt {
	return ((*RowsExt)(rows))
}

// Gets columns, with panic instead of returned error
func (rowsExt *RowsExt) Columns() ([]string) {
	columns, err := ((*sql.Rows)(rowsExt)).Columns()
	PanicIfError(err)

	return columns
}

// Scans the values of row into variables, with panic instead of returned error
func (rowsExt *RowsExt) Scan(dest ...interface{}) {
	err := ((*sql.Rows)(rowsExt)).Scan(dest...)
	PanicIfError(err)
}

// Extension for sql.Row
type RowExt sql.Row

// Converts the sql.Row to RowExt
func ToRowExt(row *sql.Row) *RowExt {
	return ((*RowExt)(row))
}

// Scans the values of row into variables, with panic instead of returned error
func (rowExt *RowExt) Scan(dest ...interface{}) {
	err := ((*sql.Row)(rowExt)).Scan(dest...)
	PanicIfError(err)
}

// Extension for sql.Stmt
type StmtExt sql.Stmt

// Converts sql.Stmt to StmtExt
func ToStmtExt(stmt *sql.Stmt) *StmtExt {
	return ((*StmtExt)(stmt))
}

// Exec with panic instead of error
func (stmtExt *StmtExt) Exec(args ...interface{}) sql.Result {
	result, err := ((*sql.Stmt)(stmtExt)).Exec(args...)
	PanicIfError(err)

	return result
}

// Query with panic instead of error
func (stmtExt *StmtExt) Query(args ...interface{}) *sql.Rows {
	rows, err := ((*sql.Stmt)(stmtExt)).Query(args...)
	PanicIfError(err)

	return rows
}

// Extnesion for sql.Tx
type TxExt sql.Tx

// Converts sql.Tx to TxExt
func ToTxExt(tx *sql.Tx) *TxExt {
	return ((*TxExt)(tx))
}

// Commit with panic instead of returned error
func (txExt *TxExt) Commit() {
	err := ((*sql.Tx)(txExt)).Commit()
	PanicIfError(err)
}

// Commit with panic instead of returned error
func (txExt *TxExt) Exec(query string, args ...interface{}) sql.Result {
	result, err := ((*sql.Tx)(txExt)).Exec(query, args...)
	PanicIfError(err)

	return result
}

// Prepare with panic instead of returned error
func (txExt *TxExt) Prepare(query string) *sql.Stmt {
	stmt, err := ((*sql.Tx)(txExt)).Prepare(query)
	PanicIfError(err)

	return stmt
}

// Query with panic instead of returned error
func (txExt *TxExt) Query(query string, args ...interface{}) *sql.Rows {
	rows, err := ((*sql.Tx)(txExt)).Query(query)
	PanicIfError(err)

	return rows
}

// Rollback with panic instead of returned error
func (txExt *TxExt) Rollback() {
	err := ((*sql.Tx)(txExt)).Rollback()
	PanicIfError(err)
}

// Extension for sql.Result
type ResultExt struct {
	sqlResult sql.Result
}

// Converts sql.Result to ResultExt
func ToResultExt (result sql.Result) *ResultExt {
	return &ResultExt{ result }
}

// Gets last id of insert with panic instead of returned error
func (resultExt *ResultExt) LastInsertId() int64 {
	insertId, err := resultExt.sqlResult.LastInsertId()
	PanicIfError(err)

	return insertId
}

// Gets last number of affected rows with panic instead of returned error
func (resultExt *ResultExt) RowsAffected() int64 {
	numberOfRowsAffected, err := resultExt.sqlResult.RowsAffected()
	PanicIfError(err)

	return numberOfRowsAffected
}

// The handler used to handler panic
//
// You should use this type with DbController.RegisterPanicHandler to
// customize your logic of panic
type PanicHandler func (panicValue interface{})

// The control of iterating
type IterateControl byte
const (
	IterateContinue = IterateControl(1)
	IterateStop = IterateControl(0)
)

// Initialize a controller for database
//
// Without RegisterPanicHandler() any PanicHandler,
// The raised panic would be re-paniced.
func NewDbController(newDbObject *sql.DB) *DbController {
	if newDbObject == nil {
		PanicIfError(fmt.Errorf("Need viable DB object(sql.DB)"))
	}

	return &DbController{
		dbObject: newDbObject,
		panicHandlers: make([]PanicHandler, 0),
	}
}

// Builds handler for error capture
//
// errHolder - The error object to holding captured one
func NewDbErrorCapture(errHolder *error) PanicHandler {
	return func (panicValue interface{}) {
		err, ok := panicValue.(error)
		if !ok {
			PanicIfError(fmt.Errorf("The panic[%v] is not a error object", panicValue))
		}

		*errHolder = err
	}
}

// Registers a handler while a panic is raised
//
// This object may register multiple handlers for panic
func (dbController *DbController) RegisterPanicHandler(panicHandler PanicHandler) {
	dbController.panicHandlers = append(dbController.panicHandlers, panicHandler)
}

// Operate on database
func (dbController *DbController) OperateOnDb(dbCallback DbCallback) {
	dbController.needInitializedOrPanic()
	defer dbController.handlePanic()

	dbCallback.OnDb(dbController.dbObject)
}

// Executes the query string or panic
func (dbController *DbController) Exec(query string, args ...interface{}) sql.Result {
	var result sql.Result
	var dbFunc DbCallbackFunc = func(db *sql.DB) {
		innerResult, err := db.Exec(query, args...)
		PanicIfError(err)

		result = innerResult
	}

	dbController.OperateOnDb(dbFunc)
	return result
}

// Query for rows and get called of rows with Next()
func (dbController *DbController) QueryForRows(
	rowsCallback RowsCallback,
	sqlQuery string, args ...interface{},
) (numberOfRows uint) {
	var dbFunc DbCallbackFunc = func(db *sql.DB) {
		rows, err := db.Query(
			sqlQuery, args...,
		)
		if err != nil {
			log.Panicf(
				"Query SQL with exception: %v SQL: [%s] Params: [%v]",
				err, sqlQuery, args,
			)
		}

		defer rows.Close()
		for rows.Next() {
			numberOfRows++

			if rowsCallback.NextRow(rows) == IterateStop {
				break
			}
		}
	}

	dbController.OperateOnDb(dbFunc)

	return
}

// Query for a row and get called if the query is not failed
func (dbController *DbController) QueryForRow(
	rowCallback RowCallback,
	sqlQuery string, args ...interface{},
) {
	var dbFunc DbCallbackFunc = func(db *sql.DB) {
		row := db.QueryRow(
			sqlQuery, args...,
		)

		rowCallback.ResultRow(row)
	}

	dbController.OperateOnDb(dbFunc)
}

// Executes in transaction.
//
// This method would commit the transaction if there is no raised panic,
// rollback it otherwise.
func (dbController *DbController) InTx(txCallback TxCallback) {
	var dbFunc DbCallbackFunc = func(db *sql.DB) {
		tx, err := db.Begin()
		PanicIfError(err)

		/**
		 * Rollback the transaction when panic is rised
		 */
		defer func() {
			p := recover()
			if p == nil {
				return
			}

			rollbackError := tx.Rollback()
			if rollbackError != nil {
				p = fmt.Errorf("Transaction has Error: %v. Rollback has error too: %v", p, rollbackError)
			}
			panic(p)
		}()
		// :~)

		switch txCallback.InTx(tx) {
		case TxCommit:
			PanicIfError(tx.Commit())
		case TxRollback:
			PanicIfError(tx.Rollback())
		}
	}

	dbController.OperateOnDb(dbFunc)
}

// Executes the complex statement in transaction
func (dbController *DbController) InTxForIf(ifCallbacks ExecuteIfByTx) {
	var txFunc TxCallbackFunc = func(tx *sql.Tx) TxFinale {
		if ifCallbacks.BootCallback(tx) {
			ifCallbacks.IfTrue(tx)
		}

		return TxCommit
	}

	dbController.InTx(txFunc)
}

// Executes in transaction
func (dbController *DbController) ExecQueriesInTx(queries... string) {
	dbController.InTx(BuildTxForSqls(queries...))
}

// Releases the database object under this object
//
// As of service application(web, daemon...), this method is rarely get called
func (dbController *DbController) Release() {
	dbController.needInitializedOrPanic()
	defer dbController.handlePanic()

	err := dbController.dbObject.Close()

	if err != nil {
		log.Panicf("Release database connection error. %v", err)
	}

	dbController.dbObject = nil
}

func (dbController *DbController) needInitializedOrPanic() {
	if dbController.dbObject == nil {
		PanicIfError(fmt.Errorf("The controller is not initialized"))
	}
}

func (dbController *DbController) handlePanic() {
	p := recover()
	if p == nil {
		return
	}

	if len(dbController.panicHandlers) == 0 {
		panic(p)
	}

	for _, handler := range dbController.panicHandlers {
		handler(p)
	}
}
