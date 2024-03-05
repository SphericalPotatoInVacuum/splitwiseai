#!/bin/bash

set -e

# The first argument is the output zip file path
output_zip=$1
# The second argument is the working directory
workdir=$2
# The second argument is the comma-separated list of files and directories
files_dirs_comma_separated=$3

# Convert the comma-separated list back into an array
IFS=',' read -r -a files_dirs <<< "$files_dirs_comma_separated"

# Change to the working directory
cd "$workdir"

# Zip the specified files and directories
zip -FSqr "$output_zip" "${files_dirs[@]}"

# Calculate the hash of the zip file
hash=$(sha256sum "$output_zip" | cut -d ' ' -f1)

# Output the hash so it can be captured by Terraform
echo "{\"hash\":\"$hash\"}"
