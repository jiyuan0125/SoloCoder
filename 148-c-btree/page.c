#include "page.h"

void page_init_internal(page_t *page, page_num_t page_num) {
    page_num_t invalid;
    memset(page, 0, PAGE_SIZE);
    page->page_num = page_num;
    page->type = PAGE_TYPE_INTERNAL;
    page->num_entries = 0;
    page->next_leaf = INVALID_PAGE_NUM;
    page->prev_leaf = INVALID_PAGE_NUM;
    invalid = INVALID_PAGE_NUM;
    memcpy(page->data, &invalid, sizeof(page_num_t));
}

void page_init_leaf(page_t *page, page_num_t page_num) {
    memset(page, 0, PAGE_SIZE);
    page->page_num = page_num;
    page->type = PAGE_TYPE_LEAF;
    page->num_entries = 0;
    page->next_leaf = INVALID_PAGE_NUM;
    page->prev_leaf = INVALID_PAGE_NUM;
}

int page_insert_internal(page_t *page, const bt_key_t *key, page_num_t left_child, page_num_t right_child) {
    uint16_t child_idx;
    int found;
    uint16_t max_entries;
    internal_entry_t *entries;

    max_entries = page_get_max_entries_internal();
    if (page->num_entries >= max_entries) {
        return -1;
    }

    found = page_find_internal(page, key, &child_idx);

    entries = (internal_entry_t *)(page->data + sizeof(page_num_t));

    if (found) {
        if (child_idx > 0 && child_idx <= page->num_entries) {
            entries[child_idx - 1].child = right_child;
        }
        return 0;
    }

    if (child_idx == 0) {
        if (page->num_entries > 0) {
            memmove(&entries[1], &entries[0],
                    page->num_entries * sizeof(internal_entry_t));
        }
        page_set_child(page, 0, left_child);
        key_copy(&entries[0].key, key);
        entries[0].child = right_child;
    } else if (child_idx > 0 && child_idx <= page->num_entries) {
        if (child_idx <= page->num_entries) {
            memmove(&entries[child_idx], &entries[child_idx - 1],
                    (page->num_entries - child_idx + 1) * sizeof(internal_entry_t));
        }
        entries[child_idx - 1].child = left_child;
        key_copy(&entries[child_idx].key, key);
        entries[child_idx].child = right_child;
    } else {
        key_copy(&entries[page->num_entries].key, key);
        entries[page->num_entries].child = right_child;
    }

    page->num_entries++;

    return 0;
}

int page_insert_leaf(page_t *page, const bt_key_t *key, page_num_t value_page, uint32_t value_len) {
    uint16_t idx;
    int found;
    uint16_t max_entries;
    leaf_entry_t *entries;

    max_entries = page_get_max_entries_leaf();
    if (page->num_entries >= max_entries) {
        return -1;
    }

    found = page_find_leaf(page, key, &idx);
    entries = (leaf_entry_t *)page->data;

    if (found) {
        entries[idx].value_page = value_page;
        entries[idx].value_len = value_len;
        return 0;
    }

    if (idx < page->num_entries) {
        memmove(&entries[idx + 1], &entries[idx],
                (page->num_entries - idx) * sizeof(leaf_entry_t));
    }

    key_copy(&entries[idx].key, key);
    entries[idx].value_page = value_page;
    entries[idx].value_len = value_len;
    page->num_entries++;

    return 0;
}

int page_update_leaf(page_t *page, uint16_t idx, page_num_t value_page, uint32_t value_len) {
    leaf_entry_t *entries;

    if (page->type != PAGE_TYPE_LEAF || idx >= page->num_entries) {
        return -1;
    }

    entries = (leaf_entry_t *)page->data;
    entries[idx].value_page = value_page;
    entries[idx].value_len = value_len;

    return 0;
}

int page_remove_internal(page_t *page, uint16_t idx) {
    internal_entry_t *entries;

    if (page->type != PAGE_TYPE_INTERNAL || idx >= page->num_entries) {
        return -1;
    }

    entries = (internal_entry_t *)(page->data + sizeof(page_num_t));
    if (idx < page->num_entries - 1) {
        memmove(&entries[idx], &entries[idx + 1],
                (page->num_entries - idx - 1) * sizeof(internal_entry_t));
    }
    page->num_entries--;

    return 0;
}

int page_remove_leaf(page_t *page, uint16_t idx) {
    leaf_entry_t *entries;

    if (page->type != PAGE_TYPE_LEAF || idx >= page->num_entries) {
        return -1;
    }

    entries = (leaf_entry_t *)page->data;
    if (idx < page->num_entries - 1) {
        memmove(&entries[idx], &entries[idx + 1],
                (page->num_entries - idx - 1) * sizeof(leaf_entry_t));
    }
    page->num_entries--;

    return 0;
}

