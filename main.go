package main

import (
	"context"
	"flag"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("WARNING: Error loading .env file: %s\n", err)
	}

	postgresHost := flag.String("postgres-host", lookupEnvOrString("POSTGRES_HOST", "localhost"), "connectionOptions host")
	postgresPort := flag.Int("postgres-port", lookupEnvOrInt("POSTGRES_PORT", 5432), "connectionOptions port")
	postgresUser := flag.String("postgres-user", lookupEnvOrString("POSTGRES_USER", "postgres"), "connectionOptions user")
	postgresPassword := flag.String("postgres-password", lookupEnvOrString("POSTGRES_PASSWORD", "postgres"), "connectionOptions password")
	postgresDB := flag.String("postgres-db", lookupEnvOrString("POSTGRES_DB", "postgres"), "connectionOptions database")

	s3Endpoint := flag.String("s3-endpoint", lookupEnvOrString("S3_ENDPOINT", ""), "S3 endpoint")
	s3Bucket := flag.String("s3-bucket", lookupEnvOrString("S3_BUCKET", "postgres-backups"), "S3 bucket")
	s3AccessKey := flag.String("s3-access-key", lookupEnvOrString("S3_ACCESS_KEY", "minio"), "S3 access key")
	s3SecretKey := flag.String("s3-secret-key", lookupEnvOrString("S3_SECRET_KEY", "minioadmin"), "S3 secret key")

	every := flag.Duration("every", lookupEnvOrDuration("EVERY", 24*time.Hour), "How often to run the backup")

	flag.Parse()

	must("postgres-host", postgresHost)
	must("s3-endpoint", s3Endpoint)
	must("s3-access-key", s3AccessKey)
	must("s3-secret-key", s3SecretKey)

	s3, err := minio.New(*s3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(*s3AccessKey, *s3SecretKey, ""),
		Secure: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	for {
		file := newFileName(*postgresDB)
		err = RunDump(&connectionOptions{
			Host:     *postgresHost,
			Port:     *postgresPort,
			Database: *postgresDB,
			Username: *postgresUser,
			Password: *postgresPassword,
		}, file)
		if err != nil {
			log.Printf("WARNING: Failed to dump database: %s", err)
		}

		log.Printf("Uploading %s to %s", file, *s3Bucket)

		if _, err := s3.FPutObject(context.Background(), *s3Bucket, file, file, minio.PutObjectOptions{}); err != nil {
			log.Printf("WARNING: Failed to upload %s to %s: %s", file, *s3Bucket, err)
		}

		log.Printf("Removing %s", file)
		if err := os.Remove(file); err != nil {
			log.Printf("WARNING: Failed to remove %s: %s", file, err)
		}

		log.Printf("Sleeping for %s", *every)
		time.Sleep(*every)
	}
}

func lookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return defaultVal
}

func lookupEnvOrInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		parsed, err := strconv.Atoi(val)
		if err != nil {
			log.Fatal(err)
		}

		return parsed
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
