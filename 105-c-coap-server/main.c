#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <errno.h>
#include <signal.h>
#include <pthread.h>

#include "coap.h"
#include "storage.h"
#include "alarm.h"

#define COAP_PORT 5683
#define BUFFER_SIZE 1024
#define MAX_URI_PATH_SEGMENTS 8
#define SIMULATION_SENSOR_COUNT 5
#define TOP_TEMP_LIMIT 5

static storage_ring_buffer_t g_storage;
static alarm_manager_t g_alarm_mgr;
static int g_running = 1;
static int g_sockfd = -1;
static uint16_t g_next_msg_id = 1;
static pthread_mutex_t g_msg_id_mutex = PTHREAD_MUTEX_INITIALIZER;

static void print_timestamp(time_t t, char *buf, size_t len) {
    struct tm *tm_info = localtime(&t);
    strftime(buf, len, "%Y-%m-%d %H:%M:%S", tm_info);
}

static const char *sensor_type_to_string(sensor_type_t type) {
    switch (type) {
        case SENSOR_TYPE_TEMPERATURE: return "temperature";
        case SENSOR_TYPE_HUMIDITY: return "humidity";
        case SENSOR_TYPE_DOOR_WINDOW: return "door_window";
        default: return "unknown";
    }
}

static int string_to_sensor_type(const char *str, sensor_type_t *type) {
    if (strcmp(str, "temperature") == 0) {
        *type = SENSOR_TYPE_TEMPERATURE;
        return 0;
    } else if (strcmp(str, "humidity") == 0) {
        *type = SENSOR_TYPE_HUMIDITY;
        return 0;
    } else if (strcmp(str, "door_window") == 0) {
        *type = SENSOR_TYPE_DOOR_WINDOW;
        return 0;
    }
    return -1;
}

static int parse_sensor_report(const uint8_t *payload, size_t payload_len, sensor_reading_t *reading) {
    char buf[COAP_MAX_PAYLOAD_SIZE];
    if (payload_len >= sizeof(buf)) {
        return -1;
    }
    memcpy(buf, payload, payload_len);
    buf[payload_len] = '\0';
    
    memset(reading, 0, sizeof(sensor_reading_t));
    
    char *token;
    char *saveptr;
    char *copy = strdup(buf);
    if (!copy) return -1;
    
    token = strtok_r(copy, ",", &saveptr);
    if (!token) { free(copy); return -1; }
    strncpy(reading->sensor_id, token, MAX_SENSOR_ID_LEN - 1);
    
    token = strtok_r(NULL, ",", &saveptr);
    if (!token) { free(copy); return -1; }
    sensor_type_t type;
    if (string_to_sensor_type(token, &type) < 0) { free(copy); return -1; }
    reading->type = type;
    
    token = strtok_r(NULL, ",", &saveptr);
    if (!token) { free(copy); return -1; }
    if (reading->type == SENSOR_TYPE_TEMPERATURE) {
        reading->value.temperature = atof(token);
    } else if (reading->type == SENSOR_TYPE_HUMIDITY) {
        reading->value.humidity = atof(token);
    } else if (reading->type == SENSOR_TYPE_DOOR_WINDOW) {
        reading->value.is_open = (strcmp(token, "1") == 0 || strcmp(token, "true") == 0);
    }
    
    token = strtok_r(NULL, ",", &saveptr);
    if (!token) { free(copy); return -1; }
    reading->battery_percent = (uint8_t)atoi(token);
    
    token = strtok_r(NULL, ",", &saveptr);
    if (token) {
        reading->timestamp = (time_t)atoll(token);
    } else {
        reading->timestamp = time(NULL);
    }
    
    free(copy);
    return 0;
}

static int parse_uri_path(const coap_message_t *msg, char **segments, uint8_t *count) {
    *count = 0;
    uint16_t current = 0;
    
    for (uint8_t i = 0; i < msg->option_count; i++) {
        current += msg->options[i].delta;
        if (current == COAP_OPTION_URI_PATH) {
            if (*count < MAX_URI_PATH_SEGMENTS && msg->options[i].value) {
                segments[*count] = malloc(msg->options[i].length + 1);
                if (!segments[*count]) return -1;
                memcpy(segments[*count], msg->options[i].value, msg->options[i].length);
                segments[*count][msg->options[i].length] = '\0';
                (*count)++;
            }
        }
    }
    return 0;
}

