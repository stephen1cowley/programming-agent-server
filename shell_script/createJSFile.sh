#!/bin/bash

# Navigate to the directory
cd ~/my-react-app/src

# Create the .js file
cat <<EOL > $1.js
$2
EOL

echo "$1.js file created at ~/my-react-app/src/$1.js"
