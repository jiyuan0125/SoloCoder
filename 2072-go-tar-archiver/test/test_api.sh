#!/bin/bash

set -e

echo "=== Starting TAR Archiver API Tests ==="

# Clean up previous test files
rm -rf test_source test_target test_archive.tar.gz test_archive2.tar.gz tar_archiver.db

# Create test source directory with some files
mkdir -p test_source/subdir

echo "Test file 1" > test_source/file1.txt
echo "Test file 2" > test_source/file2.txt
echo "Test file in subdir" > test_source/subdir/file3.txt
chmod 644 test_source/file1.txt
chmod 755 test_source/subdir/file3.txt

# Start server in background
echo "Starting server..."
./tar-archiver &
SERVER_PID=$!
sleep 2

echo ""
echo "=== Test 1: Create Archive ==="
curl -s -X POST http://localhost:8080/archive \
  -H "Content-Type: application/json" \
  -d '{
    "source_dir": "./test_source",
    "archive_path": "./test_archive.tar.gz",
    "password": "test123"
  }'

echo ""
echo ""
echo "=== Test 2: List Archive Contents ==="
curl -s -X POST http://localhost:8080/list \
  -H "Content-Type: application/json" \
  -d '{
    "archive_path": "./test_archive.tar.gz",
    "password": "test123"
  }'

echo ""
echo ""
echo "=== Test 3: Extract Archive ==="
mkdir -p test_target
curl -s -X POST http://localhost:8080/extract \
  -H "Content-Type: application/json" \
  -d '{
    "archive_path": "./test_archive.tar.gz",
    "target_dir": "./test_target",
    "password": "test123"
  }'

echo ""
echo ""
echo "=== Test 4: Verify Extracted Files ==="
echo "Extracted files:"
find test_target -type f

echo ""
echo "File permissions:"
ls -la test_source/file1.txt
ls -la test_target/file1.txt

echo ""
echo "File contents:"
cat test_target/file1.txt
cat test_target/subdir/file3.txt

echo ""
echo ""
echo "=== Test 5: Test Error - Source Directory Not Exist ==="
curl -s -X POST http://localhost:8080/archive \
  -H "Content-Type: application/json" \
  -d '{
    "source_dir": "./nonexistent_dir",
    "archive_path": "./test.tar.gz",
    "password": "test123"
  }'

echo ""
echo ""
echo "=== Test 6: Test Error - Empty Password ==="
curl -s -X POST http://localhost:8080/archive \
  -H "Content-Type: application/json" \
  -d '{
    "source_dir": "./test_source",
    "archive_path": "./test.tar.gz",
    "password": ""
  }'

echo ""
echo ""
echo "=== Test 7: Test Skip Existing Files ==="
echo "Existing file" > test_target/file1.txt
curl -s -X POST http://localhost:8080/extract \
  -H "Content-Type: application/json" \
  -d '{
    "archive_path": "./test_archive.tar.gz",
    "target_dir": "./test_target",
    "password": "test123"
  }'

echo ""
echo "Content of file1.txt (should be 'Existing file'):"
cat test_target/file1.txt

echo ""
echo ""
echo "=== Test 8: Incremental Archive ==="
sleep 1
# Touch one file to make it newer
touch -m -d "2030-01-01" test_source/file2.txt
# Archive only files modified after 2025-01-01
curl -s -X POST http://localhost:8080/archive \
  -H "Content-Type: application/json" \
  -d '{
    "source_dir": "./test_source",
    "archive_path": "./test_archive2.tar.gz",
    "incremental_date": "2025-01-01T00:00:00Z",
    "password": "test123"
  }'

echo ""
echo ""
echo "=== Cleanup ==="
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true

# Clean up test files
rm -rf test_source test_target test_archive.tar.gz test_archive2.tar.gz tar_archiver.db

echo ""
echo "=== All Tests Completed ==="