int page_find_internal(const page_t *page, const bt_key_t *key, uint16_t *out_child_idx) {
    int left, right, mid;
    int cmp;
    const internal_entry_t *entries;

    if (page->type != PAGE_TYPE_INTERNAL) {
        *out_child_idx = 0;
        return 0;
    }

    if (page->num_entries == 0) {
        *out_child_idx = 0;
        return 0;
    }

    entries = (const internal_entry_t *)(page->data + sizeof(page_num_t));
    left = 0;
    right = page->num_entries - 1;

    while (left <= right) {
        mid = (left + right) / 2;
        cmp = key_compare(key, &entries[mid].key);

        if (cmp == 0) {
            *out_child_idx = mid + 1;
            return 1;
        } else if (cmp < 0) {
            right = mid - 1;
        } else {
            left = mid + 1;
        }
    }

    *out_child_idx = left;
    return 0;
}

int page_find_leaf(const page_t *page, const bt_key_t *key, uint16_t *out_idx) {
    int left, right, mid;
    int cmp;
    const leaf_entry_t *entries;

    if (page->type != PAGE_TYPE_LEAF) {
        *out_idx = 0;
        return 0;
    }

    entries = (const leaf_entry_t *)page->data;
    left = 0;
    right = page->num_entries - 1;

    while (left <= right) {
        mid = (left + right) / 2;
        cmp = key_compare(key, &entries[mid].key);

        if (cmp == 0) {
            *out_idx = mid;
            return 1;
        } else if (cmp < 0) {
            right = mid - 1;
        } else {
            left = mid + 1;
        }
    }

    *out_idx = left;
    return 0;
}

int page_split_internal(page_t *old_page, page_t *new_page, bt_key_t *out_mid_key) {
    uint16_t mid, i;
    internal_entry_t *old_entries, *new_entries;
    page_num_t first_child;

    if (old_page->type != PAGE_TYPE_INTERNAL) {
        return -1;
    }

    mid = old_page->num_entries / 2;
    old_entries = (internal_entry_t *)(old_page->data + sizeof(page_num_t));
    new_entries = (internal_entry_t *)(new_page->data + sizeof(page_num_t));

    key_copy(out_mid_key, &old_entries[mid].key);

    first_child = old_entries[mid].child;
    memcpy(new_page->data, &first_child, sizeof(page_num_t));

    for (i = 0; i < old_page->num_entries - mid - 1; i++) {
        key_copy(&new_entries[i].key, &old_entries[mid + 1 + i].key);
        new_entries[i].child = old_entries[mid + 1 + i].child;
    }

    new_page->num_entries = old_page->num_entries - mid - 1;
    old_page->num_entries = mid;

    return 0;
}

int page_split_leaf(page_t *old_page, page_t *new_page, bt_key_t *out_mid_key) {
    uint16_t mid, i;
    leaf_entry_t *old_entries, *new_entries;

    if (old_page->type != PAGE_TYPE_LEAF) {
        return -1;
    }

    mid = old_page->num_entries / 2;
    old_entries = (leaf_entry_t *)old_page->data;
    new_entries = (leaf_entry_t *)new_page->data;

    key_copy(out_mid_key, &old_entries[mid].key);

    for (i = 0; i < old_page->num_entries - mid; i++) {
        key_copy(&new_entries[i].key, &old_entries[mid + i].key);
        new_entries[i].value_page = old_entries[mid + i].value_page;
        new_entries[i].value_len = old_entries[mid + i].value_len;
    }

    new_page->num_entries = old_page->num_entries - mid;
    old_page->num_entries = mid;

    new_page->next_leaf = old_page->next_leaf;
    new_page->prev_leaf = old_page->page_num;
    old_page->next_leaf = new_page->page_num;

    return 0;
}

int page_merge_internal(page_t *left_page, page_t *right_page, const bt_key_t *sep_key) {
    internal_entry_t *left_entries, *right_entries;
    uint16_t i;
    page_num_t right_first_child;

    if (left_page->type != PAGE_TYPE_INTERNAL || right_page->type != PAGE_TYPE_INTERNAL) {
        return -1;
    }

    left_entries = (internal_entry_t *)(left_page->data + sizeof(page_num_t));
    right_entries = (internal_entry_t *)(right_page->data + sizeof(page_num_t));

    key_copy(&left_entries[left_page->num_entries].key, sep_key);
    memcpy(&right_first_child, right_page->data, sizeof(page_num_t));
    left_entries[left_page->num_entries].child = right_first_child;
    left_page->num_entries++;

    for (i = 0; i < right_page->num_entries; i++) {
        key_copy(&left_entries[left_page->num_entries].key, &right_entries[i].key);
        left_entries[left_page->num_entries].child = right_entries[i].child;
        left_page->num_entries++;
    }

    return 0;
}

int page_merge_leaf(page_t *left_page, page_t *right_page) {
    leaf_entry_t *left_entries, *right_entries;
    uint16_t i;

    if (left_page->type != PAGE_TYPE_LEAF || right_page->type != PAGE_TYPE_LEAF) {
        return -1;
    }

    left_entries = (leaf_entry_t *)left_page->data;
    right_entries = (leaf_entry_t *)right_page->data;

    for (i = 0; i < right_page->num_entries; i++) {
        key_copy(&left_entries[left_page->num_entries].key, &right_entries[i].key);
        left_entries[left_page->num_entries].value_page = right_entries[i].value_page;
        left_entries[left_page->num_entries].value_len = right_entries[i].value_len;
        left_page->num_entries++;
    }

    left_page->next_leaf = right_page->next_leaf;

    return 0;
}

