# r_vuln.R - Deliberately unsafe R code for security testing
# DO NOT USE IN PRODUCTION

# CWE-94: Code Injection
user_input <- "system('ls')"
eval(parse(text=user_input))  # Unsafe!

# CWE-78: Command injection
filename <- readline("Enter filename:")
cmd <- paste("cat", filename)
system(cmd)  # Unsafe!

# CWE-22: Path Traversal
myfile <- readline()
readLines(myfile)  # Unsafe!

# CWE-502: Untrusted Deserialization
payload <- readline()
load(payload)  # Unsafe!

# CWE-829: Remote Code Inclusion
remote_code_url <- readline()
source(url(remote_code_url, "r"))  # Unsafe!

# CWE-798: Hardcoded Secret
db_password <- "SuperSecret1234"

# CWE-330: Weak Random Generator
pin <- sample(1000:9999, 1)