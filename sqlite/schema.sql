CREATE TABLE resrecords (
	ID INTEGER PRIMARY KEY,
	domain TEXT NOT NULL,
	data TEXT NOT NULL,
	typeID INTEGER NOT NULL,
	classID INTEGER NOT NULL,
	ttl INTEGER DEFAULT 0,
	FOREIGN KEY (typeID) REFERENCES types(id),
	FOREIGN KEY (classID) REFERENCES classes(id),
	UNIQUE(domain, data, typeID, classID) 
);

CREATE TABLE resrecords_types (
    typeID INTEGER NOT NULL,
    resrecordID INTEGER NOT NULL,
    PRIMARY KEY (typeID, resrecordID),
    FOREIGN KEY (typeID) REFERENCES types(id),
    FOREIGN KEY (resrecordID) REFERENCES resrecords(id)
);

CREATE TABLE resrecords_classes (
    classID INTEGER NOT NULL,
    resrecordID INTEGER NOT NULL,
    PRIMARY KEY (classID, resrecordID),
    FOREIGN KEY (classID) REFERENCES classes(id),
    FOREIGN KEY (resrecordID) REFERENCES resrecords(id)
);

CREATE TABLE types(
	ID INTEGER PRIMARY KEY,
	type text NOT NULL
);

CREATE TABLE classes(
	ID INTEGER PRIMARY KEY,
	class text NOT NULL
);

