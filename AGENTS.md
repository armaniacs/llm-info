# LLM-Info Agents

This document describes the various agents/components that make up the `llm-info` CLI tool for visualizing LLM gateway information.

read `DEVELOPMENT.md` for coding, testing and development.

## Core Agents

### 1. API Client Agent
**Purpose**: Handles communication with LLM gateway endpoints

**Responsibilities**:
- Send requests to `/model/info` endpoint (LiteLLM compatible mode)
- Send requests to `/v1/models` endpoint (Standard compatible mode)
- Implement automatic fallback from `/model/info` to `/v1/models`
- Handle authentication via API keys
- Manage request timeouts

**Technical Details**:
- Uses Go's `net/http` package for HTTP requests
- Implements JSON parsing with `encoding/json`
- Supports configurable base URLs and API keys

### 2. Configuration Agent
**Purpose**: Manages application settings and user preferences

**Responsibilities**:
- Load configuration from `~/.config/llm-info/llm-info.yaml`
- Parse command-line arguments for base URL and API key
- Read environment variables as alternative configuration source
- Provide default values when configuration is missing

**Configuration Structure**:
```yaml
gateways:
  - name: "default"
    url: "https://api.example.com"
    api_key: "your-api-key"
  - name: "alternative"
    url: "https://api2.example.com"
    api_key: "another-api-key"
```

### 3. Data Processing Agent
**Purpose**: Processes and normalizes model information from different API formats

**Responsibilities**:
- Parse responses from both `/model/info` and `/v1/models` endpoints
- Normalize data into a consistent internal format
- Extract relevant metadata (max_tokens, input_cost, mode, etc.)
- Handle missing or incomplete data gracefully

**Data Model**:
```go
type ModelInfo struct {
    ID         string
    Name       string
    MaxTokens  int
    Mode       string
    InputCost  float64
    // Additional metadata fields as needed
}
```

### 4. Visualization Agent
**Purpose**: Renders model information in a readable table format

**Responsibilities**:
- Generate formatted tables using the `tablewriter` library
- Dynamically adjust columns based on available data
- Ensure proper alignment and spacing for readability
- Handle terminal width constraints

**Output Format**:
```
+----------------------+------------+--------+-------------+
|      MODEL NAME      | MAX TOKENS |  MODE  | INPUT COST  |
+----------------------+------------+--------+-------------+
| gpt-4                |       8192 | chat   |     0.00003 |
| claude-3-opus        |     200000 | chat   |     0.000015|
| gemini-1.5-pro       |    1000000 | chat   |           0 |
+----------------------+------------+--------+-------------+
```

### 5. CLI Interface Agent
**Purpose**: Provides command-line interface and argument parsing

**Responsibilities**:
- Parse command-line arguments and options
- Display help information and usage examples
- Handle error messages in a user-friendly way
- Coordinate between other agents

**Command Structure**:
```bash
llm-info [flags]
  --url string        Base URL of the LLM gateway
  --api-key string    API key for authentication
  --timeout duration  Request timeout (default 10s)
  --config string     Path to config file
  --gateway string    Name of gateway configuration to use
```

## Agent Interactions

1. **CLI Interface Agent** receives user input and initializes the **Configuration Agent**
2. **Configuration Agent** loads settings and provides them to the **API Client Agent**
3. **API Client Agent** makes requests to the LLM gateway and passes responses to the **Data Processing Agent**
4. **Data Processing Agent** normalizes the data and sends it to the **Visualization Agent**
5. **Visualization Agent** formats the data and displays it to the user via the **CLI Interface Agent**

## Error Handling Strategy

Each agent implements appropriate error handling:
- **API Client Agent**: Handles network errors, timeouts, and HTTP status codes
- **Configuration Agent**: Provides defaults for missing configuration
- **Data Processing Agent**: Gracefully handles missing or malformed data
- **Visualization Agent**: Adjusts output based on available data
- **CLI Interface Agent**: Displays user-friendly error messages

## Future Extensions

The agent architecture allows for future enhancements:
- Additional output formats (JSON, CSV)
- Model comparison features
- Cost calculation tools
- Integration with more gateway types