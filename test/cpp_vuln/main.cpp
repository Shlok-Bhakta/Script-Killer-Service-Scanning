#include <iostream>
#include <cstring>
#include <climits>
#include <cstdio>    // For printf, sprintf
#include <cstdlib>   // For system
#include <fcntl.h>   // For open
#include <unistd.h>  // For close, write

// CWE-121: Stack-based Buffer Overflow
void vulnerableFunction(char* input) {
    char buffer[64];
    strcpy(buffer, input); // Vulnerable: no bounds checking
    std::cout << "Buffer: " << buffer << std::endl;
}

// CWE-134: Uncontrolled Format String
void formatStringVuln(char* input) {
    printf(input); // Vulnerable: format string (use printf("%s", input) instead)
}

// CWE-416: Use After Free
void useAfterFree() {
    int* p = new int[10];
    delete[] p;
    p[0] = 42; // Vulnerable: use after free
}

// CWE-415: Double Free
void doubleFree() {
    int* p = new int;
    delete p;
    delete p; // Vulnerable: double free
}

// CWE-476: Null Pointer Dereference
void nullDeref() {
    int* p = nullptr;
    *p = 42; // Vulnerable: null pointer dereference
}

// CWE-190: Integer Overflow/Underflow
void intOverflow(unsigned int a) {
    unsigned int b = a + UINT_MAX;
    std::cout << "Integer Overflow Result: " << b << std::endl;
}

// CWE-78: OS Command Injection
void commandInjection(char* input) {
    system(input); // Vulnerable: unsanitized system command
}

// CWE-362: Race Condition (TOCTOU - Time Of Check To Time Of Use)
void insecureFileAccess(const char* filename) {
    int fd = open(filename, O_CREAT | O_WRONLY, 0666);
    // Potential race condition between file check and use
    write(fd, "test", 4);
    close(fd);
}

// CWE-242: Dangerous Function Usage
void dangerousFunctions(char* input) {
    char buf2[64];

    strcat(buf2, input); // Vulnerable: no bounds checking
    sprintf(buf2, input); // Vulnerable: can overflow, format string risk

    std::cout << ", strcat/sprintf: " << buf2 << std::endl;
}

// Safe alternative (for negative test)
void safeFunction(char* input) {
    char buffer[64];
    strncpy(buffer, input, sizeof(buffer) - 1);
    buffer[sizeof(buffer) - 1] = '\0';
    std::cout << "Safe buffer: " << buffer << std::endl;
}

int main(int argc, char** argv) {
    std::cout << "Hello World" << std::endl;

    if (argc > 1) {
        std::cout << "Running vulnerableFunction..." << std::endl;
        vulnerableFunction(argv[1]);

        std::cout << "Running formatStringVuln..." << std::endl;
        formatStringVuln(argv[1]);

        std::cout << "Running commandInjection..." << std::endl;
        commandInjection(argv[1]);

        std::cout << "Running insecureFileAccess..." << std::endl;
        insecureFileAccess(argv[1]);

        std::cout << "Running dangerousFunctions..." << std::endl;
        dangerousFunctions(argv[1]);

        std::cout << "Running safeFunction..." << std::endl;
        safeFunction(argv[1]);
    }

    std::cout << "Running useAfterFree..." << std::endl;
    useAfterFree();

    std::cout << "Running doubleFree..." << std::endl;
    doubleFree();

    std::cout << "Running nullDeref..." << std::endl;
    nullDeref();

    std::cout << "Running intOverflow..." << std::endl;
    intOverflow(UINT_MAX);

    return 0;
}