static void free_uri_segments(char **segments, uint8_t count) {
    for (uint8_t i = 0; i < count; i++) {
        if (segments[i]) free(segments[i]);
    }
}

static const char *get_uri_query_param(const coap_message_t *msg, const char *key) {
    uint16_t current = 0;
    size_t key_len = strlen(key);
    
    for (uint8_t i = 0; i < msg->option_count; i++) {
        current += msg->options[i].delta;
        if (current == COAP_OPTION_URI_QUERY && msg->options[i].value) {
            size_t len = msg->options[i].length;
            const uint8_t *val = msg->options[i].value;
            
            if (len > key_len + 1 && 
                strncmp((const char*)val, key, key_len) == 0 && 
                val[key_len] == '=') {
                return (const char*)&val[key_len + 1];
            }
        }
    }
    return NULL;
}

static int build_response_message(coap_message_t *response, const coap_message_t *request,
                                   coap_code_t code, const char *payload) {
    coap_message_init(response);
    response->version = COAP_VERSION;
    
    if (request->type == COAP_TYPE_CON) {
        response->type = COAP_TYPE_ACK;
    } else {
        response->type = COAP_TYPE_NON;
    }
    
    response->code = code;
    response->msg_id = request->msg_id;
    response->token_len = request->token_len;
    if (request->token_len > 0) {
        memcpy(response->token, request->token, request->token_len);
    }
    
    if (payload && strlen(payload) > 0) {
        size_t len = strlen(payload);
        uint16_t format = htons(COAP_CONTENT_FORMAT_TEXT_PLAIN);
        coap_message_add_option(response, COAP_OPTION_CONTENT_FORMAT, (uint8_t*)&format, 2);
        coap_message_set_payload(response, (const uint8_t*)payload, len);
    }
    
    return 0;
}

static int handle_sensor_report(const coap_message_t *request, coap_message_t *response) {
    printf("[INFO] Received sensor report (type: %s, code: %s)\n",
           coap_type_to_string(request->type), coap_code_to_string(request->code));
    
    sensor_reading_t reading;
    if (parse_sensor_report(request->payload, request->payload_len, &reading) < 0) {
        build_response_message(response, request, COAP_CODE_400_BAD_REQUEST, "Invalid payload format");
        return -1;
    }
    
    if (storage_add_reading(&g_storage, &reading) < 0) {
        build_response_message(response, request, COAP_CODE_500_INTERNAL_SERVER_ERROR, "Storage error");
        return -1;
    }
    
    char time_buf[32];
    print_timestamp(reading.timestamp, time_buf, sizeof(time_buf));
    
    printf("[SENSOR] ID: %s, Type: %s, ",
           reading.sensor_id, sensor_type_to_string(reading.type));
    if (reading.type == SENSOR_TYPE_TEMPERATURE) {
        printf("Temp: %.2f°C, ", reading.value.temperature);
    } else if (reading.type == SENSOR_TYPE_HUMIDITY) {
        printf("Humidity: %.2f%%, ", reading.value.humidity);
    } else {
        printf("State: %s, ", reading.value.is_open ? "OPEN" : "CLOSED");
    }
    printf("Battery: %d%%, Time: %s\n", reading.battery_percent, time_buf);
    
    int alarm_result = alarm_check_reading(&g_alarm_mgr, &reading);
    if (alarm_result == 1) {
        printf("[ALARM] TRIGGERED! Sensor %s - temperature %.2f°C\n",
               reading.sensor_id, reading.value.temperature);
    } else if (alarm_result == 2) {
        printf("[ALARM] RESOLVED! Sensor %s returned to normal\n", reading.sensor_id);
    }
    
    build_response_message(response, request, COAP_CODE_204_CHANGED, NULL);
    return 0;
}

