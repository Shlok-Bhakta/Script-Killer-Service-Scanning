// CWE-79: Cross-Site Scripting (XSS) via direct DOM insertion
function badXSS() {
    const userInput = '<img src=x onerror=alert(1)>';
    document.body.innerHTML = userInput; // Vulnerable!
}

// CWE-89: SQL Injection (if using server-side JS like Node)
const sqlite3 = require('sqlite3').verbose();
const db = new sqlite3.Database(':memory:');
function sqlInjection(username) {
    // Vulnerable: direct string interpolation
    db.run(`SELECT * FROM users WHERE name = '${username}'`);
}

// CWE-78: OS Command Injection (Node.js)
const { exec } = require('child_process');
function commandInjection(cmd) {
    exec(cmd, (err, stdout, stderr) => {
        console.log(stdout);
    });
}

// CWE-502: Unsafe Deserialization (Node.js, via eval)
function unsafeDeserialize(serializedString) {
    // Vulnerable: executes arbitrary code from untrusted input!
    eval(serializedString);
}

// CWE-601: Open Redirect
function openRedirect(req, res) {
    // Vulnerable: no validation of redirect target
    const url = req.query.next;
    res.redirect(url);
}

// CWE-352: CSRF (Poor protection, example server)
const express = require('express');
const app = express();
app.post("/update-profile", (req, res) => {
    // Vulnerable: lacks CSRF token validation
    // ...update profile logic...
    res.send("Profile updated");
});

// CWE-321: Hardcoded Credentials / Secrets
const API_KEY = "123456789-FAKE-API-KEY";

// CWE-310: Weak Crypto
const crypto = require('crypto');
function weakHash(password) {
    // Vulnerable: MD5 is broken
    return crypto.createHash('md5').update(password).digest('hex');
}

// CWE-798: Hardcoded database credentials
const dbUser = "root";
const dbPass = "passw0rd";

// CWE-94: Remote Code Execution via User Input (Node.js)
function remoteCodeExec(input) {
    eval(input); // Very dangerous!
}

// CWE-20: Unsafe File Upload
app.post("/upload", (req, res) => {
    // Vulnerable: no file type/size checks!
    req.files.upload.mv("/uploads/" + req.files.upload.name);
    res.send("Uploaded");
});

// CWE-532: Information Exposure Through Logging
app.post("/login", (req, res) => {
    console.log("Login attempt:", req.body.username, req.body.password); // Logging sensitive info
    res.send("OK");
});

// Export or execute functions for scan coverage
if (require.main === module) {
    badXSS();
    sqlInjection("admin' OR '1'='1");
    commandInjection("ls; rm -rf /");
    unsafeDeserialize("alert('hacked')");
    weakHash("password123");
}