int page_read(int fd, page_num_t page_num, page_t *page) {
    off_t offset;
    ssize_t nread;

    offset = (off_t)page_num * PAGE_SIZE;
    if (lseek(fd, offset, SEEK_SET) != offset) {
        return -1;
    }

    nread = read(fd, page, PAGE_SIZE);
    if (nread != PAGE_SIZE) {
        return -1;
    }

    return 0;
}

int page_write(int fd, page_num_t page_num, const page_t *page) {
    off_t offset;
    ssize_t nwritten;

    offset = (off_t)page_num * PAGE_SIZE;
    if (lseek(fd, offset, SEEK_SET) != offset) {
        return -1;
    }

    nwritten = write(fd, page, PAGE_SIZE);
    if (nwritten != PAGE_SIZE) {
        return -1;
    }

    if (fsync(fd) < 0) {
        return -1;
    }

    return 0;
}

int value_page_read(int fd, page_num_t page_num, bt_value_t *value) {
    off_t offset;
    ssize_t nread;
    uint8_t buffer[PAGE_SIZE];

    offset = (off_t)page_num * PAGE_SIZE;
    if (lseek(fd, offset, SEEK_SET) != offset) {
        return -1;
    }

    nread = read(fd, buffer, PAGE_SIZE);
    if (nread != PAGE_SIZE) {
        return -1;
    }

    memcpy(&value->value_len, buffer, sizeof(uint32_t));
    if (value->value_len > MAX_VALUE_SIZE) {
        return -1;
    }
    memcpy(value->value_data, buffer + sizeof(uint32_t), value->value_len);

    return 0;
}

int value_page_write(int fd, page_num_t page_num, const bt_value_t *value) {
    off_t offset;
    ssize_t nwritten;
    uint8_t buffer[PAGE_SIZE];

    if (value->value_len > MAX_VALUE_SIZE) {
        return -1;
    }

    memset(buffer, 0, PAGE_SIZE);
    memcpy(buffer, &value->value_len, sizeof(uint32_t));
    memcpy(buffer + sizeof(uint32_t), value->value_data, value->value_len);

    offset = (off_t)page_num * PAGE_SIZE;
    if (lseek(fd, offset, SEEK_SET) != offset) {
        return -1;
    }

    nwritten = write(fd, buffer, PAGE_SIZE);
    if (nwritten != PAGE_SIZE) {
        return -1;
    }

    if (fsync(fd) < 0) {
        return -1;
    }

    return 0;
}

int value_page_update_partial(int fd, page_num_t page_num, uint32_t offset, const uint8_t *data, uint32_t len) {
    off_t file_offset;
    ssize_t nwritten;
    bt_value_t value;

    if (value_page_read(fd, page_num, &value) < 0) {
        return -1;
    }

    if (offset + len > value.value_len) {
        return -1;
    }

    memcpy(value.value_data + offset, data, len);

    return value_page_write(fd, page_num, &value);
}

page_num_t page_allocate(btree_t *btree) {
    page_num_t new_page;
    uint32_t next_free;
    page_t page;

    if (btree->header.freelist_head != INVALID_PAGE_NUM) {
        new_page = btree->header.freelist_head;
        if (page_read(btree->data_fd, new_page, &page) < 0) {
            return INVALID_PAGE_NUM;
        }
        next_free = *(uint32_t *)page.data;
        btree->header.freelist_head = next_free;
        btree_write_header(btree);
        return new_page;
    }

    new_page = btree->header.num_pages;
    btree->header.num_pages++;
    btree_write_header(btree);

    return new_page;
}

void page_deallocate(btree_t *btree, page_num_t page_num) {
    page_t page;
    uint32_t *free_list_ptr;

    memset(&page, 0, PAGE_SIZE);
    page.page_num = page_num;
    free_list_ptr = (uint32_t *)page.data;
    *free_list_ptr = btree->header.freelist_head;

    btree->header.freelist_head = page_num;
    page_write(btree->data_fd, page_num, &page);
    btree_write_header(btree);
}

int btree_read_header(btree_t *btree) {
    file_header_t header;
    off_t offset;
    ssize_t nread;

    offset = 0;
    if (lseek(btree->data_fd, offset, SEEK_SET) != offset) {
        return -1;
    }

    nread = read(btree->data_fd, &header, sizeof(file_header_t));
    if (nread != sizeof(file_header_t)) {
        return -1;
    }

    memcpy(&btree->header, &header, sizeof(file_header_t));
    return 0;
}

int btree_write_header(btree_t *btree) {
    off_t offset;
    ssize_t nwritten;

    offset = 0;
    if (lseek(btree->data_fd, offset, SEEK_SET) != offset) {
        return -1;
    }

    nwritten = write(btree->data_fd, &btree->header, sizeof(file_header_t));
    if (nwritten != sizeof(file_header_t)) {
        return -1;
    }

    if (fsync(btree->data_fd) < 0) {
        return -1;
    }

    return 0;
}
