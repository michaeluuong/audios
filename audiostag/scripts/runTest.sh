FILE="$1"
shift
SOURCE_FILE="${FILE}.go"
TEST_FILE="${FILE}_test.go"
go test -v -count=1 -cover ./$TEST_FILE ./$SOURCE_FILE $@ 
