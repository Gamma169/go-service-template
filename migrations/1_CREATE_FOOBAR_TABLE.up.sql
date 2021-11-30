CREATE TABLE IF NOT EXISTS foobar_models (
	id uuid PRIMARY KEY,
	user_id uuid NOT NULL,
	name varchar(511) NOT NULL,
	age int NOT NULL,
	some_prop varchar(255) NOT NULL,
	some_nullable_prop varchar(255),
	

	date_created timestamp with time zone NOT NULL DEFAULT NOW(),
	last_updated timestamp with time zone NOT NULL DEFAULT NOW()
);
