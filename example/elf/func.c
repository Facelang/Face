#include <stdio.h>

int a = 1;

int func(int x, int y) {
    return x + y;
}

int main() {
    int b , c;
    b = 2;
    c = b + c;
    c = func(a, c);
    printf("计算结果：%d\n", c);
}
