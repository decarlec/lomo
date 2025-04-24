db_container:
	docker compose -f db/docker-compose.yaml up -d	

db_setup:
	cd db && ./db_setup.sh

db_bootstrap:
	cd bootstrap && go run bootstrap.go

run:
	go run . 

db_admin:
	docker exec -it db-db-1 psql -U postgres -d postgres

drop_all_tables:
	docker exec -it db-db-1 psql -U postgres -d postgres -c 'drop table lesson; drop table users; drop table words'