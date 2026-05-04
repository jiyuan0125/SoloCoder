#define _XOPEN_SOURCE 500
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <pthread.h>
#include <unistd.h>
#include <assert.h>
#include "shared_ptr.h"
#include "ref_count.h"
#include "object_store.h"
#include "lifecycle.h"

static int g_test_passed = 0;
static int g_test_failed = 0;

#define TEST_CASE(name) \
    do { \
        printf("[TEST] %s... ", name); \
        fflush(stdout); \
    } while(0)

#define TEST_PASS() \
    do { \
        printf("PASSED\n"); \
        g_test_passed++; \
    } while(0)

#define TEST_FAIL(msg) \
    do { \
        printf("FAILED: %s\n", msg); \
        g_test_failed++; \
    } while(0)

static void test_basic_shared_ptr(void)
{
    sp_shared_ptr_t p1 = SP_SHARED_PTR_INIT;
    sp_shared_ptr_t p2 = SP_SHARED_PTR_INIT;
    sp_control_block_t *cb = NULL;
    int *data = NULL;

    TEST_CASE("Basic shared_ptr");

    data = (int *)malloc(sizeof(int));
    *data = 42;

    if (sp_control_block_create(&cb, data, free, SP_OBJ_TYPE_GENERIC, sizeof(int)) != 0) {
        free(data);
        TEST_FAIL("Failed to create control block");
        return;
    }

    p1.sp_cb = cb;

    if (sp_shared_ptr_use_count(&p1) != 1) {
        TEST_FAIL("Initial ref count should be 1");
        sp_shared_ptr_reset(&p1);
        return;
    }

    if (sp_shared_ptr_copy(&p2, &p1) != 0) {
        TEST_FAIL("Failed to copy shared_ptr");
        sp_shared_ptr_reset(&p1);
        return;
    }

    if (sp_shared_ptr_use_count(&p1) != 2) {
        TEST_FAIL("Ref count should be 2 after copy");
        goto cleanup;
    }

    if (sp_shared_ptr_use_count(&p2) != 2) {
        TEST_FAIL("Ref count should be 2 after copy (p2)");
        goto cleanup;
    }

    sp_shared_ptr_reset(&p1);
    if (sp_shared_ptr_use_count(&p2) != 1) {
        TEST_FAIL("Ref count should be 1 after reset p1");
        goto cleanup;
    }

    sp_shared_ptr_reset(&p2);
    TEST_PASS();
    return;

cleanup:
    sp_shared_ptr_reset(&p1);
    sp_shared_ptr_reset(&p2);
}

static void test_weak_ptr(void)
{
    sp_shared_ptr_t p1 = SP_SHARED_PTR_INIT;
    sp_weak_ptr_t w1 = SP_WEAK_PTR_INIT;
    sp_weak_ptr_t w2 = SP_WEAK_PTR_INIT;
    sp_shared_ptr_t locked = SP_SHARED_PTR_INIT;
    sp_control_block_t *cb = NULL;
    int *data = NULL;

    TEST_CASE("Weak_ptr operations");

    data = (int *)malloc(sizeof(int));
    *data = 100;

    if (sp_control_block_create(&cb, data, free, SP_OBJ_TYPE_GENERIC, sizeof(int)) != 0) {
        free(data);
        TEST_FAIL("Failed to create control block");
        return;
    }

    p1.sp_cb = cb;

    if (sp_weak_ptr_from_shared(&w1, &p1) != 0) {
        TEST_FAIL("Failed to create weak_ptr from shared_ptr");
        sp_shared_ptr_reset(&p1);
        return;
    }

    if (sp_weak_ptr_expired(&w1)) {
        TEST_FAIL("Weak_ptr should not be expired yet");
        goto cleanup;
    }

    if (sp_weak_ptr_copy(&w2, &w1) != 0) {
        TEST_FAIL("Failed to copy weak_ptr");
        goto cleanup;
    }

    if (sp_weak_ptr_lock(&locked, &w1) != 1) {
        TEST_FAIL("Failed to lock weak_ptr");
        goto cleanup;
    }

    if (sp_shared_ptr_use_count(&p1) != 2) {
        TEST_FAIL("Ref count should be 2 after lock (p1 + locked)");
        goto cleanup;
    }

    sp_shared_ptr_reset(&locked);
    sp_shared_ptr_reset(&p1);

    if (!sp_weak_ptr_expired(&w1)) {
        TEST_FAIL("Weak_ptr should be expired now");
        goto cleanup;
    }

    if (sp_weak_ptr_lock(&locked, &w1) != 0) {
        TEST_FAIL("Lock should fail on expired weak_ptr");
        goto cleanup;
    }

    sp_weak_ptr_reset(&w1);
    sp_weak_ptr_reset(&w2);
    TEST_PASS();
    return;

cleanup:
    sp_shared_ptr_reset(&p1);
    sp_shared_ptr_reset(&locked);
    sp_weak_ptr_reset(&w1);
    sp_weak_ptr_reset(&w2);
}

