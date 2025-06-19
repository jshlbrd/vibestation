// Package config provides structures for building configurations.
package config

import (
	"fmt"
	"strconv"
	"strings"
)

// Parser parses SUB sublang configuration
type Parser struct{}

// NewParser creates a new sublang parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses SUB sublang and returns a list of transforms
func (p *Parser) Parse(sublang string) ([]map[string]interface{}, error) {
	var transforms []map[string]interface{}
	lines := strings.Split(sublang, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		transform, err := p.parseLine(line)
		if err != nil {
			return nil, err
		}
		transforms = append(transforms, transform...)
	}

	return transforms, nil
}

// parseLine parses a single line and returns transforms
func (p *Parser) parseLine(line string) ([]map[string]interface{}, error) {
	// Handle direct assignment: $.target = $.source
	if p.isDirectAssignment(line) {
		return p.parseDirectAssignment(line)
	}

	// Handle assignment with function call: $.target = function(args)
	if p.isAssignmentWithFunction(line) {
		return p.parseAssignmentWithFunction(line)
	}

	// Handle function calls: function(args)
	if p.isFunctionCall(line) {
		return p.parseFunctionCall(line)
	}

	// If we reach here, the line is invalid
	return nil, fmt.Errorf("invalid SUB line format: %s", line)
}

// isDirectAssignment checks if line is a direct field assignment
func (p *Parser) isDirectAssignment(line string) bool {
	return strings.Contains(line, "=") && !strings.Contains(line, "(")
}

// parseDirectAssignment parses direct field assignments
func (p *Parser) parseDirectAssignment(line string) ([]map[string]interface{}, error) {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid assignment: %s", line)
	}

	target := strings.TrimSpace(parts[0])
	source := strings.TrimSpace(parts[1])
	if target == "" || source == "" {
		return nil, fmt.Errorf("invalid assignment: %s", line)
	}

	transform := map[string]interface{}{
		"id":     "assign",
		"type":   "assign",
		"source": source,
		"target": target,
	}

	return []map[string]interface{}{transform}, nil
}

// isAssignmentWithFunction checks if line is an assignment with a function call
func (p *Parser) isAssignmentWithFunction(line string) bool {
	if !strings.Contains(line, "=") || !strings.Contains(line, "(") {
		return false
	}
	eqIdx := strings.Index(line, "=")
	openParen := strings.Index(line, "(")
	return openParen > eqIdx
}

// parseAssignmentWithFunction parses assignments with function calls
func (p *Parser) parseAssignmentWithFunction(line string) ([]map[string]interface{}, error) {
	eqIdx := strings.Index(line, "=")
	target := strings.TrimSpace(line[:eqIdx])
	funcCall := strings.TrimSpace(line[eqIdx+1:])

	transforms, err := p.parseFunctionCallWithTarget(funcCall, target)
	if err != nil {
		return nil, fmt.Errorf("error parsing assignment with function: %v", err)
	}

	// Special handling for delete function in assignment context
	for _, transform := range transforms {
		if transform["type"] == "delete" {
			transform["target"] = target
		}
	}

	return transforms, nil
}

// isFunctionCall checks if line is a function call
func (p *Parser) isFunctionCall(line string) bool {
	return strings.Contains(line, "(") && strings.Contains(line, ")")
}

// parseFunctionCall parses function calls
func (p *Parser) parseFunctionCall(line string) ([]map[string]interface{}, error) {
	return p.parseFunctionCallWithTarget(line, "")
}

