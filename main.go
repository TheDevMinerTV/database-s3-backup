package main

import (
	"context"
	"flag"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("WARNING: Error loading .env file: %s\n", err)
	}

	rawUrls := flag.String("urls", lookupEnvOrString("URLS", ""), "comma separated list of urls to backup, these must be in the format postgres://<user>:<password>@<host>[:<port>]/<dbname>")

	s3Endpoint := flag.String("s3-endpoint", lookupEnvOrString("S3_ENDPOINT", ""), "S3 endpoint")
	s3Bucket := flag.String("s3-bucket", lookupEnvOrString("S3_BUCKET", "postgres-backups"), "S3 bucket")
	s3AccessKey := flag.String("s3-access-key", lookupEnvOrString("S3_ACCESS_KEY", "minio"), "S3 access key")
	s3SecretKey := flag.String("s3-secret-key", lookupEnvOrString("S3_SECRET_KEY", "minioadmin"), "S3 secret key")
	dbType := flag.String("db-type", lookupEnvOrString("DB_TYPE", "postgres"), "Type of database: postgres or mariadb")

	interval := flag.Duration("interval", lookupEnvOrDuration("INTERVAL", 24*time.Hour), "How often to run the backup")

	flag.Parse()

	must("urls", rawUrls)
	must("s3-endpoint", s3Endpoint)
	must("s3-access-key", s3AccessKey)
	must("s3-secret-key", s3SecretKey)

	urls := make([]connectionOptions, len(strings.Split(*rawUrls, ",")))
	for i, rawUrl := range strings.Split(*rawUrls, ",") {
		parsedUrl, err := url.Parse(rawUrl)
		if err != nil {
			log.Fatalf("Failed to parse url %s: %s", rawUrl, err)
		}

		port := 5432
		rawPort := parsedUrl.Port()
		if rawPort != "" {
			port, err = strconv.Atoi(rawPort)
			if err != nil {
				log.Fatalf("Failed to parse port %s: %s", parsedUrl.Port(), err)
			}
		}

		password, exist := parsedUrl.User.Password()
		if !exist {
			log.Fatalf("Failed to parse password %s: %s", parsedUrl.User, err)
		}

		urls[i] = connectionOptions{
			Host:     parsedUrl.Hostname(),
			Port:     port,
			Database: strings.TrimPrefix(parsedUrl.Path, "/"),
			Username: parsedUrl.User.Username(),
			Password: password,
			DbType:   *dbType,
		}
	}
	if len(urls) == 0 {
		log.Fatalf("No URLs specified")
	}

	s3, err := minio.New(*s3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(*s3AccessKey, *s3SecretKey, ""),
		Secure: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	for {
		for _, u := range urls {
			log.Printf("Backing up %s", u.Database)

			file := newFileName(u.Database)

			if err = RunDump(&u, file); err != nil {
				log.Printf("WARNING: Failed to dump database: %s", err)
				continue
			}

			log.Printf("Uploading %s to %s", file, *s3Bucket)

			if _, err := s3.FPutObject(context.Background(), *s3Bucket, file, file, minio.PutObjectOptions{}); err != nil {
				log.Printf("WARNING: Failed to upload %s to %s: %s", file, *s3Bucket, err)
				continue
			}

			log.Printf("Removing %s", file)
			if err := os.Remove(file); err != nil {
				log.Printf("WARNING: Failed to remove %s: %s", file, err)
				continue
			}

			log.Printf("Done")
		}

		log.Printf("Sleeping for %s", *interval)
		time.Sleep(*interval)
	}
}

func lookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return defaultVal
}

func lookupEnvOrDuration(key string, defaultVal time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok {
		parsed, err := time.ParseDuration(val)
		if err != nil {
			log.Fatal(err)
		}

		return parsed
	}

	return defaultVal
}

func must(flag string, str *string) {
	if str == nil || *str == "" {
		log.Fatalf("Missing required flag: %s", flag)
	}
}
