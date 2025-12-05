#!/bin/bash
set -e

# Check for Go
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Please install Go manually."
    exit 1
fi

# Check/Install Task
TASK_CMD="task"

if command -v task &> /dev/null; then
    echo "Task is already installed globally."
elif [ -x "./bin/task" ]; then
    echo "Task found in ./bin/task."
    TASK_CMD="./bin/task"
else
    echo "Task is not installed. Installing..."
    if [ -w /usr/local/bin ]; then
        sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin
    else
        echo "Cannot write to /usr/local/bin. Installing to ./bin"
        mkdir -p bin
        sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ./bin
        TASK_CMD="./bin/task"
        echo "Task installed to ./bin/task. Please add ./bin to your PATH."
    fi
fi

echo "Running task setup..."
$TASK_CMD setup
