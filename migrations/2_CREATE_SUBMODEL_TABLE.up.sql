CREATE TABLE IF NOT EXISTS sub_models (
	id uuid PRIMARY KEY,
	user_id uuid NOT NULL,
	foobar_model_id uuid,
	value varchar(511) NOT NULL,
	value_int int NOT NULL,

	date_created timestamp with time zone NOT NULL DEFAULT NOW(),
	last_updated timestamp with time zone NOT NULL DEFAULT NOW(),

	FOREIGN KEY (foobar_model) REFERENCES foobar_models(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE
);
