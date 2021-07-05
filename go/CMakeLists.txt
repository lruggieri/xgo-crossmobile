cmake_minimum_required(VERSION 3.17)
project(cross_compilation_test C)

set(CMAKE_C_STANDARD 99)
set(LIBRARY_DIR ${CMAKE_SOURCE_DIR})

include_directories("${LIBRARY_DIR}")

if(APPLE)
    message( "Apple system detected" )
    set(LIBRARIES "${LIBRARY_DIR}/main-darwin-10.15-amd64.dylib")
    set(CMAKE_MACOSX_RPATH 1)
elseif(UNIX)
    message( "Unix system detected" )
    set(LIBRARIES "${LIBRARY_DIR}/main-linux-amd64.so")
elseif(WIN32)
    message( "Windows system detected" )
    set(LIBRARIES "${LIBRARY_DIR}/main-windows-4.0-amd64.dll")
endif()

find_library(CGO_LIB ${LIBRARIES} HINTS .)

add_executable(cross_compilation_test
        main-darwin-10.12-amd64.h
        test/c_test.c)

TARGET_LINK_LIBRARIES(cross_compilation_test PUBLIC ${CGO_LIB})