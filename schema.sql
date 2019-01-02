DROP TABLE pairs;
DROP TABLE rounds;
DROP TABLE members;
DROP TABLE organizations;

CREATE TABLE organizations (
    name VARCHAR PRIMARY KEY,
    admin VARCHAR NOT NULL CHECK(length(admin) > 0),
    cross_match_trait VARCHAR
);

CREATE TABLE members (
    organization VARCHAR REFERENCES organizations(name),
    email VARCHAR PRIMARY KEY,
    name VARCHAR UNIQUE NOT NULL CHECK(length(name) > 0),
    metadata JSONB,
    pair_counts JSONB NOT NULL
);

CREATE TABLE rounds (
    organization VARCHAR REFERENCES organizations(name),
    id INTEGER NOT NULL CHECK (id >= 0),
    -- round dates are stored in UTC with accuracy up to seconds
    -- when retrieved, it is formatted as a string in 'YYYY-MM-DDTHH:mm:ssZ',
    -- the trailing Z indicating that that the datestring should be parsed 
    -- as UTC time (& converted into local time if a client library wants to
    -- do so e.g moment.js)
    scheduled_date TIMESTAMP NOT NULL,
    PRIMARY KEY (organization, id)
);

CREATE TABLE pairs (
    organization VARCHAR REFERENCES organizations(name),
    id1 VARCHAR REFERENCES members(email),
    id2 VARCHAR REFERENCES members(email),
    round INTEGER NOT NULL CHECK(round >= 0),
    PRIMARY KEY (organization, id1, id2, round),
    FOREIGN KEY (organization, round) REFERENCES rounds(organization, id)
);
