#!/bin/bash

README_FILE="benchmarks/README.md"

echo "=========================================="
echo "WASM Benchmark: JSON vs Stdlib JSON"
echo "=========================================="

# Check if wasmbrowsertest is installed
if ! command -v wasmbrowsertest &> /dev/null; then
    echo "⚠️  wasmbrowsertest not found. Install it with:"
    echo "   go install github.com/agnivade/wasmbrowsertest@latest"
    echo "   export PATH=\$PATH:\$(go env GOPATH)/bin"
    exit 1
fi

echo ""
echo "Running WASM benchmarks..."
echo ""

# Run WASM benchmarks and capture output
BENCH_OUTPUT=$(GOOS=js GOARCH=wasm go test -bench=. -benchmem -benchtime=10x -tags wasm 2>&1 | grep -E "(Benchmark|PASS|FAIL|ns/op|B/op|allocs/op)")

if [ ${PIPESTATUS[0]} -ne 0 ]; then
    echo ""
    echo "❌ Benchmark failed"
    exit 1
fi

# Display results
echo "$BENCH_OUTPUT"

echo ""
echo "✅ Benchmark completed!"
echo ""
echo "📊 Results interpretation:"
echo "   - Lower ns/op = faster performance"
echo "   - Lower B/op = less memory allocated"
echo "   - Lower allocs/op = fewer allocations"

# Update README with results
if [ -f "$README_FILE" ]; then
    echo ""
    echo "Updating $README_FILE with latest results..."
    
    # Create temporary file with new results
    TEMP_FILE=$(mktemp)
    
    # Extract everything before the Performance Results section
    sed -n '1,/## Performance Results/p' "$README_FILE" | head -n -1 > "$TEMP_FILE"
    
    # Add Performance Results section with current date
    echo "## Performance Results" >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
    echo "Last updated: $(date '+%Y-%m-%d %H:%M:%S')" >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
    echo '```' >> "$TEMP_FILE"
    echo "$BENCH_OUTPUT" >> "$TEMP_FILE"
    echo '```' >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
    
    # Add the rest of the file after Performance Results (if exists)
    sed -n '/## Performance Results/,$ p' "$README_FILE" | sed '1,/^```$/d' | sed '1,/^```$/d' | sed '1d' >> "$TEMP_FILE" 2>/dev/null || true
    
    # Replace original file
    mv "$TEMP_FILE" "$README_FILE"
    
    echo "✅ README updated successfully!"
else
    echo "⚠️  README file not found at $README_FILE"
fi