// parseFunctionCallWithTarget parses function calls with optional target override
func (p *Parser) parseFunctionCallWithTarget(line, target string) ([]map[string]interface{}, error) {
	openParen := strings.Index(line, "(")
	closeParen := strings.LastIndex(line, ")")
	if openParen == -1 || closeParen == -1 || closeParen <= openParen {
		return nil, fmt.Errorf("invalid function call syntax: %s", line)
	}

	funcName := strings.TrimSpace(line[:openParen])
	argsStr := line[openParen+1 : closeParen]

	args, err := p.parseArguments(argsStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing arguments: %v", err)
	}

	settings, err := p.buildTransformSettings(funcName, args)
	if err != nil {
		return nil, fmt.Errorf("error parsing settings: %v", err)
	}

	// Handle nested function calls
	var nestedTransforms []map[string]interface{}
	for key, value := range settings {
		if strings.HasPrefix(key, "nested_arg_") {
			if nestedFunc, ok := value.(string); ok {
				nested, err := p.Parse(nestedFunc)
				if err != nil {
					return nil, fmt.Errorf("error parsing nested function %s: %v", nestedFunc, err)
				}
				nestedTransforms = append(nestedTransforms, nested...)
				delete(settings, key)
				settings["source"] = "$.nested_output"
			}
		}
	}

	// Create the main transform
	transform := map[string]interface{}{
		"id":   settings["id"],
		"type": funcName,
	}

	// Add target if provided
	if target != "" {
		transform["target"] = target
	}

	// Add all settings except id (already set)
	for key, value := range settings {
		if key != "id" {
			transform[key] = value
		}
	}

	// Combine nested transforms with main transform
	result := append(nestedTransforms, transform)
	return result, nil
}

// parseArguments parses function arguments, including nested function calls
func (p *Parser) parseArguments(argsStr string) ([]string, error) {
	if strings.TrimSpace(argsStr) == "" {
		return []string{}, nil
	}

	var args []string
	var currentArg strings.Builder
	var inQuotes bool
	var quoteChar rune
	var parenDepth int

	for _, char := range argsStr {
		switch char {
		case '"', '\'':
			if !inQuotes && parenDepth == 0 {
				inQuotes = true
				quoteChar = char
				currentArg.WriteRune(char)
			} else if inQuotes && char == quoteChar {
				inQuotes = false
				currentArg.WriteRune(char)
			} else {
				currentArg.WriteRune(char)
			}
		case '(':
			if !inQuotes {
				parenDepth++
			}
			currentArg.WriteRune(char)
		case ')':
			if !inQuotes {
				parenDepth--
			}
			currentArg.WriteRune(char)
		case ',':
			if !inQuotes && parenDepth == 0 {
				arg := strings.TrimSpace(currentArg.String())
				if arg != "" {
					args = append(args, arg)
				}
				currentArg.Reset()
			} else {
				currentArg.WriteRune(char)
			}
		default:
			currentArg.WriteRune(char)
		}
	}

	if currentArg.Len() > 0 {
		arg := strings.TrimSpace(currentArg.String())
		if arg != "" {
			args = append(args, arg)
		}
	}

	return args, nil
}

// buildTransformSettings builds transform settings from arguments
func (p *Parser) buildTransformSettings(funcName string, args []string) (map[string]interface{}, error) {
	settings := make(map[string]interface{})
	nestedArgIndex := 0
	positionalIndex := 0

	for _, arg := range args {
		if err := p.processArgument(funcName, arg, settings, &nestedArgIndex, &positionalIndex); err != nil {
			return nil, err
		}
	}

	// Set default settings for known transforms
	p.setDefaultSettings(funcName, settings)

	return settings, nil
}

// processArgument processes a single argument
func (p *Parser) processArgument(funcName, arg string, settings map[string]interface{}, nestedArgIndex, positionalIndex *int) error {
	if p.isNamedArgument(arg) {
		return p.processNamedArgument(arg, settings, nestedArgIndex)
	} else if p.isLegacyNamedArgument(arg) {
		return p.processLegacyNamedArgument(arg, settings)
	} else if p.isNestedFunction(arg) {
		return p.processNestedFunction(arg, settings, nestedArgIndex)
	} else {
		return p.processPositionalArgument(funcName, arg, settings, positionalIndex)
	}
}

// isNamedArgument checks if argument is a named argument (key=value)
func (p *Parser) isNamedArgument(arg string) bool {
	return strings.Contains(arg, "=")
}

// processNamedArgument processes a named argument
func (p *Parser) processNamedArgument(arg string, settings map[string]interface{}, nestedArgIndex *int) error {
	kv := strings.SplitN(arg, "=", 2)
	if len(kv) != 2 {
		return fmt.Errorf("invalid named argument: %s", arg)
	}

	key := strings.TrimSpace(kv[0])
	value := strings.TrimSpace(kv[1])

	if p.isNestedFunction(value) {
		settings[fmt.Sprintf("nested_arg_%d", *nestedArgIndex)] = value
		*nestedArgIndex++
	} else {
		settings[key] = p.unquoteValue(value)
	}

	return nil
}