static void test_object_store(void)
{
    sp_object_store_t store;
    sp_shared_ptr_t p1 = SP_SHARED_PTR_INIT;
    sp_shared_ptr_t p2 = SP_SHARED_PTR_INIT;
    sp_shared_ptr_t p3 = SP_SHARED_PTR_INIT;
    int *data1, *data2, *data3;

    TEST_CASE("Object store operations");

    if (sp_store_init(&store, 16) != 0) {
        TEST_FAIL("Failed to init store");
        return;
    }

    data1 = (int *)malloc(sizeof(int));
    *data1 = 111;

    data2 = (int *)malloc(sizeof(int));
    *data2 = 222;

    if (sp_store_insert(&store, "key1", data1, free, SP_OBJ_TYPE_GENERIC, sizeof(int)) != 0) {
        free(data1);
        free(data2);
        sp_store_destroy(&store);
        TEST_FAIL("Failed to insert key1");
        return;
    }

    if (sp_store_insert(&store, "key2", data2, free, SP_OBJ_TYPE_GENERIC, sizeof(int)) != 0) {
        free(data2);
        sp_store_destroy(&store);
        TEST_FAIL("Failed to insert key2");
        return;
    }

    if (sp_store_size(&store) != 2) {
        TEST_FAIL("Store size should be 2");
        goto cleanup;
    }

    if (!sp_store_contains(&store, "key1")) {
        TEST_FAIL("Store should contain key1");
        goto cleanup;
    }

    if (sp_store_get(&store, "key1", &p1) != 0) {
        TEST_FAIL("Failed to get key1");
        goto cleanup;
    }

    data3 = (int *)sp_shared_ptr_get(&p1);
    if (data3 == NULL || *data3 != 111) {
        TEST_FAIL("Retrieved wrong data");
        goto cleanup;
    }

    if (sp_store_get(&store, "key2", &p2) != 0) {
        TEST_FAIL("Failed to get key2");
        goto cleanup;
    }

    if (sp_shared_ptr_copy(&p3, &p1) != 0) {
        TEST_FAIL("Failed to copy p1 to p3");
        goto cleanup;
    }

    if (sp_store_size(&store) != 2) {
        TEST_FAIL("Store size should still be 2");
        goto cleanup;
    }

    if (sp_store_remove(&store, "key1") != 0) {
        TEST_FAIL("Failed to remove key1");
        goto cleanup;
    }

    if (sp_store_size(&store) != 1) {
        TEST_FAIL("Store size should be 1 after remove");
        goto cleanup;
    }

    data3 = (int *)sp_shared_ptr_get(&p1);
    if (data3 == NULL || *data3 != 111) {
        TEST_FAIL("Data should still be accessible via existing ptr");
        goto cleanup;
    }

    sp_shared_ptr_reset(&p1);
    sp_shared_ptr_reset(&p2);
    sp_shared_ptr_reset(&p3);

    sp_store_destroy(&store);
    TEST_PASS();
    return;

cleanup:
    sp_shared_ptr_reset(&p1);
    sp_shared_ptr_reset(&p2);
    sp_shared_ptr_reset(&p3);
    sp_store_destroy(&store);
}

