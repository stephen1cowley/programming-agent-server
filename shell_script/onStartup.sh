#!/bin/bash

# Directory to clean up
DIR="/home/ubuntu/my-react-app/src"

# Files to keep (space-separated list)
KEEP_FILES=("App.js" "App.css" "index.js" "index.css" "reportWebVitals.js")

# Create a find command pattern for the files to keep
KEEP_PATTERN=""
for FILE in "${KEEP_FILES[@]}"; do
    KEEP_PATTERN+=" -name $FILE -o"
done
# Remove the trailing -o
KEEP_PATTERN=${KEEP_PATTERN::-2}

# Find and delete files not in the keep list
find "$DIR" -type f ! \( $KEEP_PATTERN \) -exec rm -f {} +

echo "Cleanup complete. Kept files: ${KEEP_FILES[*]}"
