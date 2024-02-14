# Define the table name and primary key
TABLE_NAME="UserProfileServiceTable"
PRIMARY_KEY="UserID"

# Define read and write capacity units for the table
READ_CAPACITY_UNITS=10
WRITE_CAPACITY_UNITS=10

# Command to create the DynamoDB table
aws dynamodb create-table \
    --table-name $TABLE_NAME \
    --attribute-definitions AttributeName=$PRIMARY_KEY,AttributeType=S \
    --key-schema AttributeName=$PRIMARY_KEY,KeyType=HASH \
    --provisioned-throughput ReadCapacityUnits=$READ_CAPACITY_UNITS,WriteCapacityUnits=$WRITE_CAPACITY_UNITS \
    --endpoint-url http://localhost:8000