#========================================= Postgres DB =========================================#
docker run --name cl-postgres-0 -e POSTGRES_PASSWORD=mysecretpassword -p 5432:5432 -d postgres
docker run --name cl-postgres-1 -e POSTGRES_PASSWORD=mysecretpassword -p 15432:5432 -d postgres
docker run --name cl-postgres-2 -e POSTGRES_PASSWORD=mysecretpassword -p 25432:5432 -d postgres


#======================================== Chainlink Node ========================================#
## Remove existing chainlink node data
rm -rf $HOME/.chainlink-handson/node-0
rm -rf $HOME/.chainlink-handson/node-1
rm -rf $HOME/.chainlink-handson/node-2

## Create chainlink node data directories
mkdir -p $HOME/.chainlink-handson/node-0
mkdir -p $HOME/.chainlink-handson/node-1
mkdir -p $HOME/.chainlink-handson/node-2

## Run chainlink nodes
docker run -d --platform linux/amd64 --name chainlink-0 -v $HOME/.chainlink-handson/node-0:/chainlink -it -p 6688:6688 --add-host=host.docker.internal:host-gateway smartcontract/ccip:ccip-develop node -config /chainlink/config.toml -secrets /chainlink/secrets.toml start -a /chainlink/.api
docker run -d --platform linux/amd64 --name chainlink-1 -v $HOME/.chainlink-handson/node-1:/chainlink -it -p 16688:6688 --add-host=host.docker.internal:host-gateway smartcontract/ccip:ccip-develop node -config /chainlink/config.toml -secrets /chainlink/secrets.toml start -a /chainlink/.api
docker run -d --platform linux/amd64 --name chainlink-2 -v $HOME/.chainlink-handson/node-2:/chainlink -it -p 26688:6688 --add-host=host.docker.internal:host-gateway smartcontract/ccip:ccip-develop node -config /chainlink/config.toml -secrets /chainlink/secrets.toml start -a /chainlink/.api