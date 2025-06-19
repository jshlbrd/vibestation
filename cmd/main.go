package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/josh.liburdi/vibestation"
	"github.com/josh.liburdi/vibestation/config"
	"github.com/josh.liburdi/vibestation/message"
)

func main() {
	// Parse command line flags
	var (
		configFile = flag.String("config", "", "JSON configuration file")
		inputFile  = flag.String("input", "", "Input file to process")
		gzipMode   = flag.Bool("gzip", false, "Treat input as gzipped data")
	)
	flag.Parse()

	// Determine configuration source
	var cfg vibestation.Config
	var err error

	if *configFile != "" {
		// Load configuration from JSON file
		cfg, err = loadConfigFromFile(*configFile)
		if err != nil {
			log.Fatalf("Error loading configuration file: %v", err)
		}
	} else {
		// Use default configuration based on command line flags
		cfg = createDefaultConfig(*gzipMode)
	}

	// Validate input file
	if *inputFile == "" {
		log.Fatal("Please provide an input file with -input flag")
	}

	// Read the input file
	data, err := os.ReadFile(*inputFile)
	if err != nil {
		log.Fatalf("Error reading input file: %v", err)
	}

	// Create vibestation instance
	ctx := context.Background()
	vibe, err := vibestation.New(ctx, cfg)
	if err != nil {
		log.Fatalf("Error creating vibestation: %v", err)
	}

	// Create initial message with file data
	msg := message.New().SetData(data)

	// Process the message through the transform pipeline
	results, err := vibe.Transform(ctx, msg)
	if err != nil {
		log.Fatalf("Error processing message: %v", err)
	}

	fmt.Printf("Processed %d messages\n", len(results))
}

// loadConfigFromFile loads a vibestation configuration from a JSON file
func loadConfigFromFile(filePath string) (vibestation.Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return vibestation.Config{}, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	var cfg vibestation.Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return vibestation.Config{}, fmt.Errorf("failed to decode config file: %v", err)
	}

	return cfg, nil
}

// createDefaultConfig creates a default configuration based on command line flags
func createDefaultConfig(gzipMode bool) vibestation.Config {
	cfg := vibestation.Config{
		Transforms: []config.Config{},
	}

	// Add gzip decompression if requested
	if gzipMode {
		cfg.Transforms = append(cfg.Transforms, config.Config{
			Type: "format_from_gzip",
			Settings: map[string]interface{}{
				"id": "decompress_gzip",
			},
		})
	}

	// Add string split transform to split into lines
	cfg.Transforms = append(cfg.Transforms, config.Config{
		Type: "string_split",
		Settings: map[string]interface{}{
			"separator": "\n",
			"id":        "split_lines",
		},
	})

	// Add stdout transform to print results
	cfg.Transforms = append(cfg.Transforms, config.Config{
		Type: "send_stdout",
		Settings: map[string]interface{}{
			"id": "print_to_console",
		},
	})

	return cfg
}
