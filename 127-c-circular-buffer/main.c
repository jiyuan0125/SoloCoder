#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>
#include <unistd.h>
#include <string.h>
#include <math.h>
#include "sensor_data.h"
#include "stats.h"

#define NUM_SENSORS 5
#define SIMULATION_DURATION_SECONDS 10
#define WRITE_INTERVAL_MS 100

typedef struct {
    sensor_buffer_t* buffer;
    int sensor_id;
    volatile int running;
    double base_temp;
} writer_thread_args_t;

typedef struct {
    sensor_buffer_t** buffers;
    int num_sensors;
    volatile int running;
    const char* name;
} reader_thread_args_t;

static double generate_sensor_value(double base_temp, int step, bool* inject_anomaly) {
    double noise = ((double)rand() / RAND_MAX - 0.5) * 2.0;
    double value = base_temp + noise;
    
    *inject_anomaly = false;
    if (step % 50 == 0 && step > 0) {
        value += 15.0;
        *inject_anomaly = true;
    }
    
    return value;
}

void* writer_thread(void* arg) {
    writer_thread_args_t* args = (writer_thread_args_t*)arg;
    sensor_buffer_t* buffer = args->buffer;
    int sensor_id = args->sensor_id;
    double base_temp = args->base_temp;
    
    int step = 0;
    while (args->running) {
        bool injected_anomaly;
        double value = generate_sensor_value(base_temp, step, &injected_anomaly);
        timestamp_t ts = get_timestamp_ms();
        
        if (sensor_buffer_write(buffer, ts, value) != 0) {
            fprintf(stderr, "Sensor %d: Write failed\n", sensor_id);
        }
        
        if (injected_anomaly) {
            printf("Sensor %d: Injected anomaly at step %d, value=%.2f\n", 
                   sensor_id, step, value);
        }
        
        step++;
        usleep(WRITE_INTERVAL_MS * 1000);
    }
    
    printf("Writer thread %d exiting\n", sensor_id);
    return NULL;
}

void* stats_reader_thread(void* arg) {
    reader_thread_args_t* args = (reader_thread_args_t*)arg;
    sensor_buffer_t** buffers = args->buffers;
    int num_sensors = args->num_sensors;
    
    int read_count = 0;
    while (args->running) {
        printf("\n=== %s - Stats Report #%d ===\n", args->name, ++read_count);
        
        for (int i = 0; i < num_sensors; i++) {
            stats_t stats;
            if (sensor_buffer_get_stats(buffers[i], &stats) == 0) {
                printf("Sensor %d: current=%.2f, max=%.2f, min=%.2f, avg=%.2f, valid=%zu\n",
                       i, stats.current, stats.max, stats.min, stats.avg, stats.valid_count);
            }
        }
        
        sleep(2);
    }
    
    printf("%s thread exiting\n", args->name);
    return NULL;
}

void* latest_reader_thread(void* arg) {
    reader_thread_args_t* args = (reader_thread_args_t*)arg;
    sensor_buffer_t** buffers = args->buffers;
    int num_sensors = args->num_sensors;
    
    int read_count = 0;
    while (args->running) {
        printf("\n=== %s - Latest Values #%d ===\n", args->name, ++read_count);
        
        for (int i = 0; i < num_sensors; i++) {
            sensor_data_t data;
            if (sensor_buffer_read_latest(buffers[i], &data) == 0) {
                printf("Sensor %d: value=%.2f, ts=%lu, anomaly=%s\n",
                       i, data.value, (unsigned long)data.timestamp,
                       data.is_anomaly ? "YES" : "NO");
            }
        }
        
        sleep(1);
    }
    
    printf("%s thread exiting\n", args->name);
    return NULL;
}

void* range_reader_thread(void* arg) {
    reader_thread_args_t* args = (reader_thread_args_t*)arg;
    sensor_buffer_t** buffers = args->buffers;
    int num_sensors = args->num_sensors;
    
    int read_count = 0;
    while (args->running) {
        timestamp_t now = get_timestamp_ms();
        timestamp_t five_seconds_ago = now - 5000;
        
        printf("\n=== %s - Range Query #%d (last 5 seconds) ===\n", args->name, ++read_count);
        
        for (int i = 0; i < num_sensors; i++) {
            sensor_data_t results[100];
            size_t count = sensor_buffer_read_range(buffers[i], results, 
                                                      five_seconds_ago, now, 100);
            
            printf("Sensor %d: %zu data points in range\n", i, count);
            if (count > 0) {
                printf("  First: %.2f, Last: %.2f\n", results[0].value, results[count-1].value);
            }
        }
        
        sleep(3);
    }
    
    printf("%s thread exiting\n", args->name);
    return NULL;
}