static int handle_sensor_latest(const coap_message_t *request, coap_message_t *response,
                                 const char *sensor_id) {
    sensor_reading_t *reading = storage_get_latest(&g_storage, sensor_id);
    if (!reading) {
        build_response_message(response, request, COAP_CODE_404_NOT_FOUND, "Sensor not found");
        return 0;
    }
    
    char time_buf[32];
    print_timestamp(reading->timestamp, time_buf, sizeof(time_buf));
    
    char payload[512];
    if (reading->type == SENSOR_TYPE_TEMPERATURE) {
        snprintf(payload, sizeof(payload),
                 "sensor_id=%s\ntype=temperature\nvalue=%.2f\nbattery=%d\ntimestamp=%s",
                 reading->sensor_id, reading->value.temperature,
                 reading->battery_percent, time_buf);
    } else if (reading->type == SENSOR_TYPE_HUMIDITY) {
        snprintf(payload, sizeof(payload),
                 "sensor_id=%s\ntype=humidity\nvalue=%.2f\nbattery=%d\ntimestamp=%s",
                 reading->sensor_id, reading->value.humidity,
                 reading->battery_percent, time_buf);
    } else {
        snprintf(payload, sizeof(payload),
                 "sensor_id=%s\ntype=door_window\nvalue=%s\nbattery=%d\ntimestamp=%s",
                 reading->sensor_id, reading->value.is_open ? "1" : "0",
                 reading->battery_percent, time_buf);
    }
    
    build_response_message(response, request, COAP_CODE_205_CONTENT, payload);
    return 0;
}

static int handle_sensor_history(const coap_message_t *request, coap_message_t *response,
                                  const char *sensor_id) {
    uint32_t limit = 10;
    const char *n_param = get_uri_query_param(request, "n");
    if (n_param) {
        limit = (uint32_t)atoi(n_param);
        if (limit > 100) limit = 100;
    }
    
    sensor_reading_t *history = calloc(limit, sizeof(sensor_reading_t));
    if (!history) {
        build_response_message(response, request, COAP_CODE_500_INTERNAL_SERVER_ERROR, "Memory error");
        return -1;
    }
    
    uint32_t count = 0;
    storage_get_history(&g_storage, sensor_id, limit, history, &count);
    
    if (count == 0) {
        free(history);
        build_response_message(response, request, COAP_CODE_404_NOT_FOUND, "No history found");
        return 0;
    }
    
    char payload[2048];
    int pos = 0;
    pos += snprintf(payload + pos, sizeof(payload) - pos,
                    "count=%u\n", count);
    
    for (uint32_t i = 0; i < count; i++) {
        char time_buf[32];
        print_timestamp(history[i].timestamp, time_buf, sizeof(time_buf));
        
        if (history[i].type == SENSOR_TYPE_TEMPERATURE) {
            pos += snprintf(payload + pos, sizeof(payload) - pos,
                            "\n[%u] temp=%.2f battery=%d time=%s",
                            i, history[i].value.temperature,
                            history[i].battery_percent, time_buf);
        } else if (history[i].type == SENSOR_TYPE_HUMIDITY) {
            pos += snprintf(payload + pos, sizeof(payload) - pos,
                            "\n[%u] humidity=%.2f battery=%d time=%s",
                            i, history[i].value.humidity,
                            history[i].battery_percent, time_buf);
        } else {
            pos += snprintf(payload + pos, sizeof(payload) - pos,
                            "\n[%u] state=%s battery=%d time=%s",
                            i, history[i].value.is_open ? "OPEN" : "CLOSED",
                            history[i].battery_percent, time_buf);
        }
        
        if ((size_t)pos >= sizeof(payload) - 1) break;
    }
    
    free(history);
    build_response_message(response, request, COAP_CODE_205_CONTENT, payload);
    return 0;
}

static int handle_top_temperature(const coap_message_t *request, coap_message_t *response) {
    sensor_reading_t *top = calloc(TOP_TEMP_LIMIT, sizeof(sensor_reading_t));
    if (!top) {
        build_response_message(response, request, COAP_CODE_500_INTERNAL_SERVER_ERROR, "Memory error");
        return -1;
    }
    
    uint32_t count = 0;
    storage_get_top_temperature(&g_storage, TOP_TEMP_LIMIT, top, &count);
    
    char payload[1024];
    int pos = 0;
    pos += snprintf(payload + pos, sizeof(payload) - pos,
                    "Top %d temperature readings:\n", TOP_TEMP_LIMIT);
    
    if (count == 0) {
        pos += snprintf(payload + pos, sizeof(payload) - pos,
                        "(No temperature readings available)");
    } else {
        for (uint32_t i = 0; i < count; i++) {
            char time_buf[32];
            print_timestamp(top[i].timestamp, time_buf, sizeof(time_buf));
            
            pos += snprintf(payload + pos, sizeof(payload) - pos,
                            "[%u] %s: %.2f°C (battery: %d%%, time: %s)\n",
                            i + 1, top[i].sensor_id, top[i].value.temperature,
                            top[i].battery_percent, time_buf);
            
            if ((size_t)pos >= sizeof(payload) - 1) break;
        }
    }
    
    free(top);
    build_response_message(response, request, COAP_CODE_205_CONTENT, payload);
    return 0;
}