static void test_lifecycle(void)
{
    sp_shared_ptr_t img1 = SP_SHARED_PTR_INIT;
    sp_shared_ptr_t img2 = SP_SHARED_PTR_INIT;
    sp_shared_ptr_t tmpl1 = SP_SHARED_PTR_INIT;
    sp_shared_ptr_t tmpl2 = SP_SHARED_PTR_INIT;
    sp_image_data_t *img_data;
    sp_template_data_t *tmpl_data;

    TEST_CASE("Lifecycle management");

    sp_lc_module_init();

    if (sp_lc_load_image("/test/image1.png", &img1) != 0) {
        TEST_FAIL("Failed to load image1");
        goto cleanup;
    }

    img_data = sp_lc_image_get(&img1);
    if (img_data == NULL) {
        TEST_FAIL("Failed to get image data");
        goto cleanup;
    }

    if (img_data->width != 100 || img_data->height != 100) {
        TEST_FAIL("Image dimensions are wrong");
        goto cleanup;
    }

    if (sp_lc_load_image("/test/image1.png", &img2) != 0) {
        TEST_FAIL("Failed to load image1 again");
        goto cleanup;
    }

    if (sp_shared_ptr_use_count(&img1) != 2) {
        TEST_FAIL("Ref count should be 2 for same image");
        goto cleanup;
    }

    if (sp_lc_load_template("template_main", &tmpl1) != 0) {
        TEST_FAIL("Failed to load template");
        goto cleanup;
    }

    tmpl_data = sp_lc_template_get(&tmpl1);
    if (tmpl_data == NULL) {
        TEST_FAIL("Failed to get template data");
        goto cleanup;
    }

    if (sp_lc_load_template("template_main", &tmpl2) != 0) {
        TEST_FAIL("Failed to load template again");
        goto cleanup;
    }

    if (sp_shared_ptr_use_count(&tmpl1) != 2) {
        TEST_FAIL("Ref count should be 2 for same template");
        goto cleanup;
    }

    sp_lc_release(&img1);
    sp_lc_release(&img2);
    sp_lc_release(&tmpl1);
    sp_lc_release(&tmpl2);

    sp_lc_module_cleanup();
    TEST_PASS();
    return;

cleanup:
    sp_lc_release(&img1);
    sp_lc_release(&img2);
    sp_lc_release(&tmpl1);
    sp_lc_release(&tmpl2);
    sp_lc_module_cleanup();
}

#define NUM_THREADS 10
#define NUM_ITERATIONS 1000

typedef struct {
    const char *key;
    int thread_id;
} thread_arg_t;

static pthread_mutex_t g_counter_mutex = PTHREAD_MUTEX_INITIALIZER;
static int g_error_count = 0;

static void *thread_func(void *arg)
{
    thread_arg_t *targ = (thread_arg_t *)arg;
    sp_shared_ptr_t ptr = SP_SHARED_PTR_INIT;
    int i;

    for (i = 0; i < NUM_ITERATIONS; i++) {
        if (i % 3 == 0) {
            sp_lc_release(&ptr);

            if (sp_lc_load_image(targ->key, &ptr) != 0) {
                pthread_mutex_lock(&g_counter_mutex);
                g_error_count++;
                pthread_mutex_unlock(&g_counter_mutex);
            }
        } else if (i % 3 == 1) {
            sp_image_data_t *img = sp_lc_image_get(&ptr);
            if (img != NULL) {
                if (img->width != 100) {
                    pthread_mutex_lock(&g_counter_mutex);
                    g_error_count++;
                    pthread_mutex_unlock(&g_counter_mutex);
                }
            }
        } else {
            sp_shared_ptr_t copy = SP_SHARED_PTR_INIT;
            if (sp_shared_ptr_copy(&copy, &ptr) == 0) {
                sp_lc_release(&copy);
            }
        }

        usleep(100);
    }

    sp_lc_release(&ptr);
    return NULL;
}

static void test_multithreading(void)
{
    pthread_t threads[NUM_THREADS];
    thread_arg_t args[NUM_THREADS];
    int i;

    TEST_CASE("Multi-threading safety");

    sp_lc_module_init();
    g_error_count = 0;

    for (i = 0; i < NUM_THREADS; i++) {
        args[i].key = "/test/concurrent.png";
        args[i].thread_id = i;

        if (pthread_create(&threads[i], NULL, thread_func, &args[i]) != 0) {
            TEST_FAIL("Failed to create thread");
            sp_lc_module_cleanup();
            return;
        }
    }

    for (i = 0; i < NUM_THREADS; i++) {
        pthread_join(threads[i], NULL);
    }

    if (g_error_count > 0) {
        printf("Errors: %d\n", g_error_count);
        TEST_FAIL("Multi-threaded test had errors");
        sp_lc_module_cleanup();
        return;
    }

    sp_lc_module_cleanup();
    TEST_PASS();
}

