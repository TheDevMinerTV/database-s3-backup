package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type connectionOptions struct {
	Host     string
	DbType   string
	Port     int
	Database string
	Username string
	Password string
}

var (
	PGDumpCmd           = "pg_dump"
	pgDumpStdOpts       = []string{"--no-owner", "--no-acl", "--clean", "--blobs", "-v"}
	pgDumpDefaultFormat = "c"
	ErrPgDumpNotFound   = errors.New("pg_dump not found")

	MysqlDumpCmd         = "mysqldump"
	mysqlDumpStdOpts     = []string{"--compact", "--skip-add-drop-table", "--skip-add-locks", "--skip-disable-keys", "--skip-set-charset", "-v"}
	ErrMySqlDumpNotFound = errors.New("mysqldump not found")

	ErrUnsupportedType = errors.New("unsupported database type")
)

func RunDump(connectionOpts *connectionOptions, outFile string) error {
	cmd, err := buildDumpCommand(connectionOpts, outFile)
	if err != nil {
		return err
	}

	return executeCommand(cmd)
}

func buildDumpCommand(opts *connectionOptions, outFile string) (*exec.Cmd, error) {
	switch opts.DbType {
	case "postgres":
		if !commandExist(PGDumpCmd) {
			return nil, ErrPgDumpNotFound
		}
		options := append(
			pgDumpStdOpts,
			fmt.Sprintf("-f%s", outFile),
			fmt.Sprintf("--dbname=%s", opts.Database),
			fmt.Sprintf("--host=%s", opts.Host),
			fmt.Sprintf("--port=%d", opts.Port),
			fmt.Sprintf("--username=%s", opts.Username),
			fmt.Sprintf("--format=%s", pgDumpDefaultFormat),
		)
		return exec.Command(PGDumpCmd, options...), nil

	case "mysql":
		if !commandExist(MysqlDumpCmd) {
			return nil, ErrMySqlDumpNotFound
		}
		options := append(
			mysqlDumpStdOpts,
			"-h", opts.Host,
			"-P", strconv.Itoa(opts.Port),
			"-u", opts.Username,
			fmt.Sprintf("--password=%s", opts.Password),
			"--databases", opts.Database,
			"-r", outFile,
		)

		return exec.Command(MysqlDumpCmd, options...), nil

	default:
		return nil, ErrUnsupportedType
	}
}

func executeCommand(cmd *exec.Cmd) error {
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	go io.Copy(os.Stderr, stderr)
	go io.Copy(os.Stdout, stdout)

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func commandExist(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func newFileName(db string, dbType string) string {
	switch dbType {
	case "postgres":
		return fmt.Sprintf(`%v_%v.pgdump`, db, time.Now().Unix())
	case "mysql":
		return fmt.Sprintf(`%v_%v.sql`, db, time.Now().Unix())
	}
	return fmt.Sprintf(`%v_%v`, db, time.Now().Unix())
}