static int handle_alarm_query(const coap_message_t *request, coap_message_t *response) {
    alarm_record_t active[64];
    uint8_t count = 0;
    alarm_get_all_active(&g_alarm_mgr, active, &count);
    
    char payload[1024];
    int pos = 0;
    
    if (count == 0) {
        pos += snprintf(payload + pos, sizeof(payload) - pos,
                        "No active alarms\n");
    } else {
        pos += snprintf(payload + pos, sizeof(payload) - pos,
                        "Active alarms (%u):\n", count);
        
        for (uint8_t i = 0; i < count; i++) {
            char time_buf[32];
            print_timestamp(active[i].start_time, time_buf, sizeof(time_buf));
            
            pos += snprintf(payload + pos, sizeof(payload) - pos,
                            "[%u] %s: %s (%.2f°C, since %s)\n",
                            i + 1, active[i].sensor_id,
                            active[i].type == ALARM_TYPE_TEMP_HIGH ? "HIGH TEMP" : "LOW TEMP",
                            active[i].temperature, time_buf);
            
            if ((size_t)pos >= sizeof(payload) - 1) break;
        }
    }
    
    build_response_message(response, request, COAP_CODE_205_CONTENT, payload);
    return 0;
}

static int handle_request(const coap_message_t *request, coap_message_t *response) {
    char *segments[MAX_URI_PATH_SEGMENTS] = {NULL};
    uint8_t seg_count = 0;
    
    if (parse_uri_path(request, segments, &seg_count) < 0) {
        free_uri_segments(segments, seg_count);
        build_response_message(response, request, COAP_CODE_400_BAD_REQUEST, "Invalid URI");
        return -1;
    }
    
    if (seg_count == 0) {
        const char *welcome = 
            "CoAP Smart Home Sensor Server\n"
            "Endpoints:\n"
            "  POST /sensor/report  - Submit sensor reading\n"
            "  GET /sensor/{id}/latest - Get latest reading for sensor\n"
            "  GET /sensor/{id}/history?n=N - Get last N readings\n"
            "  GET /sensors/temperature/top5 - Get top 5 highest temperatures\n"
            "  GET /alarms - Get active alarms";
        
        build_response_message(response, request, COAP_CODE_205_CONTENT, welcome);
        free_uri_segments(segments, seg_count);
        return 0;
    }
    
    int result = 0;
    
    if (seg_count >= 2 && strcmp(segments[0], "sensor") == 0) {
        if (strcmp(segments[1], "report") == 0) {
            result = handle_sensor_report(request, response);
        } else if (seg_count >= 3) {
            const char *sensor_id = segments[1];
            if (strcmp(segments[2], "latest") == 0) {
                result = handle_sensor_latest(request, response, sensor_id);
            } else if (strcmp(segments[2], "history") == 0) {
                result = handle_sensor_history(request, response, sensor_id);
            } else {
                build_response_message(response, request, COAP_CODE_404_NOT_FOUND, "Endpoint not found");
            }
        } else {
            build_response_message(response, request, COAP_CODE_404_NOT_FOUND, "Endpoint not found");
        }
    } else if (seg_count >= 3 && strcmp(segments[0], "sensors") == 0 &&
               strcmp(segments[1], "temperature") == 0 &&
               strcmp(segments[2], "top5") == 0) {
        result = handle_top_temperature(request, response);
    } else if (seg_count >= 1 && strcmp(segments[0], "alarms") == 0) {
        result = handle_alarm_query(request, response);
    } else {
        build_response_message(response, request, COAP_CODE_404_NOT_FOUND, "Endpoint not found");
    }
    
    free_uri_segments(segments, seg_count);
    return result;
}

static void signal_handler(int sig) {
    (void)sig;
    g_running = 0;
    printf("\n[INFO] Shutting down...\n");
}

static uint16_t get_next_msg_id(void) {
    uint16_t id;
    pthread_mutex_lock(&g_msg_id_mutex);
    id = g_next_msg_id++;
    pthread_mutex_unlock(&g_msg_id_mutex);
    return id;
}

