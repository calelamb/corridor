package cli

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

func initDev(filename string, entropy io.Reader, out io.Writer) error {
	bytes := make([]byte, 128)
	if _, err := io.ReadFull(entropy, bytes); err != nil {
		return fmt.Errorf("generate development credentials: %w", err)
	}
	values := []string{"POSTGRES_PASSWORD=" + hex.EncodeToString(bytes[:32]), "CORRIDOR_DB_PASSWORD=" + hex.EncodeToString(bytes[32:64]), "MINIO_ROOT_USER=corridor-admin", "MINIO_ROOT_PASSWORD=" + hex.EncodeToString(bytes[64:96]), "CORRIDOR_S3_ACCESS_KEY=corridor-app", "CORRIDOR_S3_SECRET_KEY=" + hex.EncodeToString(bytes[96:]), "CORRIDOR_S3_BUCKET=corridor-raw"}
	// #nosec G304 -- production caller supplies the fixed local .env path; O_EXCL refuses existing files/symlinks. Tests supply isolated temporary paths.
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("create development environment: %w", err)
	}
	_, writeErr := io.WriteString(file, strings.Join(values, "\n")+"\n")
	closeErr := file.Close()
	if writeErr != nil {
		return fmt.Errorf("write development environment: %w", writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close development environment: %w", closeErr)
	}
	if _, err := fmt.Fprintln(out, "Created", filename); err != nil {
		return fmt.Errorf("report environment creation: %w", err)
	}
	return nil
}
