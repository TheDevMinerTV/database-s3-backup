package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type connectionOptions struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	DbType   string
}

var (
	PGDumpCmd           = "pg_dump"
	pgDumpStdOpts       = []string{"--no-owner", "--no-acl", "--clean", "--blobs", "-v"}
	pgDumpDefaultFormat = "c"

	ErrPgDumpNotFound    = errors.New("pg_dump not found")
	ErrMySqlDumpNotFound = errors.New("mysqldump not found")
)

func RunDump(connectionOpts *connectionOptions, outFile string) error {
	var cmd *exec.Cmd

	if connectionOpts.DbType == "postgres" {
		if !commandExist(PGDumpCmd) {
			return ErrPgDumpNotFound
		}

		options := append(
			pgDumpStdOpts,
			fmt.Sprintf(`-f%s`, outFile),
			fmt.Sprintf(`--dbname=%v`, connectionOpts.Database),
			fmt.Sprintf(`--host=%v`, connectionOpts.Host),
			fmt.Sprintf(`--port=%v`, connectionOpts.Port),
			fmt.Sprintf(`--username=%v`, connectionOpts.Username),
			fmt.Sprintf(`--format=%v`, pgDumpDefaultFormat),
		)
		cmd = exec.Command(PGDumpCmd, options...)
	} else if connectionOpts.DbType == "mariadb" {
		mysqldumpCmd := "mysqldump"
		if !commandExist(mysqldumpCmd) {
			return ErrMySqlDumpNotFound
		}

		options := []string{
			"-h", connectionOpts.Host,
			"-P", strconv.Itoa(connectionOpts.Port),
			"-u", connectionOpts.Username,
			fmt.Sprintf(`--password=%s`, connectionOpts.Password),
			"--databases", connectionOpts.Database,
			"-r", outFile,
		}
		cmd = exec.Command(mysqldumpCmd, options...)
	} else {
		return errors.New("unsupported database type")
	}

	cmd.Env = append(os.Environ(), fmt.Sprintf(`PGPASSWORD=%v`, connectionOpts.Password))

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err = cmd.Start(); err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	if err = cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func commandExist(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func newFileName(db string) string {
	return fmt.Sprintf(`%v_%v.pgdump`, db, time.Now().Unix())
}
