// file: service/utils/file.go
package utils

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"vtarchitect/config"
)

// CheckAndConvertCSV checks for a CSV file in the shared directory, converts it to YAML, and deletes the source CSV.
func CheckAndConvertCSV() error {
	yamlPath := filepath.Join(config.SharedDir, "architect.yaml")

	log.Println("STARTUP: Checking for CSV file in shared directory...")
	files, err := filepath.Glob(filepath.Join(config.SharedDir, "*.csv"))
	if err != nil {
		log.Printf("[ERROR] Could not search for CSV: %v", err)
		return err
	}
	if len(files) == 0 {
		log.Println("STARTUP: No CSV file found. Skipping conversion.")
		return nil // No CSV to process
	}

	csvPath := files[0] // Use the first CSV found
	log.Printf("STARTUP: Found CSV: %s. Converting to YAML...", csvPath)
	cmd := exec.Command("go", "run", filepath.Join(config.SharedDir, "csv-to-yaml.go"), csvPath, yamlPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("[ERROR] CSV to YAML conversion failed: %v", err)
		return err
	}

	log.Printf("STARTUP: Conversion complete. Deleting CSV: %s", csvPath)
	if err := os.Remove(csvPath); err != nil {
		log.Printf("[ERROR] Could not delete CSV: %v", err)
		return err
	}
	log.Printf("STARTUP: Converted %s to %s and deleted the CSV.", csvPath, yamlPath)

	return nil
}
