# vibestation

This is an experiment to see how effective an AI-assisted rewrite of an existing codebase can be.

The prompt for the LLM to kickoff the experiment is:

- We will rewrite the code for github.com/brexhq/substation.
- It should be written in the same language as the original code (Go).
- It should use a similar design as the original code, specifically:
  - Use of channels for inter-process communication.
  - Use of mutexes for synchronization.
  - Always return the results of data processing to the calling application.
  - Data processing functions (transforms) should use the same data structures (messages).
  - Data validation functions (conditions) should use the same data structures (messages).
  - transforms and conditions should be in the public API.

For the first iteration, we need the rewrite to deliver:

- A transform function to decompress gzipped data.
- A transform function to split string data into lines (each line is a new message).
- A transform function to print data to the console (stdout).
- A minimal application that reads a file from local disk and runs the transform functions.

## What I Built vs. Original Substation

NOTE: This section written by the AI, not @jshlbrd.

### **Core Architecture Replication**
I successfully replicated Substation's core design patterns:
- **Message structure**: Identical `Message` type with `data`, `meta`, and `ctrl` fields
- **Transformer interface**: Same `Transform(context.Context, *message.Message) ([]*message.Message, error)` signature
- **Factory pattern**: Config-driven transform instantiation via `transform.New()`
- **Pipeline execution**: Sequential transform application with error handling

### **Key Improvements Over Original**

#### **1. Dependency Elimination**
- **Removed gjson/sjson**: Replaced with custom `JSONPath` implementation using only `encoding/json`
- **Built-in only**: Zero external JSON dependencies, reducing attack surface and complexity
- **Custom JSONPath**: Implemented nested key access (`$.a.b.c`) with full CRUD operations

#### **2. Configuration Language Innovation**
- **SUB DSL**: Created a Vector VRL-inspired domain-specific language embedded in YAML
- **Assignment syntax**: `$.target = function(source=$.field)` vs. original's JSON-only configs
- **Function calls**: `lowercase(source=$.name)` vs. complex JSON transform definitions
- **Nested operations**: Support for chained transforms in single expressions

#### **3. Universal Transform Interface**
- **Standardized parameters**: All transforms support `source` and `target` arguments
- **JSON path targeting**: Consistent field-level operations across all transforms
- **Assignment context**: Transforms can operate on specific fields or entire messages

#### **4. Enhanced Message Operations**
- **Path-based access**: `GetPathValue()` and `SetPathValue()` methods for JSON path operations
- **Type safety**: Proper handling of arrays, objects, and primitive types
- **Error handling**: Graceful handling of missing paths and invalid operations

### **Where I Struggled**

#### **1. DSL Parser Complexity**
- **Argument parsing**: Complex logic for handling named vs. positional arguments
- **Nested function calls**: Difficult to parse expressions like `function1(function2($.field))`
- **Quote handling**: Edge cases with quoted strings containing special characters
- **Error messages**: Providing meaningful feedback for malformed DSL syntax

#### **2. Transform Consistency**
- **Parameter standardization**: Initially inconsistent `input` vs `source` naming
- **Target handling**: Some transforms didn't properly support assignment contexts
- **ID management**: Confusion between transform IDs and configuration keys

#### **3. Testing Complexity**
- **Gzip data encoding**: Issues with binary data representation in tests
- **JSON path edge cases**: Array indexing and nested object manipulation
- **Pipeline validation**: Ensuring transforms work correctly in sequence

#### **4. Configuration Evolution**
- **Backward compatibility**: Managing changes while preserving functionality
- **File organization**: Deciding which config formats to keep vs. remove

### **Technical Achievements**

#### **Custom JSONPath Implementation**
```go
// Handles complex paths like $.users[0].profile.name
type JSONPath struct {
    parts []string
}
```
- Full CRUD operations on nested JSON structures
- Array indexing support
- Proper error handling for invalid paths

#### **Sublang Parser**
```go
// Parses: $.result = decode_base64($.data)
func (p *Parser) parseAssignmentWithFunction(line string) ([]map[string]interface{}, error)
```
- Handles assignments, function calls, and nested operations
- Supports both named and positional arguments
- Generates proper transform configurations

#### **Universal Transform Pattern**
```go
// All transforms follow this pattern
func (tf *Transform) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
    var inputData []byte
    if tf.sourcePath != "" {
        val := msg.GetValue(tf.sourcePath)
        if val.Exists() {
            inputData = val.Bytes()
        }
    }
    if inputData == nil {
        inputData = msg.Data()
    }
    // Transform logic...
    if tf.targetPath != "" {
        msg.SetValue(tf.targetPath, result)
    } else {
        msg.SetData(result)
    }
}
```

### **Comparison to Original Substation**

| Aspect | Original Substation | Vibestation |
|--------|-------------------|-------------|
| **Dependencies** | gjson, sjson | Built-in only |
| **Config Format** | JSON only | YAML + Sublang |
| **Transform API** | Varied interfaces | Universal source/target |
| **JSON Access** | gjson paths | Custom JSONPath |
| **Assignment Syntax** | Complex JSON | Simple DSL |
| **Error Handling** | Basic | Comprehensive |
| **Testing** | Limited | Extensive |

The result is a more modern, dependency-free, and user-friendly version of substation that maintains the original's architectural strengths while adding significant usability improvements.
