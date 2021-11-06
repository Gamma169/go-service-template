CREATE TABLE IF NOT EXISTS foobar_models (
	id uuid PRIMARY KEY,
	name varchar(511) NOT NULL,
	

	date_created timestamp with time zone NOT NULL DEFAULT NOW(),
	last_updated timestamp with time zone NOT NULL DEFAULT NOW()
);