static int create_coap_message(coap_message_t *msg, coap_type_t type, coap_code_t code,
                                const char *path, const char *payload) {
    coap_message_init(msg);
    msg->version = COAP_VERSION;
    msg->type = type;
    msg->code = code;
    msg->msg_id = get_next_msg_id();
    
    if (path) {
        char path_copy[256];
        strncpy(path_copy, path, sizeof(path_copy) - 1);
        path_copy[sizeof(path_copy) - 1] = '\0';
        
        char *token;
        char *saveptr;
        token = strtok_r(path_copy, "/", &saveptr);
        while (token) {
            coap_message_add_option(msg, COAP_OPTION_URI_PATH,
                                    (const uint8_t*)token, strlen(token));
            token = strtok_r(NULL, "/", &saveptr);
        }
    }
    
    if (payload) {
        coap_message_set_payload(msg, (const uint8_t*)payload, strlen(payload));
    }
    
    return 0;
}

typedef struct {
    const char *sensor_id;
    sensor_type_t type;
    float base_value;
    struct sockaddr_in server_addr;
} simulation_config_t;

static void *simulate_sensor(void *arg) {
    simulation_config_t *config = (simulation_config_t*)arg;
    
    int sock = socket(AF_INET, SOCK_DGRAM, 0);
    if (sock < 0) {
        perror("Simulation socket");
        return NULL;
    }
    
    float current_value = config->base_value;
    int cycle_count = 0;
    
    while (g_running) {
        if (config->type == SENSOR_TYPE_TEMPERATURE) {
            float change = ((float)rand() / RAND_MAX - 0.5f) * 2.0f;
            current_value += change;
            
            if (cycle_count % 20 == 5) {
                if (strcmp(config->sensor_id, "temp_001") == 0) {
                    current_value = 45.0f;
                    printf("[SIM] %s: Simulating high temperature (45.0°C) for alarm test...\n",
                           config->sensor_id);
                } else if (strcmp(config->sensor_id, "temp_002") == 0) {
                    current_value = -5.0f;
                    printf("[SIM] %s: Simulating low temperature (-5.0°C) for alarm test...\n",
                           config->sensor_id);
                }
            } else if (cycle_count % 20 == 15) {
                current_value = config->base_value;
                printf("[SIM] %s: Returning to normal temperature...\n", config->sensor_id);
            }
            
            if (current_value > 60.0f) current_value = 60.0f;
            if (current_value < -20.0f) current_value = -20.0f;
        } else if (config->type == SENSOR_TYPE_HUMIDITY) {
            float change = ((float)rand() / RAND_MAX - 0.5f) * 5.0f;
            current_value += change;
            if (current_value > 95.0f) current_value = 95.0f;
            if (current_value < 30.0f) current_value = 30.0f;
        }
        
        uint8_t battery = 85 + (uint8_t)(rand() % 15);
        time_t now = time(NULL);
        
        char payload[256];
        if (config->type == SENSOR_TYPE_TEMPERATURE) {
            snprintf(payload, sizeof(payload),
                     "%s,temperature,%.2f,%d,%lld",
                     config->sensor_id, current_value, battery, (long long)now);
        } else if (config->type == SENSOR_TYPE_HUMIDITY) {
            snprintf(payload, sizeof(payload),
                     "%s,humidity,%.2f,%d,%lld",
                     config->sensor_id, current_value, battery, (long long)now);
        } else {
            bool is_open = (cycle_count % 10 < 5);
            snprintf(payload, sizeof(payload),
                     "%s,door_window,%s,%d,%lld",
                     config->sensor_id, is_open ? "1" : "0", battery, (long long)now);
        }
        
        coap_message_t msg;
        create_coap_message(&msg, COAP_TYPE_NON, COAP_CODE_POST, "/sensor/report", payload);
        
        uint8_t buffer[BUFFER_SIZE];
        int len = coap_message_encode(&msg, buffer, sizeof(buffer));
        if (len > 0) {
            sendto(sock, buffer, len, 0,
                   (struct sockaddr*)&config->server_addr,
                   sizeof(config->server_addr));
        }
        
        coap_message_free(&msg);
        cycle_count++;
        
        sleep(3);
    }
    
    close(sock);
    free(config);
    return NULL;
}

