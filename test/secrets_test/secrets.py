# Secret Scanner Test File
# This file contains intentionally planted fake secrets for testing gitleaks

import os

# Fake AWS credentials (not real - test patterns)
AWS_ACCESS_KEY_ID = "AKIAIOSFODNN7EXAMPLE"
AWS_SECRET_ACCESS_KEY = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"

# Fake GitHub token (not real - obviously fake)
GITHUB_TOKEN = "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# Fake API key patterns
API_KEY = "sk-1234567890abcdefghijklmnopqrstuvwxyz"

# Fake database connection string with password
DATABASE_URL = "postgresql://admin:SuperSecretPassword123!@localhost:5432/mydb"

# Fake private key (truncated, not real)
PRIVATE_KEY = """-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA0Z3VS5JJcds3xfn/ygWyF8PbnGy0AHB7MASjgKz
-----END RSA PRIVATE KEY-----"""

def get_secret():
    # Hardcoded password (bad practice)
    password = "MyHardcodedPassword123!"
    return password

# Removed Slack/Stripe patterns - GitHub push protection blocks them
# Use generic webhook pattern instead for testing
WEBHOOK_URL = "https://example.com/webhook/secret_token_abc123"

# Generic API key pattern for testing
PAYMENT_API_KEY = "pk_test_abcdef1234567890ghijklmnop"
