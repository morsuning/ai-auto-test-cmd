# ATC - API Automation Testing Command Line Tool

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey.svg)](#installation)

> **Language**: [中文文档](README_zh.md) | English

ATC (API Test Command) is a powerful API automation testing command-line tool designed to simplify API testing workflows. It supports intelligent test case generation, batch interface testing, multiple data format processing, and more.

## ✨ Core Features

### 🎯 Intelligent Test Case Generation
- **Local Generation**: Rapidly generate diverse test cases based on positive examples
- **Smart Constraint System**: Supports 11 constraint types, generating realistic and valid Chinese test data
- **Multi-format Support**: Supports JSON and XML input/output formats
- **Data Variation Rules**: 50% fluctuation for numbers, 10% length change for strings

### 🚀 Batch Interface Testing
- **Multiple HTTP Methods**: Supports POST, GET, and other HTTP request methods
- **Multiple Authentication**: Bearer Token, Basic Auth, API Key, etc.
- **Custom Headers**: Flexible addition of HTTP header information
- **Concurrent Execution**: Improves testing execution efficiency
- **Result Export**: Supports CSV format result export

### 🛡️ Configuration Validation
- **Format Validation**: Constraint configuration file integrity checking
- **Content Validation**: Data type and range reasonableness validation
- **Error Reporting**: Detailed error information and precise location identification

## 📦 Installation

### Pre-compiled Binary Files

Download the pre-compiled version suitable for your system from the [Releases](https://github.com/morsuning/ai-auto-test-cmd/releases) page:

- **Windows (amd64)**: `atc-windows-amd64.exe`
- **macOS (ARM)**: `atc-darwin-arm64`
- **Linux (ARM)**: `atc-linux-arm64`
- **Linux (amd64)**: `atc-linux-amd64`

### Compile from Source

```bash
# Clone repository
git clone https://github.com/morsuning/ai-auto-test-cmd.git
cd ai-auto-test-cmd

# Compile
go build -o atc

# Or use build scripts
# Windows
build\build.bat

# macOS/Linux
bash build/build.sh
```

## 🚀 Quick Start

### 1. Generate Test Cases

```bash
# Generate test cases from JSON positive example
atc local-gen '{"name":"John","age":25,"email":"test@example.com"}' --json --count 10

# Generate test cases from XML positive example
atc local-gen '<user><name>John</name><age>25</age></user>' --xml --count 5

# Read positive example from file and generate
atc local-gen -f examples/json_example.json --json --count 20

# Generate using smart constraint system
atc local-gen -f examples/json_example.json --json --count 10 --constraints
```

### 2. Execute Interface Testing

```bash
# POST request sending JSON data
atc request -u https://api.example.com/users -m post -f testcases.csv --json

# GET request
atc request -u https://api.example.com/users -m get -f testcases.csv --json

# Use Bearer Token authentication
atc request -u https://api.example.com/users -m post -f testcases.csv --json --auth-bearer "your_token"

# Add custom request headers
atc request -u https://api.example.com/users -m post -f testcases.csv --json --header "X-API-Key: key123"

# Save test results
atc request -u https://api.example.com/users -m post -f testcases.csv --json -s results.csv
```

### 3. Validate Constraint Configuration

```bash
# Validate default configuration file
atc validate

# Validate specified configuration file
atc validate my-constraints.toml

# Show detailed validation information
atc validate --verbose
```

## 📋 Command Reference

### `local-gen` - Local Test Case Generation

Generate diverse test cases based on positive examples.

```bash
atc local-gen [positive_example] [flags]
```

**Main Parameters:**
- `--json`: Specify JSON format
- `--xml`: Specify XML format
- `--count, -c`: Generation count (default 5)
- `--file, -f`: Read positive example from file
- `--output, -o`: Output file path
- `--constraints`: Enable smart constraint system
- `--constraints-file`: Specify constraint configuration file

**Examples:**
```bash
# Generate 10 JSON test cases
atc local-gen '{"name":"John","age":25}' --json -c 10

# Use constraint system to generate realistic data
atc local-gen '{"name":"张三","phone":"13800138000"}' --json --constraints -c 5

# Generate from file and save to specified location
atc local-gen -f input.json --json -c 20 -o testcases.csv
```

### `request` - Batch Interface Testing

Execute HTTP requests in batches based on CSV test case files.

```bash
atc request -u [URL] -m [METHOD] -f [CSV_FILE] [flags]
```

**Main Parameters:**
- `--url, -u`: Target interface URL (required)
- `--method, -m`: HTTP method (post/get)
- `--file, -f`: CSV test case file (required)
- `--json`: JSON format request body
- `--xml`: XML format request body
- `--save, -s`: Save results to file
- `--timeout`: Request timeout (default 30 seconds)
- `--debug`: Enable debug mode

**Authentication Parameters:**
- `--auth-bearer`: Bearer Token authentication
- `--auth-basic`: Basic Auth authentication (format: username:password)
- `--header`: Custom HTTP headers (can be used multiple times)

**Examples:**
```bash
# Basic POST request
atc request -u https://api.example.com/users -m post -f users.csv --json

# Use authentication and custom headers
atc request -u https://api.example.com/users -m post -f users.csv --json \
  --auth-bearer "eyJhbGciOiJIUzI1NiIs..." \
  --header "X-Request-ID: 12345" \
  --header "X-Client-Version: 1.0"

# GET request (automatically converts to query parameters)
atc request -u https://api.example.com/users -m get -f users.csv --json

# Enable debug mode and save results
atc request -u https://api.example.com/users -m post -f users.csv --json --debug -s results.csv
```

### `validate` - Configuration Validation

Validate the format and content correctness of constraint configuration files.

```bash
atc validate [config_file] [flags]
```

**Main Parameters:**
- `--verbose, -v`: Show detailed validation information

**Examples:**
```bash
# Validate default configuration
atc validate

# Validate specified configuration file
atc validate my-constraints.toml

# Show detailed information
atc validate --verbose
```

## 🎯 Smart Constraint System

The smart constraint system is ATC's core feature, capable of automatically identifying field names and generating realistic and valid test data.

### Supported Constraint Types

| Constraint Type | Description | Example Field Names | Generation Example |
|----------------|-------------|--------------------|-----------------|
| `date` | Date type | date, time, created_at | 20230101 |
| `chinese_name` | Chinese name | name, username, author | 周桂兰 |
| `phone` | Phone number | phone, mobile, tel | 17234495798 |
| `email` | Email address | email, mail | test473@189.cn |
| `chinese_address` | Chinese address | address, location | 武汉市武昌区中南路99号 |
| `id_card` | ID card number | id_card, identity | 500101198909148195 |
| `integer` | Integer type | age, count, quantity | 64 |
| `float` | Float type | price, amount, rate | 161782.59 |

### Configuration File Example

Create a `constraints.toml` file:

```toml
# Date field constraint
[date]
type = "date"
format = "20060102"  # Go time format
min_date = "20200101"
max_date = "20301231"
description = "Date field, format YYYYMMDD"

# Name field constraint
[name]
type = "chinese_name"
description = "Chinese name"

# Age field constraint
[age]
type = "integer"
min = 1
max = 120
description = "Age range 1-120"

# Price field constraint
[price]
type = "float"
min = 0.01
max = 999999.99
precision = 2
description = "Price field, 2 decimal places"

# Built-in datasets
[builtin_data]
first_names = ["张", "王", "李", "赵", "刘"]
last_names = ["伟", "芳", "娜", "敏", "静"]
addresses = ["北京市朝阳区建国门外大街1号", "上海市浦东新区陆家嘴环路1000号"]
email_domains = ["qq.com", "163.com", "126.com", "gmail.com"]
```

### Generation Effect Comparison

**Before using constraint system (random variation):**
```json
{"date":"27388202","name":"p","age":18,"phone":"11684695289","email":"1haDgsai8xOmpyU.C0m","price":122,"address":"K京w区"}
```

**After using constraint system (smart constraints):**
```json
{"date":"20230101","name":"周桂兰","age":64,"phone":"17234495798","email":"test473@189.cn","price":161782.59,"address":"武汉市武昌区中南路99号"}
```

## 📁 CSV File Format

### Generation Phase Format

- **JSON Format**: Single-column CSV with column name "JSON", one JSON string per row
- **XML Format**: Single-column CSV with column name "XML", one XML string per row

### Testing Phase Format Recognition

- **Single-column JSON**: Column name "JSON", directly uses JSON content as request body
- **Single-column XML**: Column name "XML", directly uses XML content as request body
- **Multi-column Format**: Combines column data into JSON object
- **GET Requests**: Only supports JSON format, automatically converts to query parameters

## 🔧 Advanced Features

### Debug Mode

Use the `--debug` parameter to enable detailed debug output:

```bash
atc request -u https://api.example.com/users -m post -f users.csv --json --debug
```

Debug mode displays:
- Detailed information for each request (URL, method, headers, request body)
- Complete response information (status code, response time, response body)
- Formatted JSON response content

### Concurrency Control

The system automatically adjusts concurrency based on the number of test cases, improving execution efficiency while avoiding excessive pressure on target servers.

### Error Handling

- Detailed error information prompts
- Batch error reporting
- Precise error location identification
- User-friendly prompts

## 📊 Example Project

The `examples/` directory contains complete usage examples:

- `constraints.toml`: Constraint configuration example
- `json_example.json`: JSON positive example input
- `xml_example.xml`: XML positive example input
- `input.xml`: Complex XML structure example

## 🤝 Contributing

Welcome to submit Issues and Pull Requests!

1. Fork this repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Related Links

- [Project Requirements Document](项目需求文档.md)
- [Issue Feedback](https://github.com/morsuning/ai-auto-test-cmd/issues)
- [Feature Requests](https://github.com/morsuning/ai-auto-test-cmd/issues/new?template=feature_request.md)
- [中文文档](README_zh.md)

---

**ATC** - Making API testing simpler, smarter, and more efficient!