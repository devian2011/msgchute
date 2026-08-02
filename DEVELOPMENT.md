# Makefile Main commands
```shell
make build # Build application
make test # Run tests
make lint # Run linter - must be run before each feature done
make migrate # Run migrations see ./migrations
make swagger # Generate swagger api doc - access on {{host}}/payment/v1/swagger
```

## Migrations

Migration package [DbMate](https://github.com/amacneil/dbmate)

```shell
make migrate command='new create_table' # Create new migration command
make migrate command=up # Make migrations
```

### Manual UUIDv7 generator
[Go here](https://www.uuidgenerator.net/version7)
