package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type connectionOptions struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

var (
	PGDumpCmd           = "pg_dump"
	pgDumpStdOpts       = []string{"--no-owner", "--no-acl", "--clean", "--blobs", "-v"}
	pgDumpDefaultFormat = "c"

	ErrPgDumpNotFound = errors.New("pg_dump not found")
)

func RunDump(pg *connectionOptions, outFile string) error {
	if !commandExist(PGDumpCmd) {
		return ErrPgDumpNotFound
	}

	options := append(
		pgDumpStdOpts,
		fmt.Sprintf(`-f%s`, outFile),
		fmt.Sprintf(`--dbname=%v`, pg.Database),
		fmt.Sprintf(`--host=%v`, pg.Host),
		fmt.Sprintf(`--port=%v`, pg.Port),
		fmt.Sprintf(`--username=%v`, pg.Username),
		fmt.Sprintf(`--format=%v`, pgDumpDefaultFormat),
	)

	cmd := exec.Command(PGDumpCmd, options...)
	cmd.Env = append(os.Environ(), fmt.Sprintf(`PGPASSWORD=%v`, pg.Password))

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
