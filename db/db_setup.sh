#!/bin/bash

# Configuration variables
CONTAINER_NAME="db-db-1"  # Replace with your Docker container name or ID
DB_NAME="postgres"                  # Replace with your database name
DB_USER="postgres"                  # Replace with your database user
DB_PASSWORD="example"              # Replace with your database password
SQL_FILE="create_tables.sql"              # Name of the SQL file

# Check if the SQL file exists
if [ ! -f "$SQL_FILE" ]; then
  echo "Error: $SQL_FILE not found in the current directory."
  exit 1
fi

# Copy the SQL file to the container
echo "Copying $SQL_FILE to container..."
docker cp "$SQL_FILE" "$CONTAINER_NAME:/tmp/$SQL_FILE"

# Execute the SQL file in the container
echo "Executing $SQL_FILE in the database..."
docker exec "$CONTAINER_NAME" bash -c "PGPASSWORD='$DB_PASSWORD' psql -U '$DB_USER' -d '$DB_NAME' -f /tmp/$SQL_FILE"

# Check if the command was successful
if [ $? -eq 0 ]; then
  echo "Tables created successfully!"
else
  echo "Error: Failed to execute $SQL_FILE."
  exit 1
fi