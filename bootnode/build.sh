echo "Start build bootnode"

echo "go get"
go get -d

APP_NAME="incognito-bootnode"

echo "go build -o $APP_NAME"
go build -o $APP_NAME

echo "Build bootnode success!"