// isLegacyNamedArgument checks if argument is a legacy named argument (key:value)
func (p *Parser) isLegacyNamedArgument(arg string) bool {
	return strings.Contains(arg, ":")
}

// processLegacyNamedArgument processes a legacy named argument
func (p *Parser) processLegacyNamedArgument(arg string, settings map[string]interface{}) error {
	kv := strings.SplitN(arg, ":", 2)
	if len(kv) != 2 {
		return fmt.Errorf("invalid legacy named argument: %s", arg)
	}

	key := strings.TrimSpace(kv[0])
	value := strings.TrimSpace(kv[1])

	if value == "true" || value == "false" {
		settings[key] = value == "true"
	} else if num, err := strconv.Atoi(value); err == nil {
		settings[key] = num
	} else {
		settings[key] = strings.Trim(value, `"'`)
	}

	return nil
}

// isNestedFunction checks if argument is a nested function call
func (p *Parser) isNestedFunction(arg string) bool {
	return strings.Contains(arg, "(") && strings.Contains(arg, ")")
}

// processNestedFunction processes a nested function call
func (p *Parser) processNestedFunction(arg string, settings map[string]interface{}, nestedArgIndex *int) error {
	settings[fmt.Sprintf("nested_arg_%d", *nestedArgIndex)] = arg
	*nestedArgIndex++
	return nil
}

// processPositionalArgument processes a positional argument
func (p *Parser) processPositionalArgument(funcName, arg string, settings map[string]interface{}, positionalIndex *int) error {
	if p.isBuiltinTransform(funcName) {
		return p.processBuiltinPositionalArgument(funcName, arg, settings, positionalIndex)
	} else {
		return p.processCustomPositionalArgument(arg, settings, positionalIndex)
	}
}

// isBuiltinTransform checks if function name is a built-in transform
func (p *Parser) isBuiltinTransform(funcName string) bool {
	builtins := map[string]bool{
		"split_string":     true,
		"decompress_gzip":  true,
		"send_stdout":      true,
		"decode_base64":    true,
		"lowercase_string": true,
		"delete":           true,
	}
	return builtins[funcName]
}

// processBuiltinPositionalArgument processes positional arguments for built-in transforms
func (p *Parser) processBuiltinPositionalArgument(funcName, arg string, settings map[string]interface{}, positionalIndex *int) error {
	if *positionalIndex == 0 {
		if arg == "$" || strings.HasPrefix(arg, "$.") {
			settings["source"] = arg
		} else if p.isNestedFunction(arg) {
			settings["nested_arg_0"] = arg
		} else {
			return fmt.Errorf("first positional argument must be a JSON path (starting with $ or $.) or a function call (containing parentheses); got: %q", arg)
		}
	} else {
		return fmt.Errorf("only the first positional argument is allowed for built-in transforms; use named arguments for additional parameters (got: %q)", arg)
	}
	*positionalIndex++
	return nil
}

// processCustomPositionalArgument processes positional arguments for custom functions
func (p *Parser) processCustomPositionalArgument(arg string, settings map[string]interface{}, positionalIndex *int) error {
	settings[fmt.Sprintf("arg%d", *positionalIndex)] = p.unquoteValue(arg)
	*positionalIndex++
	return nil
}

// unquoteValue unquotes a value if it's quoted
func (p *Parser) unquoteValue(value string) interface{} {
	if len(value) > 1 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
		unq, err := strconv.Unquote(value)
		if err == nil {
			return unq
		}
	}
	return value
}

// setDefaultSettings sets default settings for known transforms
func (p *Parser) setDefaultSettings(funcName string, settings map[string]interface{}) {
	defaults := map[string]map[string]interface{}{
		"decompress_gzip": {
			"id": "decompress_gzip",
		},
		"split_string": {
			"separator": "\n",
			"id":        "split_string",
		},
		"send_stdout": {
			"id": "send_stdout",
		},
		"decode_base64": {
			"id":   "decode_base64",
			"type": "decode_base64",
		},
		"lowercase_string": {
			"id": "lowercase_string",
		},
		"delete": {
			"id": "delete",
		},
	}

	if defaults, ok := defaults[funcName]; ok {
		for key, value := range defaults {
			if _, exists := settings[key]; !exists {
				settings[key] = value
			}
		}
	}
}