static void *simulate_queries(void *arg) {
    struct sockaddr_in *server_addr = (struct sockaddr_in*)arg;
    
    int sock = socket(AF_INET, SOCK_DGRAM, 0);
    if (sock < 0) {
        perror("Query simulation socket");
        return NULL;
    }
    
    struct timeval tv;
    tv.tv_sec = 2;
    tv.tv_usec = 0;
    setsockopt(sock, SOL_SOCKET, SO_RCVTIMEO, (const char*)&tv, sizeof(tv));
    
    int query_type = 0;
    const char *sensors[] = {"temp_001", "temp_002", "hum_001"};
    int sensor_count = 3;
    
    while (g_running) {
        sleep(8);
        
        coap_message_t msg;
        uint8_t buffer[BUFFER_SIZE];
        char path[128];
        
        switch (query_type % 4) {
            case 0:
                snprintf(path, sizeof(path), "/sensor/%s/latest",
                         sensors[query_type % sensor_count]);
                printf("\n[QUERY] GET %s (CON)\n", path);
                create_coap_message(&msg, COAP_TYPE_CON, COAP_CODE_GET, path, NULL);
                break;
                
            case 1:
                snprintf(path, sizeof(path), "/sensor/%s/history",
                         sensors[(query_type + 1) % sensor_count]);
                printf("\n[QUERY] GET %s?n=5 (CON)\n", path);
                create_coap_message(&msg, COAP_TYPE_CON, COAP_CODE_GET, path, NULL);
                coap_message_add_option(&msg, COAP_OPTION_URI_QUERY,
                                        (const uint8_t*)"n=5", 3);
                break;
                
            case 2:
                printf("\n[QUERY] GET /sensors/temperature/top5 (CON)\n");
                create_coap_message(&msg, COAP_TYPE_CON, COAP_CODE_GET,
                                    "/sensors/temperature/top5", NULL);
                break;
                
            case 3:
                printf("\n[QUERY] GET /alarms (CON)\n");
                create_coap_message(&msg, COAP_TYPE_CON, COAP_CODE_GET, "/alarms", NULL);
                break;
        }
        
        int len = coap_message_encode(&msg, buffer, sizeof(buffer));
        if (len > 0) {
            sendto(sock, buffer, len, 0,
                   (struct sockaddr*)server_addr,
                   sizeof(*server_addr));
            
            uint8_t resp_buffer[BUFFER_SIZE];
            socklen_t addr_len = sizeof(*server_addr);
            int recv_len = recvfrom(sock, resp_buffer, sizeof(resp_buffer), 0,
                                     (struct sockaddr*)server_addr, &addr_len);
            
            if (recv_len > 0) {
                coap_message_t response;
                if (coap_message_decode(&response, resp_buffer, recv_len) >= 0) {
                    printf("[RESPONSE] Type: %s, Code: %s\n",
                           coap_type_to_string(response.type),
                           coap_code_to_string(response.code));
                    if (response.payload_len > 0) {
                        printf("--- Response Body ---\n");
                        fwrite(response.payload, 1, response.payload_len, stdout);
                        printf("\n----------------------\n");
                    }
                    coap_message_free(&response);
                }
            } else {
                printf("[QUERY] Timeout or error\n");
            }
        }
        
        coap_message_free(&msg);
        query_type++;
    }
    
    close(sock);
    return NULL;
}

