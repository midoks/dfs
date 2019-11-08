package database

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type DB struct {
	db *sql.DB
}

type BinFile struct {
	Id      int
	Md5     string
	NodeNum int
	Path    string
	Node    string
	Attr    string
	Created string
}

func (this *DB) Init(path string) {
	this.db, _ = sql.Open("sqlite3", path)

	sql_table := `
    CREATE TABLE IF NOT EXISTS bin_file (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        md5 CHAR(32) NULL UNIQUE,
        path TEXT NULL,
        node_num INT NULL,
        node TEXT NULL,
        attr TEXT NULL,
        created DATE NULL
    );
`
	this.db.Exec(sql_table)

	this.db.Exec("create index bin_file_md5 on bin_file(md5);")
}

func (this *DB) FindFileByMd5(md5 string) (*BinFile, bool) {
	var (
		err  error
		rows *sql.Rows
	)
	di := &BinFile{}

	rows, err = this.db.Query(fmt.Sprintf("select id,md5,path,attr,created from bin_file where md5='%s' limit 1", md5))
	if err != nil {
		return di, false
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&di.Id, &di.Md5, &di.Path, &di.Attr, &di.Created)
		if err != nil {
			return di, false
		}
		return di, true
	}

	return di, false
}

func (this *DB) AddFileRow(md5, path string, node_num int, node, attr string) error {
	var (
		err  error
		stmt *sql.Stmt
	)
	stmt, err = this.db.Prepare("INSERT INTO bin_file(md5,path,node_num,node,attr,created) values(?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	created := time.Now().Format("2006-01-02T15:04:05Z")
	_, err = stmt.Exec(md5, path, node_num, node, attr, created)
	if err != nil {
		return err
	}
	return nil
}

func (this *DB) DeleteRowById(id int) error {
	var (
		err  error
		stmt *sql.Stmt
	)
	stmt, err = this.db.Prepare("DELETE FROM bin_file where id=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}
	return nil
}

func (this *DB) DeleteRowByMd5(md5 int) error {
	var (
		err  error
		stmt *sql.Stmt
	)
	stmt, err = this.db.Prepare("DELETE FROM bin_file where md5=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(md5)
	if err != nil {
		return err
	}
	return nil
}

func Open(path string) *DB {
	ds := &DB{}
	ds.Init(path)
	return ds
}
