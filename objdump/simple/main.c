#include <stdio.h>

int add(int a, int b) {
    return a + b;
}

int main() {
    int result = add(5, 10);
    printf("Result: %d\n", result);
    return 0;
} 