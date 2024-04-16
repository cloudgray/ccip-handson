CHAINLINK_DIR=${CHAINLINK_DIR}

# Postgres DB
docker run --name cl-postgres -e POSTGRES_PASSWORD=mysecretpassword -p 5432:5432 -d postgres

sleep 3

# Chainlink Node 
chainlink node -config $CHAINLINK_DIR/config.toml -secrets $CHAINLINK_DIR/secrets.toml start -a $CHAINLINK_DIR/.api