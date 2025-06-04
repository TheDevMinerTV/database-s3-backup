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

	CompressCmd         = "zstd"
	compressStdOpts     = []string{"--rm", "--no-progress", "--rsyncable", "-T0", "-11"}
	ErrCompressNotFound = errors.New("zstd not found")

	ErrUnsupportedType = errors.New("unsupported database type")
)

func RunDump(connectionOpts *connectionOptions) (string, error) {
	dumpFile := newFileName(connectionOpts.Database, connectionOpts.DbType, false)
	cmd, err := buildDumpCommand(connectionOpts, dumpFile)
	if err != nil {
		return "", err
	}
	if err = executeCommand(cmd); err != nil {
		return "", err
	}

	outFile := newFileName(connectionOpts.Database, connectionOpts.DbType, true)
	cmd, err = buildCompressCommand(dumpFile, outFile)
	if err != nil {
		return "", err
	}
	return outFile, executeCommand(cmd)
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
		mysqldumpCmd := "mysqldump"
		if !commandExist(mysqldumpCmd) {
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

		return exec.Command(mysqldumpCmd, options...), nil

	default:
		return nil, ErrUnsupportedType
	}
}

func buildCompressCommand(inFile string, outFile string) (*exec.Cmd, error) {
	if !commandExist(CompressCmd) {
		return nil, ErrCompressNotFound
	}
	options := append(
		compressStdOpts,
		inFile,
		"-o",
		outFile,
	)
	return exec.Command(CompressCmd, options...), nil
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

func newFileName(db string, dbType string, compress bool) string {
	ext := ""

	switch dbType {
	case "postgres":
		ext = ".pgdump"
	case "mysql":
		ext = ".sql"
	}

	if compress {
		ext = fmt.Sprintf("%s.zst", ext)
	}

	return fmt.Sprintf(`%v_%v%s`, db, time.Now().Unix(), ext)
}
