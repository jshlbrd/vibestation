package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jshlbrd/vibestation"
	"github.com/jshlbrd/vibestation/config"
	"github.com/jshlbrd/vibestation/message"
	"gopkg.in/yaml.v3"
)

func main() {
	// Parse command line flags
	var (
		configFile = flag.String("config", "", "Configuration file (YAML or SUB)")
		inputFile  = flag.String("input", "", "Input file to process")
	)
	flag.Parse()

	// Validate required arguments
	if *configFile == "" {
		log.Fatal("Please provide a configuration file with -config flag")
	}
	if *inputFile == "" {
		log.Fatal("Please provide an input file with -input flag")
	}

	// Load configuration from file
	cfg, err := loadConfigFromFile(*configFile)
	if err != nil {
		log.Fatalf("Error loading configuration file: %v", err)
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

// loadYAMLConfig loads a YAML configuration file with embedded SUB sublang
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
	parser := config.NewParser()
	transformMaps, err := parser.Parse(yamlConfig.Transforms)
	if err != nil {
		return vibestation.Config{}, fmt.Errorf("failed to parse SUB script in YAML: %v", err)
	}

	// Convert map[string]interface{} to config.Config
	var transforms []config.Config
	for _, tmap := range transformMaps {
		transformType, ok := tmap["type"].(string)
		if !ok {
			return vibestation.Config{}, fmt.Errorf("transform missing type field")
		}

		// Remove type and id from settings, keep everything else
		settings := make(map[string]interface{})
		for k, v := range tmap {
			if k != "type" && k != "id" {
				settings[k] = v
			}
		}

		// Add id to settings if it exists
		if id, ok := tmap["id"].(string); ok {
			settings["id"] = id
		}

		transforms = append(transforms, config.Config{
			Type:     transformType,
			Settings: settings,
		})
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
	parser := config.NewParser()
	transformMaps, err := parser.Parse(string(content))
	if err != nil {
		return vibestation.Config{}, fmt.Errorf("failed to parse SUB config: %v", err)
	}

	// Convert map[string]interface{} to config.Config
	var transforms []config.Config
	for _, tmap := range transformMaps {
		transformType, ok := tmap["type"].(string)
		if !ok {
			return vibestation.Config{}, fmt.Errorf("transform missing type field")
		}

		// Remove type and id from settings, keep everything else
		settings := make(map[string]interface{})
		for k, v := range tmap {
			if k != "type" && k != "id" {
				settings[k] = v
			}
		}

		// Add id to settings if it exists
		if id, ok := tmap["id"].(string); ok {
			settings["id"] = id
		}

		transforms = append(transforms, config.Config{
			Type:     transformType,
			Settings: settings,
		})
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
