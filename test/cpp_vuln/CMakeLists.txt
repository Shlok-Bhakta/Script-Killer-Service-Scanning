cmake_minimum_required(VERSION 3.10)
project(cpp_vuln_test)
set(CMAKE_CXX_STANDARD 17)
add_executable(app main.cpp)

# Optional: Uncomment to enable AddressSanitizer
# target_compile_options(app PRIVATE -fsanitize=address -g)
# target_link_options(app PRIVATE -fsanitize=address)

# Optional: Static analysis
find_program(CPPCHECK cppcheck)
if(CPPCHECK)
    add_custom_target(
        cppcheck ALL
        COMMAND ${CPPCHECK} --enable=all --inconclusive --std=c++17 main.cpp
        COMMENT "Running cppcheck static analysis"
    )
endif()