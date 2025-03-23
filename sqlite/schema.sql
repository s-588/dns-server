CREATE TABLE resourceRecords (
	id INTEGER PRIMARY KEY,
	domain text NOT NULL,
	result text NOT NULL,
	type text NOT NULL,
	class text NOT NULL,
	ttl INTEGER 
);