int main(int argc, char *argv[]) {
    (void)argc;
    (void)argv;
    
    printf("========================================\n");
    printf("  CoAP Smart Home Sensor Server\n");
    printf("========================================\n\n");
    
    storage_init(&g_storage);
    alarm_manager_init(&g_alarm_mgr);
    srand((unsigned int)time(NULL));
    
    g_sockfd = socket(AF_INET, SOCK_DGRAM, 0);
    if (g_sockfd < 0) {
        perror("socket");
        return 1;
    }
    
    int optval = 1;
    setsockopt(g_sockfd, SOL_SOCKET, SO_REUSEADDR, &optval, sizeof(optval));
    
    struct sockaddr_in server_addr;
    memset(&server_addr, 0, sizeof(server_addr));
    server_addr.sin_family = AF_INET;
    server_addr.sin_port = htons(COAP_PORT);
    server_addr.sin_addr.s_addr = INADDR_ANY;
    
    if (bind(g_sockfd, (struct sockaddr*)&server_addr, sizeof(server_addr)) < 0) {
        perror("bind");
        close(g_sockfd);
        return 1;
    }
    
    printf("[INFO] Server listening on UDP port %d\n", COAP_PORT);
    printf("[INFO] Storage capacity: %d readings\n", MAX_SENSOR_READINGS);
    printf("[INFO] Alarm threshold: temp > %.1f°C or < %.1f°C\n",
           TEMP_ALARM_HIGH_THRESHOLD, TEMP_ALARM_LOW_THRESHOLD);
    printf("[INFO] Consecutive readings for alarm: %d\n\n", ALARM_CONSECUTIVE_COUNT);
    
    signal(SIGINT, signal_handler);
    signal(SIGTERM, signal_handler);
    
    pthread_t sim_threads[SIMULATION_SENSOR_COUNT];
    pthread_t query_thread;
    
    const char *sensor_ids[] = {"temp_001", "temp_002", "hum_001", "hum_002", "door_001"};
    sensor_type_t types[] = {
        SENSOR_TYPE_TEMPERATURE,
        SENSOR_TYPE_TEMPERATURE,
        SENSOR_TYPE_HUMIDITY,
        SENSOR_TYPE_HUMIDITY,
        SENSOR_TYPE_DOOR_WINDOW
    };
    float base_values[] = {22.5f, 24.0f, 60.0f, 55.0f, 0.0f};
    
    for (int i = 0; i < SIMULATION_SENSOR_COUNT; i++) {
        simulation_config_t *config = malloc(sizeof(simulation_config_t));
        config->sensor_id = sensor_ids[i];
        config->type = types[i];
        config->base_value = base_values[i];
        memcpy(&config->server_addr, &server_addr, sizeof(server_addr));
        config->server_addr.sin_addr.s_addr = inet_addr("127.0.0.1");
        
        pthread_create(&sim_threads[i], NULL, simulate_sensor, config);
        printf("[INFO] Started simulation for sensor: %s (%s)\n",
               sensor_ids[i], sensor_type_to_string(types[i]));
    }
    
    struct sockaddr_in query_addr;
    memcpy(&query_addr, &server_addr, sizeof(query_addr));
    query_addr.sin_addr.s_addr = inet_addr("127.0.0.1");
    pthread_create(&query_thread, NULL, simulate_queries, &query_addr);
    printf("[INFO] Started query simulation thread\n\n");
    printf("----------------------------------------\n");
    printf("Press Ctrl+C to stop the server\n");
    printf("----------------------------------------\n\n");
    
    uint8_t buffer[BUFFER_SIZE];
    struct sockaddr_in client_addr;
    socklen_t addr_len = sizeof(client_addr);
    
    while (g_running) {
        fd_set readfds;
        struct timeval tv;
        
        FD_ZERO(&readfds);
        FD_SET(g_sockfd, &readfds);
        
        tv.tv_sec = 1;
        tv.tv_usec = 0;
        
        int sel = select(g_sockfd + 1, &readfds, NULL, NULL, &tv);
        if (sel < 0) {
            if (errno != EINTR) perror("select");
            continue;
        }
        if (sel == 0) continue;
        
        int recv_len = recvfrom(g_sockfd, buffer, sizeof(buffer), 0,
                                 (struct sockaddr*)&client_addr, &addr_len);
        if (recv_len < 0) {
            if (errno != EINTR && g_running) perror("recvfrom");
            continue;
        }
        
        coap_message_t request;
        if (coap_message_decode(&request, buffer, (size_t)recv_len) < 0) {
            printf("[WARN] Failed to decode CoAP message\n");
            continue;
        }
        
        coap_message_t response;
        handle_request(&request, &response);
        
        if (request.type == COAP_TYPE_CON) {
            uint8_t resp_buffer[BUFFER_SIZE];
            int resp_len = coap_message_encode(&response, resp_buffer, sizeof(resp_buffer));
            if (resp_len > 0) {
                sendto(g_sockfd, resp_buffer, resp_len, 0,
                       (struct sockaddr*)&client_addr, addr_len);
            }
        }
        
        coap_message_free(&request);
        coap_message_free(&response);
    }
    
    printf("[INFO] Waiting for threads to finish...\n");
    for (int i = 0; i < SIMULATION_SENSOR_COUNT; i++) {
        pthread_join(sim_threads[i], NULL);
    }
    pthread_join(query_thread, NULL);
    
    close(g_sockfd);
    printf("[INFO] Server stopped. Total readings received: %u\n", g_storage.total);
    
    return 0;
}
