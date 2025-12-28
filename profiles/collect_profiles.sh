#!/bin/bash
# profiles/collect_profiles.sh

set -e

echo "=== Memory Profiling Script ==="
echo

# Создаем директорию если не существует
mkdir -p profiles

echo "Step 1: Cleaning old profiles..."
rm -f profiles/*.pprof
echo

echo "Step 2: Starting server in background..."
go run ./cmd/shortener -a ":8081" &
SERVER_PID=$!

# Ждем запуска сервера
echo "Waiting for server to start..."
sleep 5

# Проверяем, что сервер запущен
if ! curl -s http://localhost:8081/ping > /dev/null; then
    echo "ERROR: Server failed to start"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

echo "Server started with PID: $SERVER_PID"
echo

echo "Step 3: Generating initial load..."
# Запускаем нагрузочный тест
go run ./scripts/load_generator.go
echo

echo "Step 4: Collecting BASELINE memory profile..."
curl -s "http://localhost:8081/debug/pprof/heap" > profiles/base.pprof
echo "Baseline profile saved: profiles/base.pprof"
echo

echo "Step 5: Collecting additional profiles..."
# Собираем различные типы профилей для полного анализа
curl -s "http://localhost:8081/debug/pprof/allocs" > profiles/base_allocs.pprof
curl -s "http://localhost:8081/debug/pprof/goroutine" > profiles/goroutine.pprof
echo

echo "Step 6: Generating more load for stress testing..."
# Дополнительная нагрузка
for i in {1..500}; do
    curl -s -X POST "http://localhost:8081/api/shorten/batch" \
        -H "Content-Type: application/json" \
        -d "[
            {\"correlation_id\": \"$i-1\", \"original_url\": \"https://batch1.example.com/test/$i\"},
            {\"correlation_id\": \"$i-2\", \"original_url\": \"https://batch2.example.com/test/$i\"}
        ]" > /dev/null
done
echo

echo "Step 7: Collecting STRESS memory profile..."
curl -s "http://localhost:8081/debug/pprof/heap" > profiles/stress.pprof
echo "Stress profile saved: profiles/stress.pprof"
echo

echo "Step 8: Stopping server..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null
echo

echo "Step 9: Analyzing baseline profile..."
echo "--- Top 20 memory allocations ---"
go tool pprof -top profiles/base.pprof 2>/dev/null | head -30 || true
echo

echo "Step 10: Creating analysis report..."
cat > profiles/ANALYSIS.md << 'EOF'
# Memory Profile Analysis

## Collected Profiles

1. `base.pprof` - Baseline memory profile after initial load
2. `base_allocs.pprof` - Allocation profile
3. `goroutine.pprof` - Goroutine profile
4. `stress.pprof` - Memory profile after stress testing

## Quick Analysis Commands

```bash
# Analyze heap profile
go tool pprof profiles/base.pprof

# Compare baseline and stress profiles
go tool pprof -top -diff_base=profiles/base.pprof profiles/stress.pprof

# Web interface
go tool pprof -http=:8082 profiles/base.pprof

# Look at specific functions
go tool pprof -list=<function_name> profiles/base.pprof
