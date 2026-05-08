package seed

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strings"

	"eva_trainer/internal/models"
)

func SeedMaps(db *sql.DB, mapsDir string) error {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM maps`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	entries, err := os.ReadDir(mapsDir)
	if err != nil {
		log.Printf("No maps directory found at %s, skipping seed", mapsDir)
		return nil
	}

	mapRepo := models.NewMapRepository(db)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := extractMapName(entry.Name())
		if name == "" {
			continue
		}

		blobData, err := os.ReadFile(filepath.Join(mapsDir, entry.Name()))
		if err != nil {
			log.Printf("Warning: failed to read map file %s: %v", entry.Name(), err)
			continue
		}

		if _, err := mapRepo.Create(name, blobData); err != nil {
			log.Printf("Warning: failed to seed map %s (%s): %v", name, entry.Name(), err)
			continue
		}

		log.Printf("Seeded map: %s (%s)", name, entry.Name())
	}

	return nil
}

func extractMapName(filename string) string {
	if !strings.HasSuffix(strings.ToLower(filename), ".webp") {
		return ""
	}

	name := filename

	if strings.HasSuffix(strings.ToLower(name), "-map.webp") {
		name = name[:len(name)-len("-map.webp")]
	} else {
		name = name[:len(name)-len(".webp")]
	}

	name = strings.ReplaceAll(name, "-", " ")

	words := strings.Fields(name)
	if len(words) == 0 {
		return ""
	}

	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}

	return strings.Join(words, " ")
}
