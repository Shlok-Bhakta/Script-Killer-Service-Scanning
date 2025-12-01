# Script Killer - Universal Security Scanner CLI

An everything scanner CLI tool that makes security scanning beautiful and effortless across all your projects.

## Project Overview

In a world where the threat of cyberattacks is growing faster than ever, developers face increased pressure to secure their projects and services. Many projects utilize more than one programming language, and it can be quite costly to pursue the idea of having a dedicated cyber security team to manage vulnerabilities within the code. To address this challenge, our group came together to produce a universal language security scanner that is designed to bring simplicity to the risk management within developers repos. Through the use of various CLI scanning tools like OSVScanner, Bandit, GoSec, and more, this tool provides a high quality static analysis of projects by pooling these tools into a single interface. This makes it easier for smaller developers to analyze and find vulnerabilities within their project code and manage the risks appropriately for a cheaper alternative than to have designated security professionals

## Project Idea

Stop bad code before it even gets committed. Script Killer is a universal security scanner that:

- **Multi-language support**: Scans projects in any language (JavaScript, Python, Go, Rust, Java, etc.)
- **Zero config**: Uses Nix to automatically fetch the right scanning tools based on your project files
- **Beautiful output**: Built with Go + Bubbletea for a gorgeous terminal UI because appearances matter
- **Comprehensive reports**: Analyzes dependencies, vulnerabilities, code quality, and security issues
- **CI/CD ready**: Drop-in GitHub Action to run scans in your pipeline
- **Git hook integration**: Run as a pre-commit hook to catch issues before they're committed
- **Live scanning** (stretch goal): Watch mode that scans as you code for real-time feedback

## Philosophy

Making security scanning so easy, accessible, and beautiful that developers actually want to use it. We're essentially building the world's prettiest bash script - but one that actually stops vulnerable code from making it into your repo.

## How It Works

1. Detect project files and languages
2. Use Nix to provision the right scanning tools (no manual setup!)
3. Run all relevant scanners in parallel
4. Generate a beautiful, actionable report
5. Integrate seamlessly into your workflow (CI, git hooks, or standalone)

## Use Cases

- **Local Development**: Run scans before committing
- **Pre-commit Hook**: Automatic scanning on every commit
- **CI/CD Pipeline**: GitHub Action integration for automated checks
- **Live Mode**: Continuous scanning during development (stretch goal)

## MCP Integration (opencode)

Script Killer can run as an MCP (Model Context Protocol) server, allowing AI assistants like opencode to scan your codebase for vulnerabilities.

### Setup

1. Build the binary:
   ```bash
   make build
   ```

2. Add to your `opencode.json` in your project root:
   ```json
   {
     "mcp": {
       "scriptkiller": {
         "type": "local",
         "command": ["./bin/scriptkiller", "--mcp"]
       }
     }
   }
   ```

3. If the binary is not in your project directory, specify the full path or add it to PATH:
   ```bash
   PATH=/path/to/scriptkiller/bin:$PATH opencode
   ```

### Available MCP Tools

- **scan**: Run security scan on a directory
- **list_findings**: List current scan findings (optionally filter by severity)
- **detect_languages**: Detect programming languages in a directory

### Available MCP Resources

- `scriptkiller://findings` - Current scan findings
- `scriptkiller://languages` - Detected languages

## Implemented Tools
- GoSec - https://github.com/securego/gosec
- Grype - https://github.com/anchore/grype
- OSVScanner - https://github.com/google/osv-scanner
- Bandit - https://github.com/PyCQA/bandit
- Gitleaks - https://github.com/gitleaks/gitleaks
