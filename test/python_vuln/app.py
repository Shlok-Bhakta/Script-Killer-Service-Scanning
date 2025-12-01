from flask import Flask, request, jsonify, render_template_string
import yaml
import requests
import os
import sqlite3

app = Flask(__name__)

# --- CWE-798: Hardcoded Credentials ---
API_KEY = "hardcodedapikey123456"

# --- CWE-502: Unsafe YAML Deserialization ---
@app.route('/config', methods=['POST'])
def load_config():
    config_data = request.data
    config = yaml.load(config_data)  # Vulnerable: unsafe deserialization
    return jsonify(config)

# --- CWE-918: Server-Side Request Forgery (SSRF) and CWE-295: Insecure TLS ---
@app.route('/fetch')
def fetch_data():
    url = request.args.get('url')
    response = requests.get(url, verify=False)  # Dangerous: SSRF and disables TLS verification
    return response.text

# --- CWE-78: OS Command Injection ---
@app.route('/exec')
def exec_cmd():
    cmd = request.args.get('cmd')
    os.system(cmd)  # Dangerous: command injection
    return "ok"

# --- CWE-89: SQL Injection ---
@app.route('/user')
def get_user():
    username = request.args.get('username')
    conn = sqlite3.connect('test.db')
    cursor = conn.cursor()
    # Vulnerable: SQL Injection due to unsanitized input!
    cursor.execute(f"SELECT * FROM users WHERE username = '{username}'")
    result = cursor.fetchone()
    return jsonify(result if result else {"error": "User not found"})

# --- CWE-79: Cross-Site Scripting (XSS) via Template Injection ---
@app.route('/hello')
def hello():
    name = request.args.get('name', '')
    # Vulnerable: user input directly rendered as template!
    template = f"Hello, {name}"
    # Even more dangerous (actual template injection):
    return render_template_string(template)

# --- CWE-601: Open Redirect ---
@app.route('/redirect')
def insecure_redirect():
    next_url = request.args.get('next', '/')
    # Vulnerable: open redirect
    return f'<a href="{next_url}">Continue</a>'

# --- CWE-22: Path Traversal ---
@app.route('/read')
def read_file():
    filename = request.args.get('filename')
    with open(filename, 'r') as f:  # Vulnerable: path traversal
        content = f.read()
    return content

# --- CWE-306: Missing Authentication for Critical Function ---
@app.route('/admin')
def admin_panel():
    # No auth on admin endpoint
    return "Welcome to admin panel! (No authentication!)"

# --- CWE-532: Information Exposure Through Log Files ---
@app.route('/login', methods=['POST'])
def login():
    user = request.form.get('user')
    password = request.form.get('pass')
    print(f"User attempted login: {user}, password: {password}")  # Logging sensitive info
    # Always returns "success" for testing
    return "Login attempted."

if __name__ == '__main__':
    # CWE-489: Info Disclosure via debug mode
    app.run(debug=True)