static void test_circular_reference_prevention(void)
{
    sp_shared_ptr_t objA = SP_SHARED_PTR_INIT;
    sp_shared_ptr_t objB = SP_SHARED_PTR_INIT;
    sp_template_data_t *dataA, *dataB;

    TEST_CASE("Circular reference prevention (weak_ptr)");

    sp_lc_module_init();

    if (sp_lc_load_template("objA", &objA) != 0) {
        TEST_FAIL("Failed to load objA");
        goto cleanup;
    }

    if (sp_lc_load_template("objB", &objB) != 0) {
        TEST_FAIL("Failed to load objB");
        goto cleanup;
    }

    dataA = sp_lc_template_get(&objA);
    dataB = sp_lc_template_get(&objB);

    if (dataA == NULL || dataB == NULL) {
        TEST_FAIL("Failed to get template data");
        goto cleanup;
    }

    if (sp_lc_add_weak_dependency(&objA, &objB) != 0) {
        TEST_FAIL("Failed to add weak dependency A->B");
        goto cleanup;
    }

    if (sp_lc_add_weak_dependency(&objB, &objA) != 0) {
        TEST_FAIL("Failed to add weak dependency B->A");
        goto cleanup;
    }

    if (!sp_weak_ptr_expired(&dataA->dependency)) {
        sp_shared_ptr_t locked = SP_SHARED_PTR_INIT;
        if (sp_weak_ptr_lock(&locked, &dataA->dependency) == 1) {
            sp_lc_release(&locked);
        }
    }

    if (sp_shared_ptr_use_count(&objA) != 1) {
        TEST_FAIL("Ref count for A should be 1 (weak refs don't count)");
        goto cleanup;
    }

    if (sp_shared_ptr_use_count(&objB) != 1) {
        TEST_FAIL("Ref count for B should be 1 (weak refs don't count)");
        goto cleanup;
    }

    sp_lc_release(&objA);
    sp_lc_release(&objB);

    sp_lc_module_cleanup();
    TEST_PASS();
    return;

cleanup:
    sp_lc_release(&objA);
    sp_lc_release(&objB);
    sp_lc_module_cleanup();
}

static void test_object_singleton(void)
{
    sp_shared_ptr_t ptr1 = SP_SHARED_PTR_INIT;
    sp_shared_ptr_t ptr2 = SP_SHARED_PTR_INIT;
    sp_shared_ptr_t ptr3 = SP_SHARED_PTR_INIT;
    sp_image_data_t *img1, *img2, *img3;

    TEST_CASE("Object singleton in memory");

    sp_lc_module_init();

    if (sp_lc_load_image("/test/unique.png", &ptr1) != 0) {
        TEST_FAIL("Failed to load first time");
        goto cleanup;
    }

    if (sp_lc_load_image("/test/unique.png", &ptr2) != 0) {
        TEST_FAIL("Failed to load second time");
        goto cleanup;
    }

    if (sp_lc_load_image("/test/unique.png", &ptr3) != 0) {
        TEST_FAIL("Failed to load third time");
        goto cleanup;
    }

    img1 = sp_lc_image_get(&ptr1);
    img2 = sp_lc_image_get(&ptr2);
    img3 = sp_lc_image_get(&ptr3);

    if (img1 == NULL || img2 == NULL || img3 == NULL) {
        TEST_FAIL("Failed to get image data");
        goto cleanup;
    }

    if (img1 != img2 || img2 != img3) {
        TEST_FAIL("All pointers should point to the same object");
        goto cleanup;
    }

    if (sp_shared_ptr_use_count(&ptr1) != 3) {
        TEST_FAIL("Ref count should be 3");
        goto cleanup;
    }

    sp_lc_release(&ptr1);
    sp_lc_release(&ptr2);
    sp_lc_release(&ptr3);

    sp_lc_module_cleanup();
    TEST_PASS();
    return;

cleanup:
    sp_lc_release(&ptr1);
    sp_lc_release(&ptr2);
    sp_lc_release(&ptr3);
    sp_lc_module_cleanup();
}

int main(void)
{
    printf("\n");
    printf("========================================\n");
    printf("  Shared Resource Manager Test Suite\n");
    printf("========================================\n");
    printf("\n");

    test_basic_shared_ptr();
    test_weak_ptr();
    test_object_store();
    test_lifecycle();
    test_object_singleton();
    test_circular_reference_prevention();
    test_multithreading();

    printf("\n");
    printf("========================================\n");
    printf("  Test Results\n");
    printf("========================================\n");
    printf("  PASSED: %d\n", g_test_passed);
    printf("  FAILED: %d\n", g_test_failed);
    printf("========================================\n");
    printf("\n");

    if (g_test_failed == 0) {
        sp_lc_module_init();
        printf("Dumping final state after all tests:\n\n");
        sp_lc_dump_state();
        sp_lc_module_cleanup();
        printf("\nAll tests passed!\n");
        return 0;
    } else {
        printf("Some tests failed!\n");
        return 1;
    }
}
