#!/bin/bash

# 환경 변수
SCRIPT_DIR=$(dirname "$0")
TEMPLATES_DIR=$SCRIPT_DIR/../templates

CHAINLINK_DIR=${CHAINLINK_DIR}
CHAIN_ID_A=${CHAIN_ID_A}
HTTP_PORT_A=${HTTP_PORT_A}
WS_PORT_A=${WS_PORT_A}
CHAIN_ID_B=${CHAIN_ID_B}
HTTP_PORT_B=${HTTP_PORT_B}
WS_PORT_B=${WS_PORT_B}
TRANSMITTER_ADDRESS=${TRANSMITTER_ADDRESS}

# 디렉토리 확인 및 조건부 삭제
if [ -d "$CHAINLINK_DIR" ]; then
    echo "The directory $CHAINLINK_DIR already exists."
    read -p "Do you want to remove existing config files and recreate new ones? (y/n): " answer
    if [[ "$answer" == [Yy]* ]]; then
        echo "Removing $CHAINLINK_DIR..."
        rm -rf "$CHAINLINK_DIR"
        echo "Creating $CHAINLINK_DIR..."
        mkdir -p "$CHAINLINK_DIR"
    else
        echo "Exiting without making changes."
        exit 1
    fi
else
    echo "Creating $CHAINLINK_DIR..."
    mkdir -p "$CHAINLINK_DIR"
fi

# Copy and modify the template/example.config.toml file to config.toml
cp $TEMPLATES_DIR/example.config.toml "$CHAINLINK_DIR/config.toml"
sed -i '' "s/\$CHAIN_ID_A/$CHAIN_ID_A/g" "$CHAINLINK_DIR/config.toml"
sed -i '' "s/\$HTTP_PORT_A/$HTTP_PORT_A/g" "$CHAINLINK_DIR/config.toml"
sed -i '' "s/\$WS_PORT_A/$WS_PORT_A/g" "$CHAINLINK_DIR/config.toml"
sed -i '' "s/\$CHAIN_ID_B/$CHAIN_ID_B/g" "$CHAINLINK_DIR/config.toml"
sed -i '' "s/\$HTTP_PORT_B/$HTTP_PORT_B/g" "$CHAINLINK_DIR/config.toml"
sed -i '' "s/\$WS_PORT_B/$WS_PORT_B/g" "$CHAINLINK_DIR/config.toml"
sed -i '' "s/\$TRANSMITTER_ADDRESS/$TRANSMITTER_ADDRESS/g" "$CHAINLINK_DIR/config.toml"

# .api 파일 생성 및 이름 변경
cp $TEMPLATES_DIR/example.api "$CHAINLINK_DIR/.api"

# secrets.toml 파일 생성 및 이름 변경
cp $TEMPLATES_DIR/example.secrets.toml "$CHAINLINK_DIR/secrets.toml"

echo "Configuration files have been successfully created and modified in $CHAINLINK_DIR."
