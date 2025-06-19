// Package config provides structures for building configurations.
package config

import (
	"fmt"
	"strconv"
	"strings"
)

// SUBParser parses a SUB script into a configuration
// (formerly DSLParser)
type SUBParser struct {
	lines []string
}

// NewSUBParser creates a new SUB parser
func NewSUBParser(sub string) *SUBParser {
	lines := strings.Split(sub, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			cleanLines = append(cleanLines, line)
		}
	}
	return &SUBParser{lines: cleanLines}
}

// Parse parses the SUB script into a list of transforms
func (p *SUBParser) Parse() ([]Config, error) {
	var configs []Config

	for _, line := range p.lines {
		config, err := p.parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("error parsing line '%s': %v", line, err)
		}
		configs = append(configs, config)
	}

	return configs, nil
}

// parseLine parses a single SUB line into a transform config
func (p *SUBParser) parseLine(line string) (Config, error) {
	// Check if it's an assignment (contains = and = comes before the first ()
	if strings.Contains(line, "=") {
		eqIdx := strings.Index(line, "=")
		openIdx := strings.Index(line, "(")
		if openIdx == -1 || eqIdx < openIdx {
			return p.parseAssignment(line)
		}
	}

	// Check if it's a function call (contains parentheses)
	if strings.Contains(line, "(") {
		return p.parseFunctionCall(line)
	}

	return Config{}, fmt.Errorf("invalid SUB line format: %s", line)
}

// parseFunctionCall parses a function call like "function_name!(arg1, arg2)" or with named arguments
func (p *SUBParser) parseFunctionCall(line string) (Config, error) {
	// Find the first '(' and last ')'
	openIdx := strings.Index(line, "(")
	closeIdx := strings.LastIndex(line, ")")
	if openIdx == -1 || closeIdx == -1 || closeIdx < openIdx {
		return Config{}, fmt.Errorf("invalid function call format: %s", line)
	}

	funcName := strings.TrimSpace(line[:openIdx])
	argsStr := line[openIdx+1 : closeIdx]

	// Ensure nothing but whitespace after the closing parenthesis
	if strings.TrimSpace(line[closeIdx+1:]) != "" {
		return Config{}, fmt.Errorf("invalid function call format: %s", line)
	}

	// Parse arguments (robustly, including named/positional)
	args, err := p.parseArguments(argsStr)
	if err != nil {
		return Config{}, err
	}

	// Convert function name to transform type
	transformType := p.functionToTransformType(funcName)

	// Convert arguments to settings
	settings := p.argsToSettings(funcName, args)

	return Config{
		Type:     transformType,
		Settings: settings,
	}, nil
}

// parseAssignment parses an assignment like "$.foo = lowercase($.bar)"
func (p *SUBParser) parseAssignment(line string) (Config, error) {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return Config{}, fmt.Errorf("invalid assignment format: %s", line)
	}

	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])

	// Parse the right side as a function call
	config, err := p.parseFunctionCall(right)
	if err != nil {
		return Config{}, err
	}

	// Always override any target in the function call with the assignment target
	if config.Settings == nil {
		config.Settings = make(map[string]interface{})
	}
	config.Settings["target"] = left

	return config, nil
}

// parseArguments parses function arguments
func (p *SUBParser) parseArguments(argsStr string) ([]string, error) {
	if strings.TrimSpace(argsStr) == "" {
		return []string{}, nil
	}
	var args []string
	var currentArg strings.Builder
	var inQuotes bool
	var quoteChar rune
	for i, char := range argsStr {
		switch char {
		case '"', '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
				currentArg.WriteRune(char)
			} else if char == quoteChar {
				inQuotes = false
				currentArg.WriteRune(char)
			} else {
				currentArg.WriteRune(char)
			}
		case ',':
			if !inQuotes {
				args = append(args, strings.TrimSpace(currentArg.String()))
				currentArg.Reset()
			} else {
				currentArg.WriteRune(char)
			}
		default:
			currentArg.WriteRune(char)
		}
		// If at the end, flush the last argument
		if i == len(argsStr)-1 && currentArg.Len() > 0 {
			args = append(args, strings.TrimSpace(currentArg.String()))
		}
	}
	// Unescape quoted arguments
	for i, arg := range args {
		if len(arg) > 1 && (arg[0] == '"' && arg[len(arg)-1] == '"' || arg[0] == '\'' && arg[len(arg)-1] == '\'') {
			unq, err := strconv.Unquote(arg)
			if err == nil {
				args[i] = unq
			}
		}
	}
	return args, nil
}

// functionToTransformType converts a function name to a transform type
func (p *SUBParser) functionToTransformType(funcName string) string {
	switch funcName {
	case "decompress_gzip":
		return "decompress_gzip"
	case "split_string":
		return "split_string"
	case "send_stdout":
		return "send_stdout"
	case "decode_base64":
		return "decode_base64"
	case "lowercase_string":
		return "lowercase_string"
	default:
		return funcName
	}
}

// argsToSettings converts function arguments to settings
func (p *SUBParser) argsToSettings(funcName string, args []string) map[string]interface{} {
	settings := make(map[string]interface{})

	for _, arg := range args {
		if strings.Contains(arg, "=") {
			// Named argument: key=value
			kv := strings.SplitN(arg, "=", 2)
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				value := strings.TrimSpace(kv[1])
				// Remove quotes from value if present
				if len(value) > 1 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
					unq, err := strconv.Unquote(value)
					if err == nil {
						value = unq
					}
				}
				settings[key] = value
			}
		} else if strings.Contains(arg, ":") {
			// Legacy: key:value
			kv := strings.SplitN(arg, ":", 2)
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				value := strings.TrimSpace(kv[1])
				if value == "true" || value == "false" {
					settings[key] = value == "true"
				} else if num, err := strconv.Atoi(value); err == nil {
					settings[key] = num
				} else {
					settings[key] = strings.Trim(value, `"'`)
				}
			}
		} else {
			panic("Positional arguments are not supported. Use only named arguments (key=value or key:value). Argument: " + arg)
		}
	}

	// For known transforms, set default id if not already set by named args
	switch funcName {
	case "decompress_gzip":
		if _, ok := settings["id"]; !ok {
			settings["id"] = "decompress_gzip"
		}
	case "split_string":
		if _, ok := settings["separator"]; !ok {
			settings["separator"] = "\n" // default
		}
		if _, ok := settings["id"]; !ok {
			settings["id"] = "split_string"
		}
	case "send_stdout":
		if _, ok := settings["id"]; !ok {
			settings["id"] = "send_stdout"
		}
	case "decode_base64":
		if _, ok := settings["id"]; !ok {
			settings["id"] = "decode_base64"
		}
		settings["type"] = "decode_base64"
	case "lowercase_string":
		if _, ok := settings["id"]; !ok {
			settings["id"] = "lowercase_string"
		}
	}

	return settings
}
