#include <iostream>
#include <cstring>

void vulnerableFunction(char* input) {
    char buffer[64];
    strcpy(buffer, input);
    std::cout << "Buffer: " << buffer << std::endl;
}

int main(int argc, char** argv) {
    if (argc > 1) {
        vulnerableFunction(argv[1]);
    }
    
    std::cout << "Hello World" << std::endl;
    return 0;
}
