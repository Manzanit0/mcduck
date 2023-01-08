rw-db-migrate:
	docker run --rm -v `pwd`/migrations:/flyway/sql flyway/flyway:7.14.0 \
	-url=jdbc:postgresql://`railway variables get PGHOST -s mcduck`:`railway variables get PGPORT -s mcduck`/`railway variables get PGDATABASE -s mcduck` \
	-user=`railway variables get PGUSER -s mcduck` \
	-password=`railway variables get PGPASSWORD -s mcduck` \
	-schemas=public \
	-connectRetries=60 \
	migrate

rw-db-repair:
	docker run --rm -v `pwd`/migrations:/flyway/sql flyway/flyway:7.14.0 \
	-url=jdbc:postgresql://`railway variables get PGHOST -s mcduck`:`railway variables get PGPORT -s mcduck`/`railway variables get PGDATABASE -s mcduck` \
	-user=`railway variables get PGUSER -s mcduck` \
	-password=`railway variables get PGPASSWORD -s mcduck` \
	-schemas=public \
	-connectRetries=60 \
	repair

rw-pgcli:
	pgcli `railway variables get DATABASE_URL -s mcduck`
