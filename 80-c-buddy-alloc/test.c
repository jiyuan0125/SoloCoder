#include <stdio.h>
#include "buddy_alloc.h"

int main() {
    printf("=== Buddy Allocator Test ===\n\n");
    
    buddy_allocator_t* alloc = buddy_init();
    if (alloc == NULL) {
        printf("Failed to initialize allocator\n");
        return 1;
    }
    
    printf("Initial state:\n");
    buddy_stats(alloc);
    
    printf("\n--- Test 1: Allocate 100 bytes (should round to 128) ---\n");
    void* p1 = buddy_alloc(alloc, 100);
    printf("Allocated 100 bytes at %p\n", p1);
    buddy_stats(alloc);
    
    printf("\n--- Test 2: Allocate 200 bytes (should round to 256) ---\n");
    void* p2 = buddy_alloc(alloc, 200);
    printf("Allocated 200 bytes at %p\n", p2);
    buddy_stats(alloc);
    
    printf("\n--- Test 3: Allocate 64 bytes ---\n");
    void* p3 = buddy_alloc(alloc, 64);
    printf("Allocated 64 bytes at %p\n", p3);
    buddy_stats(alloc);
    
    printf("\n--- Test 4: Free p1 (128 bytes) ---\n");
    buddy_free(alloc, p1);
    printf("Freed p1\n");
    buddy_stats(alloc);
    
    printf("\n--- Test 5: Free p3 (64 bytes), should merge with buddy if possible ---\n");
    buddy_free(alloc, p3);
    printf("Freed p3\n");
    buddy_stats(alloc);
    
    printf("\n--- Test 6: Free p2 (256 bytes) ---\n");
    buddy_free(alloc, p2);
    printf("Freed p2\n");
    buddy_stats(alloc);
    
    printf("\n--- Test 7: Allocate 0 bytes (should return NULL) ---\n");
    void* p4 = buddy_alloc(alloc, 0);
    printf("Allocated 0 bytes: %p (expected: NULL)\n", p4);
    
    printf("\n--- Test 8: Free invalid pointer (outside range) ---\n");
    int x = 42;
    buddy_error_t err = buddy_free(alloc, &x);
    printf("Free invalid pointer result: %d (expected: -1)\n", err);
    
    buddy_destroy(alloc);
    printf("\n=== Test Complete ===\n");
    
    return 0;
}