int main(void) {
    srand((unsigned int)time(NULL));
    
    printf("=== Industrial Sensor Data Buffer Demo ===\n");
    printf("Number of sensors: %d\n", NUM_SENSORS);
    printf("Buffer size per sensor: %d\n", BUFFER_SIZE);
    printf("Anomaly threshold: %.1f degrees\n", ANOMALY_THRESHOLD);
    printf("Simulation duration: %d seconds\n", SIMULATION_DURATION_SECONDS);
    printf("Write interval: %d ms\n\n", WRITE_INTERVAL_MS);
    
    sensor_buffer_t* buffers[NUM_SENSORS];
    double base_temps[NUM_SENSORS] = {25.0, 30.0, 22.5, 28.0, 35.0};
    
    for (int i = 0; i < NUM_SENSORS; i++) {
        buffers[i] = sensor_buffer_create(BUFFER_SIZE);
        if (!buffers[i]) {
            fprintf(stderr, "Failed to create buffer for sensor %d\n", i);
            return 1;
        }
        printf("Created buffer for sensor %d (base temp: %.1f)\n", i, base_temps[i]);
    }
    
    pthread_t writer_threads[NUM_SENSORS];
    writer_thread_args_t writer_args[NUM_SENSORS];
    
    for (int i = 0; i < NUM_SENSORS; i++) {
        writer_args[i].buffer = buffers[i];
        writer_args[i].sensor_id = i;
        writer_args[i].running = 1;
        writer_args[i].base_temp = base_temps[i];
        
        if (pthread_create(&writer_threads[i], NULL, writer_thread, &writer_args[i]) != 0) {
            fprintf(stderr, "Failed to create writer thread %d\n", i);
            return 1;
        }
    }
    
    pthread_t reader_threads[3];
    reader_thread_args_t reader_args[3];
    
    reader_args[0].buffers = buffers;
    reader_args[0].num_sensors = NUM_SENSORS;
    reader_args[0].running = 1;
    reader_args[0].name = "Stats Reader";
    pthread_create(&reader_threads[0], NULL, stats_reader_thread, &reader_args[0]);
    
    reader_args[1].buffers = buffers;
    reader_args[1].num_sensors = NUM_SENSORS;
    reader_args[1].running = 1;
    reader_args[1].name = "Latest Reader";
    pthread_create(&reader_threads[1], NULL, latest_reader_thread, &reader_args[1]);
    
    reader_args[2].buffers = buffers;
    reader_args[2].num_sensors = NUM_SENSORS;
    reader_args[2].running = 1;
    reader_args[2].name = "Range Reader";
    pthread_create(&reader_threads[2], NULL, range_reader_thread, &reader_args[2]);
    
    printf("\n=== All threads started, running for %d seconds ===\n\n", SIMULATION_DURATION_SECONDS);
    
    sleep(SIMULATION_DURATION_SECONDS);
    
    printf("\n=== Stopping all threads ===\n");
    
    for (int i = 0; i < NUM_SENSORS; i++) {
        writer_args[i].running = 0;
    }
    for (int i = 0; i < 3; i++) {
        reader_args[i].running = 0;
    }
    
    for (int i = 0; i < NUM_SENSORS; i++) {
        pthread_join(writer_threads[i], NULL);
    }
    for (int i = 0; i < 3; i++) {
        pthread_join(reader_threads[i], NULL);
    }
    
    printf("\n=== Final Stats ===\n");
    for (int i = 0; i < NUM_SENSORS; i++) {
        stats_t stats;
        size_t size = sensor_buffer_size(buffers[i]);
        
        if (sensor_buffer_get_stats(buffers[i], &stats) == 0) {
            printf("Sensor %d: buffer_size=%zu, current=%.2f, max=%.2f, min=%.2f, avg=%.2f, valid=%zu\n",
                   i, size, stats.current, stats.max, stats.min, stats.avg, stats.valid_count);
        }
    }
    
    printf("\n=== Reading recent 10 values from sensor 0 ===\n");
    sensor_data_t recent[10];
    size_t count = sensor_buffer_read_recent(buffers[0], recent, 10);
    printf("Read %zu recent values:\n", count);
    for (size_t i = 0; i < count; i++) {
        printf("  [%zu] value=%.2f, anomaly=%s\n", i, recent[i].value, 
               recent[i].is_anomaly ? "YES" : "NO");
    }
    
    for (int i = 0; i < NUM_SENSORS; i++) {
        sensor_buffer_destroy(buffers[i]);
    }
    
    printf("\n=== Demo completed successfully ===\n");
    return 0;
}
