package main

import (
	"fmt"
	"log"

	"automatic_text_processing/lab_2/internal/domain/corpus"
	"automatic_text_processing/lab_2/internal/infrastructure/config"
	"automatic_text_processing/lab_2/internal/infrastructure/filesystem"
	"automatic_text_processing/lab_2/internal/infrastructure/openai/client"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	folderReports, total, err := filesystem.FindReportsByFolder(cfg.InputDir)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d folders with %d total reports\n", len(folderReports), total)

	c, err := client.NewClient(cfg.AITunnelAPIKey, cfg.AITunnelBaseUrl)
	if err != nil {
		log.Fatal(err)
	}

	corporaByFolder := make(map[string][]*corpus.Corpus)
	totalProcessed := 0

	for _, folderReport := range folderReports {
		fmt.Printf("Processing folder: %s (%d reports)\n", folderReport.FolderPath, len(folderReport.Reports))

		corpora, err := client.GetChatCompletion(c, folderReport.Reports)
		if err != nil {
			log.Printf("Error processing folder %s: %v", folderReport.FolderPath, err)
			if len(corpora) > 0 {
				corporaByFolder[folderReport.FolderPath] = corpora
				totalProcessed += len(corpora)
				fmt.Printf("Processed %d reports in folder %s (with some errors)\n", len(corpora), folderReport.FolderPath)
			}
			continue
		}

		corporaByFolder[folderReport.FolderPath] = corpora
		totalProcessed += len(corpora)

		fmt.Printf("Processed %d reports in folder: %s\n", len(corpora), folderReport.FolderPath)
	}

	if err = filesystem.SaveCorporaByFolder(corporaByFolder); err != nil {
		log.Printf("Warning: Some save operations failed: %v", err)
		fmt.Printf("Processed %d total reports across %d folders (with some save errors)\n", totalProcessed, len(corporaByFolder))
	} else {
		fmt.Printf("Successfully processed %d total reports across %d folders\n", totalProcessed, len(corporaByFolder))
	}
}
