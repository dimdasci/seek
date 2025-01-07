# seek

A CLI utility for searching the web directly from your terminal.

Given an information request, the utility will build and execute a search plan and return the answer in markdown format.

## Project Structure

```
seek/
├── bin/           # Built binaries
├── cmd/           # Command definitions
├── internal/      # Private application code
├── main.go        # Entry point
└── Makefile       # Build automation
```

## Building & Installation
To build the project locally, run:
```
make build
```
This will compile the utility into the bin/ directory.

## Dependencies

**[OpenAI](https://openai.com)** models are used for building a search plan, search results analysis and answer compilation.

**[Tavily Search](https://tavily.com)** is used for web search.

You need API keys from these services to use seek.

## Configuration
**seek** can be configured via a `.seek.yaml` file located in user home directory or directory of launching **seek**. 

Openai and tavily api_keys are required. You can also use environment variables with the `SEEK_` prefix. For example, `SEEK_OPENAI_API_KEY` for OpenAI API key. 

`.seek.yaml` example:

```yaml
openai:
  api_key: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
  reasoning: 
    timeout: 60s
    max_tokens: 5000
    model: o1-mini
  completion: 
    timeout: 60s
    max_tokens: 7000
    model: gpt4o-mini

websearch:
  tavily:
    api_key: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    search_url: https://api.tavily.com/search
    extract_url: https://api.tavily.com/extract
    max_results: 10
    timeout: 20s

webreader:
  timeout: 20s

logging:
  level: "error"
  file: "~/logs/seek.log"
```

Place config file in the same directory as the binary or in your home directory.

## Running the Binary

After building the binary, you can run it directly from the `bin/` directory or place it in a directory included in your system's PATH.

## Usage Examples

Run the CLI with a subcommand, such as:
```
seek answer "What is the capital of France?"
```

You can skip quotes if the request does not contain special symbols:

```
seek answer 2025 public holidays in Madrid Spain
```

Use `-o` flag to specify the output file:
```
seek answer "compare 2025 public holidays in UK, \
Spain and Argentina. Which country provides \
more opportunities to celebrate?" \
-o holidays.md
```

Use the `--help` flag for more details.
