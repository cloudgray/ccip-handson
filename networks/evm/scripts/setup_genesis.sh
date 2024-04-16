#!/bin/bash

# Get the directory of the script
SCRIPT_DIR=$(dirname "$0")
TEMPLATE_DIR="$SCRIPT_DIR/../templates"

# Create arrays for environment variables for chains A and B
chain_ids=("CHAIN_ID_A" "CHAIN_ID_B")
etherbases=("ETHERBASE_A" "ETHERBASE_B")
chain_dirs=("CHAIN_DIR_A" "CHAIN_DIR_B")

# Process chains A and B in a loop
for i in {0..1}
do
    # Read environment variables
    chain_id=${!chain_ids[$i]}
    etherbase=${!etherbases[$i]}
    chain_dir=${!chain_dirs[$i]}

    # Remove '0x' prefix from ETHERBASE
    etherbase_trimmed=${etherbase:2}

    # Check if the directory exists
    if [ -d "$chain_dir" ]; then
        echo "The directory $chain_dir already exists."
        read -p "Do you want to remove it and recreate it? (y/n): " answer
        if [[ "$answer" == [Yy]* ]]; then
            echo "Removing $chain_dir..."
            rm -rf "$chain_dir"
            echo "Creating $chain_dir..."
            mkdir -p "$chain_dir"
        else
            echo "Exiting without making changes."
            exit 1
        fi
    else
        echo "Creating $chain_dir..."
        mkdir -p "$chain_dir"
    fi

    # Set the location for creating the genesis.json file
    genesis_file="$chain_dir/genesis.json"

    # Copy the template genesis.json file to the target location
    cp $TEMPLATE_DIR/example.genesis.json $genesis_file

    # Modify the copied genesis file using sed
    sed -i '' "s/\"\\\$CHAIN_ID\"/$chain_id/g" $genesis_file
    sed -i '' "s/\$ETHERBASE_TRIMMED/$etherbase_trimmed/g" $genesis_file
    sed -i '' "s/\$ETHERBASE/$etherbase/g" $genesis_file

    echo "genesis.json has been created at $genesis_file."
done
