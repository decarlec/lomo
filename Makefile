db_container:
	docker compose -f db/docker-compose.yaml up -d	

db_setup:
	cd db && ./db_setup.sh

run:
	go run .

db_admin:
	docker exec -it db-db-1 psql -U postgres -d postgres