#!/bin/bash

# Navigate to the directory
cd /home/ubuntu/my-react-app/src

# Create the App.css file
cat <<EOL > App.css
$1
EOL

echo "App.css file updated at ~/my-react-app/src/App.css"
# TEST CHANGE