package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/josh.liburdi/vibestation"
	"github.com/josh.liburdi/vibestation/config"
	"github.com/josh.liburdi/vibestation/message"
	"gopkg.in/yaml.v3"
)

func main() {
	// Parse command line flags
	var (
		configFile = flag.String("config", "", "Configuration file (YAML or SUB)")
		inputFile  = flag.String("input", "", "Input file to process")
		gzipMode   = flag.Bool("gzip", false, "Treat input as gzipped data")
	)
	flag.Parse()

	// Determine configuration source
	var cfg vibestation.Config
	var err error

	if *configFile != "" {
		// Load configuration from file (YAML or SUB)
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

// loadConfigFromFile loads a vibestation configuration from a file (YAML or SUB)
func loadConfigFromFile(filePath string) (vibestation.Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return vibestation.Config{}, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	// Determine file type based on extension
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".yaml", ".yml":
		return loadYAMLConfig(file)
	case ".sub":
		return loadSUBConfig(file)
	default:
		// Try to detect format by reading first few bytes
		return loadAutoDetectConfig(file)
	}
}

// loadYAMLConfig loads a YAML configuration file with embedded SUB DSL
func loadYAMLConfig(file *os.File) (vibestation.Config, error) {
	// Read the entire file content
	content, err := os.ReadFile(file.Name())
	if err != nil {
		return vibestation.Config{}, fmt.Errorf("failed to read YAML config file: %v", err)
	}

	// Parse YAML
	var yamlConfig struct {
		Transforms string `yaml:"transforms"`
	}

	if err := yaml.Unmarshal(content, &yamlConfig); err != nil {
		return vibestation.Config{}, fmt.Errorf("failed to parse YAML config: %v", err)
	}

	// Parse the embedded SUB script
	parser := config.NewSUBParser(yamlConfig.Transforms)
	transforms, err := parser.Parse()
	if err != nil {
		return vibestation.Config{}, fmt.Errorf("failed to parse SUB script in YAML: %v", err)
	}

	return vibestation.Config{
		Transforms: transforms,
	}, nil
}

// loadSUBConfig loads a SUB-style configuration file
func loadSUBConfig(file *os.File) (vibestation.Config, error) {
	// Read the entire file content
	content, err := os.ReadFile(file.Name())
	if err != nil {
		return vibestation.Config{}, fmt.Errorf("failed to read SUB config file: %v", err)
	}

	// Parse the SUB script
	parser := config.NewSUBParser(string(content))
	transforms, err := parser.Parse()
	if err != nil {
		return vibestation.Config{}, fmt.Errorf("failed to parse SUB config: %v", err)
	}

	return vibestation.Config{
		Transforms: transforms,
	}, nil
}

// loadAutoDetectConfig tries to auto-detect the configuration format
func loadAutoDetectConfig(file *os.File) (vibestation.Config, error) {
	// Read first few bytes to detect format
	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil {
		return vibestation.Config{}, fmt.Errorf("failed to read config file: %v", err)
	}

	content := string(buffer[:n])

	// Check if it looks like YAML (contains "transforms:" and "|")
	if strings.Contains(content, "transforms:") && strings.Contains(content, "|") {
		// Reset file position and try YAML
		file.Seek(0, 0)
		return loadYAMLConfig(file)
	}

	// Check if it looks like SUB (contains function calls or assignments)
	if strings.Contains(content, "(") || strings.Contains(content, "=") {
		// Reset file position and try SUB
		file.Seek(0, 0)
		return loadSUBConfig(file)
	}

	return vibestation.Config{}, fmt.Errorf("unable to detect configuration format")
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
