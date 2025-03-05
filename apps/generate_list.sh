#!/bin/sh

INPUT_DIR="packages"
OUTPUT_FILE="list.json"

results=""

# Loop through all JSON files in the input directory
for app_dir in "$INPUT_DIR"/*; do
    # Check if the file exists
    json_file="$app_dir/schema.json"
    if [ -f "$json_file" ]; then
        # Extract id, name, version, author, and description using a single jq call
        parsed_data=$(jq -c '{id: .id, name: .name, version: .version, author: .author, description: .description, categories: .categories}' "$json_file")

        if [ -z "$results" ]; then
            results="$parsed_data"
        else
            results="$results,$parsed_data"
        fi
    fi
done

# Create the output JSON array
output_json="[$results]"

# Write the output to the specified file using echo
echo "$output_json" | jq . > "$OUTPUT_FILE"

echo "Parsed data written to $OUTPUT_FILE"
