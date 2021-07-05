#include <stdio.h>
#include <stdlib.h>
#include <main-darwin-10.12-amd64.h>

int main() {
    const int maxBufferLen = 4096;
    char *errorBuffer = (char*) malloc(sizeof(char) * maxBufferLen);
    ServerClientSingleCommunication(1234, errorBuffer, maxBufferLen);

    printf("%s\n", errorBuffer);

    return 0;
}
