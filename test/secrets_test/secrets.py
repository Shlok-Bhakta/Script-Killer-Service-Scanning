# Secret Scanner Test File
# This file contains intentionally planted fake secrets for testing gitleaks

import os

# Fake private key (truncated, not real)
PRIVATE_KEY = """-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA0Z3VS5JJcds3xfn/ygWyF8PbnGy0AHB7MASjgKz
-----END RSA PRIVATE KEY-----"""

# Generic API key pattern for testing
PAYMENT_API_KEY = "pk_test_abcdef1234567890ghijklmnop"

def get_secrets():
    return {
        "private_key": PRIVATE_KEY,
        "payment_api_key": PAYMENT_API_KEY
    }