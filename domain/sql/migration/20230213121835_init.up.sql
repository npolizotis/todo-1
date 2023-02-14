create table if not exists todo (
	id varchar(64) primary key,
	task varchar(255) not null,
	created integer NOT NULL,
	updated integer NOT NULL,
	complete  CHARACTER(1) NOT NULL,
	rank	integer not null
)