#!/bin/bash

# Navigate to the directory
cd ~/my-react-app/src

# Create the App.js file
cat <<EOL > App.js
$1
EOL

echo "App.js file updated at ~/my-react-app/src/App.js"
